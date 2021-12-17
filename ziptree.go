package pix

// This file implements a zip tree:
// https://arxiv.org/abs/1806.06726

import (
	"math"
	"math/rand"
)

type Handle uint32

const nilHandle = Handle(0)

type zipNode struct {
	rankAndKey  uint32 // 8-bit rank, then 24-bit morton code
	left, right Handle // handles to the left and right children in the pool
}

type zipTree struct {
	root  Handle    // handle to the root node
	nodes []zipNode // pool of pre-allocated nodes
	free  []Handle  // free list
	rng   *rand.Rand
}

func newZipTree(rng *rand.Rand) *zipTree {
	nodes := make([]zipNode, 1, 250_000)
	free := make([]Handle, 0, 100_000)
	return &zipTree{nilHandle, nodes, free, rng}
}

func (t *zipTree) Insert(key MortonCode) {
	handle := t.newZipNode(key)
	t.root = t.InsertRec(t.root, handle)
}

func (t *zipTree) Delete(key MortonCode) {
	t.root = t.DeleteRec(t.root, key)
}

func (t *zipTree) Reset() {
	t.root = nilHandle
	t.nodes = t.nodes[:1]
	t.free = t.free[:0]
}

// take some bits from each of the r, g, b channels and use them to break rank ties.
// tarjan calls these "fractional ranks" and suggests their use to improve the balance of
// the tree, which are otherwise right-heavy                     rgbrgbrgbrgbrgbrgbrgbrgb
func (x zipNode) Rank() uint32 { return x.rankAndKey & 0b11111111000000000000000000111000 }

func (x zipNode) Key() MortonCode {
	return MortonCode(x.rankAndKey & 0b00000000111111111111111111111111)
}

func (t *zipTree) newZipNode(key MortonCode) Handle {
	rank := uint32(0)
	for t.rng.Int63()&1 == 0 {
		rank++
	}
	rankAndKey := rank<<24 | uint32(key)
	handle := t.Get(rankAndKey)
	return handle
}

func (t *zipTree) MinKey(x zipNode) MortonCode {
	for x.left != nilHandle {
		x = t.Node(x.left)
	}
	return x.Key()
}

func (t *zipTree) MaxKey(x zipNode) MortonCode {
	for x.right != nilHandle {
		x = t.Node(x.right)
	}
	return x.Key()
}

func (t *zipTree) InsertRec(hroot, hx Handle) Handle {
	if hroot == nilHandle {
		return hx
	}
	// since the pool's backing buffer remains unmodified during deletion, we can take this reference safely
	x := t.NodeRef(hx)
	root := t.NodeRef(hroot)
	if x.Key() < root.Key() {
		if t.InsertRec(root.left, hx) == hx {
			if x.Rank() < root.Rank() {
				root.left = hx
			} else {
				root.left = x.right
				x.right = hroot
				return hx
			}
		}
	} else {
		if t.InsertRec(root.right, hx) == hx {
			if x.Rank() <= root.Rank() {
				root.right = hx
			} else {
				root.right = x.left
				x.left = hroot
				return hx
			}
		}
	}
	return hroot
}

func (t *zipTree) DeleteRec(hroot Handle, key MortonCode) Handle {
	// since the pool's backing buffer remains unmodified during deletion, we can take this reference safely
	root := t.NodeRef(hroot)
	if key == root.Key() {
		t.Put(hroot)
		return t.zip(root.left, root.right)
	}
	if key < root.Key() {
		left := t.Node(root.left)
		if key == left.Key() {
			t.Put(root.left)
			root.left = t.zip(left.left, left.right)
		} else {
			t.DeleteRec(root.left, key)
		}
	} else {
		right := t.Node(root.right)
		if key == right.Key() {
			t.Put(root.right)
			root.right = t.zip(right.left, right.right)
		} else {
			t.DeleteRec(root.right, key)
		}
	}
	return hroot
}

func (t *zipTree) zip(hx, hy Handle) Handle {
	if hx == nilHandle {
		return hy
	}
	if hy == nilHandle {
		return hx
	}
	x, y := t.Node(hx), t.Node(hy)
	if x.Rank() < y.Rank() {
		t.SetLeft(hy, t.zip(hx, y.left))
		return hy
	} else {
		t.SetRight(hx, t.zip(x.right, hy))
		return hx
	}
}

func (t *zipTree) Node(handle Handle) zipNode {
	if handle == nilHandle {
		panic("Got a node with handle zero in Handle.Node()")
	}
	return t.nodes[handle]
}

func (t *zipTree) SetLeft(handle Handle, x Handle) { t.nodes[handle].left = x }

func (t *zipTree) SetRight(handle Handle, x Handle) { t.nodes[handle].right = x }

func (t *zipTree) NodeRef(handle Handle) *zipNode {
	if handle == nilHandle {
		panic("Got a node with handle zero in Handle.NodeRef()")
	}
	return &t.nodes[handle]
}

// Put the handle back into the pool
func (t *zipTree) Put(handle Handle) {
	t.free = append(t.free, handle)
}

// Get an unused handle from the pool
func (t *zipTree) Get(rankAndKey uint32) Handle {
	n := len(t.free)
	if n > 0 {
		handle := t.free[n-1]
		t.free = t.free[:n-1]
		t.nodes[handle] = zipNode{rankAndKey, nilHandle, nilHandle}
		return handle
	}
	handle := Handle(len(t.nodes))
	t.nodes = append(t.nodes, zipNode{rankAndKey, nilHandle, nilHandle})
	return handle
}

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
		// We exclude searching intervals if the distance from the query point to the snug power-of-2 bounding box
		// enclosing the interval is farther away than our best distance so far.
		if a.left == a.right || midCode == qCode || distSqToBBox(qCode, t.MinKey(a), t.MaxKey(a), q) >= rSq {
			return
		}
		// If we can't exclude the interval, go ahead with a recursive search.
		// Recurse into one or both halves of the array. We can avoid recursing
		// into the second half when the box enclosing our "best radius" circle.
		// The points qPos and qNeg are the smallest and largest morton codes of
		// points within that box – thanks to properties of the z-order curve,
		// any points below qNeg or above qPos are definitely more distant from q
		// than the best radius r.
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
