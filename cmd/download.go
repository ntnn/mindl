package cmd

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/ntnn/mindl/pkg/mindl"
	"github.com/ntnn/mindl/pkg/sum"
)

// Download fetches and extracts an executable from a URL.
//
//nolint:cyclop // sequential steps, not real complexity and will be refactored
func Download(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("", flag.ExitOnError)
	fURL := fs.String("url", "", "URL Template")
	fVersion := fs.String("version", "", "Version to download")
	fExtract := fs.String("extract", "", "File to extract from archive")
	fOut := fs.String("out", "", "Where to place the extracted file")
	if err := fs.Parse(args); err != nil {
		return err
	}

	td := mindl.NewTemplateData()
	td.Version = *fVersion
	templatedURL, err := mindl.Template(*fURL, td)
	if err != nil {
		return err
	}

	templatedExe, err := mindl.Template(*fExtract, td)
	if err != nil {
		return err
	}

	db, err := sum.Open("mindl.sum")
	if err != nil {
		return err
	}

	sumKey := fmt.Sprintf("%s#%s", templatedURL, templatedExe)

	urlEntry, ok := db.Get(sumKey)
	if !ok && !mindl.ShouldUpdate() {
		return fmt.Errorf("mindl should not update, but the required hash is not in the sumdb: %q", sumKey)
	}

	matches, err := sum.PathMatchesHash(*fOut, sum.Fnv128aPath, urlEntry.Sum)
	if err != nil {
		return err
	}
	if matches {
		return nil
	}

	u, err := url.Parse(templatedURL)
	if err != nil {
		return err
	}

	basefilename := filepath.Base(u.Path)

	tmpdir := os.TempDir()
	outfile := filepath.Join(tmpdir, basefilename)

	if err := mindl.Download(ctx, templatedURL, outfile); err != nil {
		return err
	}

	if err := mindl.Unarchive(outfile, templatedExe, *fOut); err != nil {
		return err
	}

	if err := mindl.MakeExecutable(*fOut); err != nil {
		return err
	}

	newHashOnDisk, err := sum.Fnv128aPath(*fOut)
	if err != nil {
		return err
	}

	db.Set(sumKey, newHashOnDisk, "fnv128a")
	return db.Save()
}
