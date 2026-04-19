package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/ntnn/mindl/pkg/mindl"
	"github.com/ntnn/mindl/pkg/simplcli"
)

// Common dispatches common subcommands.
var Common = simplcli.SimplCLI{
	SubCmds: map[string]simplcli.SubCmd{
		"list":   {CommonList, "List available common tools"},
		"detail": {CommonDetail, "Output details of common tools"},
	},
}

// CommonList writes the list of known common tools to stdout.
func CommonList(_ context.Context, stdout, _ io.Writer, _ []string) error {
	if _, err := fmt.Fprintln(stdout, "Available common tools:"); err != nil {
		return err
	}
	for name := range mindl.CommonTools {
		if _, err := fmt.Fprintf(stdout, "- %s\n", name); err != nil {
			return err
		}
	}
	return nil
}

// CommonDetail writes the details of the given tools to stdout.
func CommonDetail(_ context.Context, stdout, stderr io.Writer, args []string) error {
	if len(args) == 0 {
		return errors.New("no tool passed")
	}

	erroredTools := []string{}
	for _, name := range args {
		tool, ok := mindl.CommonTools[name]
		if !ok {
			if _, err := fmt.Fprintf(stderr, "tool %q no found\n", name); err != nil {
				return err
			}
			erroredTools = append(erroredTools, name)
			continue
		}
		if _, err := fmt.Fprintf(stdout, "%s:\n", name); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(stdout, "  URL: %q\n", tool.URL); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(stdout, "  InArchive: %q\n", tool.InArchive); err != nil {
			return err
		}
	}

	if len(erroredTools) > 0 {
		return fmt.Errorf("some tools were not found: %s", strings.Join(erroredTools, ", "))
	}

	return nil
}
