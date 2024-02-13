package cli

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/go-git/go-git/v5/plumbing"

	"github.com/prskr/git-age/internal/fsx"

	"github.com/go-git/go-git/v5/plumbing/object"
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

	if !h.OpenSealer.CanSeal() {
		logger.Warn("No recipients specified - file will be staged as plain text")
		if _, err := io.Copy(os.Stdout, os.Stdin); err != nil {
			return fmt.Errorf("failed to copy file to stdout: %w", err)
		}
		return nil
	}

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
	if err != nil {
		if errors.Is(err, plumbing.ErrObjectNotFound) {
			logger.Info("File not found at HEAD, file is apparently new")
			return h.copyEncryptedFileToStdout(fileToClean)
		}

		return fmt.Errorf("failed to hash file at HEAD: %w", err)
	}

	logger = logger.With(slog.String("orig_hash", hex.EncodeToString(headHash)))

	logger.Info("Hashing file at current state to determine whether it has changed")
	currentHash, err := h.hashFileAt(fileToClean)
	if err != nil {
		return err
	}

	logger = logger.With(slog.String("current_hash", hex.EncodeToString(currentHash)))

	if bytes.Equal(headHash, currentHash) {
		logger.Info("File has not changed, returning original")
		return h.copyGitObjectToStdout(obj)
	}

	logger.Info("File has changed since last commit")
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

func (h *CleanCliHandler) hashFileAtHead(
	path string,
	expectToBeEncrypted bool,
) (obj *object.File, hash []byte, err error) {
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

	if expectToBeEncrypted {
		bufReader := bufio.NewReader(reader)
		isEncrypted, err := h.OpenSealer.IsEncrypted(bufReader)
		if err != nil {
			return nil, nil, err
		} else if !isEncrypted {
			slog.Warn("Expected encrypted file but age header is missing", slog.String("path", path))
		} else if r, err := h.OpenSealer.OpenFile(bufReader); err != nil {
			return nil, nil, err
		} else {
			reader = r
		}
	}

	hashBytes, err := fsx.HashFile(reader)
	return fileObjAtHead, hashBytes, err
}

func (h *CleanCliHandler) hashFileAt(f io.ReadSeeker) (hash []byte, err error) {
	defer func() {
		_, seekErr := f.Seek(0, io.SeekStart)
		err = errors.Join(err, seekErr)
	}()

	return fsx.HashFile(f)
}
