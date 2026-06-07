// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package lists_test

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/mdhender/olyg6/pkg/lists"
	"github.com/mdhender/olyg6/pkg/prng"
)

// TestLifecycle mirrors the C ilist_test/plist_test sequence: build a list,
// copy it, delete, prepend, look up, clear, and reclaim, checking length and
// contents at each step.
func TestLifecycle(t *testing.T) {
	const N = 100

	var l lists.List[int]
	if got := len(l); got != 0 {
		t.Fatalf("empty len = %d, want 0", got)
	}
	if got := l.Lookup(1); got != -1 {
		t.Fatalf("lookup on empty = %d, want -1", got)
	}

	items := make([]int, N)
	for i := range items {
		items[i] = i + 1
	}
	for _, v := range items {
		l.Append(v)
	}
	if len(l) != N {
		t.Fatalf("after append len = %d, want %d", len(l), N)
	}
	for i := range items {
		if l[i] != items[i] {
			t.Fatalf("l[%d] = %d, want %d", i, l[i], items[i])
		}
	}

	c := l.Copy()
	if len(c) != N {
		t.Fatalf("copy len = %d, want %d", len(c), N)
	}
	for i := range c {
		if c[i] != l[i] {
			t.Fatalf("copy[%d] = %d, want %d", i, c[i], l[i])
		}
	}
	// Copy must be independent: mutating it does not touch the original.
	c[0] = -999
	if l[0] == c[0] {
		t.Fatalf("copy is not independent of original")
	}

	l.Delete(50)
	if len(l) != N-1 {
		t.Fatalf("after delete len = %d, want %d", len(l), N-1)
	}
	if l.Lookup(items[50]) != -1 {
		t.Fatalf("deleted value %d still present", items[50])
	}

	l.Prepend(items[50])
	if len(l) != N {
		t.Fatalf("after prepend len = %d, want %d", len(l), N)
	}
	if l[0] != items[50] {
		t.Fatalf("prepend head = %d, want %d", l[0], items[50])
	}
	if l.Lookup(items[50]) != 0 {
		t.Fatalf("lookup after prepend = %d, want 0", l.Lookup(items[50]))
	}

	l.Clear()
	if len(l) != 0 {
		t.Fatalf("after clear len = %d, want 0", len(l))
	}

	l.Reclaim()
	if l != nil {
		t.Fatalf("after reclaim list is not nil")
	}
}

// TestAddDedups checks Add's set-style behavior (ilist_add).
func TestAddDedups(t *testing.T) {
	var l lists.List[int]
	l.Add(7)
	l.Add(7)
	l.Add(9)
	if len(l) != 2 {
		t.Fatalf("len = %d, want 2 (%v)", len(l), l)
	}
	if l.Lookup(7) != 0 || l.Lookup(9) != 1 {
		t.Fatalf("unexpected contents %v", l)
	}
}

// TestRemValue removes all matches; TestRemValueUniq removes only the first.
func TestRemValue(t *testing.T) {
	l := lists.List[int]{1, 2, 3, 2, 2, 4}
	l.RemValue(2)
	want := []int{1, 3, 4}
	if len(l) != len(want) {
		t.Fatalf("len = %d, want %d (%v)", len(l), len(want), l)
	}
	for i := range want {
		if l[i] != want[i] {
			t.Fatalf("l = %v, want %v", l, want)
		}
	}

	u := lists.List[int]{1, 2, 3, 2, 2, 4}
	u.RemValueUniq(2)
	wantU := []int{1, 3, 2, 2, 4}
	for i := range wantU {
		if u[i] != wantU[i] {
			t.Fatalf("RemValueUniq = %v, want %v", u, wantU)
		}
	}
}

// TestDeletePanics confirms out-of-range Delete panics, matching the C assert.
func TestDeletePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatalf("Delete out of range did not panic")
		}
	}()
	l := lists.List[int]{1, 2, 3}
	l.Delete(5)
}

// TestPointerList exercises the plist case: a List of pointers.
func TestPointerList(t *testing.T) {
	a, b, c := new(int), new(int), new(int)
	var l lists.List[*int]
	l.Append(a)
	l.Append(b)
	l.Append(c)
	if l.Lookup(b) != 1 {
		t.Fatalf("lookup b = %d, want 1", l.Lookup(b))
	}
	l.RemValue(b)
	if l.Lookup(b) != -1 || len(l) != 2 {
		t.Fatalf("after RemValue: len=%d lookup=%d", len(l), l.Lookup(b))
	}
	// Delete must zero the vacated trailing slot so the removed pointer is not
	// pinned in the backing array (the var zero T write in Delete).
	l.Append(c)
	l.Delete(2)
	if got := l[:cap(l)][2]; got != nil {
		t.Fatalf("Delete left trailing slot non-nil: %v", got)
	}
}

// TestShuffleDeterministic checks that Shuffle reproduces the C contract: it
// is a permutation, and identical RNG state yields an identical ordering (the
// property that keeps seeded turn outcomes faithful).
func TestShuffleDeterministic(t *testing.T) {
	seed := make([]byte, prng.SeedLen)
	for i := range seed {
		seed[i] = byte(i + 1)
	}

	build := func() lists.List[int] {
		l := make(lists.List[int], 0, 50)
		for i := range 50 {
			l.Append(i)
		}
		return l
	}

	r1 := prng.NewRNG()
	r1.Load(seed)
	a := build()
	a.Shuffle(r1)

	r2 := prng.NewRNG()
	r2.Load(seed)
	b := build()
	b.Shuffle(r2)

	// Same seed -> same ordering.
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("shuffle not deterministic at %d: %d vs %d", i, a[i], b[i])
		}
	}

	// Result is a permutation of 0..49.
	sorted := append(lists.List[int](nil), a...)
	sort.Ints(sorted)
	for i := range sorted {
		if sorted[i] != i {
			t.Fatalf("shuffle dropped/duplicated values: %v", sorted)
		}
	}
}

// TestShuffleGoldenParity asserts byte-for-byte ordering parity with the
// reference C ilist_shuffle. The seed and expected ordering were produced by
// the C program in testdata/golden/cases/lists (see its README). This pins the
// RNG draw order, which is the only part of the port that affects seeded game
// outcomes.
func TestShuffleGoldenParity(t *testing.T) {
	const caseDir = "../../testdata/golden/cases/lists"

	seed, err := os.ReadFile(filepath.Join(caseDir, "fixtures", "seed"))
	if err != nil {
		t.Fatalf("read seed: %v", err)
	}

	want := readInts(t, filepath.Join(caseDir, "golden", "shuffle-50"))

	r := prng.NewRNG()
	r.Load(seed)

	l := make(lists.List[int], 0, len(want))
	for i := range len(want) {
		l.Append(i)
	}
	l.Shuffle(r)

	if len(l) != len(want) {
		t.Fatalf("len = %d, want %d", len(l), len(want))
	}
	for i := range want {
		if l[i] != want[i] {
			t.Fatalf("ordering diverges from C golden at index %d: got %d, want %d\n got:  %v\n want: %v",
				i, l[i], want[i], []int(l), want)
		}
	}
}

// readInts reads a whitespace/newline-separated list of ints from a file.
func readInts(t *testing.T, path string) []int {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()

	var out []int
	sc := bufio.NewScanner(f)
	sc.Split(bufio.ScanWords)
	for sc.Scan() {
		n, err := strconv.Atoi(strings.TrimSpace(sc.Text()))
		if err != nil {
			t.Fatalf("parse %q in %s: %v", sc.Text(), path, err)
		}
		out = append(out, n)
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("scan %s: %v", path, err)
	}
	return out
}
