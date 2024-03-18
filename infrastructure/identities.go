package infrastructure

import (
	"errors"
	"fmt"
	"os"
	"time"

	"filippo.io/age"

	"github.com/prskr/git-age/core/ports"
)

var _ ports.Identities = (*Identities)(nil)

func NewIdentities(r ports.KeysStore) *Identities {
	return &Identities{
		KeysReader: r,
	}
}

type Identities struct {
	KeysReader ports.KeysStore
}

func (i Identities) All() ([]age.Identity, error) {
	f, err := i.KeysReader.Reader()
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

	if comment == "" {
		comment = "# generated on " + time.Now().Format(time.RFC3339)
	}

	return identity.Recipient().String(), i.KeysReader.Write(identity, comment)
}
