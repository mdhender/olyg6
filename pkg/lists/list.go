// Copyright (c) 2026 Michael D Henderson. All rights reserved.

// Package lists is a Go port of the G3 reallocing-array helpers ilist and
// plist (see lib/ilist.c and lib/plist.c in the reference C source).
//
// The C versions are bare C arrays with two hidden header slots in front of
// the user-visible pointer: base[0] holds the length, base[1] the capacity,
// and callers are handed &base[2] so they can index the list from 0. That
// whole apparatus exists only to fake a growable array in C. A Go slice is
// already pointer+len+cap, so this port keeps the *operations* (and, crucially,
// the RNG draw order in Shuffle) but backs them with an ordinary slice instead
// of reproducing the hidden-header layout.
//
// ilist (int *) and plist (void **) collapse into a single generic List[T]:
// ilist is List[int], plist is List of some pointer type.
package lists

import "github.com/mdhender/olyg6/pkg/prng"

// List is a growable, slice-backed list. T must be comparable so Lookup, Add,
// and RemValue can test values with ==, matching the C == comparisons.
type List[T comparable] []T

// Append adds n to the end of the list. Ports ilist_append/plist_append.
func (l *List[T]) Append(n T) {
	*l = append(*l, n)
}

// Add appends n only if it is not already present (set-style insert). Ports
// ilist_add. (plist has no add in the C, but it is harmless to share here.)
func (l *List[T]) Add(n T) {
	if l.Lookup(n) == -1 {
		*l = append(*l, n)
	}
}

// Prepend inserts n at the front, shifting the rest up one. Ports
// ilist_prepend/plist_prepend.
func (l *List[T]) Prepend(n T) {
	*l = append(*l, n) // grow by one; value overwritten below
	copy((*l)[1:], (*l)[:len(*l)-1])
	(*l)[0] = n
}

// Lookup returns the index of the first element equal to n, or -1 if absent.
// A nil list returns -1. Ports ilist_lookup/plist_lookup.
func (l List[T]) Lookup(n T) int {
	for i, v := range l {
		if v == n {
			return i
		}
	}
	return -1
}

// Delete removes the element at index i, preserving order. It panics if i is
// out of range, matching the C assert. Ports ilist_delete/plist_delete.
func (l *List[T]) Delete(i int) {
	s := *l
	if i < 0 || i >= len(s) {
		panic("lists: Delete index out of range")
	}
	copy(s[i:], s[i+1:])
	var zero T
	s[len(s)-1] = zero // drop reference so pointers can be collected
	*l = s[:len(s)-1]
}

// RemValue removes every element equal to n. Ports ilist_rem_value/
// plist_rem_value (both remove all matches).
func (l *List[T]) RemValue(n T) {
	for i := len(*l) - 1; i >= 0; i-- {
		if (*l)[i] == n {
			l.Delete(i)
		}
	}
}

// RemValueUniq removes only the first element equal to n. Ports
// ilist_rem_value_uniq. (Declared but never defined for plist in the C.)
func (l *List[T]) RemValueUniq(n T) {
	if i := l.Lookup(n); i != -1 {
		l.Delete(i)
	}
}

// Clear truncates the list to length 0 while retaining capacity. Ports
// ilist_clear/plist_clear.
func (l *List[T]) Clear() {
	*l = (*l)[:0]
}

// Copy returns an independent copy of the list. A nil list copies to nil,
// matching the C (which returns NULL). Ports ilist_copy/plist_copy.
func (l List[T]) Copy() List[T] {
	if l == nil {
		return nil
	}
	c := make(List[T], len(l))
	copy(c, l)
	return c
}

// Reclaim releases the list, setting it to nil. Ports ilist_reclaim/
// plist_reclaim.
func (l *List[T]) Reclaim() {
	*l = nil
}

// Shuffle randomly permutes the list in place using the game RNG. It is a
// direct port of ilist_shuffle/plist_shuffle: a Fisher-Yates shuffle that
// draws r = rnd(i, len-1) for each position, so given an identical RNG state
// it reproduces the C ordering exactly.
func (l List[T]) Shuffle(r *prng.RNG) {
	last := len(l) - 1
	for i := range last {
		j := r.Rnd(i, last)
		if j != i {
			l[i], l[j] = l[j], l[i]
		}
	}
}

// Scramble is an alias for Shuffle, matching the C ilist_scramble/
// plist_scramble (which simply call shuffle).
func (l List[T]) Scramble(r *prng.RNG) {
	l.Shuffle(r)
}
