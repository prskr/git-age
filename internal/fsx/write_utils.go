package fsx

import (
	"errors"
	"path/filepath"

	"github.com/prskr/git-age/core/ports"
)

func WriteTo(rwfs ports.ReadWriteFS, filePath string, data []byte) (err error) {
	dir, _ := filepath.Split(filePath)

	if dir != "" {
		if err := rwfs.Mkdir(dir, true, 0o755); err != nil {
			return err
		}
	}

	file, err := rwfs.OpenRW(filePath)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, file.Close())
	}()

	_, err = file.Write(data)
	return err
}
