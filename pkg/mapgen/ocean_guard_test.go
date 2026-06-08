// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package mapgen

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRequireOcean checks the threshold: < 3 ocean provinces is rejected.
func TestRequireOcean(t *testing.T) {
	g := New(Options{Log: io.Discard})
	for _, n := range []int{0, 1, 2} {
		g.WaterCount = n
		if err := g.requireOcean(); err == nil {
			t.Errorf("WaterCount=%d: want error, got nil", n)
		}
	}
	for _, n := range []int{3, 100} {
		g.WaterCount = n
		if err := g.requireOcean(); err != nil {
			t.Errorf("WaterCount=%d: unexpected error: %v", n, err)
		}
	}
}

// TestRunRejectsOceanlessMap proves Run returns a clear error (instead of the
// old randomIsland nil-pointer panic) when the map has no ocean.
func TestRunRejectsOceanlessMap(t *testing.T) {
	tmp := t.TempDir()
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(tmp, name), []byte(body), 0644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	// a 20x20 all-plains map: no ocean at all
	var b strings.Builder
	for range 20 {
		b.WriteString(strings.Repeat("p", 20))
		b.WriteByte('\n')
	}
	write("ascii-map.txt", b.String())
	write("regions.json", "[]")
	write("lands.json", "[]")
	write("cities.txt", "Alpha\nBravo\nCharlie\n")
	if err := os.WriteFile(filepath.Join(tmp, "randseed"), make([]byte, 16), 0644); err != nil {
		t.Fatalf("write seed: %v", err)
	}

	g := New(Options{InputDir: tmp, OutputDir: tmp, Log: io.Discard})
	err := g.Run()
	if err == nil {
		t.Fatalf("Run on an ocean-free map returned nil; want an error")
	}
	if !strings.Contains(err.Error(), "ocean") {
		t.Errorf("error = %q, want it to mention ocean", err)
	}
}
