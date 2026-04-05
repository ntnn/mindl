package sum

import (
	"crypto/sha512"
	"io"
	"os"
)

// Sha512Path returns the base64-encoded SHA-512 hash of the file at p.
func Sha512Path(p string) (string, error) {
	f, err := os.Open(p)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hasher := sha512.New()

	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}

	return Base64(hasher.Sum(nil)), nil
}
