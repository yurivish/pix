package pix

import (
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path"
)

// Returns `ImageColor`s from the source in row major order.
func LoadImage(path string) ([]ImageColor, error) {
	img, err := loadRGBA(path)
	if err != nil {
		return nil, fmt.Errorf("error loading image: %w", err)
	}
	sz := img.Bounds().Max
	index := 0
	colors := make([]ImageColor, sz.X*sz.Y)
	pix, stride := img.Pix, img.Stride
	for y := 0; y < sz.Y; y++ {
		for x := 0; x < sz.X; x++ {
			i := y*stride + x*4
			r, g, b := pix[i], pix[i+1], pix[i+2]
			colors[index] = ImageColor{x, y, r, g, b}
			index++
		}
	}
	return colors, nil
}

type ImageColor struct {
	X, Y    int
	R, G, B uint8
}

func loadRGBA(filepath string) (*image.RGBA, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer f.Close()
	var img image.Image
	ext := path.Ext(filepath)
	if ext == ".png" {
		img, err = png.Decode(f)
	} else if ext == ".jpg" || ext == ".jpeg" {
		img, err = jpeg.Decode(f)
	} else {
		return nil, fmt.Errorf("unknown image extension (we understand .png, .jpg, .jpeg): %v", ext)
	}
	if err != nil {
		return nil, fmt.Errorf("error decoding image: %w", err)
	}
	if rgba, ok := img.(*image.RGBA); ok {
		return rgba, nil
	} else {
		b := img.Bounds()
		rgba := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
		draw.Draw(rgba, rgba.Bounds(), img, b.Min, draw.Src)
		return rgba, nil
	}
}
