package mindl

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/ntnn/mindl/pkg/sum"
)

// Tool is the definition of a tool to download.
type Tool struct {
	URLTemplate string
	InArchive   string
	ExtractTo   string
}

// ToolHandler is the handler for a tool download.
type ToolHandler struct {
	tool      Tool
	url       string
	inArchive string
	tmpdir    string
	extractTo string
}

// Handle sets up a [ToolHandler].
func Handle(tool Tool, td *TemplateData) (*ToolHandler, error) {
	th := &ToolHandler{tool: tool}

	var err error
	th.url, err = Template(th.tool.URLTemplate, td)
	if err != nil {
		return nil, fmt.Errorf("error templating url template %q: %w", th.tool.URLTemplate, err)
	}

	th.inArchive, err = Template(th.tool.InArchive, td)
	if err != nil {
		return nil, fmt.Errorf("error templating extract path %q: %w", th.tool.InArchive, err)
	}

	return th, nil
}

// Download downloads and extracts a tool.
func (th *ToolHandler) Download(ctx context.Context) error {
	th.tmpdir = os.TempDir()
	th.extractTo = th.tool.ExtractTo
	if th.extractTo == "" {
		// If .ExractTo is empty the hash for a different OS/Arch is
		// being updated so extract to inside of the temporary
		// directory.
		th.extractTo = filepath.Join(th.tmpdir, "extracted")
	}

	u, err := url.Parse(th.url)
	if err != nil {
		return fmt.Errorf("error parsing URL %q: %w", th.url, err)
	}

	basefilename := filepath.Base(u.Path)
	outfile := filepath.Join(th.tmpdir, basefilename)
	if err := Download(ctx, th.url, outfile); err != nil {
		return fmt.Errorf("error downloading %q to %q: %w", th.url, outfile, err)
	}

	if err := Unarchive(outfile, th.inArchive, th.extractTo); err != nil {
		return fmt.Errorf("error extracting %q from %q to %q: %w", th.inArchive, outfile, th.extractTo, err)
	}

	return nil
}

// Hash runs the given hash func on the tool and returns the result.
func (th *ToolHandler) Hash(hasher sum.HashPathFunc) (string, error) {
	return hasher(th.extractTo)
}

// Cleanup deletes the temporary files leftover by the download and extraction.
func (th *ToolHandler) Cleanup() {
	if th.tmpdir != "" {
		_ = os.RemoveAll(th.tmpdir)
	}
}

// DownloadAndHash downloads the tool for the given target/version combination.
func DownloadAndHash(ctx context.Context, tool Tool, t Target, version string) (string, error) {
	td := NewTemplateData(t.OS, t.Arch)
	td.Version = version

	th, err := Handle(tool, td)
	if err != nil {
		return "", err
	}
	defer th.Cleanup()

	if err := th.Download(ctx); err != nil {
		return "", err
	}

	return th.Hash(sum.Fnv128aPath)
}
