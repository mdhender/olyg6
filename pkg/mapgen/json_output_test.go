package mapgen_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"testing"

	"github.com/mdhender/olyg6/pkg/mapgen"
	"github.com/mdhender/olyg6/pkg/store"
)

// generateIntoTemp runs the generator against the committed fixtures in a fresh
// temp dir and returns that dir, which then holds both the flat and JSON output.
func generateIntoTemp(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	for _, name := range []string{"ascii-map.txt", "cities.txt", "lands.json", "regions.json", "randseed"} {
		data, err := os.ReadFile(filepath.Join(fixturesDir, name))
		if err != nil {
			t.Fatalf("read fixture %s: %v", name, err)
		}
		if err := os.WriteFile(filepath.Join(tmp, name), data, 0644); err != nil {
			t.Fatalf("write fixture %s: %v", name, err)
		}
	}
	g := mapgen.New(mapgen.Options{InputDir: tmp, OutputDir: tmp, Log: io.Discard})
	if err := g.Run(); err != nil {
		t.Fatalf("Run: %v", err)
	}
	return tmp
}

// flatIDs returns the set of entity ids declared in a flat-file store. Every
// box begins with a header line "<id> <kind> <subkind>" in the first column;
// value/section lines are indented or alphabetic, so a leading-digit line is a
// header.
func flatIDs(t *testing.T, path string) map[int]bool {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()

	ids := map[int]bool{}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || line[0] < '0' || line[0] > '9' {
			continue
		}
		fields := bytes.Fields([]byte(line))
		id, err := strconv.Atoi(string(fields[0]))
		if err != nil {
			t.Fatalf("bad header %q in %s: %v", line, path, err)
		}
		ids[id] = true
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("scan %s: %v", path, err)
	}
	return ids
}

// TestJSONStoreConsistency proves the JSON store carries exactly the same
// entities as the flat store: identical id sets per kind (loc/gate/road).
func TestJSONStoreConsistency(t *testing.T) {
	tmp := generateIntoTemp(t)

	// loc
	var locs []store.Location
	readJSON(t, filepath.Join(tmp, "loc.json"), &locs)
	gotLoc := map[int]bool{}
	for _, l := range locs {
		if l.Kind != "loc" {
			t.Errorf("loc.json entry %d has kind %q, want loc", l.Id, l.Kind)
		}
		gotLoc[l.Id] = true
	}
	assertSameIDs(t, "loc", flatIDs(t, filepath.Join(tmp, "loc")), gotLoc)

	// gate
	var gates []store.Gate
	readJSON(t, filepath.Join(tmp, "gate.json"), &gates)
	gotGate := map[int]bool{}
	for _, g := range gates {
		if g.Kind != "gate" {
			t.Errorf("gate.json entry %d has kind %q, want gate", g.Id, g.Kind)
		}
		gotGate[g.Id] = true
	}
	assertSameIDs(t, "gate", flatIDs(t, filepath.Join(tmp, "gate")), gotGate)

	// road
	var roads []store.Road
	readJSON(t, filepath.Join(tmp, "road.json"), &roads)
	gotRoad := map[int]bool{}
	for _, r := range roads {
		if r.Kind != "road" {
			t.Errorf("road.json entry %d has kind %q, want road", r.Id, r.Kind)
		}
		gotRoad[r.Id] = true
	}
	assertSameIDs(t, "road", flatIDs(t, filepath.Join(tmp, "road")), gotRoad)
}

// TestJSONStoreGolden locks the exact JSON output (schema + formatting) against
// committed goldens. Unlike the flat goldens (which came from the reference C),
// these are produced by our Go and were spot-checked against the flat store at
// creation; the test guards against accidental drift.
func TestJSONStoreGolden(t *testing.T) {
	tmp := generateIntoTemp(t)
	for _, name := range []string{"loc.json", "gate.json", "road.json"} {
		got, err := os.ReadFile(filepath.Join(tmp, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
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

func readJSON(t *testing.T, path string, v any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}
}

func assertSameIDs(t *testing.T, kind string, want, got map[int]bool) {
	t.Helper()
	if len(want) != len(got) {
		t.Errorf("%s: id count flat=%d json=%d", kind, len(want), len(got))
	}
	var missing, extra []int
	for id := range want {
		if !got[id] {
			missing = append(missing, id)
		}
	}
	for id := range got {
		if !want[id] {
			extra = append(extra, id)
		}
	}
	sort.Ints(missing)
	sort.Ints(extra)
	if len(missing) > 0 {
		t.Errorf("%s: ids in flat but not json: %v", kind, missing)
	}
	if len(extra) > 0 {
		t.Errorf("%s: ids in json but not flat: %v", kind, extra)
	}
}
