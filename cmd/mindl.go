package cmd

import (
	"context"
	"errors"
	"fmt"
)

var ErrNoArgs = errors.New("no arguments passed")

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
