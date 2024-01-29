package ports

import (
	"errors"
	"io"
	"io/fs"

	"github.com/go-git/go-git/v5/plumbing"
)

type ReadWriteFile interface {
	fs.File
	io.WriteSeeker
	WriteString(s string) (int, error)
}

type ReadWriteFS interface {
	fs.FS
	Create(filePath string) (ReadWriteFile, error)
	Append(filePath string) (ReadWriteFile, error)
}

func ReEncryptWalkFunc(repo GitRepository, rwfs ReadWriteFS, sealer FileOpenSealer) fs.WalkDirFunc {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		fileObj, err := repo.OpenObjectAtHead(path)
		if err != nil {
			if errors.Is(err, plumbing.ErrObjectNotFound) {
				return nil
			}
			return err
		}

		objReader, err := fileObj.Reader()
		if err != nil {
			return err
		}

		defer func() {
			_ = objReader.Close()
		}()

		f, err := rwfs.Create(path)
		if err != nil {
			return err
		}

		defer func() {
			err = errors.Join(err, f.Close())
		}()

		plainTextReader, err := sealer.OpenFile(objReader)
		if err != nil {
			return err
		}

		encryptWriter, err := sealer.SealFile(f)
		if err != nil {
			return err
		}

		defer func() {
			err = errors.Join(err, encryptWriter.Close())
		}()

		_, err = io.Copy(encryptWriter, plainTextReader)
		if err != nil {
			return err
		}

		return repo.StageFile(path)
	}
}
