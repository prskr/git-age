package infrastructure

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"filippo.io/age"

	"github.com/prskr/git-age/core/dto"
	"github.com/prskr/git-age/core/ports"
)

var (
	_ ports.IdentitiesStore = (*FileIdentityStore)(nil)
	_ identityStoreSource   = (*FileIdentityStoreSource)(nil)
)

func NewFileIdentityStoreSource(url *url.URL) *FileIdentityStoreSource {
	return (*FileIdentityStoreSource)(url)
}

type FileIdentityStoreSource url.URL

func (f *FileIdentityStoreSource) IsValid(context.Context) (bool, error) {
	return f != nil && f.Path != "", nil
}

func (f *FileIdentityStoreSource) GetStore() (ports.IdentitiesStore, error) {
	return (*FileIdentityStore)(f), nil
}

type FileIdentityStore url.URL

func (f *FileIdentityStore) Identities(context.Context, dto.IdentitiesQuery) ([]age.Identity, error) {
	keysFile, err := os.Open(f.identitiesFilePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to open identities file: %w", err)
	}

	defer func() {
		_ = keysFile.Close()
	}()

	return age.ParseIdentities(keysFile)
}

func (f *FileIdentityStore) Generate(_ context.Context, cmd dto.GenerateIdentityCommand) (publicKey string, err error) {
	newId, err := age.GenerateX25519Identity()
	if err != nil {
		return "", err
	}

	publicKey = newId.Recipient().String()

	if cmd.Comment == "" {
		cmd.Comment = "# generated on " + time.Now().Format(time.RFC3339)
	}

	ifp := f.identitiesFilePath()
	identitiesDir, _ := filepath.Split(ifp)
	if err := os.MkdirAll(identitiesDir, 0o700); err != nil {
		return "", fmt.Errorf("failed to create identities directory: %w", err)
	}

	identitiesFile, err := os.OpenFile(ifp, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return "", fmt.Errorf("failed to open identities file: %w", err)
	}

	defer func() {
		err = errors.Join(err, identitiesFile.Close())
	}()

	scanner := bufio.NewScanner(strings.NewReader(cmd.Comment))
	for scanner.Scan() {
		if _, err := identitiesFile.WriteString(fmt.Sprintf("# %s\n", scanner.Text())); err != nil {
			return "", fmt.Errorf("failed to write comment to identities file: %w", err)
		}
	}

	if scanner.Err() != nil {
		return "", fmt.Errorf("failed to write comment: %w", err)
	}

	if _, err := identitiesFile.WriteString(fmt.Sprintf("# public key: %s\n", publicKey)); err != nil {
		return "", fmt.Errorf("failed to write public key to identities file: %w", err)
	}

	if _, err := identitiesFile.WriteString(newId.String() + "\n"); err != nil {
		return "", fmt.Errorf("failed to write public key to identities file: %w", err)
	}

	return publicKey, nil
}

func (f *FileIdentityStore) identitiesFilePath() string {
	if runtime.GOOS == "windows" {
		return strings.TrimLeft(f.Path, "/")
	}

	return f.Path
}
