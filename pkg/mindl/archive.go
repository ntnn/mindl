package mindl

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// Unarchive extracts target from archive to output.
// For paths ending in .zip [Unzip] is used.
// Otherwise [Untar] is used.
func Unarchive(archive, target, output string) error {
	if strings.HasSuffix(archive, ".zip") {
		return Unzip(archive, target, output)
	}
	return Untar(archive, target, output)
}

// Unzip extracts target from a zip archive to output.
func Unzip(archive, target, output string) error {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer reader.Close()

	f, err := reader.Open(target)
	if err != nil {
		return err
	}
	defer f.Close()

	out, err := os.Create(output)
	if err != nil {
		return err
	}

	_, err = io.Copy(out, f)
	return err
}

// Untar extracts target from a tar archive to output.
//
//nolint:cyclop // not that complex, mostly compression + for loop
func Untar(archive, target, output string) error {
	raw, err := os.Open(archive)
	if err != nil {
		return fmt.Errorf("error opening archive %q: %w", archive, err)
	}
	defer raw.Close()

	var f io.Reader = raw

	// handle compression
	switch {
	case strings.HasSuffix(archive, ".gz"):
		gzreader, err := gzip.NewReader(raw)
		if err != nil {
			return fmt.Errorf("error adding gzip reader: %w", err)
		}
		f = gzreader
	case strings.HasSuffix(archive, ".bz2"):
		f = bzip2.NewReader(raw)
	}

	reader := tar.NewReader(f)

	for {
		header, err := reader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return fmt.Errorf("hit end of archive without finding %q", target)
			}
			return fmt.Errorf("error reading tar: %w", err)
		}

		if header.Name != target {
			continue
		}

		out, err := os.Create(output)
		if err != nil {
			return fmt.Errorf("error creating output %q: %w", output, err)
		}
		defer out.Close()

		if _, err := io.Copy(out, reader); err != nil {
			return fmt.Errorf("error writing to output %q: %w", output, err)
		}
		return nil
	}
}
