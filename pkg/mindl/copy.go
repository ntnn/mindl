package mindl

import (
	"fmt"
	"io"
	"os"
)

func iocopy(srcpath, dstpath string) error {
	src, err := os.Open(srcpath)
	if err != nil {
		return fmt.Errorf("error opening %q for reading: %w", srcpath, err)
	}
	defer src.Close()

	dst, err := os.Create(dstpath)
	if err != nil {
		return fmt.Errorf("error opening and truncating %q for writing: %w", dstpath, err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}
