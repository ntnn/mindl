package sum

import (
	"encoding/base64"
	"os"
)

// Base64 encodes in as a base64 string.
func Base64(in []byte) string {
	return base64.StdEncoding.EncodeToString(in)
}

// HashPathFunc is a function that hashes the file at the given path and
// returns the base64 encoded hash.
type HashPathFunc func(string) (string, error)

// PathMatchesHash check if the file at the given path matches the given hash when hashed with hasher.
// If the target does not exist the functions returns (false, nil) as if
// the hash did not match.
func PathMatchesHash(p string, hasher HashPathFunc, expected string) (bool, error) {
	_, err := os.Stat(p)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}
	if os.IsNotExist(err) {
		return false, nil
	}

	hash, err := hasher(p)
	if err != nil {
		return false, err
	}
	return hash == expected, nil
}
