// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package main

import (
	"context"
	"fmt"

	"github.com/mdhender/olyg6"
	"github.com/peterbourgon/ff/v4"
)

// versionCmd creates the "version" subcommand, which prints the engine's
// semantic version (major.minor.patch).
func versionCmd(parentFlags *ff.FlagSet) *ff.Command {
	versionFlags := ff.NewFlagSet("version").SetParent(parentFlags)
	return &ff.Command{
		Name:  "version",
		Usage: "oly-g6 version",
		Flags: versionFlags,
		Exec: func(context.Context, []string) error {
			fmt.Println(olyg6.Version().Core())
			return nil
		},
	}
}
