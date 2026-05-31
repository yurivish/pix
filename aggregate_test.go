package pix

import (
	"math/rand"
	"testing"
)

// TestAggregateConsistency stresses the zip tree with a long randomized mix of
// inserts and deletes and verifies, against an independent leaf-walk
// (MinKey/MaxKey), that every live node's cached subtree min/max keys
// (subMin/subMax) stay correct. This is the safety net for the augmented-tree
// maintenance that the O(1) bbox prune relies on.
func TestAggregateConsistency(t *testing.T) {
	tr := newZipTree(rand.New(rand.NewSource(7)))
	rng := rand.New(rand.NewSource(42))
	present := map[MortonCode]bool{}

	var verify func(h Handle)
	verify = func(h Handle) {
		if h == nilHandle {
			return
		}
		n := tr.nodes[h]
		wantMin := tr.MinKey(n) // independent O(height) leaf walk
		wantMax := tr.MaxKey(n)
		if n.subMin != wantMin || n.subMax != wantMax {
			t.Fatalf("node key=%d: subMin=%d (want %d) subMax=%d (want %d)",
				n.Key(), n.subMin, wantMin, n.subMax, wantMax)
		}
		verify(n.left)
		verify(n.right)
	}

	for i := 0; i < 8000; i++ {
		k := MortonCode(rng.Intn(1 << 24))
		if present[k] {
			tr.Delete(k)
			delete(present, k)
		} else {
			tr.Insert(k)
			present[k] = true
		}
		if i%37 == 0 {
			verify(tr.root)
		}
	}
	verify(tr.root)
}
