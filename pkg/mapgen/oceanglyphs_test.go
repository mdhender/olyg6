// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package mapgen

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/mdhender/olyg6/pkg/asciimap"
)

// TestReadMapOceanGlyphsMatchShared guards against drift between mapgen's
// readMap glyph switch and the shared asciimap ocean set: every glyph in
// asciimap.OceanGlyphs must read as ocean, and exactly the asciimap sea-lane
// glyphs must set the sea-lane flag.
func TestReadMapOceanGlyphsMatchShared(t *testing.T) {
	tmp := t.TempDir()
	row := asciimap.OceanGlyphs // all eight ocean glyphs on one line
	if err := os.WriteFile(filepath.Join(tmp, "ascii-map.txt"), []byte(row+"\n"), 0644); err != nil {
		t.Fatalf("write map: %v", err)
	}

	g := New(Options{InputDir: tmp, InputMap: "ascii-map.txt", Log: io.Discard})
	if err := g.readMap(); err != nil {
		t.Fatalf("readMap: %v", err)
	}

	for col := 0; col < len(row); col++ {
		ch := row[col]
		tile := g.Map[0][col]
		if tile == nil {
			t.Fatalf("no tile for glyph %q", rune(ch))
		}
		if tile.Terrain != TerrOcean {
			t.Errorf("glyph %q: terrain = %d, want ocean (%d)", rune(ch), tile.Terrain, TerrOcean)
		}
		if !asciimap.IsOcean(ch) {
			t.Errorf("glyph %q read as ocean by mapgen but asciimap.IsOcean is false", rune(ch))
		}
		wantSeaLane := asciimap.IsSeaLane(ch)
		gotSeaLane := tile.SeaLane != 0
		if gotSeaLane != wantSeaLane {
			t.Errorf("glyph %q: mapgen sea-lane = %v, asciimap = %v", rune(ch), gotSeaLane, wantSeaLane)
		}
	}
}
