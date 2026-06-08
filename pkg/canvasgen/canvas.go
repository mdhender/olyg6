// Copyright (c) 2026 Michael D Henderson. All rights reserved.

// Package canvasgen generates a square, all-ocean starting map for the
// map-building pipeline (canvas -> island -> map).
//
// The map is partitioned into several distinct oceans grown from random seed
// points, then 4-colored with the plain ocean glyphs so that mapgen reads each
// ocean as its own region. The output contains no sea lanes, so generate island
// will accept it.
package canvasgen

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mdhender/olyg6/pkg/asciimap"
	"github.com/mdhender/olyg6/pkg/prng"
	"github.com/mdhender/olyg6/pkg/store"
)

// Size and ocean-count bounds.
const (
	MinSize   = 9
	MaxSize   = 99
	MinOceans = 1
	MaxOceans = 20
	maxColors = 4 // number of plain ocean glyphs
)

// Options configures a Canvas generator.
type Options struct {
	InputDir  string // directory holding the input seed; defaults to "."
	OutputDir string // directory to write the map and seed; defaults to "."
	AsciiMap  string // map filename to create; defaults to "ascii-map.txt"
	InputSeed string // RNG seed filename (binary or .json); defaults to "randseed"
	Size      int    // square map edge length, 9..99 (required; CLI default 99)
	Oceans    int    // number of oceans, 1..20 and < Size/2 (required; CLI default 7)
	Log       io.Writer
}

// Canvas generates a multi-ocean starting map.
type Canvas struct {
	inputDir, outputDir string
	asciiMap, inputSeed string
	size, oceans        int
	log                 io.Writer
	rng                 *prng.RNG
}

type cell struct{ r, c int }

// New returns a Canvas generator with defaults applied.
func New(opts Options) *Canvas {
	g := &Canvas{
		inputDir:  opts.InputDir,
		outputDir: opts.OutputDir,
		asciiMap:  opts.AsciiMap,
		inputSeed: opts.InputSeed,
		size:      opts.Size,
		oceans:    opts.Oceans,
		log:       opts.Log,
		rng:       prng.NewRNG(),
	}
	if g.inputDir == "" {
		g.inputDir = "."
	}
	if g.outputDir == "" {
		g.outputDir = "."
	}
	if g.asciiMap == "" {
		g.asciiMap = "ascii-map.txt"
	}
	if g.inputSeed == "" {
		g.inputSeed = "randseed"
	}
	// Size and Oceans are not defaulted here; the CLI supplies 99/7 as flag
	// defaults, and validate() rejects out-of-range values (including 0).
	return g
}

func (g *Canvas) logf(format string, a ...any) {
	if g.log != nil {
		_, _ = fmt.Fprintf(g.log, format, a...)
	}
}

// Run validates the parameters, generates the canvas, and writes the map and
// the advanced seed to the output directory. It refuses to overwrite an
// existing map file.
func (g *Canvas) Run() error {
	if err := g.validate(); err != nil {
		return err
	}

	mapPath := filepath.Join(g.outputDir, g.asciiMap)
	if _, err := os.Stat(mapPath); err == nil {
		return fmt.Errorf("canvas: %s already exists; refusing to overwrite", mapPath)
	} else if !os.IsNotExist(err) {
		return err
	}

	if err := store.LoadSeed(g.rng, filepath.Join(g.inputDir, g.inputSeed)); err != nil {
		return err
	}

	region := g.partition()
	colors := g.colorOceans(region)

	if err := g.writeMap(region, colors); err != nil {
		return err
	}
	if err := store.SaveSeed(g.rng, g.outputDir); err != nil {
		return err
	}

	g.logf("Generated %dx%d canvas with %d oceans.\n", g.size, g.size, g.oceans)
	return nil
}

func (g *Canvas) validate() error {
	if g.size < MinSize || g.size > MaxSize {
		return fmt.Errorf("canvas: size %d out of range [%d,%d]", g.size, MinSize, MaxSize)
	}
	if g.oceans < MinOceans || g.oceans > MaxOceans {
		return fmt.Errorf("canvas: oceans %d out of range [%d,%d]", g.oceans, MinOceans, MaxOceans)
	}
	if g.oceans >= g.size/2 {
		return fmt.Errorf("canvas: oceans %d must be less than size/2 (%d)", g.oceans, g.size/2)
	}
	return nil
}

