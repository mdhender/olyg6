package mapgen_test

import (
	"bufio"
	"bytes"
	"encoding/hex"
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
	for _, name := range []string{"loc.json", "gate.json", "road.json", "randseed.json"} {
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

// TestRandSeedJSON checks that randseed.json's hex string decodes to exactly
// the bytes in the flat randseed file.
func TestRandSeedJSON(t *testing.T) {
	tmp := generateIntoTemp(t)

	var seed store.RandSeed
	readJSON(t, filepath.Join(tmp, "randseed.json"), &seed)
	got, err := hex.DecodeString(seed.Seed)
	if err != nil {
		t.Fatalf("decode seed hex %q: %v", seed.Seed, err)
	}

	want, err := os.ReadFile(filepath.Join(tmp, "randseed"))
	if err != nil {
		t.Fatalf("read flat randseed: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("randseed.json decodes to %x, flat randseed is %x", got, want)
	}
}

// TestSeedInputParity proves that loading the seed from randseed.json (hex)
// produces byte-identical output to loading the legacy binary randseed.
func TestSeedInputParity(t *testing.T) {
	// Reference run: binary seed (the default input name).
	binDir := generateIntoTemp(t)

	// JSON-seed run: same world, but the seed supplied as randseed.json.
	jsonDir := t.TempDir()
	for _, name := range []string{"ascii-map.txt", "cities.txt", "lands.json", "regions.json"} {
		data, err := os.ReadFile(filepath.Join(fixturesDir, name))
		if err != nil {
			t.Fatalf("read fixture %s: %v", name, err)
		}
		if err := os.WriteFile(filepath.Join(jsonDir, name), data, 0644); err != nil {
			t.Fatalf("write fixture %s: %v", name, err)
		}
	}
	seedBytes, err := os.ReadFile(filepath.Join(fixturesDir, "randseed"))
	if err != nil {
		t.Fatalf("read fixture randseed: %v", err)
	}
	seedJSON, err := json.MarshalIndent(store.RandSeed{Seed: hex.EncodeToString(seedBytes)}, "", "  ")
	if err != nil {
		t.Fatalf("marshal seed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(jsonDir, "randseed.json"), seedJSON, 0644); err != nil {
		t.Fatalf("write randseed.json: %v", err)
	}

	g := mapgen.New(mapgen.Options{
		InputDir: jsonDir, OutputDir: jsonDir, InputSeed: "randseed.json", Log: io.Discard,
	})
	if err := g.Run(); err != nil {
		t.Fatalf("Run with json seed: %v", err)
	}

	for _, name := range []string{"loc", "gate", "road", "randseed", "loc.json", "gate.json", "road.json", "randseed.json"} {
		a, err := os.ReadFile(filepath.Join(binDir, name))
		if err != nil {
			t.Fatalf("read %s (bin): %v", name, err)
		}
		b, err := os.ReadFile(filepath.Join(jsonDir, name))
		if err != nil {
			t.Fatalf("read %s (json): %v", name, err)
		}
		if !bytes.Equal(a, b) {
			t.Errorf("%s differs between binary-seed and json-seed runs", name)
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
