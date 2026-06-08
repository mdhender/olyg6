// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package mapgen

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

// newReadMap builds a generator over a tiny map string and runs readMap +
// fixTerrainLand, the two passes that resolve a '?' hidden province.
func newReadMap(t *testing.T, mapText string) *Generator {
	t.Helper()
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "ascii-map.txt"), []byte(mapText), 0644); err != nil {
		t.Fatalf("write map: %v", err)
	}
	g := New(Options{InputDir: tmp, InputMap: "ascii-map.txt", Log: io.Discard})
	seed := make([]byte, 16)
	for i := range seed {
		seed[i] = byte(i + 1)
	}
	g.RNG.Load(seed) // fixTerrainLand draws via randomizeDirVector
	if err := g.readMap(); err != nil {
		t.Fatalf("readMap: %v", err)
	}
	g.fixTerrainLand()
	return g
}

// TestHiddenProvinceInfersTerrain: a '?' surrounded by a definite terrain
// becomes that terrain and is marked hidden.
func TestHiddenProvinceInfersTerrain(t *testing.T) {
	g := newReadMap(t, "ppp\np?p\nppp\n")
	h := g.Map[1][1]
	if h.Hidden != 1 {
		t.Errorf("hidden flag = %d, want 1", h.Hidden)
	}
	if h.Terrain != TerrPlain {
		t.Errorf("terrain = %d, want plain (%d)", h.Terrain, TerrPlain)
	}
	if h.SaveChar != '?' {
		t.Errorf("save char = %q, want '?'", rune(h.SaveChar))
	}
}

// TestHiddenAdjacentNoRecursion: two adjacent '?'s both resolve to the
// surrounding definite terrain (they never pick each other).
func TestHiddenAdjacentNoRecursion(t *testing.T) {
	g := newReadMap(t, "ffff\nf??f\nffff\n")
	for _, c := range [][2]int{{1, 1}, {1, 2}} {
		h := g.Map[c[0]][c[1]]
		if h.Hidden != 1 {
			t.Errorf("(%d,%d) hidden = %d, want 1", c[0], c[1], h.Hidden)
		}
		if h.Terrain == TerrLand || h.Terrain == TerrOcean {
			t.Errorf("(%d,%d) terrain = %d, want a definite land terrain", c[0], c[1], h.Terrain)
		}
		if h.Terrain != TerrForest {
			t.Errorf("(%d,%d) terrain = %d, want forest (%d)", c[0], c[1], h.Terrain, TerrForest)
		}
	}
}

// TestHiddenForestFallback: a '?' with no definite-terrain neighbor (only ocean
// here) falls back to forest, still hidden, and never becomes ocean.
func TestHiddenForestFallback(t *testing.T) {
	g := newReadMap(t, "...\n.?.\n...\n")
	h := g.Map[1][1]
	if h.Hidden != 1 {
		t.Errorf("hidden flag = %d, want 1", h.Hidden)
	}
	if h.Terrain == TerrOcean {
		t.Errorf("hidden province became ocean")
	}
	if h.Terrain != TerrForest {
		t.Errorf("terrain = %d, want forest fallback (%d)", h.Terrain, TerrForest)
	}
}
