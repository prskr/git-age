package infrastructure

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"filippo.io/age"

	"github.com/prskr/git-age/core/ports"
)

var _ ports.KeysStore = (*FileKeysStore)(nil)

type FileKeysStore url.URL

func (f *FileKeysStore) Reader() (io.ReadCloser, error) {
	return os.Open(f.Path)
}

func (f *FileKeysStore) Write(id *age.X25519Identity, comment string) (err error) {
	identitiesDir, _ := filepath.Split(f.Path)
	if err := os.MkdirAll(identitiesDir, 0o700); err != nil {
		return fmt.Errorf("failed to create identities directory: %w", err)
	}

	identitiesFile, err := os.OpenFile(f.Path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open identities file: %w", err)
	}

	defer func() {
		err = errors.Join(err, identitiesFile.Close())
	}()

	scanner := bufio.NewScanner(strings.NewReader(comment))
	for scanner.Scan() {
		if _, err := identitiesFile.WriteString(fmt.Sprintf("# %s\n", scanner.Text())); err != nil {
			return fmt.Errorf("failed to write comment to identities file: %w", err)
		}
	}

	if scanner.Err() != nil {
		return fmt.Errorf("failed to write comment: %w", err)
	}

	if _, err := identitiesFile.WriteString(fmt.Sprintf("# public key: %s\n", id.Recipient().String())); err != nil {
		return fmt.Errorf("failed to write public key to identities file: %w", err)
	}

	if _, err := identitiesFile.WriteString(id.String() + "\n"); err != nil {
		return fmt.Errorf("failed to write public key to identities file: %w", err)
	}

	return nil
}

func (f *FileKeysStore) Clear() error {
	err := os.Remove(f.Path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
