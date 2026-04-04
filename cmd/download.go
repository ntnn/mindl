package cmd

import (
	"context"
	"flag"
	"fmt"

	"github.com/ntnn/mindl/pkg/mindl"
	"github.com/ntnn/mindl/pkg/sum"
)

// Download fetches and extracts an executable from a URL.
//
//nolint:cyclop // sequential steps, not real complexity and will be refactored
func Download(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("", flag.ExitOnError)
	fURL := fs.String("url", "", "URL Template")
	fInArchive := fs.String("inarchive", "", "File to extract from archive")
	fVersion := fs.String("version", "", "Version to download")
	fOut := fs.String("out", "", "Where to place the extracted file")
	if err := fs.Parse(args); err != nil {
		return err
	}

	mindlOS := mindl.OS()
	mindlArch := mindl.Arch()

	td := mindl.NewTemplateData(mindlOS, mindlArch)
	td.Version = *fVersion

	db, err := sum.Open("mindl.sum")
	if err != nil {
		return err
	}

	urlEntry, ok := db.Get(*fURL, *fInArchive, mindlOS, mindlArch)
	if !ok && !mindl.ShouldUpdate() {
		return fmt.Errorf("could not update, required hash not found in mindl.sum")
	}

	matches, err := sum.PathMatchesHash(*fOut, sum.Fnv128aPath, urlEntry.Sum)
	if err != nil {
		return err
	}
	if matches {
		return nil
	}

	tool := mindl.Tool{
		URLTemplate: *fURL,
		InArchive:   *fInArchive,
		ExtractTo:   *fOut,
	}

	th, err := mindl.Handle(tool, td)
	if err != nil {
		return err
	}

	if err := th.Download(ctx); err != nil {
		return err
	}

	hash, err := th.Hash(sum.Fnv128aPath)
	if err != nil {
		return err
	}

	// TODO not everything is executable
	if err := mindl.MakeExecutable(*fOut); err != nil {
		return fmt.Errorf("error marking %q as executable: %w", *fOut, err)
	}

	db.Set(*fURL, *fInArchive, mindlOS, mindlArch, hash, "fnv128a")
	return db.Save()
}
