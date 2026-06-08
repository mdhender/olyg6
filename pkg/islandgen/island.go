// Copyright (c) 2026 Michael D Henderson. All rights reserved.

// Package islandgen is a Go port of the Olympia G3 island generator
// (g3/mapgen/island.c). It adds one randomly-shaped, terrain-filled island to
// an ASCII art world map.
//
// island runs BEFORE the map generator (pkg/mapgen): you chain it to grow a
// world's land masses, then feed the finished ASCII map to mapgen. Because
// adding land reshapes the world, island refuses to edit a map that already
// contains sea lanes — those signal the author has finalized the land masses
// and is adding finishing touches (see Run).
package islandgen

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mdhender/olyg6/pkg/asciimap"
	"github.com/mdhender/olyg6/pkg/prng"
	"github.com/mdhender/olyg6/pkg/store"
)

// distanceCap bounds the distance-from-land field, matching the C DISTANCE_CAP.
const distanceCap = 9

// Options configures an Island generator.
type Options struct {
	InputDir  string // directory holding the input map and seed; defaults to "."
	OutputDir string // directory to write the updated map and seed; defaults to "."
	InputMap  string // ASCII map filename; defaults to "ascii-map.txt"
	InputSeed string // RNG seed filename (binary or .json); defaults to "randseed"
	Border    int    // edge border kept clear of new land (C default 2)
	Shelf     int    // continental shelf kept around existing land (C default 3)
	Size      int    // target island size; 0 selects a random size
	Log       io.Writer
}

// Island generates a single island onto an ASCII map.
type Island struct {
	inputDir, outputDir string
	inputMap, inputSeed string
	border, shelf, size int
	log                 io.Writer
	rng                 *prng.RNG

	grid         [][]byte // the map, mutated in place
	work         [][]byte // availability classification
	ids          [][]int  // grid cell -> index in the island slice
	dist         [][]int  // distance from nearest non-water cell
	ySize, xSize int
}

type loc struct{ x, y int }

type terrain struct {
	symbol                     byte
	min, max, targetProb, prob int
}

// New returns an Island generator with defaults applied.
func New(opts Options) *Island {
	g := &Island{
		inputDir:  opts.InputDir,
		outputDir: opts.OutputDir,
		inputMap:  opts.InputMap,
		inputSeed: opts.InputSeed,
		border:    opts.Border,
		shelf:     opts.Shelf,
		size:      opts.Size,
		log:       opts.Log,
		rng:       prng.NewRNG(),
	}
	if g.inputDir == "" {
		g.inputDir = "."
	}
	if g.outputDir == "" {
		g.outputDir = "."
	}
	if g.inputMap == "" {
		g.inputMap = "ascii-map.txt"
	}
	if g.inputSeed == "" {
		g.inputSeed = "randseed"
	}
	if g.border < 0 {
		g.border = 0
	}
	if g.shelf < 0 {
		g.shelf = 0
	}
	return g
}

func (g *Island) logf(format string, a ...any) {
	if g.log != nil {
		_, _ = fmt.Fprintf(g.log, format, a...)
	}
}

// Run adds one island to the input map and writes the updated map and seed to
// the output directory. It refuses (with an error, before consuming the seed)
// any map that already contains a sea lane.
func (g *Island) Run() error {
	if err := g.readMap(); err != nil {
		return err
	}
	if err := g.checkSeaLanes(); err != nil {
		return err
	}
	if err := store.LoadSeed(g.rng, filepath.Join(g.inputDir, g.inputSeed)); err != nil {
		return err
	}

	g.classify()
	g.makeShelves()
	g.excludeBorders()
	g.computeDistance()

	target := g.size
	if target < 1 {
		// Sum of two geometric draws (mean ~100 each) for a ~200 mean.
		for range 2 {
			for g.rng.Rnd(0, 99) != 0 {
				target++
			}
		}
	}

	island, islandSize := g.growIsland(target)
	g.assignTerrain(island, islandSize)

	if err := store.SaveSeed(g.rng, g.outputDir); err != nil {
		return err
	}
	if err := g.writeMap(); err != nil {
		return err
	}

	g.logf("Added island of %d provinces.\n", islandSize)
	return nil
}

