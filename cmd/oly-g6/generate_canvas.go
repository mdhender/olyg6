// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/mdhender/olyg6/pkg/canvasgen"
	"github.com/peterbourgon/ff/v4"
)

// generateCanvasConfig holds the resolved flags for "generate canvas".
type generateCanvasConfig struct {
	inputPath  string
	outputPath string
	asciiMap   string
	seed       string
	size       int
	oceans     int
}

// generateCanvas runs the "generate canvas" subcommand.
func generateCanvas(_ context.Context, cfg generateCanvasConfig) error {
	g := canvasgen.New(canvasgen.Options{
		InputDir:  cfg.inputPath,
		OutputDir: cfg.outputPath,
		AsciiMap:  cfg.asciiMap,
		InputSeed: cfg.seed,
		Size:      cfg.size,
		Oceans:    cfg.oceans,
		Log:       os.Stderr,
	})
	if err := g.Run(); err != nil {
		return fmt.Errorf("generate canvas: %v", err)
	}
	return nil
}

func canvasCmd(parentFlags *ff.FlagSet) *ff.Command {
	var cfg generateCanvasConfig
	fs := ff.NewFlagSet("canvas").SetParent(parentFlags)
	fs.StringVar(&cfg.inputPath, 'i', "input-path", ".", "directory containing the seed file")
	fs.StringVar(&cfg.outputPath, 'o', "output-path", ".", "directory to write output files")
	fs.StringVar(&cfg.asciiMap, 'M', "ascii-map", "ascii-map.txt", "ascii map file to create")
	fs.StringVar(&cfg.seed, 'S', "seed", "randseed", "random seed input file")
	fs.IntVar(&cfg.size, 0, "size", 99, "square map edge length (9-99)")
	fs.IntVar(&cfg.oceans, 0, "oceans", 7, "number of oceans (1-20, < size/2)")
	return &ff.Command{
		Name:  "canvas",
		Usage: "oly-g6 generate canvas [flags]",
		Flags: fs,
		Exec: func(ctx context.Context, args []string) error {
			return generateCanvas(ctx, cfg)
		},
	}
}
