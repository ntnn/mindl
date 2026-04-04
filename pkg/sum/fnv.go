package sum

import (
	"encoding/base64"
	"hash/fnv"
	"io"
	"os"
)

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

func Base64(in []byte) string {
	return base64.StdEncoding.EncodeToString(in)
}
