package cli

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"os"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/minio/sha256-simd"
	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/core/services"
	"github.com/prskr/git-age/infrastructure"
	"github.com/urfave/cli/v2"
)

type CleanCliHandler struct {
	Repository ports.GitRepository
	OpenSealer ports.FileOpenSealer
}

func (h *CleanCliHandler) CleanFile(ctx *cli.Context) error {
	if err := requireStdin(); err != nil {
		return err
	}

	fileToCleanPath := ctx.Args().First()
	logger := slog.Default().With("path", fileToCleanPath)

	logger.Info("Copying file to temp")
	fileToClean, err := copyToTemp(os.Stdin)
	if err != nil {
		return err
	}

	defer func() {
		_ = fileToClean.Close()
		_ = os.Remove(fileToClean.Name())
	}()

	logger.Info("Hashing file at HEAD")
	obj, headHash, err := h.hashFileAtHead(fileToCleanPath, true)
	if errors.Is(err, plumbing.ErrObjectNotFound) {
		logger.Info("File not found at HEAD, file is apparently new")
		return h.copyEncryptedFileToStdout(fileToClean)
	}

	logger.Info("Hashing file at current state to determine whether it has changed")
	currentHash, err := h.hashFileAt(fileToClean)
	if err != nil {
		return err
	}

	if bytes.Equal(headHash, currentHash) {
		logger.Info("File has not changed, returning original")
		return h.copyGitObjectToStdout(obj)
	}

	return h.copyEncryptedFileToStdout(fileToClean)
}

func (h *CleanCliHandler) Command() *cli.Command {
	return &cli.Command{
		Name:   "clean",
		Usage:  "clean should only be invoked by Git",
		Action: h.CleanFile,
		Args:   true,
		Hidden: true,
		Before: func(context *cli.Context) error {
			wd, err := os.Getwd()
			if err != nil {
				return err
			}

			var repoFS ports.ReadWriteFS

			h.Repository, repoFS, err = infrastructure.NewGitRepositoryFromPath(wd)
			if err != nil {
				return err
			}

			h.OpenSealer, err = services.NewAgeSealer(
				services.WithRecipients(infrastructure.NewRecipientsFile(repoFS)),
				services.WithIdentities(infrastructure.NewIdentities(context.String("keys"))),
			)

			return err
		},
		Flags: []cli.Flag{
			&keysFlag,
		},
	}
}

func (h *CleanCliHandler) copyEncryptedFileToStdout(reader io.Reader) (err error) {
	encryptWriter, err := h.OpenSealer.SealFile(os.Stdout)
	if err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, encryptWriter.Close())
	}()

	_, err = io.Copy(encryptWriter, reader)

	return err
}

func (h *CleanCliHandler) copyGitObjectToStdout(obj *object.File) error {
	r, err := obj.Blob.Reader()
	if err != nil {
		return err
	}

	defer func() {
		_ = r.Close()
	}()

	_, err = io.Copy(os.Stdout, r)
	return err
}

func (h *CleanCliHandler) hashFileAtHead(path string, encrypted bool) (obj *object.File, hash []byte, err error) {
	fileObjAtHead, err := h.Repository.OpenObjectAtHead(path)
	if err != nil {
		return nil, nil, err
	}

	fileObjReader, err := fileObjAtHead.Blob.Reader()
	if err != nil {
		return nil, nil, err
	}

	defer func() {
		_ = fileObjReader.Close()
	}()

	var reader io.Reader = fileObjReader
	if encrypted {
		if r, err := h.OpenSealer.OpenFile(fileObjReader); err != nil {
			slog.Warn("Expected encrypted file but failed to decrypt", slog.String("path", path), slog.String("err", err.Error()))
		} else {
			reader = r
		}
	}

	hashBytes, err := hashFile(reader)
	return fileObjAtHead, hashBytes, err
}

func (h *CleanCliHandler) hashFileAt(f io.ReadSeeker) (hash []byte, err error) {
	defer func() {
		_, seekErr := f.Seek(0, io.SeekStart)
		err = errors.Join(err, seekErr)
	}()

	return hashFile(f)
}

func hashFile(reader io.Reader) ([]byte, error) {
	h := sha256.New()
	if _, err := io.Copy(h, reader); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}
