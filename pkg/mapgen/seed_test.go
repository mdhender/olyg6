// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package mapgen

import (
	"bytes"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/mdhender/olyg6/pkg/prng"
)

// TestLoadSeedJSONDecoding locks the randseed.json decode contract, which must
// match the legacy binary randseed: terminate on malformed hex, zero-fill short
// values, and truncate long values to the 16-byte state.
func TestLoadSeedJSONDecoding(t *testing.T) {
	tests := []struct {
		name    string
		hexSeed string
		wantErr bool
		want    []byte // expected 16-byte digest when no error
	}{
		{
			name:    "exact 16 bytes",
			hexSeed: "000102030405060708090a0b0c0d0e0f",
			want:    []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		},
		{
			name:    "short zero-fills",
			hexSeed: "0102", // 2 bytes -> rest zero
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
			body := []byte(`{"seed":"` + tt.hexSeed + `"}`)
			if err := os.WriteFile(filepath.Join(tmp, "randseed.json"), body, 0644); err != nil {
				t.Fatalf("write: %v", err)
			}

			g := New(Options{InputDir: tmp, InputSeed: "randseed.json", Log: io.Discard})
			err := g.loadSeed()

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got nil", tt.hexSeed)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := g.RNG.Seed(); !bytes.Equal(got, tt.want) {
				t.Errorf("digest = %x, want %x", got, tt.want)
			}
		})
	}
}

// TestLoadSeedBinaryMatchesJSON confirms the binary and JSON paths produce the
// same RNG state for the same 16 bytes (they share prng.Load).
func TestLoadSeedBinaryMatchesJSON(t *testing.T) {
	raw := []byte{9, 8, 7, 6, 5, 4, 3, 2, 1, 0, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "randseed"), raw, 0644); err != nil {
		t.Fatalf("write binary: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "randseed.json"), []byte(`{"seed":"`+hex.EncodeToString(raw)+`"}`), 0644); err != nil {
		t.Fatalf("write json: %v", err)
	}

	gb := New(Options{InputDir: tmp, InputSeed: "randseed", Log: io.Discard})
	if err := gb.loadSeed(); err != nil {
		t.Fatalf("binary loadSeed: %v", err)
	}
	gj := New(Options{InputDir: tmp, InputSeed: "randseed.json", Log: io.Discard})
	if err := gj.loadSeed(); err != nil {
		t.Fatalf("json loadSeed: %v", err)
	}
	if !bytes.Equal(gb.RNG.Seed(), gj.RNG.Seed()) {
		t.Errorf("binary %x != json %x", gb.RNG.Seed(), gj.RNG.Seed())
	}
}
