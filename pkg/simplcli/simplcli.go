// Package simplcli provides a very simplistic implementation of a CLI structure.
package simplcli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"
	"strconv"
)

const (
	// Help is the special "help" subcommand. By default simplcli prints
	// all available subcommands. If a SimplCLI has a custom help
	// subcommand set it will be executed instead.
	// The Help function is a [Runner] but only gets a valid stdout.
	// io.Discard is passed as stderr.
	Help = "help"
)

// Runner is the interface expected for functions being executed as a subcommand.
type Runner func(ctx context.Context, stdout, stderr io.Writer, args []string) error

// SubCmd is a CLI subcommand implementation.
type SubCmd struct {
	Runner Runner
	Doc    string
	// Executed before Runner if set. If it returns an error, Runner is skipped.
	PreRun Runner
	// Executed after Runner if set.
	// Executed regardless of whether Runner returns an error.
	// If Runner and PostRun errored the Runner error takes precedence.
	PostRun Runner
}

// SimplCLI contains multiple [SubCmd]s.
type SimplCLI struct {
	SubCmds map[string]SubCmd
}

var (
	// ErrNoArgs is returned when no arguments are passed.
	ErrNoArgs = fmt.Errorf(`no arguments passed, pass %q as the first argument `+
		`to get the list of available subcommands`, Help)
)

// Run runs subcommand indicated by the first argument.
// If the first argument is "help" all registered subcommands are printed to out.
func (s SimplCLI) Run(ctx context.Context, stdout, stderr io.Writer, args []string) error {
	if len(args) == 0 {
		return ErrNoArgs
	}

	cmd := args[0]
	if cmd == Help {
		return s.PrintHelp(ctx, stdout)
	}

	subCmd, ok := s.SubCmds[cmd]
	if !ok {
		return fmt.Errorf("unknown subcommand %q", cmd)
	}

	if subCmd.PreRun != nil {
		if err := subCmd.PreRun(ctx, stdout, stderr, args[1:]); err != nil {
			return fmt.Errorf("error in pre run: %w", err)
		}
	}

	runnerErr := subCmd.Runner(ctx, stdout, stderr, args[1:])

	var postRunErr error
	if subCmd.PostRun != nil {
		postRunErr = subCmd.PostRun(ctx, stdout, stderr, args[1:])
	}

	if runnerErr != nil {
		return runnerErr
	}

	if postRunErr != nil {
		return fmt.Errorf("error in post run: %w", postRunErr)
	}

	return nil
}

// PrintHelp prints the available subcommands to out.
func (s SimplCLI) PrintHelp(ctx context.Context, out io.Writer) error {
	if h, ok := s.SubCmds[Help]; ok {
		return h.Runner(ctx, out, io.Discard, []string{})
	}
	return PrintDefaultHelp(out, s)
}

// PrintDefaultHelp prints all available subcommands to out.
func PrintDefaultHelp(out io.Writer, s SimplCLI) error {
	if len(s.SubCmds) == 0 {
		_, err := fmt.Fprintln(out, "No subcommands available")
		return err
	}
	subCmds := slices.Collect(maps.Keys(s.SubCmds))
	slices.Sort(subCmds)

	longestKey := slices.MaxFunc(subCmds, func(a, b string) int { return len(a) - len(b) })

	// right-align and left-pad all subcommands, e.g.:
	//   template-data   Print the template data with example values
	//        template   Template the given string with example values
	fmtstring := "  %" + strconv.Itoa(len(longestKey)) + "s   %s\n"

	var errs error
	_, err := fmt.Fprintln(out, "Available subcommands:")
	errs = errors.Join(errs, err)
	for _, key := range subCmds {
		_, err := fmt.Fprintf(out, fmtstring, key, s.SubCmds[key].Doc)
		errs = errors.Join(errs, err)
	}
	return errs
}
