package cmd

import (
	"context"
	"flag"

	"github.com/ntnn/mindl/pkg/mindl"
	"github.com/ntnn/mindl/pkg/sum"
)

// Update updates hashes in the sumdb.
func Update(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("", flag.ExitOnError)
	fURL := fs.String("url", "", "URL Template")
	fInArchive := fs.String("inarchive", "", "File to extract from archive")
	fVersion := fs.String("version", "", "Version to download")
	fCommon := fs.Bool("common", false, "Also update hashes for common OS/Arch combinations")
	if err := fs.Parse(args); err != nil {
		return err
	}

	db, err := sum.Open(sumDBPath)
	if err != nil {
		return err
	}

	entries := db.GetAll(*fURL, *fInArchive)
	if *fCommon {
		// TODO deduplicate
		for _, cc := range commonCombs {
			entries = append(entries, sum.Entry{
				URLTemplate: *fURL,
				InArchive:   *fInArchive,
				OS:          cc[0],
				Arch:        cc[1],
			})
		}
	}
	for _, entry := range entries {
		if err := update(ctx, db, entry, *fVersion); err != nil {
			return err
		}
	}

	return db.Save()
}

var commonCombs = [][]string{
	{"linux", "amd64"},
	{"linux", "arm64"},
	{"darwin", "amd64"},
	{"darwin", "arm64"},
	{"windows", "amd64"},
}

func update(ctx context.Context, db *sum.DB, entry sum.Entry, version string) error {
	tool := mindl.Tool{
		URLTemplate: entry.URLTemplate,
		InArchive:   entry.InArchive,
	}

	hash, err := mindl.DownloadAndHash(ctx, tool, entry.OS, entry.Arch, version)
	if err != nil {
		return err
	}

	db.Set(entry.URLTemplate, entry.InArchive, entry.OS, entry.Arch, hash, "fnv128a")
	return nil
}
