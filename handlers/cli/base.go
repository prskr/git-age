package cli

import (
	"filippo.io/age"
	"io"
	"os"
	"path/filepath"
)

const recipientsFileName = ".agerecipients"

type baseHandler struct {
	Recipients []age.Recipient
	Identities []age.Identity
}

func (h *baseHandler) AddRecipientsFrom(reader io.Reader) error {
	r, err := age.ParseRecipients(reader)
	if err != nil {
		return err
	}

	h.Recipients = append(h.Recipients, r...)
	return nil
}

func (h *baseHandler) AddRecipientsFromPath(repoPath string) error {
	f, err := os.Open(filepath.Join(repoPath, recipientsFileName))
	if err != nil {
		return err
	}

	defer func() {
		_ = f.Close()
	}()

	return h.AddRecipientsFrom(f)
}

func (h *baseHandler) AddIdentitiesFrom(reader io.Reader) error {
	i, err := age.ParseIdentities(reader)
	if err != nil {
		return err
	}

	h.Identities = append(h.Identities, i...)
	return nil
}

func (h *baseHandler) AddIdentitiesFromPath(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	defer func() {
		_ = f.Close()
	}()

	return h.AddIdentitiesFrom(f)
}
