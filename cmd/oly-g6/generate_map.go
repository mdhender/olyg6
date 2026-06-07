// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/mdhender/olyg6/pkg/mapgen"
	"github.com/peterbourgon/ff/v4"
)

// generateMap is a stub for the "generate map" subcommand.
func generateMap(_ context.Context, cfg generateMapConfig) error {
	// TODO: wire this up to pkg/mapgen.

	g := mapgen.New(mapgen.Options{
		InputDir:     cfg.inputPath,
		InputCities:  cfg.cities,
		InputLands:   cfg.lands,
		InputMap:     cfg.asciiMap,
		InputRegions: cfg.regions,
		InputSeed:    cfg.randseed,
		OutputDir:    cfg.outputPath,
		Log:          os.Stderr,
	})

	if err := g.Run(); err != nil {
		return fmt.Errorf("generate map: %v", err)
	}

	return nil
}

// generateMapConfig holds the resolved flags for "generate map".
type generateMapConfig struct {
	inputPath  string
	outputPath string
	asciiMap   string
	cities     string
	lands      string
	randseed   string
	regions    string
}

func mapCmd(parentFlags *ff.FlagSet) *ff.Command {
	var cfg generateMapConfig
	mapFlags := ff.NewFlagSet("map").SetParent(parentFlags)
	mapFlags.StringVar(&cfg.inputPath, 'i', "input-path", ".", "directory containing input files")
	mapFlags.StringVar(&cfg.outputPath, 'o', "output-path", ".", "directory to write output files")
	mapFlags.StringVar(&cfg.asciiMap, 'M', "ascii-map", "ascii-map.txt", "ascii map input file")
	mapFlags.StringVar(&cfg.cities, 'C', "cities", "cities.txt", "city names input file")
	mapFlags.StringVar(&cfg.lands, 'L', "lands", "lands.json", "land-area names input file")
	mapFlags.StringVar(&cfg.regions, 'R', "regions", "regions.json", "region names input file")
	mapFlags.StringVar(&cfg.randseed, 'S', "seed", "randseed", "random seed input file")
	return &ff.Command{
		Name:  "map",
		Usage: "oly-g6 generate map [flags]",
		Flags: mapFlags,
		Exec: func(ctx context.Context, args []string) error {
			return generateMap(ctx, cfg)
		},
	}
}
