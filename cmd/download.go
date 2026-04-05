package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/ntnn/mindl/pkg/mindl"
	"github.com/ntnn/mindl/pkg/sum"
)

var (
	hasher      = sum.Sha512Path
	hasherDescr = "sha512"
)

// Download fetches and extracts an executable from a URL.
//
// When MINDL_UPDATE is set the hashes for all known OS/Arch
// combinations for this tool are updated in the sumdb.
//
// If -common is passed hashes for common OS/Arch combinations are added
// as well if they are not set yet.
func Download(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("", flag.ExitOnError)
	fURL := fs.String("url", "", "URL Template")
	fInArchive := fs.String("inarchive", "", "File to extract from archive")
	fVersion := fs.String("version", "", "Version to download")
	fOut := fs.String("out", "", "Where to place the extracted file")
	fCommon := fs.Bool("common", false, "Also update hashes for common OS/Arch combinations")
	if err := fs.Parse(args); err != nil {
		return err
	}

	db, err := sum.Open(sumDBPath)
	if err != nil {
		return err
	}

	tool := mindl.Tool{
		URLTemplate: *fURL,
		InArchive:   *fInArchive,
	}
	current := mindl.CurrentTarget()

	if err := downloadCurrent(ctx, db, tool, current, *fVersion, *fOut); err != nil {
		return err
	}

	if mindl.ShouldUpdate() {
		targets := targetsFromEntries(db.GetAll(tool.URLTemplate, tool.InArchive))
		if *fCommon {
			targets = append(targets, mindl.CommonTargets...)
		}
		if err := updateTargets(ctx, db, tool, *fVersion, mindl.DeduplicateTargets(targets), current); err != nil {
			return err
		}
	}

	return db.Save()
}

func targetsFromEntries(entries []sum.Entry) []mindl.Target {
	targets := make([]mindl.Target, len(entries))
	for i, e := range entries {
		targets[i] = mindl.Target{OS: e.OS, Arch: e.Arch}
	}
	return targets
}

// downloadCurrent downloads the tool for the current target if necessary.
func downloadCurrent(
	ctx context.Context, db *sum.DB, tool mindl.Tool,
	current mindl.Target, version, out string,
) error {
	entry, ok := db.Get(tool.URLTemplate, tool.InArchive, current.OS, current.Arch)
	if !ok {
		if !mindl.ShouldUpdate() {
			return errors.New("required hash not found in mindl.sum")
		}
	}

	if entry.Sum != "" {
		matches, err := sum.PathMatchesHash(out, hasher, entry.Sum)
		if err != nil {
			return err
		}
		if matches {
			return nil
		}
	}

	hash, err := downloadAndHash(ctx, tool, current, version, out, hasher)
	if err != nil {
		return err
	}

	// TODO not everything is executable
	if err := mindl.MakeExecutable(out); err != nil {
		return fmt.Errorf("error marking %q as executable: %w", out, err)
	}

	db.Set(tool.URLTemplate, tool.InArchive, current.OS, current.Arch, hash, hasherDescr)
	return nil
}

// updateTargets downloads and hashes the tool for all targets except skip.
func updateTargets(
	ctx context.Context, db *sum.DB, tool mindl.Tool,
	version string, targets []mindl.Target, skip mindl.Target,
) error {
	for _, t := range targets {
		if t == skip {
			continue
		}
		h, err := downloadAndHash(ctx, tool, t, version, "", hasher)
		if err != nil {
			return err
		}
		db.Set(tool.URLTemplate, tool.InArchive, t.OS, t.Arch, h, hasherDescr)
	}
	return nil
}

func downloadAndHash(
	ctx context.Context,
	tool mindl.Tool,
	t mindl.Target,
	version, extractTo string,
	hasher sum.HashPathFunc,
) (string, error) {
	td := mindl.NewTemplateData(t.OS, t.Arch)
	td.Version = version

	th, err := mindl.Handle(tool, td)
	if err != nil {
		return "", fmt.Errorf("error creating tool handler for %q: %w", tool, err)
	}
	defer th.Cleanup()

	if err := th.Download(ctx, extractTo); err != nil {
		return "", fmt.Errorf("error downloading tool %q: %w", tool, err)
	}

	return th.Hash(hasher)
}
