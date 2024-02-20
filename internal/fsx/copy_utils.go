package fsx

import (
	"errors"
	"os"
)

func CopyFile(src, dst string) (err error) {
	var (
		srcFile *os.File
		dstFile *os.File
	)

	if srcFile, err = os.Open(src); err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, srcFile.Close())
	}()

	if dstFile, err = os.Create(dst); err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, dstFile.Close())
	}()

	if _, err = srcFile.WriteTo(dstFile); err != nil {
		return err
	}

	return nil
}
