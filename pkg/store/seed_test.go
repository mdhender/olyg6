// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package store_test

import (
	"bytes"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/mdhender/olyg6/pkg/prng"
	"github.com/mdhender/olyg6/pkg/store"
)

// TestLoadSeedJSONDecoding locks the randseed.json decode contract, which must
// match the legacy binary randseed: terminate on malformed hex, zero-fill short
// values, and truncate long values to the 16-byte state.
func TestLoadSeedJSONDecoding(t *testing.T) {
	tests := []struct {
		name    string
		hexSeed string
		wantErr bool
		want    []byte
	}{
		{
			name:    "exact 16 bytes",
			hexSeed: "000102030405060708090a0b0c0d0e0f",
			want:    []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		},
		{
			name:    "short zero-fills",
			hexSeed: "0102",
			want:    append([]byte{1, 2}, make([]byte, prng.SeedLen-2)...),
		},
		{
			name:    "long truncates",
			hexSeed: "000102030405060708090a0b0c0d0e0fff", // 17 bytes -> 17th dropped
			want:    []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		},
		{
			name:    "invalid hex char terminates",
			hexSeed: "zz0102030405060708090a0b0c0d0e0f",
			wantErr: true,
		},
		{
			name:    "odd length terminates",
			hexSeed: "abc",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := t.TempDir()
			path := filepath.Join(tmp, "randseed.json")
			if err := os.WriteFile(path, []byte(`{"seed":"`+tt.hexSeed+`"}`), 0644); err != nil {
				t.Fatalf("write: %v", err)
			}

			rng := prng.NewRNG()
			err := store.LoadSeed(rng, path)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got nil", tt.hexSeed)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := rng.Seed(); !bytes.Equal(got, tt.want) {
				t.Errorf("digest = %x, want %x", got, tt.want)
			}
		})
	}
}

// TestLoadSeedBinaryMatchesJSON confirms the binary and JSON paths produce the
// same RNG state for the same 16 bytes.
func TestLoadSeedBinaryMatchesJSON(t *testing.T) {
	raw := []byte{9, 8, 7, 6, 5, 4, 3, 2, 1, 0, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "randseed"), raw, 0644); err != nil {
		t.Fatalf("write binary: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "randseed.json"), []byte(`{"seed":"`+hex.EncodeToString(raw)+`"}`), 0644); err != nil {
		t.Fatalf("write json: %v", err)
	}

	rb := prng.NewRNG()
	if err := store.LoadSeed(rb, filepath.Join(tmp, "randseed")); err != nil {
		t.Fatalf("binary LoadSeed: %v", err)
	}
	rj := prng.NewRNG()
	if err := store.LoadSeed(rj, filepath.Join(tmp, "randseed.json")); err != nil {
		t.Fatalf("json LoadSeed: %v", err)
	}
	if !bytes.Equal(rb.Seed(), rj.Seed()) {
		t.Errorf("binary %x != json %x", rb.Seed(), rj.Seed())
	}
}

// TestSaveSeedRoundTrip writes both seed forms and reads each back unchanged.
func TestSaveSeedRoundTrip(t *testing.T) {
	raw := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	tmp := t.TempDir()
	src := prng.NewRNG()
	src.Load(raw)

	if err := store.SaveSeed(src, tmp); err != nil {
		t.Fatalf("SaveSeed: %v", err)
	}

	for _, name := range []string{"randseed", "randseed.json"} {
		rng := prng.NewRNG()
		if err := store.LoadSeed(rng, filepath.Join(tmp, name)); err != nil {
			t.Fatalf("LoadSeed %s: %v", name, err)
		}
		if !bytes.Equal(rng.Seed(), raw) {
			t.Errorf("%s round-trip = %x, want %x", name, rng.Seed(), raw)
		}
	}
}
