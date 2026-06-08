// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/mdhender/olyg6/pkg/islandgen"
	"github.com/peterbourgon/ff/v4"
)

// generateIslandConfig holds the resolved flags for "generate island".
type generateIslandConfig struct {
	inputPath  string
	outputPath string
	asciiMap   string
	seed       string
	border     int
	shelf      int
	size       int
}

// generateIsland runs the "generate island" subcommand.
func generateIsland(_ context.Context, cfg generateIslandConfig) error {
	g := islandgen.New(islandgen.Options{
		InputDir:  cfg.inputPath,
		OutputDir: cfg.outputPath,
		InputMap:  cfg.asciiMap,
		InputSeed: cfg.seed,
		Border:    cfg.border,
		Shelf:     cfg.shelf,
		Size:      cfg.size,
		Log:       os.Stderr,
	})
	if err := g.Run(); err != nil {
		return fmt.Errorf("generate island: %v", err)
	}
	return nil
}

func islandCmd(parentFlags *ff.FlagSet) *ff.Command {
	var cfg generateIslandConfig
	fs := ff.NewFlagSet("island").SetParent(parentFlags)
	fs.StringVar(&cfg.inputPath, 'i', "input-path", ".", "directory containing input files")
	fs.StringVar(&cfg.outputPath, 'o', "output-path", ".", "directory to write output files")
	fs.StringVar(&cfg.asciiMap, 'M', "ascii-map", "ascii-map.txt", "ascii map input/output file")
	fs.StringVar(&cfg.seed, 'S', "seed", "randseed", "random seed input file")
	fs.IntVar(&cfg.border, 'b', "border", 2, "edge border kept clear of new land")
	fs.IntVar(&cfg.shelf, 'c', "shelf", 3, "continental shelf kept around existing land")
	fs.IntVar(&cfg.size, 0, "size", 0, "target island size (0 = random)")
	return &ff.Command{
		Name:  "island",
		Usage: "oly-g6 generate island [flags]",
		Flags: fs,
		Exec: func(ctx context.Context, args []string) error {
			return generateIsland(ctx, cfg)
		},
	}
}
