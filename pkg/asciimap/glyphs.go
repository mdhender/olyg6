// Copyright (c) 2026 Michael D Henderson. All rights reserved.

// Package asciimap holds the shared definitions for the ASCII art world map
// that both the map generator (pkg/mapgen) and the island generator
// (pkg/islandgen) read.
//
// Ocean tiles are written with one of eight glyphs. They come in four
// "color" pairs; within each pair one glyph marks a plain ocean tile and the
// other marks a sea lane (fast ocean travel). Keeping the set in one place
// guarantees the two generators agree on what counts as ocean.
package asciimap

// Ocean glyphs, by color pair: (plain, sea-lane).
//
//	color 1: ','  ';'
//	color 2: '.'  ':'
//	color 3: ' '  '~'
//	color 4: '\'' '"'
const (
	OceanGlyphs   = ",;.: ~'\""
	SeaLaneGlyphs = ";:~\""
)

// IsOcean reports whether b is one of the eight ocean glyphs.
func IsOcean(b byte) bool {
	for i := range len(OceanGlyphs) {
		if OceanGlyphs[i] == b {
			return true
		}
	}
	return false
}

// IsSeaLane reports whether b is one of the four sea-lane ocean glyphs.
func IsSeaLane(b byte) bool {
	for i := range len(SeaLaneGlyphs) {
		if SeaLaneGlyphs[i] == b {
			return true
		}
	}
	return false
}
