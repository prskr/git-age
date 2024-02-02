package services

import (
	"io"
	"os"
	"path/filepath"

	"filippo.io/age"
	"github.com/prskr/git-age/core/ports"
)

var (
	_ ports.FileOpener = (*AgeSealer)(nil)
	_ ports.FileSealer = (*AgeSealer)(nil)
)

type AgeSealerOption func(*AgeSealer) error

func WithRecipients(r ports.Recipients) AgeSealerOption {
	return func(sealer *AgeSealer) error {
		recipients, err := r.All()
		if err != nil {
			return err
		}

		sealer.Recipients = recipients
		return nil
	}
}

func WithIdentities(id ports.Identities) AgeSealerOption {
	return func(sealer *AgeSealer) error {
		identities, err := id.All()
		if err != nil {
			return err
		}

		sealer.Identities = identities
		return nil
	}
}

func NewAgeSealer(opts ...AgeSealerOption) (*AgeSealer, error) {
	sealer := new(AgeSealer)

	for _, opt := range opts {
		if err := opt(sealer); err != nil {
			return nil, err
		}
	}

	return sealer, nil
}

type AgeSealer struct {
	Recipients []age.Recipient
	Identities []age.Identity
}

func (h *AgeSealer) CanSeal() bool {
	return len(h.Recipients) > 0
}

func (h *AgeSealer) AddRecipients(r ...age.Recipient) {
	h.Recipients = append(h.Recipients, r...)
}

func (h *AgeSealer) AddIdentity(identity age.Identity) {
	h.Identities = append(h.Identities, identity)
}

func (h *AgeSealer) OpenFile(reader io.Reader) (io.Reader, error) {
	return age.Decrypt(reader, h.Identities...)
}

func (h *AgeSealer) SealFile(dst io.Writer) (io.WriteCloser, error) {
	return age.Encrypt(dst, h.Recipients...)
}

func (h *AgeSealer) AddRecipientsFrom(reader io.Reader) error {
	r, err := age.ParseRecipients(reader)
	if err != nil {
		return err
	}

	h.Recipients = append(h.Recipients, r...)
	return nil
}

func (h *AgeSealer) AddRecipientsFromPath(repoPath string) error {
	f, err := os.Open(filepath.Join(repoPath, ports.RecipientsFileName))
	if err != nil {
		return err
	}

	defer func() {
		_ = f.Close()
	}()

	return h.AddRecipientsFrom(f)
}

func (h *AgeSealer) AddIdentitiesFrom(reader io.Reader) error {
	i, err := age.ParseIdentities(reader)
	if err != nil {
		return err
	}

	h.Identities = append(h.Identities, i...)
	return nil
}

func (h *AgeSealer) AddIdentitiesFromPath(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	defer func() {
		_ = f.Close()
	}()

	return h.AddIdentitiesFrom(f)
}
