package services

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"path/filepath"

	"github.com/go-git/go-git/v5/plumbing"

	"github.com/prskr/git-age/core/ports"
)

//nolint:funlen // cannot split this function further
func ReEncryptWalkFunc(repo ports.GitRepository, rwfs ports.ReadWriteFS, sealer ports.FileOpenSealer) fs.WalkDirFunc {
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
			return fmt.Errorf("opening object at path %s at HEAD: %w", path, err)
		}

		objReader, err := fileObj.Reader()
		if err != nil {
			return fmt.Errorf("opening object reader at path %s at HEAD: %w", path, err)
		}

		defer func() {
			_ = objReader.Close()
		}()

		dir, fileName := filepath.Split(path)
		tmp, err := rwfs.TempFile(dir, fileName)
		if err != nil {
			return fmt.Errorf("creating temp file: %w", err)
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

		f, err := rwfs.Create(path)
		if err != nil {
			return fmt.Errorf("creating file at path %s: %w", path, err)
		}

		defer func() {
			err = errors.Join(err, f.Close())
		}()

		plainTextReader, err := sealer.OpenFile(objReader)
		if err != nil {
			return fmt.Errorf("opening file for decryption: %w", err)
		}

		encryptWriter, err := sealer.SealFile(f)
		if err != nil {
			return fmt.Errorf("opening file for encryption: %w", err)
		}

		defer func() {
			err = errors.Join(err, encryptWriter.Close(), repo.StageFile(path))
		}()

		_, err = io.Copy(encryptWriter, io.TeeReader(plainTextReader, tmp))
		return err
	}
}
