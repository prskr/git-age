package infrastructure

import (
	"errors"
	"fmt"
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

func (r RecipientsFile) Append(pubKey string, comment string) (err error) {
	_, err = age.ParseRecipients(strings.NewReader(pubKey))
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	existingRecipients, err := r.readExistingRecipients()
	if err != nil {
		return err
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
		return nil
	}

	recipientsFile, err := r.FS.Append(ports.RecipientsFileName)
	if err != nil {
		return fmt.Errorf("failed to open recipients file: %w", err)
	}

	defer func() {
		err = errors.Join(err, recipientsFile.Close())
	}()

	if comment != "" {
		if _, err := recipientsFile.WriteString(fmt.Sprintf("# %s\n", comment)); err != nil {
			return fmt.Errorf("failed to write comment to recipients file: %w", err)
		}
	}

	if _, err := recipientsFile.WriteString(pubKey + "\n"); err != nil {
		return fmt.Errorf("failed to write public key to recipients file: %w", err)
	}

	return nil
}

func (r RecipientsFile) readExistingRecipients() ([]age.Recipient, error) {
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
