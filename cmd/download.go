package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"

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
//
//nolint:cyclop // Just flag handling, not that complex.
func Download(ctx context.Context, _ io.Writer, args []string) error {
	fs := flag.NewFlagSet("", flag.ExitOnError)
	fURL := fs.String("url", "", "URL for the archive to download, templated")
	fInArchive := fs.String("inarchive", "", "Path of the file to extract from the archive, templated")
	fVersion := fs.String("version", "", "Version string substituted into templates")
	fOut := fs.String("out", "", "Destination path for the extracted binary")
	fCommon := fs.Bool("common", false, "Also update hashes for common OS/Arch combinations")
	fTool := fs.String("tool", "", "Default -url and -inarchive to the values of this common tool")
	if err := fs.Parse(args); err != nil {
		return err
	}

	db, err := sum.Open(sumDBPath)
	if err != nil {
		return err
	}

	tool := mindl.Tool{}
	if *fTool != "" {
		ct, ok := mindl.CommonTools[*fTool]
		if !ok {
			return fmt.Errorf("unknown tool: %q", *fTool)
		}
		tool.URLTemplate = ct.URL
		tool.InArchive = ct.InArchive
	}
	if *fURL != "" {
		tool.URLTemplate = *fURL
	}
	if *fInArchive != "" {
		tool.InArchive = *fInArchive
	}
	current := mindl.CurrentTarget()

	if mindl.ShouldUpdate() {
		targets := targetsFromEntries(db.GetAll(tool.URLTemplate, tool.InArchive))
		if *fCommon {
			targets = append(targets, mindl.CommonTargets...)
		}
		if err := updateTargets(ctx, db, tool, *fVersion, mindl.DeduplicateTargets(targets)); err != nil {
			return err
		}
	}

	// TODO download re-download the same file
	if err := downloadCurrent(ctx, db, tool, current, *fVersion, *fOut); err != nil {
		return err
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
	switch ok {
	case false:
		if !mindl.ShouldUpdate() {
			return errors.New("required hash not found in mindl.sum")
		}
	case true:
		matches, err := sum.PathMatchesHash(out, hasher, entry.Sum)
		if err != nil {
			return err
		}
		if matches {
			return nil
		}
	}

	hash, err := downloadAndHash(ctx, tool, current, version, out, entry.Sum, hasher)
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

// updateTargets downloads and hashes the tool for all targets.
func updateTargets(
	ctx context.Context, db *sum.DB, tool mindl.Tool,
	version string, targets []mindl.Target,
) error {
	for _, t := range targets {
		h, err := downloadAndHash(ctx, tool, t, version, "", "", hasher)
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
	expectedHash string, hasher sum.HashPathFunc,
) (string, error) {
	td := mindl.NewTemplateData(t.OS, t.Arch)
	td.Version = version

	th, err := mindl.Handle(tool, td)
	if err != nil {
		return "", fmt.Errorf("error creating tool handler for %q: %w", tool, err)
	}
	defer th.Cleanup()

	if err := th.Download(ctx); err != nil {
		return "", fmt.Errorf("error downloading tool %q: %w", tool, err)
	}

	hash, err := th.Hash(hasher)
	if err != nil {
		return "", fmt.Errorf("error hashing extracted file %q: %w", tool.InArchive, err)
	}

	if expectedHash != "" && hash != expectedHash {
		return "", fmt.Errorf("hash %q does not match expected hash %q", hash, expectedHash)
	}

	if extractTo != "" {
		if err := th.Move(extractTo); err != nil {
			return "", fmt.Errorf("error moving extracted archive file %q to %q: %w", tool.InArchive, extractTo, err)
		}
	}

	return hash, nil
}
