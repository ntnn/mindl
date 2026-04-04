package sum

import (
	"encoding/base64"
	"hash/fnv"
	"io"
	"os"
)

// Fnv128aPath returns the base64-encoded FNV-128a hash of the file at p.
func Fnv128aPath(p string) (string, error) {
	f, err := os.Open(p)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hasher := fnv.New128a()

	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}

	return Base64(hasher.Sum(nil)), nil
}

// Base64 encodes in as a base64 string.
func Base64(in []byte) string {
	return base64.StdEncoding.EncodeToString(in)
}