// readMap loads the ASCII map. Following the C reader, the width is taken from
// the first non-empty line; only lines of exactly that width are kept, and
// empty lines are skipped.
func (g *Island) readMap() error {
	path := filepath.Join(g.inputDir, g.inputMap)
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("island: can't read %s: %w", path, err)
	}

	xSize := 0
	var grid [][]byte
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		if xSize == 0 {
			xSize = len(line)
		}
		if len(line) == xSize {
			grid = append(grid, []byte(line))
		}
	}
	if len(grid) == 0 {
		return fmt.Errorf("island: empty map %s", path)
	}

	g.grid = grid
	g.xSize = xSize
	g.ySize = len(grid)
	return nil
}

// checkSeaLanes refuses maps that already contain a sea-lane glyph.
func (g *Island) checkSeaLanes() error {
	for y, row := range g.grid {
		for x := range row {
			if asciimap.IsSeaLane(row[x]) {
				return fmt.Errorf("island: map %s contains a sea lane %q at row %d col %d; "+
					"island only edits maps before sea lanes are added", g.inputMap, rune(row[x]), y, x)
			}
		}
	}
	return nil
}

// classify divides the map into available water ('~') and unavailable land
// ('p'). Every ocean glyph counts as water (sea lanes are already rejected).
func (g *Island) classify() {
	g.work = make([][]byte, g.ySize)
	g.ids = make([][]int, g.ySize)
	for y := 0; y < g.ySize; y++ {
		g.work[y] = make([]byte, g.xSize)
		g.ids[y] = make([]int, g.xSize)
		for x := 0; x < g.xSize; x++ {
			if asciimap.IsOcean(g.grid[y][x]) {
				g.work[y][x] = '~'
			} else {
				g.work[y][x] = 'p'
			}
		}
	}
}

// makeShelves keeps a continental shelf of clear water around existing land.
func (g *Island) makeShelves() {
	for y := 0; y < g.ySize; y++ {
		for x := 0; x < g.xSize; x++ {
			if g.work[y][x] == 'p' {
				g.makeShelf(y, x, g.shelf)
			}
		}
	}
}

func (g *Island) makeShelf(y, x, shelf int) {
	if g.work[y][x] == '~' {
		g.work[y][x] = '_'
	}
	if shelf < 1 {
		return
	}
	if y > 0 {
		g.makeShelf(y-1, x, shelf-1)
	}
	if y < g.ySize-1 {
		g.makeShelf(y+1, x, shelf-1)
	}
	if x > 0 {
		g.makeShelf(y, x-1, shelf-1)
	}
	if x < g.xSize-1 {
		g.makeShelf(y, x+1, shelf-1)
	}
}

// excludeBorders clears a border of water around each edge.
func (g *Island) excludeBorders() {
	for y := 0; y < g.ySize; y++ {
		for x := 0; x < g.border; x++ {
			if g.work[y][x] == '~' {
				g.work[y][x] = '_'
			}
			if g.work[y][g.xSize-x-1] == '~' {
				g.work[y][g.xSize-x-1] = '_'
			}
		}
	}
	for x := 0; x < g.xSize; x++ {
		for y := 0; y < g.border; y++ {
			if g.work[y][x] == '~' {
				g.work[y][x] = '_'
			}
			if g.work[g.ySize-y-1][x] == '~' {
				g.work[g.ySize-y-1][x] = '_'
			}
		}
	}
}

// computeDistance fills dist with each water cell's distance from the nearest
// non-water cell, capped at distanceCap.
func (g *Island) computeDistance() {
	g.dist = make([][]int, g.ySize)
	for y := 0; y < g.ySize; y++ {
		g.dist[y] = make([]int, g.xSize)
		for x := 0; x < g.xSize; x++ {
			if g.work[y][x] == '~' {
				g.dist[y][x] = distanceCap
			} else {
				g.dist[y][x] = 0
			}
		}
	}
	for y := 0; y < g.ySize; y++ {
		for x := 0; x < g.xSize; x++ {
			g.extendDistance(y, x)
		}
	}
}

