package fsx

import (
	"io"
	"io/fs"

	"github.com/minio/sha256-simd"
)

func HashFile(r io.Reader) ([]byte, error) {
	shaHash := sha256.New()
	_, err := io.Copy(shaHash, r)
	if err != nil {
		return nil, err
	}

	return shaHash.Sum(nil), nil
}

func HashFSFile(fs fs.FS, path string) ([]byte, error) {
	f, err := fs.Open(path)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = f.Close()
	}()

	shaHash := sha256.New()
	_, err = io.Copy(shaHash, f)
	if err != nil {
		return nil, err
	}

	return shaHash.Sum(nil), nil
}
