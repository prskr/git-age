package infrastructure

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"filippo.io/age"
	"github.com/prskr/git-age/core/ports"
)

var _ ports.Identities = (*Identities)(nil)

func NewIdentities(identitiesFile string) *Identities {
	return &Identities{
		IdentitiesFile: identitiesFile,
	}
}

type Identities struct {
	IdentitiesFile string
}

func (i Identities) All() ([]age.Identity, error) {
	f, err := os.Open(i.IdentitiesFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	defer func() {
		_ = f.Close()
	}()

	return age.ParseIdentities(f)
}

func (i Identities) Generate(comment string) (pubKey string, err error) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return "", fmt.Errorf("failed to generate identity: %w", err)
	}

	identitiesDir, _ := filepath.Split(i.IdentitiesFile)
	if err := os.MkdirAll(identitiesDir, 0o700); err != nil {
		return "", fmt.Errorf("failed to create identities directory: %w", err)
	}

	identitiesFile, err := os.OpenFile(i.IdentitiesFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return "", fmt.Errorf("failed to open identities file: %w", err)
	}

	defer func() {
		err = errors.Join(err, identitiesFile.Close())
	}()

	if comment == "" {
		comment = "# generated on " + time.Now().Format(time.RFC3339)
	}

	scanner := bufio.NewScanner(strings.NewReader(comment))
	for scanner.Scan() {
		if _, err := identitiesFile.WriteString(fmt.Sprintf("# %s\n", scanner.Text())); err != nil {
			return "", fmt.Errorf("failed to write comment to identities file: %w", err)
		}
	}

	if scanner.Err() != nil {
		return "", fmt.Errorf("failed to write comment: %w", err)
	}

	if _, err := identitiesFile.WriteString(fmt.Sprintf("# public key: %s\n", identity.Recipient().String())); err != nil {
		return "", fmt.Errorf("failed to write public key to identities file: %w", err)
	}

	if _, err := identitiesFile.WriteString(identity.String() + "\n"); err != nil {
		return "", fmt.Errorf("failed to write public key to identities file: %w", err)
	}

	return identity.Recipient().String(), nil
}