func (g *Island) extendDistance(y, x int) {
	if y > 0 && g.dist[y-1][x] > g.dist[y][x]+1 {
		g.dist[y-1][x] = g.dist[y][x] + 1
		g.extendDistance(y-1, x)
	}
	if y < g.ySize-1 && g.dist[y+1][x] > g.dist[y][x]+1 {
		g.dist[y+1][x] = g.dist[y][x] + 1
		g.extendDistance(y+1, x)
	}
	if x > 0 && g.dist[y][x-1] > g.dist[y][x]+1 {
		g.dist[y][x-1] = g.dist[y][x] + 1
		g.extendDistance(y, x-1)
	}
	if x < g.xSize-1 && g.dist[y][x+1] > g.dist[y][x]+1 {
		g.dist[y][x+1] = g.dist[y][x] + 1
		g.extendDistance(y, x+1)
	}
}

// growIsland seeds the island at the cell farthest from land, then grows it to
// target by random 4-neighbor expansion (multiplicity favors filling
// interiors). Cells are marked 'o' in both work and grid.
func (g *Island) growIsland(target int) ([]loc, int) {
	island := make([]loc, target+1)

	max, count := 0, 0
	for y := 0; y < g.ySize; y++ {
		for x := 0; x < g.xSize; x++ {
			if g.dist[y][x] > max {
				max = g.dist[y][x]
				count = 1
			} else if g.dist[y][x] == max {
				count++
			}
		}
	}

	d := g.rng.Rnd(1, count)
	islandSize := 0
	for y := 0; d > 0 && y < g.ySize; y++ {
		for x := 0; d > 0 && x < g.xSize; x++ {
			if g.dist[y][x] == max {
				d--
				if d == 0 {
					island[0] = loc{x, y}
					islandSize = 1
					g.work[y][x] = 'o'
					g.grid[y][x] = 'o'
					g.ids[y][x] = 0
				}
			}
		}
	}

	for islandSize < target {
		count = 0
		for i := 0; i < islandSize; i++ {
			ix, iy := island[i].x, island[i].y
			if ix > 0 && g.work[iy][ix-1] == '~' {
				count++
			}
			if ix < g.xSize-1 && g.work[iy][ix+1] == '~' {
				count++
			}
			if iy > 0 && g.work[iy-1][ix] == '~' {
				count++
			}
			if iy < g.ySize-1 && g.work[iy+1][ix] == '~' {
				count++
			}
		}
		if count < 1 {
			g.logf("Not enough room to expand island!\n")
			break
		}

		d = g.rng.Rnd(0, count-1)
		for i := 0; i < islandSize; i++ {
			ix, iy := island[i].x, island[i].y
			if ix > 0 && g.work[iy][ix-1] == '~' {
				if d == 0 {
					island[islandSize] = loc{ix - 1, iy}
				}
				d--
			}
			if ix < g.xSize-1 && g.work[iy][ix+1] == '~' {
				if d == 0 {
					island[islandSize] = loc{ix + 1, iy}
				}
				d--
			}
			if iy > 0 && g.work[iy-1][ix] == '~' {
				if d == 0 {
					island[islandSize] = loc{ix, iy - 1}
				}
				d--
			}
			if iy < g.ySize-1 && g.work[iy+1][ix] == '~' {
				if d == 0 {
					island[islandSize] = loc{ix, iy + 1}
				}
				d--
			}
		}

		nx, ny := island[islandSize].x, island[islandSize].y
		g.work[ny][nx] = 'o'
		g.grid[ny][nx] = 'o'
		g.ids[ny][nx] = islandSize
		islandSize++
	}

	return island, islandSize
}

