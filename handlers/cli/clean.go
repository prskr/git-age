package cli

import (
	"bytes"
	"errors"
	"filippo.io/age"
	"fmt"
	"github.com/adrg/xdg"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/minio/sha256-simd"
	"github.com/urfave/cli/v2"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

type CleanCliHandler struct {
	baseHandler
	Repository *git.Repository
}

func (h *CleanCliHandler) CleanFile(ctx *cli.Context) error {
	if err := requireStdin(); err != nil {
		return err
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	alwaysEncrypt, err := h.checkIfRecipientsChanged(wd)
	if err != nil {
		return err
	}

	fileToCleanPath := ctx.Args().First()

	fileToClean, err := copyToTemp(os.Stdin)
	if err != nil {
		return err
	}

	defer func() {
		_ = fileToClean.Close()
		_ = os.Remove(fileToClean.Name())
	}()

	if !alwaysEncrypt {
		obj, headHash, err := h.hashFileAtHead(fileToCleanPath, false)
		if errors.Is(err, plumbing.ErrObjectNotFound) {
			return h.copyEncryptedFileToStdout(fileToClean)
		}

		currentHash, err := h.hashFileAt(fileToClean)
		if err != nil {
			return err
		}

		if bytes.Equal(headHash, currentHash) {
			return h.copyGitObjectToStdout(obj)
		}
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

			h.Repository, err = git.PlainOpen(wd)
			if err != nil {
				return err
			}

			keysPath := filepath.Join(xdg.ConfigHome, "git-age", "keys.txt")
			if flagPath := context.String("keys"); flagPath != "" {
				keysPath = flagPath
			}

			return errors.Join(h.AddRecipientsFromPath(wd), h.AddIdentitiesFromPath(keysPath))
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "keys",
				DefaultText: "By default keys are read from $XDG_CONFIG_HOME/git-age/keys.txt i.e. $HOME/.config/git-age/keys.txt on most systems",
				EnvVars: []string{
					"GIT_AGE_KEYS",
				},
			},
		},
	}
}

func (h *CleanCliHandler) checkIfRecipientsChanged(repoRoot string) (alwaysEncrypt bool, err error) {
	_, recipientsHeadHash, err := h.hashFileAtHead(recipientsFileName, true)
	if errors.Is(err, plumbing.ErrObjectNotFound) {
		alwaysEncrypt = true
	} else if err != nil {
		return false, err
	}

	recipientsFile, err := os.Open(filepath.Join(repoRoot, recipientsFileName))
	if err != nil {
		return false, fmt.Errorf("no recipients file found, please run 'git age add' first")
	}

	defer func() {
		_ = recipientsFile.Close()
	}()

	currentRecipientsHash, err := h.hashFileAt(recipientsFile)
	if err != nil {
		return false, err
	}

	if !bytes.Equal(recipientsHeadHash, currentRecipientsHash) {
		alwaysEncrypt = true
	}

	return alwaysEncrypt, nil
}

func (h *CleanCliHandler) copyEncryptedFileToStdout(reader io.Reader) (err error) {
	encryptWriter, err := age.Encrypt(os.Stdout, h.Recipients...)
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
	if r, err := obj.Blob.Reader(); err != nil {
		return err
	} else {
		defer func() {
			_ = r.Close()
		}()

		_, err = io.Copy(os.Stdout, r)
		return err
	}
}

func (h *CleanCliHandler) hashFileAtHead(path string, encrypted bool) (obj *object.File, hash []byte, err error) {
	head, err := h.Repository.Head()
	if err != nil {
		return nil, nil, err
	}

	latestCommit, err := h.Repository.CommitObject(head.Hash())
	if err != nil {
		return nil, nil, err
	}

	headTree, err := latestCommit.Tree()
	if err != nil {
		return nil, nil, err
	}

	fileObjAtHead, err := headTree.File(path)
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
		if r, err := age.Decrypt(fileObjReader, h.Identities...); err != nil {
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
