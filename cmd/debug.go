package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/ntnn/mindl/pkg/mindl"
)

func Debug(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return ErrNoArgs
	}
	switch args[0] {
	case "template-data":
		return DebugTemplateData(ctx, os.Stdout)
	default:
		return fmt.Errorf("unknown command: %q", args[0])
	}
}

func DebugTemplateData(ctx context.Context, out io.Writer) error {
	td := mindl.NewTemplateData()

	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(td)
}
