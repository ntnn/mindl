package cmd

import (
	"context"
	"flag"
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
	fExecutable := fs.String("executable", "", "Executable to extract from archive")
	fOut := fs.String("out", "", "Where to extract the executable to")
	if err := fs.Parse(args); err != nil {
		return err
	}

	td := mindl.NewTemplateData()
	td.Version = *fVersion
	result, err := mindl.Template(*fURL, td)
	if err != nil {
		return err
	}

	db, err := sum.Open("mindl.sum")
	if err != nil {
		return err
	}

	_, err = os.Stat(*fOut)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if !os.IsNotExist(err) {
		hashOnDisk, err := sum.Fnv128aPath(*fOut)
		if err != nil && !os.IsNotExist(err) {
			return err
		}

		urlEntry, _ := db.Get(sumKey)
		if hashOnDisk == urlEntry.Sum {
			return nil
		}
	}

	u, err := url.Parse(result)
	if err != nil {
		return err
	}

	basefilename := filepath.Base(u.Path)

	tmpdir := os.TempDir()
	outfile := filepath.Join(tmpdir, basefilename)

	if err := mindl.Download(ctx, result, outfile); err != nil {
		return err
	}

	if err := mindl.Unarchive(outfile, *fExecutable, *fOut); err != nil {
		return err
	}

	if err := mindl.MakeExecutable(*fOut); err != nil {
		return err
	}

	newHashOnDisk, err := sum.Fnv128aPath(*fOut)
	if err != nil {
		return err
	}

	if err := db.Set(result, newHashOnDisk, "fnv128a"); err != nil {
		return err
	}

	return nil
}
