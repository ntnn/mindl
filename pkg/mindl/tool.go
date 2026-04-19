package mindl

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"

	"github.com/ntnn/mindl/pkg/sum"
)

// Tool is the definition of a tool to download.
type Tool struct {
	// URLTemplate is templated with [TemplateData].
	// See [CommonTools] for examples.
	URLTemplate string
	// InArchive is the location within the archive.
	// If the downloaded file is the file and not an archive leave
	// InArchive empty.
	InArchive string
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
// If extractTo is empty the file is extracted to a temporary directory.
func (th *ToolHandler) Download(ctx context.Context) error {
	tmpdir, err := os.MkdirTemp(os.TempDir(), "mindl-")
	if err != nil {
		return fmt.Errorf("error creating temporary directory: %w", err)
	}
	th.tmpdir = tmpdir
	th.extractTo = filepath.Join(th.tmpdir, "extracted")

	u, err := url.Parse(th.url)
	if err != nil {
		return fmt.Errorf("error parsing URL %q: %w", th.url, err)
	}

	basefilename := filepath.Base(u.Path)
	outfile := filepath.Join(th.tmpdir, basefilename)
	if err := download(ctx, th.url, outfile); err != nil {
		return fmt.Errorf("error downloading %q to %q: %w", th.url, outfile, err)
	}

	if th.inArchive == "" {
		return iocopy(outfile, th.extractTo)
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

// Move moves the extracted file to dst.
// It attempts [os.Rename] first, falling back to a copy for
// cross-filesystem moves.
func (th *ToolHandler) Move(dst string) error {
	if err := os.Rename(th.extractTo, dst); err == nil {
		return nil
	}

	in, err := os.Open(th.extractTo)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return out.Close()
}

// Cleanup deletes the temporary files leftover by the download and extraction.
func (th *ToolHandler) Cleanup() {
	if th.tmpdir != "" {
		_ = os.RemoveAll(th.tmpdir)
	}
}
