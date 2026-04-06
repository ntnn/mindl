package mindl

import (
	"fmt"
	"io"

	"github.com/ntnn/mindl/pkg/sum"
)

type Cache struct {
	cacheDir string
	db       *sum.DB
}

func OpenCache(cacheDir string) (*Cache, error) {
	c := new(Cache)
	c.cacheDir = cacheDir

	var err error
	c.db, err = sum.Open(c.cacheDir)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (cache *Cache) HasCached(url, inArchive, os, arch string) (string, bool) {
	return cache.db.Get(url, inArchive, os, arch)
}

func (cache *Cache) CopyTo(url, inArchive, os, arch, dst string) error {
	if _, present := cache.db.Get(url, inArchive, os, arch); !present {
		return fmt.Errorf("file not cached for url %q, in-archive %q, os %q, arch %q",
			url, inArchive, os, arch)
	}

	// TODO calculate from
	var from string

	in, err := os.Open(from)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return out.Close()
}
