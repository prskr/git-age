package infrastructure

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"slices"
	"strings"

	"filippo.io/age"
	"github.com/prskr/git-age/core/ports"
)

var _ ports.Recipients = (*RecipientsFile)(nil)

func NewRecipientsFile(fs ports.ReadWriteFS) *RecipientsFile {
	return &RecipientsFile{FS: fs}
}

type RecipientsFile struct {
	FS ports.ReadWriteFS
}

func (r RecipientsFile) All() ([]age.Recipient, error) {
	existingRecipients, err := r.FS.Open(ports.RecipientsFileName)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
		return nil, nil
	}

	defer func() {
		_ = existingRecipients.Close()
	}()

	return age.ParseRecipients(existingRecipients)
}

func (r RecipientsFile) Append(pubKey string, comment string) (recipients []age.Recipient, err error) {
	recipients, err = age.ParseRecipients(strings.NewReader(pubKey))
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	existingRecipients, err := r.All()
	if err != nil {
		return nil, err
	}

	alreadyInRecipients := slices.ContainsFunc(existingRecipients, func(recipient age.Recipient) bool {
		switch actual := recipient.(type) {
		// currently there are only X25519 recipients
		case *age.X25519Recipient:
			return actual.String() == pubKey
		default:
			return false
		}
	})

	if alreadyInRecipients {
		return nil, nil
	}

	recipientsFile, err := r.FS.OpenRW(ports.RecipientsFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open recipients file: %w", err)
	}

	if _, err := recipientsFile.Seek(0, io.SeekEnd); err != nil {
		return nil, err
	}

	defer func() {
		err = errors.Join(err, recipientsFile.Close())
	}()

	if comment != "" {
		if _, err := recipientsFile.WriteString(fmt.Sprintf("# %s\n", comment)); err != nil {
			return nil, fmt.Errorf("failed to write comment to recipients file: %w", err)
		}
	}

	if _, err := recipientsFile.WriteString(pubKey + "\n"); err != nil {
		return nil, fmt.Errorf("failed to write public key to recipients file: %w", err)
	}

	return recipients, nil
}
