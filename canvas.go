package pix

import (
	"fmt"
	"image"
	"image/png"
	"math/rand"
	"os"
	"sort"
)

// A canvas represents a specific pixel-placed drawing
type Canvas struct {
	tree             *zipTree                // frontier colors sorted in z-order
	positions        map[MortonCode]*posList // candidate positions for every color in the tree
	rng              *rand.Rand              // rng for reproducibility (used in RandEmptyNeighbor)
	img              []MortonCode            // placed colors
	ns               neighbors               // neighborhood-tracking structure
	nPlaced          int                     // number of pixels placed
	inpaintCutoff    int                     // number of pixels beyond which to reject poor matches
	w, h, wPad, hPad int                     // width and height, along with their 1-padded versions
}

func NewCanvas(w, h int, seed int64) *Canvas {
	rng := rand.New(rand.NewSource(seed))
	tree := newZipTree(rng)
	wPad, hPad := w+2, h+2
	img := make([]MortonCode, wPad*hPad) // init image data
	ns := NewNeighbors(wPad, hPad)       // init empty neighbor-tracking structure
	nPlaced := 0
	inpaintCutoff := (w * h * 95) / 100
	positions := make(map[MortonCode]*posList)
	return &Canvas{tree, positions, rng, img, ns, nPlaced, inpaintCutoff, w, h, wPad, hPad}
}

func (c *Canvas) Reset() {
	c.tree.Reset()
	c.positions = make(map[MortonCode]*posList)
	c.img = make([]MortonCode, c.wPad*c.hPad)
	c.ns = NewNeighbors(c.wPad, c.hPad)
	c.nPlaced = 0
}

// Represents a color sample in the RGB and OkLab color spaces,
// along with its Morton and Hilbert codes.
type SampledColor struct {
	rgb, lab         Color
	rgbCode, labCode MortonCode
	xyCode           uint32
	sortScore        float64
}

func SampleColors(src []ImageColor, nPixels int) []SampledColor {
	nSrc, nDst := len(src), nPixels
	ret := make([]SampledColor, nDst)
	// center the hilbert curve in its power-of-2 bounding box so as
	// not to unduly privilege one corner of the image over the others
	srcW, srcH := src[nSrc-1].X, src[nSrc-1].Y
	wOffset := int((pow2MoreThan(srcW) - uint32(srcW)) / 2)
	hOffset := int((pow2MoreThan(srcH) - uint32(srcH)) / 2)
	if nDst < nSrc {
		// do expensive stuff once per dst
		for i := 0; i < nDst; i++ {
			pc := float64(i) / float64(nDst)
			index := int(float64(nSrc) * pc)
			c := src[index]
			rgb := Color{c.R, c.G, c.B}
			lab := rgbToOkLab(rgb)
			rgbCode := mortonCode(rgb.x, rgb.y, rgb.z)
			labCode := mortonCode(lab.x, lab.y, lab.z)
			xyCode := xyToHilbert(uint32(c.X+wOffset), uint32(c.Y+hOffset), 16)
			ret[i] = SampledColor{rgb, lab, rgbCode, labCode, xyCode, 0}
		}
	} else {
		nMultiples := nDst / nSrc
		index := 0
		for _, c := range src {
			rgb := Color{c.R, c.G, c.B}
			lab := rgbToOkLab(rgb)
			rgbCode := mortonCode(rgb.x, rgb.y, rgb.z)
			labCode := mortonCode(lab.x, lab.y, lab.z)
			xyCode := xyToHilbert(uint32(c.X+wOffset), uint32(c.Y+hOffset), 16)
			x := SampledColor{rgb, lab, rgbCode, labCode, xyCode, 0}
			for s := 0; s < nMultiples; s++ {
				ret[index] = x
				index++
			}
		}
		nPlaced := nSrc * nMultiples
		nRemaining := nDst - nPlaced
		for i := 0; i < nRemaining; i++ {
			pc := float64(i) / float64(nRemaining)
			index := int(float64(nSrc) * pc)
			ret[nPlaced+i] = ret[index]
		}
	}
	return ret
}

func (c *Canvas) PlaceAt(code MortonCode, pos Pos) {
	c.img[pos] = code
	c.ns.Fill(pos, func(pos Pos) {
		code := c.img[pos]
		if c.positions[code].delete(pos) {
			c.tree.Delete(code)
			delete(c.positions, code)
		}
	})
	if c.ns.Count(pos) < 9 {
		if plist, ok := c.positions[code]; ok {
			plist.insert(pos)
		} else {
			c.positions[code] = &posList{nil, pos}
			c.tree.Insert(code)
		}
	}
	c.nPlaced++
}

func (c *Canvas) PlaceSeed(color SampledColor, x, y int) {
	// todo: check xy bounds
	c.PlaceAt(color.labCode, Pos(rowMajorIndex(x+1, y+1, c.wPad)))
}

