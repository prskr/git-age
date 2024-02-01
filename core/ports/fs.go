package ports

import (
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"path/filepath"

	"github.com/go-git/go-git/v5/plumbing"
)

type ReadWriteFile interface {
	fs.File
	io.WriteSeeker
	WriteString(s string) (int, error)
	Name() string
}

type ReadWriteFS interface {
	fs.FS
	OpenRW(filePath string) (ReadWriteFile, error)
	TempFile(dir, pattern string) (ReadWriteFile, error)
	Rename(oldPath, newPath string) error
	Remove(filePath string) error
}

func ReEncryptWalkFunc(repo GitRepository, rwfs ReadWriteFS, sealer FileOpenSealer) fs.WalkDirFunc {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		logger := slog.Default().With(slog.String("path", path))

		logger.Info("Re-encrypting file")

		logger.Debug("Checking if file was already present at HEAD")
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

		dir, fileName := filepath.Split(path)
		tmp, err := rwfs.TempFile(dir, fileName)
		if err != nil {
			return err
		}
		logger.Debug("Preserve decrypted file in temp file", slog.String("tmp_file_path", tmp.Name()))

		defer func() {
			if err == nil {
				err = errors.Join(tmp.Close(), rwfs.Rename(tmp.Name(), path))
			} else {
				_ = tmp.Close()
				_ = rwfs.Remove(tmp.Name())
			}
		}()

		f, err := rwfs.OpenRW(path)
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
			err = errors.Join(err, encryptWriter.Close(), repo.StageFile(path))
		}()

		_, err = io.Copy(encryptWriter, io.TeeReader(plainTextReader, tmp))
		return err
	}
}
