// Copyright (c) 2026 Michael D Henderson. All rights reserved.

// Package prng implements.
// The RNG type reproduces the original MD5-based pseudo random
// number generator so that, given an identical seed, this package produces
// byte-for-byte identical output to the legacy C program.
package prng

import (
	"crypto/md5"
	"encoding/binary"
	"os"
)

// SeedLen is the size, in bytes, of the RNG seed/state.
const SeedLen = 16

// RNG is the deterministic, MD5-based random number generator used by the
// original Olympia map generator. Its internal state is a 16-byte digest
// that is re-hashed on every draw.
type RNG struct {
	Digest [SeedLen]byte
}

// NewRNG returns an RNG with a zeroed state. Callers normally seed it with
// Load or LoadSeed before drawing numbers.
func NewRNG() *RNG {
	return &RNG{}
}

// Load sets the RNG state from the first SeedLen bytes of b. Bytes beyond
// SeedLen are ignored; a short slice leaves the remaining state bytes zero.
// This mirrors the C load_seed behaviour of fread()-ing up to 16 bytes.
func (r *RNG) Load(b []byte) {
	r.Digest = [SeedLen]byte{}
	copy(r.Digest[:], b)
}

// LoadSeed reads the RNG state from the named file, matching the C
// load_seed function (which reads up to sizeof(digest) == 16 bytes).
func (r *RNG) LoadSeed(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	r.Load(b)
	return nil
}

// Seed returns a copy of the current RNG state.
func (r *RNG) Seed() []byte {
	out := make([]byte, SeedLen)
	copy(out, r.Digest[:])
	return out
}

// SaveSeed writes the current RNG state to the named file, matching the C
// save_seed function.
func (r *RNG) SaveSeed(path string) error {
	return os.WriteFile(path, r.Digest[:], 0644)
}

// Rnd returns a uniformly distributed random integer in the inclusive range
// [low, high]. It reproduces the legacy C rnd() function exactly: the digest
// is re-hashed and its first little-endian 32-bit word is masked and
// rejection-sampled into the requested range.
func (r *RNG) Rnd(low, high int) int {
	rng := uint32(high - low)

	var mask uint32
	for v := rng; v != 0; v >>= 1 {
		mask |= v
	}

	var num uint32
	for {
		sum := md5.Sum(r.Digest[:])
		copy(r.Digest[:], sum[:])
		num = binary.LittleEndian.Uint32(r.Digest[0:4]) & mask
		if num <= rng {
			break
		}
	}

	return int(num) + low
}
