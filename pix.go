package pix

import (
	"image/png"
)

type Options struct {
	Width, Height    int
	RandomSeed       int64
	Sort             SortOptions
	Seeds            []int
	Output           string
	CompressionLevel png.CompressionLevel
}

func Place(colors []SampledColor, opts Options) error {

	// Create a canvas object
	canvas := NewCanvas(opts.Width, opts.Height, opts.RandomSeed)

	// Place an initial seed color in the middle of the canvas
	seeds := opts.Seeds
	if seeds == nil {
		seeds = []int{opts.Width / 2, opts.Height / 2}
	}
	rest, err := canvas.PlaceSeeds(colors, seeds...)
	if err != nil {
		return err
	}

	// Place the rest of the colors using the growth algorithm
	for _, color := range rest {
		canvas.Place(color)
	}

	// Save the output image
	outPath := opts.Output
	if outPath == "" {
		outPath = "out.png"
	}
	// fmt.Println("saving", outPath)
	return canvas.SaveImage(outPath, opts.CompressionLevel)
}
