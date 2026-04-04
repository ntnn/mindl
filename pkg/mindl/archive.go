package mindl

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func Unarchive(archive, target, output string) error {
	if strings.HasSuffix(archive, ".zip") {
		return Unzip(archive, target, output)
	}
	return Untar(archive, target, output)
}

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
