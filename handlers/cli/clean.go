package cli

import (
	"bufio"
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/go-git/go-git/v5/plumbing"

	"github.com/prskr/git-age/core/dto"
	"github.com/prskr/git-age/internal/fsx"

	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/core/services"
	"github.com/prskr/git-age/infrastructure"
)

type CleanCliHandler struct {
	KeysFlag        `embed:""`
	Repository      ports.GitRepository  `kong:"-"`
	OpenSealer      ports.FileOpenSealer `kong:"-"`
	FileToCleanPath string               `arg:"" name:"file" help:"Path to the file to clean"`
}

func (h *CleanCliHandler) Run(stdin ports.STDIN, stdout ports.STDOUT) error {
	if err := requireStdin(stdin); err != nil {
		return err
	}

	logger := slog.Default().With("path", h.FileToCleanPath)

	if !h.OpenSealer.CanSeal() {
		logger.Warn("No recipients specified - file will be staged as plain text")
		if _, err := io.Copy(stdout, stdin); err != nil {
			return fmt.Errorf("failed to copy file to stdout: %w", err)
		}
		return nil
	}

	logger.Info("Copying file to temp")
	fileToClean, err := copyToTemp(stdin)
	if err != nil {
		return err
	}

	defer func() {
		_ = fileToClean.Close()
		_ = os.Remove(fileToClean.Name())
	}()

	logger.Info("Hashing file at HEAD")
	obj, headHash, err := h.hashFileAtHead(h.FileToCleanPath, true)
	if err != nil {
		if isFileNotFound(err) {
			logger.Info("Could not compare file to HEAD, handling as new")
			return h.copyEncryptedFileToStdout(fileToClean, stdout)
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
		return h.copyGitObjectTo(obj, stdout)
	}

	logger.Info("File has changed since last commit")
	return h.copyEncryptedFileToStdout(fileToClean, stdout)
}

func (h *CleanCliHandler) AfterApply(ctx context.Context, cwd ports.CWD, env ports.OSEnv) (err error) {
	var repoFS ports.ReadWriteFS

	h.Repository, repoFS, err = infrastructure.NewGitRepositoryFromPath(cwd)
	if err != nil {
		return err
	}

	idStore, err := infrastructure.IdentitiesStore(
		ctx,
		infrastructure.NewAgentIdentitiesStoreSource(env),
		infrastructure.NewFileIdentityStoreSource(h.Keys),
	)
	if err != nil {
		return fmt.Errorf("failed to init identities store: %w", err)
	}

	remotes, err := h.Repository.Remotes()
	if err != nil {
		return fmt.Errorf("failed to determine Git remotes: %w", err)
	}

	query := dto.IdentitiesQuery{
		Remotes: remotes,
	}

	ids, err := idStore.Identities(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to get identities: %w", err)
	}

	h.OpenSealer, err = services.NewAgeSealer(
		services.WithRecipients(infrastructure.NewRecipientsFile(repoFS)),
		services.WithIdentities(ids...),
	)

	return err
}

func (h *CleanCliHandler) copyEncryptedFileToStdout(reader io.Reader, out io.Writer) (err error) {
	encryptWriter, err := h.OpenSealer.SealFile(out)
	if err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, encryptWriter.Close())
	}()

	_, err = io.Copy(encryptWriter, reader)

	return err
}

func (h *CleanCliHandler) copyGitObjectTo(obj *object.File, out io.Writer) error {
	r, err := obj.Blob.Reader()
	if err != nil {
		return err
	}

	defer func() {
		_ = r.Close()
	}()

	_, err = io.Copy(out, r)

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

func isFileNotFound(err error) bool {
	switch {
	case errors.Is(err, plumbing.ErrObjectNotFound):
		return true
	case errors.Is(err, plumbing.ErrReferenceNotFound):
		return true
	case errors.Is(err, object.ErrFileNotFound):
		return true
	default:
		return false
	}
}
