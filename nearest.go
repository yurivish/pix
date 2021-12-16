package pix

import (
	"math"
	"math/bits"
)

// Nearest-neighbor search in a 3D color space using an approach described in
// A Minimalist's Implementation of an Approximate Nearest Search in Fixed Dimensions:
// http://cs.uwaterloo.ca/~tmchan/sss.ps
//
// The algorithm is a variant of binary search through a Morton-ordered list of points
// which alternately prunes the search space in Euclidean space and along the curve.
// In our case we stores the points in a zip tree for dynamic updates, an perform the
// search by recursively traversing the tree.
func (t *zipTree) Nearest(q Color, qCode MortonCode) MortonCode {
	var rSq uint32 = 1 << 30
	var best MortonCode
	var qPosCode, qNegCode MortonCode
	// todo: figure out why epsilon can be set to eg. 100000 with no ill effect
	// ε := float64(100000)
	// approxFactor := (1.0 + ε) * (1.0 + ε)
	// float64(distSqToBBox(qCode, t.MinKey(a), t.MaxKey(a), q))*approxFactor >= float64(rSq) {
	var query func(q Color, ah Handle)
	query = func(q Color, ah Handle) {
		if ah == 0 {
			return
		}
		a := t.Node(ah)
		midCode := a.Key()
		mid := mortonCodeToColor(midCode)
		dSq := sqDist(q, mid)
		if dSq < rSq {
			rSq = dSq
			var r uint8
			if dSq >= 255*255 {
				r = 255
			} else {
				r = uint8(math.Ceil(math.Sqrt(float64(dSq))))
			}
			qPosCode = mortonCode(satAdd(q.x, r), satAdd(q.y, r), satAdd(q.z, r))
			qNegCode = mortonCode(satSub(q.x, r), satSub(q.y, r), satSub(q.z, r))
			best = midCode
		}
		// a.left is only equal to a.right if both are nilHandle
		if a.left == a.right || midCode == qCode || distSqToBBox(qCode, t.MinKey(a), t.MaxKey(a), q) >= rSq {
			return
		}
		if qCode <= midCode {
			query(q, a.left)
			if qPosCode >= midCode {
				query(q, a.right)
			}
		} else {
			query(q, a.right)
			if qNegCode <= midCode {
				query(q, a.left)
			}
		}
	}
	query(q, t.root)
	return best
}

func distSqToBBox(q, a, b MortonCode, c Color) uint32 { // takes morton codes as arguments
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
		dSq += sqDiff(c.x, mortonX(lo))
	} else if gtMortonX(q, hi) {
		dSq += sqDiff(c.x, mortonX(hi))
	}

	// y coordinate
	if ltMortonY(q, lo) {
		dSq += sqDiff(c.y, mortonY(lo))
	} else if gtMortonY(q, hi) {
		dSq += sqDiff(c.y, mortonY(hi))
	}

	// z coordinate
	if ltMortonZ(q, lo) {
		dSq += sqDiff(c.z, mortonZ(lo))
	} else if gtMortonZ(q, hi) {
		dSq += sqDiff(c.z, mortonZ(hi))
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