func (c *Canvas) PlaceSeeds(colors []SampledColor, xys ...int) ([]SampledColor, error) {
	n := len(xys)
	if n%2 == 1 {
		return nil, fmt.Errorf("attempting to place seeds with an odd number of coordinates") // todo: throw error
	}
	rest := colors
	for i := 0; i < n; i += 2 {
		color := rest[0]
		x, y := xys[i], xys[i+1]
		if x < 0 || x >= c.w || y < 0 || y >= c.h {
			return nil, fmt.Errorf("attempting to place out-of-bound seed: (%v, %v) with width %v and height %v", x, y, c.w, c.h)
		}
		// todo: error if seed is out of bounds
		c.PlaceSeed(color, x, y)
		rest = rest[1:]
	}
	return rest, nil
}

func mortonCodeToColor(code MortonCode) Color {
	return Color{mortonX(code), mortonY(code), mortonZ(code)}
}

type SortOptions struct {
	Image, Color, Random float64
	Reverse              bool
}

func SortBySimilarity(colors []SampledColor, opts SortOptions) {
	rgbMax := float64(mortonCode(255, 255, 255))

	// compute the smallest and largest Hilbert codes in order to
	// properly normalize the XY component of the sort score
	var _xyMax uint32 = 0 // compute over the codes actually used
	var _xyMin uint32 = ^uint32(0)
	for _, e := range colors {
		if e.xyCode > _xyMax {
			_xyMax = e.xyCode
		}
		if e.xyCode < _xyMin {
			_xyMin = e.xyCode
		}
	}
	xyMax := float64(_xyMax)
	xyMin := float64(_xyMin)
	xyDiff := xyMax - xyMin

	order := 1.0
	if opts.Reverse {
		order = -1
	}
	random := opts.Random > 0
	for i, e := range colors {
		rgb := float64(e.rgbCode) / rgbMax
		xy := (float64(e.xyCode) - xyMin) / xyDiff
		score := float64(opts.Image*xy + opts.Color*rgb)
		if random {
			score += opts.Random * rand.Float64()
		}
		colors[i].sortScore = order * score
	}
	sort.Slice(colors, func(i, j int) bool { return colors[i].sortScore < colors[j].sortScore })

}

func (c *Canvas) Place(x SampledColor) {
	color, code := x.lab, x.labCode
	nearest := c.tree.Nearest(color, code)
	inpaint := c.nPlaced > c.inpaintCutoff
	if inpaint {
		nearestColor := mortonCodeToColor(nearest)
		// have low tolerance for discrepancies in color
		const maxDist = 10
		if sqDist(color, nearestColor) > maxDist*maxDist {
			code = nearest
		}
	}
	pos := c.positions[nearest].arbitrary()
	targetPos := c.ns.RandEmptyNeighbor(pos, c.rng)
	c.PlaceAt(code, targetPos)
}

func (c *Canvas) ImageData() []uint8 {
	// create a new buffer with an alpha channel then copy data over
	nPixels := c.w * c.h
	data := make([]uint8, 4*nPixels)
	for y := 0; y < c.h; y++ {
		for x := 0; x < c.w; x++ {
			isrc := rowMajorIndex(x+1, y+1, c.wPad) // img:  account for padding
			idst := 4 * rowMajorIndex(x, y, c.w)    // data: account for the flat structure of 4 uint8s per color
			code := c.img[isrc]
			if c.ns.Empty(Pos(isrc)) {
				data[idst], data[idst+1], data[idst+2], data[idst+3] = 0, 0, 0, 0
			} else {
				// note: we do round-trip through srgb -> linear srgb -> oklab -> linear rgb -> srgb.
				// this handles the general case when placed colors do not correspond to a source image.
				data[idst], data[idst+1], data[idst+2] = okLabCodeToRgb(code)
				data[idst+3] = 255
			}
		}
	}
	return data
}

func (c *Canvas) SaveImage(path string, compressionLevel png.CompressionLevel) error {
	data, w, h := c.ImageData(), c.w, c.h
	r := image.Rectangle{image.Point{0, 0}, image.Point{w, h}}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error opening output image: %w", err)
	}
	defer f.Close()

	enc := &png.Encoder{CompressionLevel: compressionLevel}
	err = enc.Encode(f, &image.RGBA{Pix: data, Stride: 4 * r.Dx(), Rect: r})
	if err != nil {
		return fmt.Errorf("error writing output image: %w", err)
	}
	return nil
}

// Pos represents an (x, y index) pair as a single uint32 index into a (padded) array.
type Pos int32

func rowMajorIndex(x, y, w int) int { return y*w + x }

func pow2MoreThan(x int) uint32 {
	for i := 0; i < 31; i++ {
		if x < 1<<i {
			return 1 << i
		}
	}
	return 1 << 31
}