// partition assigns every cell to an ocean (0..oceans-1) via randomized region
// growth from random seed cells.
func (g *Canvas) partition() [][]int {
	region := make([][]int, g.size)
	for r := range region {
		region[r] = make([]int, g.size)
		for c := range region[r] {
			region[r][c] = -1
		}
	}

	var frontier []cell
	for ocean := 0; ocean < g.oceans; ocean++ {
		for {
			r, c := g.rng.Rnd(0, g.size-1), g.rng.Rnd(0, g.size-1)
			if region[r][c] == -1 {
				region[r][c] = ocean
				frontier = append(frontier, cell{r, c})
				break
			}
		}
	}

	unclaimed := g.size*g.size - g.oceans
	dirs := [4]cell{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	for unclaimed > 0 && len(frontier) > 0 {
		i := g.rng.Rnd(0, len(frontier)-1)
		cur := frontier[i]

		var open []cell
		for _, d := range dirs {
			nr, nc := cur.r+d.r, cur.c+d.c
			if nr < 0 || nr >= g.size || nc < 0 || nc >= g.size {
				continue
			}
			if region[nr][nc] == -1 {
				open = append(open, cell{nr, nc})
			}
		}
		if len(open) == 0 {
			// no room to grow here; drop from the frontier
			frontier[i] = frontier[len(frontier)-1]
			frontier = frontier[:len(frontier)-1]
			continue
		}

		pick := open[g.rng.Rnd(0, len(open)-1)]
		region[pick.r][pick.c] = region[cur.r][cur.c]
		frontier = append(frontier, pick)
		unclaimed--
	}

	return region
}

// colorOceans builds the ocean-adjacency graph using mapgen's flood rule
// (8-neighbor, columns wrap, rows do not) and returns a proper <=4 coloring.
func (g *Canvas) colorOceans(region [][]int) []int {
	adj := make([][]bool, g.oceans)
	for i := range adj {
		adj[i] = make([]bool, g.oceans)
	}
	for r := 0; r < g.size; r++ {
		for c := 0; c < g.size; c++ {
			a := region[r][c]
			for dr := -1; dr <= 1; dr++ {
				for dc := -1; dc <= 1; dc++ {
					if dr == 0 && dc == 0 {
						continue
					}
					nr := r + dr
					if nr < 0 || nr >= g.size {
						continue
					}
					nc := (c + dc + g.size) % g.size // columns wrap
					b := region[nr][nc]
					if a != b {
						adj[a][b] = true
						adj[b][a] = true
					}
				}
			}
		}
	}

	colors := make([]int, g.oceans)
	for i := range colors {
		colors[i] = -1
	}
	if !colorVertex(0, g.oceans, adj, colors) {
		// Unreachable: the region-adjacency graph of a contiguous partition is
		// planar and therefore 4-colorable.
		panic("canvas: could not 4-color oceans")
	}
	return colors
}

// colorVertex is a deterministic backtracking graph colorer over maxColors.
func colorVertex(v, n int, adj [][]bool, colors []int) bool {
	if v == n {
		return true
	}
	for color := range maxColors {
		ok := true
		for u := range n {
			if adj[v][u] && colors[u] == color {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}
		colors[v] = color
		if colorVertex(v+1, n, adj, colors) {
			return true
		}
		colors[v] = -1
	}
	return false
}

// writeMap renders the colored partition to the output map file: each cell
// becomes the plain ocean glyph for its ocean's color.
func (g *Canvas) writeMap(region [][]int, colors []int) error {
	buf := make([]byte, 0, g.size*(g.size+1))
	for r := range g.size {
		for c := range g.size {
			buf = append(buf, asciimap.PlainOcean[colors[region[r][c]]])
		}
		buf = append(buf, '\n')
	}
	return os.WriteFile(filepath.Join(g.outputDir, g.asciiMap), buf, 0644)
}
