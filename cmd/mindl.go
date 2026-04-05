package cmd

import (
	"context"
	"errors"
	"fmt"
)

// ErrNoArgs is returned when no arguments are passed.
var ErrNoArgs = errors.New("no arguments passed")

const sumDBPath = "mindl.sum"

// Main dispatches CLI commands.
func Main(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return ErrNoArgs
	}
	switch args[0] {
	case "debug":
		return Debug(ctx, args[1:])
	case "download":
		return Download(ctx, args[1:])
	default:
		return fmt.Errorf("unknown command: %q", args[0])
	}
}
