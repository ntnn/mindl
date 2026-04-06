package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/ntnn/mindl/pkg/mindl"
	"github.com/ntnn/mindl/pkg/simplcli"
)

// Debug dispatches debug subcommands.
var Debug = simplcli.SimplCLI{
	SubCmds: map[string]simplcli.SubCmd{
		"template-data": {DebugTemplateData, "Print the template data with example values"},
		"template":      {DebugTemplate, "Template the given string with example values"},
	},
}

// DebugTemplateData writes template data as JSON to out.
func DebugTemplateData(_ context.Context, stdout, _ io.Writer, _ []string) error {
	td := mindl.NewTemplateData("<OS>", "<ARCH>")

	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(td)
}

// DebugTemplate renders text as a template and writes the result to out.
func DebugTemplate(_ context.Context, stdout, _ io.Writer, args []string) error {
	td := mindl.NewTemplateData("<OS>", "<ARCH>")
	td.Version = "<VERSION>"
	result, err := mindl.Template(strings.Join(args, " "), td)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(stdout, result)
	return err
}
