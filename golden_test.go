package pix

import (
	"hash/fnv"
	"testing"
)

// goldenConfig is a fixed placement scenario. Its output must never change
// across the nearest-neighbor optimization PRs (cache color / subtree bbox /
// de-closure are all pure performance changes), so we pin an FNV-1a hash of the
// rendered pixels. If an optimization alters output, this test fails.
//
// The expected hash is the baseline (pre-optimization) value, recorded from
// main. To regenerate after an intentional output change, set goldenHash to 0,
// run `go test -run TestGoldenOutput -v`, and copy the printed value.
const goldenHash = 0x706de251f9b2e47b

func renderGolden(tb testing.TB) []uint8 {
	src, err := LoadImage("img/winter.png")
	if err != nil {
		tb.Fatalf("loading image: %v", err)
	}
	const w, h = 160, 160
	colors := SampleColors(src, w*h)
	SortBySimilarity(colors, SortOptions{Image: 10, Color: 90, Random: 0, Reverse: true})

	c := NewCanvas(w, h, 1)
	rest, err := c.PlaceSeeds(colors, w/2, h/2)
	if err != nil {
		tb.Fatal(err)
	}
	for _, col := range rest {
		c.Place(col)
	}
	return c.ImageData()
}

func TestGoldenOutput(t *testing.T) {
	h := fnv.New64a()
	h.Write(renderGolden(t))
	got := h.Sum64()
	if got != goldenHash {
		t.Fatalf("golden output hash changed: got 0x%016x, want 0x%016x\n"+
			"(if this change is intentional, update goldenHash)", got, goldenHash)
	}
}
