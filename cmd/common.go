package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/ntnn/mindl/pkg/mindl"
	"github.com/ntnn/mindl/pkg/simplcli"
)

// Common dispatches common subcommands.
var Common = simplcli.SimplCLI{
	SubCmds: map[string]simplcli.SubCmd{
		"list":      {CommonList, "List available common tools"},
		"detail":    {CommonDetail, "Output details of common tools"},
		"bootstrap": {CommonBootstrap, "Bootstrap a common tool"},
	},
}

// CommonList writes the list of known common tools to stdout.
func CommonList(_ context.Context, stdout, _ io.Writer, _ []string) error {
	if _, err := fmt.Fprintln(stdout, "Available common tools:"); err != nil {
		return err
	}
	for name := range mindl.CommonTools {
		if _, err := fmt.Fprintf(stdout, "- %s\n", name); err != nil {
			return err
		}
	}
	return nil
}

// CommonDetail writes the details of the given tools to stdout.
func CommonDetail(_ context.Context, stdout, stderr io.Writer, args []string) error {
	if len(args) == 0 {
		return errors.New("no tool passed")
	}

	erroredTools := []string{}
	for _, name := range args {
		tool, ok := mindl.CommonTools[name]
		if !ok {
			if _, err := fmt.Fprintf(stderr, "tool %q no found\n", name); err != nil {
				return err
			}
			erroredTools = append(erroredTools, name)
			continue
		}
		if _, err := fmt.Fprintf(stdout, "%s:\n", name); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(stdout, "  URL: %q\n", tool.URL); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(stdout, "  InArchive: %q\n", tool.InArchive); err != nil {
			return err
		}
	}

	if len(erroredTools) > 0 {
		return fmt.Errorf("some tools were not found: %s", strings.Join(erroredTools, ", "))
	}

	return nil
}

var toolTemplate = `
{{.Var}}_VER := 0.0.0
{{.Var}} := {{.ToolsDir}}/{{.Tool}}-$({{.Var}}_VER)

$({{.Var}}):
	mkdir -p {{.ToolsDir}}
	{{.Mindl}} download -common -out $@ -tool {{.Tool}} -version $({{.Var}}_VER)
`

var defaultPerm os.FileMode = 0600

// CommonBootstrap writes the boilerplate to ensure a common tool to Makefile.
func CommonBootstrap(_ context.Context, _, stderr io.Writer, args []string) error {
	fs := flag.NewFlagSet("", flag.ExitOnError)
	fMakefile := fs.String("makefile", "Makefile", "Makefile to append to")
	fToolsdir := fs.String("toolsdir", "$(TOOLS_DIR)", "The tools dir to reference")
	fMindl := fs.String("mindl", "$(GO) tool github.com/ntnn/mindl", "How to run mindl")
	if err := fs.Parse(args); err != nil {
		return err
	}

	t, err := template.New("").Parse(toolTemplate)
	if err != nil {
		return fmt.Errorf("error parsing template: %w", err)
	}

	f, err := os.OpenFile(*fMakefile, os.O_RDWR|os.O_CREATE|os.O_APPEND, defaultPerm)
	if err != nil {
		return fmt.Errorf("error opening %q to append: %w", *fMakefile, err)
	}
	defer f.Close()

	for _, tool := range args {
		varname := strings.ReplaceAll(strings.ToUpper(tool), "-", "_")

		data := map[string]string{
			// golangci-lint => GOLANGCI_LINT
			"Var": varname,
			// golangci-lint
			"Tool": tool,
			// $(TOOLS_DIR)
			"ToolsDir": *fToolsdir,
			// $(GO) tool run github.com/ntnn/mindl
			"Mindl": *fMindl,
		}

		if err := t.Execute(f, data); err != nil {
			_, _ = fmt.Fprintf(stderr, "error templating %q: %v", tool, err)
		}
	}

	return nil
}
