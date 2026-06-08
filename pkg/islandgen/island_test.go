// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package islandgen_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/mdhender/olyg6/pkg/islandgen"
)

const (
	fixturesDir = "../../testdata/golden/cases/island/fixtures"
	goldenDir   = "../../testdata/golden/cases/island/golden"
)

// goldenOpts are the options used to produce the committed golden output.
func goldenOpts(dir string) islandgen.Options {
	return islandgen.Options{
		InputDir: dir, OutputDir: dir,
		InputMap: "ascii-map.txt", InputSeed: "randseed",
		Border: 2, Shelf: 3, Size: 60, Log: io.Discard,
	}
}

// copyFixtures copies the named fixtures into a fresh temp dir and returns it.
func copyFixtures(t *testing.T, names ...string) string {
	t.Helper()
	tmp := t.TempDir()
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(fixturesDir, name))
		if err != nil {
			t.Fatalf("read fixture %s: %v", name, err)
		}
		if err := os.WriteFile(filepath.Join(tmp, name), data, 0644); err != nil {
			t.Fatalf("write fixture %s: %v", name, err)
		}
	}
	return tmp
}

// TestSeaLaneGuard: a map with any sea-lane glyph is refused, and nothing is
// written to the output dir.
func TestSeaLaneGuard(t *testing.T) {
	for _, sl := range []byte(";:~\"") {
		in := t.TempDir()
		out := t.TempDir()
		row := []byte("....................")
		row[9] = sl
		var b bytes.Buffer
		for i := range 20 {
			if i == 10 {
				b.Write(row)
			} else {
				b.WriteString("....................")
			}
			b.WriteByte('\n')
		}
		if err := os.WriteFile(filepath.Join(in, "ascii-map.txt"), b.Bytes(), 0644); err != nil {
			t.Fatalf("write map: %v", err)
		}
		if err := os.WriteFile(filepath.Join(in, "randseed"), make([]byte, 16), 0644); err != nil {
			t.Fatalf("write seed: %v", err)
		}

		g := islandgen.New(islandgen.Options{InputDir: in, OutputDir: out, Log: io.Discard})
		if err := g.Run(); err == nil {
			t.Errorf("sea lane %q: Run returned nil, want error", rune(sl))
		}
		if entries, _ := os.ReadDir(out); len(entries) != 0 {
			t.Errorf("sea lane %q: output dir not empty (%d entries)", rune(sl), len(entries))
		}
	}
}

// TestDeterministic: same input + options produce identical output, and the
// island actually adds land (terrain letters appear, water is preserved
// elsewhere).
func TestDeterministic(t *testing.T) {
	a := copyFixtures(t, "ascii-map.txt", "randseed")
	b := copyFixtures(t, "ascii-map.txt", "randseed")

	if err := islandgen.New(goldenOpts(a)).Run(); err != nil {
		t.Fatalf("run a: %v", err)
	}
	if err := islandgen.New(goldenOpts(b)).Run(); err != nil {
		t.Fatalf("run b: %v", err)
	}

	mapA, _ := os.ReadFile(filepath.Join(a, "ascii-map.txt"))
	mapB, _ := os.ReadFile(filepath.Join(b, "ascii-map.txt"))
	if !bytes.Equal(mapA, mapB) {
		t.Fatalf("non-deterministic: two runs differ")
	}

	in, _ := os.ReadFile(filepath.Join(fixturesDir, "ascii-map.txt"))
	if bytes.Equal(in, mapA) {
		t.Errorf("output identical to input: no island was added")
	}
	if !bytes.ContainsAny(mapA, "pfmds") {
		t.Errorf("output has no terrain letters")
	}
}

// TestGolden locks the output map and seed against committed goldens.
func TestGolden(t *testing.T) {
	dir := copyFixtures(t, "ascii-map.txt", "randseed")
	if err := islandgen.New(goldenOpts(dir)).Run(); err != nil {
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
