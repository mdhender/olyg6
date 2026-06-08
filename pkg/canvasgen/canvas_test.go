// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package canvasgen_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/mdhender/olyg6/pkg/asciimap"
	"github.com/mdhender/olyg6/pkg/canvasgen"
)

const (
	fixturesDir = "../../testdata/golden/cases/canvas/fixtures"
	goldenDir   = "../../testdata/golden/cases/canvas/golden"
)

// goldenOpts are the options used to produce the committed golden output.
func goldenOpts(dir string) canvasgen.Options {
	return canvasgen.Options{
		InputDir: dir, OutputDir: dir,
		AsciiMap: "ascii-map.txt", InputSeed: "randseed",
		Size: 40, Oceans: 5, Log: io.Discard,
	}
}

// seedDir returns a temp dir holding a copy of the fixture seed.
func seedDir(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	data, err := os.ReadFile(filepath.Join(fixturesDir, "randseed"))
	if err != nil {
		t.Fatalf("read fixture seed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "randseed"), data, 0644); err != nil {
		t.Fatalf("write seed: %v", err)
	}
	return tmp
}

func TestValidation(t *testing.T) {
	cases := []struct {
		name         string
		size, oceans int
	}{
		{"size too small", 8, 3},
		{"size too large", 100, 3},
		{"oceans too few", 40, 0},
		{"oceans too many", 99, 21},
		{"oceans >= size/2", 10, 5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := seedDir(t)
			g := canvasgen.New(canvasgen.Options{
				InputDir: dir, OutputDir: dir, InputSeed: "randseed",
				Size: tc.size, Oceans: tc.oceans, Log: io.Discard,
			})
			if err := g.Run(); err == nil {
				t.Errorf("expected error for size=%d oceans=%d", tc.size, tc.oceans)
			}
		})
	}
}

func TestOverwriteGuard(t *testing.T) {
	dir := seedDir(t)
	mapPath := filepath.Join(dir, "ascii-map.txt")
	if err := os.WriteFile(mapPath, []byte("PRECIOUS\n"), 0644); err != nil {
		t.Fatalf("seed map: %v", err)
	}
	if err := canvasgen.New(goldenOpts(dir)).Run(); err == nil {
		t.Fatalf("expected error: map already exists")
	}
	got, _ := os.ReadFile(mapPath)
	if string(got) != "PRECIOUS\n" {
		t.Errorf("existing map was overwritten")
	}
}

// TestRegionCount asserts the colored output yields exactly Oceans regions under
// mapgen's flood rule (8-neighbor, equal glyph, columns wrap), i.e. no merges.
func TestRegionCount(t *testing.T) {
	dir := seedDir(t)
	opts := goldenOpts(dir)
	if err := canvasgen.New(opts).Run(); err != nil {
		t.Fatalf("run: %v", err)
	}
	rows := readGrid(t, filepath.Join(dir, "ascii-map.txt"), opts.Size)

	if got := countRegions(rows); got != opts.Oceans {
		t.Errorf("region count = %d, want %d", got, opts.Oceans)
	}

	// no sea lanes; every glyph is a plain ocean glyph
	for _, row := range rows {
		for _, b := range row {
			if asciimap.IsSeaLane(b) {
				t.Fatalf("output contains a sea lane %q", rune(b))
			}
			if bytes.IndexByte([]byte(asciimap.PlainOcean), b) < 0 {
				t.Fatalf("output contains non-plain-ocean glyph %q", rune(b))
			}
		}
	}
}

func TestDeterministic(t *testing.T) {
	a := seedDir(t)
	b := seedDir(t)
	if err := canvasgen.New(goldenOpts(a)).Run(); err != nil {
		t.Fatalf("run a: %v", err)
	}
	if err := canvasgen.New(goldenOpts(b)).Run(); err != nil {
		t.Fatalf("run b: %v", err)
	}
	ma, _ := os.ReadFile(filepath.Join(a, "ascii-map.txt"))
	mb, _ := os.ReadFile(filepath.Join(b, "ascii-map.txt"))
	if !bytes.Equal(ma, mb) {
		t.Errorf("non-deterministic: two runs differ")
	}
}

func TestGolden(t *testing.T) {
	dir := seedDir(t)
	if err := canvasgen.New(goldenOpts(dir)).Run(); err != nil {
		t.Fatalf("run: %v", err)
	}
	for _, name := range []string{"ascii-map.txt", "randseed", "randseed.json"} {
		got, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("read output %s: %v", name, err)
		}
		want, err := os.ReadFile(filepath.Join(goldenDir, name))
		if err != nil {
			t.Fatalf("read golden %s: %v", name, err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("%s does not match golden (got %d bytes, want %d)", name, len(got), len(want))
		}
	}
}

func readGrid(t *testing.T, path string, size int) [][]byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var rows [][]byte
	for line := range bytes.SplitSeq(data, []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		rows = append(rows, line)
	}
	if len(rows) != size {
		t.Fatalf("got %d rows, want %d", len(rows), size)
	}
	return rows
}

// countRegions counts connected same-glyph components using mapgen's adjacency:
// 8-neighbor, columns wrap, rows do not.
func countRegions(rows [][]byte) int {
	h, w := len(rows), len(rows[0])
	seen := make([][]bool, h)
	for i := range seen {
		seen[i] = make([]bool, w)
	}
	count := 0
	for r := range h {
		for c := range w {
			if seen[r][c] {
				continue
			}
			count++
			glyph := rows[r][c]
			stack := []int{r*w + c}
			seen[r][c] = true
			for len(stack) > 0 {
				p := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				y, x := p/w, p%w
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						if dy == 0 && dx == 0 {
							continue
						}
						ny := y + dy
						if ny < 0 || ny >= h {
							continue
						}
						nx := (x + dx + w) % w
						if !seen[ny][nx] && rows[ny][nx] == glyph {
							seen[ny][nx] = true
							stack = append(stack, ny*w+nx)
						}
					}
				}
			}
		}
	}
	return count
}
