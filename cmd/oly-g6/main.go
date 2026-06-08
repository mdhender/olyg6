// Command oly-g6 is the command-line interface for the Olympia G6 engine.
//
// It is organized as a tree of commands. It implements:
//
//	oly-g6 generate map   - generate map data files
//	oly-g6 version         - print the engine version
package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
)

func main() {
	cmd, err := run(context.Background(), os.Args[1:])
	if errors.Is(err, ff.ErrHelp) {
		// Render help for the most specific command the user selected.
		selected := cmd
		if sub := cmd.GetSelected(); sub != nil {
			selected = sub
		}
		_, _ = fmt.Fprintln(os.Stderr, ffhelp.Command(selected))
		os.Exit(0)
	} else if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "oly-g6: %v\n", err)
		os.Exit(1)
	}
}

// run builds the command tree and executes it, returning the selected command
// (for help rendering) along with any error.
func run(ctx context.Context, args []string) (*ff.Command, error) {
	// root: oly-g6
	rootFlags := ff.NewFlagSet("oly-g6")
	rootCmd := &ff.Command{
		Name:  "oly-g6",
		Usage: "oly-g6 <command> [flags]",
		Flags: rootFlags,
		Exec: func(context.Context, []string) error {
			return ff.ErrHelp
		},
	}

	// generate: oly-g6 generate
	generateFlags := ff.NewFlagSet("generate").SetParent(rootFlags)
	generateCmd := &ff.Command{
		Name:  "generate",
		Usage: "oly-g6 generate <command> [flags]",
		Flags: generateFlags,
		Exec: func(context.Context, []string) error {
			return ff.ErrHelp
		},
	}
	rootCmd.Subcommands = append(rootCmd.Subcommands, generateCmd)

	// map: oly-g6 generate map
	generateCmd.Subcommands = append(generateCmd.Subcommands, mapCmd(generateFlags))

	// island: oly-g6 generate island
	generateCmd.Subcommands = append(generateCmd.Subcommands, islandCmd(generateFlags))

	// version: oly-g6 version
	rootCmd.Subcommands = append(rootCmd.Subcommands, versionCmd(rootFlags))

	return rootCmd, rootCmd.ParseAndRun(ctx, args)
}
