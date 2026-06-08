// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package store

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mdhender/olyg6/pkg/prng"
)

// LoadSeed loads the RNG state from path, choosing the decoder by file
// extension: a ".json" file holds the hex-string RandSeed form; anything else
// is the legacy raw 16-byte binary file (kept for backwards compatibility).
//
// Decoding matches the binary path (prng.Load semantics): malformed hex
// terminates with an error, short values zero-fill, long values truncate to the
// 16-byte state.
func LoadSeed(rng *prng.RNG, path string) error {
	if filepath.Ext(path) != ".json" {
		return rng.LoadSeed(path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var rs RandSeed
	if err := json.Unmarshal(data, &rs); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	b, err := hex.DecodeString(rs.Seed)
	if err != nil {
		return fmt.Errorf("%s: bad seed hex: %w", path, err)
	}
	rng.Load(b)
	return nil
}

// SaveSeed writes the RNG state to dir in both forms: the legacy binary
// "randseed" and the G6 native "randseed.json" (lowercase hex).
func SaveSeed(rng *prng.RNG, dir string) error {
	if err := rng.SaveSeed(filepath.Join(dir, "randseed")); err != nil {
		return err
	}
	data, err := json.MarshalIndent(RandSeed{Seed: hex.EncodeToString(rng.Seed())}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "randseed.json"), append(data, '\n'), 0644)
}
