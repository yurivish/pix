package pix

import (
	"math/rand"
)

// used to track empty/fullness of cells along with the fill status of their neighborhoods
type neighbors struct {
	empty   []bool  // whether a cell is full or empty
	count   []uint8 // filled neighbor count (including oneself), 0-9
	offsets [9]Pos  // index offsets to reach a cell's immediate Cartesian neighborhood
	w, h    int     // width, height
}

func NewNeighbors(w, h int) neighbors {
	// w*h in row-major order
	empty := make([]bool, w*h)
	count := make([]uint8, w*h)

	// add 1px of empty padding on each side
	// leaving a border of `false`
	for x := 1; x < w-1; x++ {
		for y := 1; y < h-1; y++ {
			empty[rowMajorIndex(x, y, w)] = true
		}
	}

	tl := Pos(-w - 1) // 0 is the middle, so move up one row (-w) and left one (-1)
	ml := tl + Pos(w)
	bl := ml + Pos(w)
	offsets := [9]Pos{tl, tl + 1, tl + 2, ml, ml + 1, ml + 2, bl, bl + 1, bl + 2}
	n := neighbors{empty, count, offsets, w, h}

	// since the padding positions are flagged as nonempty, they are never added to the frontier.
	// since their counts never reach 9, so we never wind up trying to remove them from the frontier.
	for i := 0; i < w; i++ {
		n.SafeIncrCount(i, 0)
		n.SafeIncrCount(i, h-1)
	}

	// don't double count the corners; loop from 1 to h-1
	for i := 1; i < h-1; i++ {
		n.SafeIncrCount(0, i)
		n.SafeIncrCount(w-1, i)
	}

	return n
}

func (n neighbors) Count(pos Pos) uint8 {
	return n.count[pos]
}

func (n neighbors) Empty(pos Pos) bool {
	return n.empty[pos]
}

// Fill `pos`, and call `cb` with the positions of any of its neighbors
// that are now full, so they can be removed from the frontier.
func (n neighbors) Fill(pos Pos, cb func(pos Pos)) {
	c := n.count
	for i, o := range n.offsets {
		index := pos + o
		v := c[index]
		c[index] = v + 1
		// at i == 4, `index` corresponds to the index for `pos`,
		// which we want to skip — the callback should be called
		// for its neighbors only.
		// v == 8 when all neighbors of  `index` are full.
		if v == 8 && i != 4 {
			cb(index)
		}
	}
	n.empty[pos] = false
}

// return a random empty neighbor of `pos`
func (n neighbors) RandEmptyNeighbor(pos Pos, rng *rand.Rand) Pos {
	var empties [8]Pos
	var index int
	e := n.empty
	for _, o := range n.offsets {
		// e[pos] is always false, so no need to skip i == 4
		if e[pos+o] {
			empties[index] = pos + o
			index++
		}
	}
	if index == 1 {
		return empties[0]
	} else {
		return empties[rng.Int31n(int32(index))]
	}
}

// used to initialize cells at the border
func (n neighbors) SafeIncrCount(x, y int) {
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			px, py := x+dx, y+dy
			if px >= 0 && px < n.w && py >= 0 && py < n.h {
				n.count[rowMajorIndex(px, py, n.w)]++
			}
		}
	}
}
