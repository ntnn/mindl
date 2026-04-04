package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ntnn/mindl/pkg/mindl"
)

// Debug dispatches debug subcommands.
func Debug(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return ErrNoArgs
	}
	switch args[0] {
	case "template-data":
		return DebugTemplateData(ctx, os.Stdout)
	case "template":
		return DebugTemplate(ctx, os.Stdout, strings.Join(args[1:], " "))
	default:
		return fmt.Errorf("unknown command: %q", args[0])
	}
}

// DebugTemplateData writes template data as JSON to out.
func DebugTemplateData(_ context.Context, out io.Writer) error {
	td := mindl.NewTemplateData("", "")

	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(td)
}

// DebugTemplate renders text as a template and writes the result to out.
func DebugTemplate(_ context.Context, out io.Writer, text string) error {
	td := mindl.NewTemplateData("", "")
	td.Version = "<VERSION>"
	result, err := mindl.Template(text, td)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, result)
	return err
}
