package mindl

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/ntnn/mindl/pkg/sum"
)

type Tool struct {
	URLTemplate string
	InArchive   string
	ExtractTo   string
}

type ToolHandler struct {
	tool      Tool
	url       string
	inArchive string
	tmpdir    string
	extractTo string
}

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

func (th *ToolHandler) Hash(hasher sum.HashPathFunc) (string, error) {
	return hasher(th.extractTo)
}

func (th *ToolHandler) Cleanup() error {
	if th.tmpdir == "" {
		return nil
	}
	return os.RemoveAll(th.tmpdir)
}
