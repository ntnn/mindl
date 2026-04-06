// Package simplcli provides a very simplistic implementation of a CLI structure.
package simplcli

import (
	"context"
	"errors"
	"fmt"
	"strconv"
)

const (
	// Help is the special "help" subcommand. By default simplcli prints
	// all available subcommands. If a SimplCLI has a custom help
	// subcommand set it will be executed instead.
	Help = "help"
)

// Runner is the interface expected for functions being executed as a subcommand.
type Runner func(ctx context.Context, args []string) error

// SubCmd is a CLI subcommand implementation.
type SubCmd struct {
	Runner Runner
	Doc    string
}

// SimplCLI contains multiple [SubCmd]s.
type SimplCLI struct {
	SubCmds map[string]SubCmd
}

var (
	// ErrNoArgs is returned when no arguments are passed.
	ErrNoArgs = errors.New("no arguments passed")
)

// Run runs subcommand indicated by the first argument.
// If the first argument is "help" all registered subcommands are printed to stdout.
func (s SimplCLI) Run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		s.Help(ctx)
		return ErrNoArgs
	}

	cmd := args[0]
	if cmd == Help {
		s.Help(ctx)
		return nil
	}

	subCmd, ok := s.SubCmds[cmd]
	if !ok {
		return fmt.Errorf("unknown subcommand %q", cmd)
	}

	return subCmd.Runner(ctx, args[1:])
}

// Help prints the available subcommands to stdout.
func (s SimplCLI) Help(ctx context.Context) {
	if h, ok := s.SubCmds[Help]; ok {
		_ = h.Runner(ctx, []string{})
		return
	}
	DefaultHelp(s)
}

// DefaultHelp prints all available subcommands to stdout.
func DefaultHelp(s SimplCLI) {
	leftLength := 0
	out := map[string]string{}
	for key, subCmd := range s.SubCmds {
		out[key] = subCmd.Doc
		if len(key) > leftLength {
			leftLength = len(key)
		}
	}

	longestKey := slices.MaxFunc(maps.Keys(out), func(a, b string) int { return len(a) - len(b) })

	// right-align and left-pad all subcommands, e.g.:
	//   template-data   Print the template data with example values
	//        template   Template the given string with example values
	fmtstring := "  %" + strconv.Itoa(len(longestKey)) + "s   %s\n"

	fmt.Println("Available subcommands:")
	for key, subCmd := range s.SubCmds {
		fmt.Printf(fmtstring, key, subCmd.Doc)
	}
}
