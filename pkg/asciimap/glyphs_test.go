// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package asciimap_test

import (
	"testing"

	"github.com/mdhender/olyg6/pkg/asciimap"
)

func TestIsOcean(t *testing.T) {
	for _, b := range []byte(",;.: ~'\"") {
		if !asciimap.IsOcean(b) {
			t.Errorf("IsOcean(%q) = false, want true", rune(b))
		}
	}
	for _, b := range []byte("pPfmdsoO#xX0") {
		if asciimap.IsOcean(b) {
			t.Errorf("IsOcean(%q) = true, want false", rune(b))
		}
	}
}

func TestIsSeaLane(t *testing.T) {
	for _, b := range []byte(";:~\"") {
		if !asciimap.IsSeaLane(b) {
			t.Errorf("IsSeaLane(%q) = false, want true", rune(b))
		}
		if !asciimap.IsOcean(b) {
			t.Errorf("sea lane %q must also be ocean", rune(b))
		}
	}
	// The plain ocean glyphs are ocean but not sea lanes.
	for _, b := range []byte(",. '") {
		if asciimap.IsSeaLane(b) {
			t.Errorf("IsSeaLane(%q) = true, want false", rune(b))
		}
	}
}

func TestPlainOcean(t *testing.T) {
	if len(asciimap.PlainOcean) != 4 {
		t.Fatalf("PlainOcean has %d glyphs, want 4", len(asciimap.PlainOcean))
	}
	for i := range len(asciimap.PlainOcean) {
		b := asciimap.PlainOcean[i]
		if !asciimap.IsOcean(b) {
			t.Errorf("PlainOcean[%d]=%q is not ocean", i, rune(b))
		}
		if asciimap.IsSeaLane(b) {
			t.Errorf("PlainOcean[%d]=%q is a sea lane", i, rune(b))
		}
	}
}
