package pix

import (
	"math/bits"
)

// This file holds supporting functions for nearest-neighbor search.
// See ziptree.go for the actual search implementation.

func distSqToBBox(q, a, b MortonCode, qColor Color) uint32 { // takes morton codes as arguments
	// use the code representation to compute the binary bbox
	// get the most significant differing bit between the morton codes a and b
	msb := bits.Len32(uint32(a ^ b))

	// lo and hi will give us the morton codes for the square or binary (2 contiguous squares)
	// octree-aligned bounding box containing a and b.

	// zero out the bits below that; that's our lowest morton code
	lo := (a >> msb) << msb
	// set those bits to one; that's our highest morton code
	hi := lo + (1 << msb) - 1

	// accumulate squared distance to the bounding box
	dSq := uint32(0)

	// x coordinate
	if ltMortonX(q, lo) {
		dSq += sqDiff(qColor.x, mortonX(lo))
	} else if gtMortonX(q, hi) {
		dSq += sqDiff(qColor.x, mortonX(hi))
	}

	// y coordinate
	if ltMortonY(q, lo) {
		dSq += sqDiff(qColor.y, mortonY(lo))
	} else if gtMortonY(q, hi) {
		dSq += sqDiff(qColor.y, mortonY(hi))
	}

	// z coordinate
	if ltMortonZ(q, lo) {
		dSq += sqDiff(qColor.z, mortonZ(lo))
	} else if gtMortonZ(q, hi) {
		dSq += sqDiff(qColor.z, mortonZ(hi))
	}
	return dSq
}

// saturating 8-bit addition
func satAdd(a, b uint8) uint8 {
	r := a + b
	if r < a {
		return 255
	}
	return r
}

// saturating 8-bit subtraction
func satSub(a, b uint8) uint8 {
	if b > a {
		return 0
	}
	return a - b
}

func sqDist(a, b Color) uint32 {
	return sqDiff(a.x, b.x) + sqDiff(a.y, b.y) + sqDiff(a.z, b.z)
}

func sqDiff(x uint8, y uint8) uint32 {
	diff := uint32(x) - uint32(y)
	return diff * diff
}
