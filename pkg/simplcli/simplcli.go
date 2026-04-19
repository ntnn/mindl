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

// SubCmd is a CLI subcommand.
type SubCmd struct {
	// Runner is the function to be executed when this sub command is run.
	Runner Runner
	// Doc is the doc string used when printing the help output.
	Doc string
}

// Middleware is a function that wraps [Runner].
type Middleware func(ctx context.Context, stdout, stderr io.Writer, args []string, next Runner) error

// SimplCLI contains multiple [SubCmd]'s.
type SimplCLI struct {
	// The [SubCmd]'s to execute.
	SubCmds map[string]SubCmd
	// A list of middlewares to be applied to all subcommands.
	Middlewares []Middleware
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

	// Build the execution chain by wrapping with each middleware from bottom to top.
	// CLI:
	//   Runner: R
	//   MWs:
	//     - A
	//     - B
	//     - C
	// -> C runs R
	// -> B runs C
	// -> A runs B
	// -> A is executed
	runner := subCmd.Runner
	middlewares := slices.Clone(s.Middlewares)
	slices.Reverse(middlewares)
	for _, mw := range middlewares {
		next := runner
		runner = func(ctx context.Context, stdout, stderr io.Writer, args []string) error {
			return mw(ctx, stdout, stderr, args, next)
		}
	}

	return runner(ctx, stdout, stderr, args[1:])
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
