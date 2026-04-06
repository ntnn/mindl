// Package simplcli provides a very simplistic implementation of a CLI structure.
package simplcli

import (
	"context"
	"errors"
	"fmt"
)

const (
	// Help is the special "help" subcommand. By default simplcli prints
	// all available subcommands. If a SimplCLI has a custom help
	// subcommand set it will be executed instead.
	Help = "help"
)

// SubCmd is a CLI subcommand implementation.
type SubCmd func(ctx context.Context, args []string) error

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

	return subCmd(ctx, args[1:])
}

// Help prints the available subcommands to stdout.
func (s SimplCLI) Help(ctx context.Context) {
	if h, ok := s.SubCmds[Help]; ok {
		_ = h(ctx, []string{})
		return
	}
	DefaultHelp(s)
}

// DefaultHelp prints all available subcommands to stdout.
func DefaultHelp(s SimplCLI) {
	fmt.Println("Available subcommands:")
	for key := range s.SubCmds {
		fmt.Printf("  %s\n", key)
	}
}
