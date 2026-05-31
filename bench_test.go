package pix

import (
	"fmt"
	"testing"
)

// benchPlace times the isolated placement loop (canvas alloc + seed + growth),
// excluding image load, color sampling, sort, and PNG encode. The ns/px metric
// is the scaling signal for the nearest-neighbor search work.
func benchPlace(b *testing.B, sz int) {
	src, err := LoadImage("img/winter.png")
	if err != nil {
		b.Fatalf("loading image: %v", err)
	}
	colors := SampleColors(src, sz*sz)
	SortBySimilarity(colors, SortOptions{Image: 10, Color: 90, Random: 0, Reverse: true})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c := NewCanvas(sz, sz, 1)
		rest, err := c.PlaceSeeds(colors, sz/2, sz/2)
		if err != nil {
			b.Fatal(err)
		}
		for _, col := range rest {
			c.Place(col)
		}
	}
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/float64(sz*sz), "ns/px")
}

func BenchmarkPlace(b *testing.B) {
	for _, sz := range []int{256, 512, 1024} {
		b.Run(fmt.Sprintf("%dx%d", sz, sz), func(b *testing.B) { benchPlace(b, sz) })
	}
}
