package pix

import "testing"

func TestDistSqToBBox(t *testing.T) {
	colorToMortonCode := func(color Color) MortonCode {
		return mortonCode(color.x, color.y, color.z)
	}

	// todo: more tests for the other components, multiple components, binary bbox.
	tests := []struct {
		q, a, b Color
		result  uint32
	}{
		{}, // zero  colors are zero distance apart
		{Color{0, 0, 0}, Color{0, 0, 0}, Color{0, 0, 2}, 0},
		{Color{0, 0, 0}, Color{0, 0, 3}, Color{0, 0, 3}, 9},
		{Color{0, 0, 0}, Color{0, 0, 0}, Color{0, 0, 3}, 0},
		{Color{0, 0, 4}, Color{0, 0, 0}, Color{0, 0, 3}, 1},
	}
	for _, test := range tests {
		q, a, b := colorToMortonCode(test.q), colorToMortonCode(test.a), colorToMortonCode(test.b)
		got := distSqToBBox(q, a, b, test.q)
		if got != test.result {
			t.Errorf("%v: got %v; want %v", test, got, test.result)
		}
	}
}

func TestSatAdd(t *testing.T) {
	tests := []struct {
		a, b, c uint8
	}{{0, 0, 0}, {1, 1, 2}, {255, 0, 255}, {255, 1, 255}, {1, 255, 255},
		{100, 155, 255}, {200, 200, 255}, {255, 255, 255}, {128, 128, 255}}
	for _, test := range tests {
		got := satAdd(test.a, test.b)
		if got != test.c {
			t.Errorf("%v: got %v; want %v", test, got, test.c)
		}
	}
}

func TestSatSub(t *testing.T) {
	tests := []struct {
		a, b, c uint8
	}{{0, 0, 0}, {0, 1, 0}, {1, 0, 1}, {255, 254, 1}, {254, 255, 0},
		{100, 155, 0}, {200, 100, 100}, {0, 100, 0}, {10, 10, 0}}
	for _, test := range tests {
		got := satSub(test.a, test.b)
		if got != test.c {
			t.Errorf("%v: got %v; want %v", test, got, test.c)
		}
	}
}

func TestSqDist(t *testing.T) {
	tests := []struct {
		a, b Color
		c    uint32
	}{
		{Color{0, 0, 0}, Color{0, 0, 0}, 0},
		{Color{2, 0, 0}, Color{0, 0, 0}, 4},
		{Color{0, 2, 0}, Color{0, 0, 0}, 4},
		{Color{0, 0, 2}, Color{0, 0, 0}, 4},
		{Color{255, 255, 255}, Color{1, 1, 1}, 254 * 254 * 3},
	}
	for _, test := range tests {
		got := sqDist(test.a, test.b)
		if got != test.c {
			t.Errorf("%v: got %v; want %v", test, got, test.c)
		}
		got = sqDist(test.b, test.a)
		if got != test.c {
			t.Errorf("%v: got %v; want %v", test, got, test.c)
		}
	}
}

func TestSqDiff(t *testing.T) {
	tests := []struct {
		a, b uint8
		c    uint32
	}{{0, 0, 0}, {1, 1, 0}, {2, 1, 1}, {1, 2, 1}, {2, 0, 4}, {0, 2, 4},
		{255, 0, 255 * 255}, {128, 255, 127 * 127}, {255, 128, 127 * 127}}
	for _, test := range tests {
		got := sqDiff(test.a, test.b)
		if got != test.c {
			t.Errorf("%v: got %v; want %v", test, got, test.c)
		}
	}
}
