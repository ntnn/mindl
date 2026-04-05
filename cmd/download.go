package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/ntnn/mindl/pkg/mindl"
	"github.com/ntnn/mindl/pkg/sum"
)

// Download fetches and extracts an executable from a URL.
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

	db, err := sum.Open(sumDBPath)
	if err != nil {
		return err
	}

	urlEntry, ok := db.Get(*fURL, *fInArchive, mindlOS, mindlArch)
	if !ok && !mindl.ShouldUpdate() {
		return errors.New("could not update, required hash not found in mindl.sum")
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

	hash, err := mindl.DownloadAndHash(ctx, tool, mindlOS, mindlArch, *fVersion)
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
