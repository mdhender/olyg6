package mapgen_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/mdhender/olyg6/pkg/mapgen"
)

// Test fixtures and the golden files produced by the original C program live
// under the shared testdata tree at the module root.
const (
	fixturesDir = "../../testdata/golden/cases/mapgen/fixtures"
	goldenDir   = "../../testdata/golden/cases/mapgen/golden"
)

// TestGoldenParity runs the generator against the committed G3 fixtures and
// asserts that loc, gate, road, and randseed match the golden files produced
// by the original C program byte-for-byte.
func TestGoldenParity(t *testing.T) {
	tmp := t.TempDir()
	// The fixture filenames match mapgen's default input names, so copying
	// them into tmp lets the generator find them with no extra configuration.
	for _, name := range []string{"ascii-map.txt", "cities.txt", "lands.json", "regions.json", "randseed"} {
		data, err := os.ReadFile(filepath.Join(fixturesDir, name))
		if err != nil {
			t.Fatalf("read fixture %s: %v", name, err)
		}
		if err := os.WriteFile(filepath.Join(tmp, name), data, 0644); err != nil {
			t.Fatalf("write fixture %s: %v", name, err)
		}
	}

	g := mapgen.New(mapgen.Options{
		InputDir:  tmp,
		OutputDir: tmp,
		Log:       io.Discard,
	})
	if err := g.Run(); err != nil {
		t.Fatalf("Run: %v", err)
	}

	for _, name := range []string{"loc", "gate", "road", "randseed"} {
		got, err := os.ReadFile(filepath.Join(tmp, name))
		if err != nil {
			t.Fatalf("read output %s: %v", name, err)
		}
		want, err := os.ReadFile(filepath.Join(goldenDir, name))
		if err != nil {
			t.Fatalf("read golden %s: %v", name, err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("%s: output does not match golden (got %d bytes, want %d bytes)",
				name, len(got), len(want))
		}
	}
}