// assignTerrain colors the island's cells with terrain letters, growing
// contiguous clusters of each type in proportion to the terrain ratios.
func (g *Island) assignTerrain(island []loc, islandSize int) {
	terrains := []terrain{
		{'p', 12, 30, 30, 0},
		{'f', 6, 14, 30, 0},
		{'m', 6, 10, 20, 0},
		{'d', 15, 30, 10, 0},
		{'s', 1, 3, 10, 0},
	}

	// Render the desired ratios down to integer weights via GCD/LCM.
	for i := range terrains {
		terrains[i].prob = terrains[i].min + terrains[i].max
		gg := gcd(terrains[i].targetProb, terrains[i].prob)
		terrains[i].targetProb /= gg
		terrains[i].prob /= gg
	}
	lcmv := 1
	for i := range terrains {
		lcmv = lcm(lcmv, terrains[i].prob)
	}
	total := 0
	for i := range terrains {
		terrains[i].targetProb *= lcmv / terrains[i].prob
		total += terrains[i].targetProb
	}

	o := islandSize
	size := 0
	terr := 0
	for o > 0 {
		clusterEnd := o
		o--
		d := g.rng.Rnd(0, o)
		island[d], island[o] = island[o], island[d]
		g.ids[island[o].y][island[o].x] = o
		g.ids[island[d].y][island[d].x] = d

		if size < 1 {
			d = g.rng.Rnd(1, total)
			for terr = 0; d > terrains[terr].targetProb; terr++ {
				d -= terrains[terr].targetProb
			}
			size = g.rng.Rnd(terrains[terr].min, terrains[terr].max)
		}

		sym := terrains[terr].symbol
		g.grid[island[o].y][island[o].x] = sym
		g.work[island[o].y][island[o].x] = sym
		size--
		if size > o {
			size = o
		}

		for size > 0 {
			opt := 0
			for i := o; i < clusterEnd; i++ {
				ix, iy := island[i].x, island[i].y
				if iy > 0 && g.work[iy-1][ix] == 'o' {
					opt++
				}
				if iy < g.ySize-1 && g.work[iy+1][ix] == 'o' {
					opt++
				}
				if ix > 0 && g.work[iy][ix-1] == 'o' {
					opt++
				}
				if ix < g.xSize-1 && g.work[iy][ix+1] == 'o' {
					opt++
				}
			}
			if opt < 1 {
				break
			}

			d = g.rng.Rnd(0, opt-1)
			id := 0
			for i := o; i < clusterEnd; i++ {
				ix, iy := island[i].x, island[i].y
				if iy > 0 && g.work[iy-1][ix] == 'o' {
					if d == 0 {
						id = g.ids[iy-1][ix]
						break
					}
					d--
				}
				if iy < g.ySize-1 && g.work[iy+1][ix] == 'o' {
					if d == 0 {
						id = g.ids[iy+1][ix]
						break
					}
					d--
				}
				if ix > 0 && g.work[iy][ix-1] == 'o' {
					if d == 0 {
						id = g.ids[iy][ix-1]
						break
					}
					d--
				}
				if ix < g.xSize-1 && g.work[iy][ix+1] == 'o' {
					if d == 0 {
						id = g.ids[iy][ix+1]
						break
					}
					d--
				}
			}

			o--
			size--
			island[id], island[o] = island[o], island[id]
			g.grid[island[o].y][island[o].x] = sym
			g.work[island[o].y][island[o].x] = sym
			g.ids[island[o].y][island[o].x] = o
			g.ids[island[id].y][island[id].x] = id
		}
	}
}

// writeMap writes the updated ASCII map to the output directory under the same
// basename as the input map.
func (g *Island) writeMap() error {
	var b bytes.Buffer
	for _, row := range g.grid {
		b.Write(row)
		b.WriteByte('\n')
	}
	return os.WriteFile(filepath.Join(g.outputDir, g.inputMap), b.Bytes(), 0644)
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func lcm(a, b int) int {
	return a * b / gcd(a, b)
}
