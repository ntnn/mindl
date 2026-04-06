package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ntnn/mindl/pkg/mindl"
	"github.com/ntnn/mindl/pkg/simplcli"
)

// Debug dispatches debug subcommands.
var Debug = simplcli.SimplCLI{
	SubCmds: map[string]simplcli.SubCmd{
		"template-data": {
			func(ctx context.Context, _ []string) error {
				return DebugTemplateData(ctx, os.Stdout)
			},
			"Print the template data with example values",
		},
		"template": {
			func(ctx context.Context, args []string) error {
				return DebugTemplate(ctx, os.Stdout, strings.Join(args, " "))
			},
			"Template the given string with example values",
		},
	},
}

// DebugTemplateData writes template data as JSON to out.
func DebugTemplateData(_ context.Context, out io.Writer) error {
	td := mindl.NewTemplateData("<OS>", "<ARCH>")

	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(td)
}

// DebugTemplate renders text as a template and writes the result to out.
func DebugTemplate(_ context.Context, out io.Writer, text string) error {
	td := mindl.NewTemplateData("<OS>", "<ARCH>")
	td.Version = "<VERSION>"
	result, err := mindl.Template(text, td)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(out, result)
	return err
}
