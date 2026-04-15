package ports

import (
	"encoding"
	"errors"
	"fmt"
	"strings"

	"filippo.io/age"
)

var ErrUnknownAlgorithm = errors.New("unknown algorithm")

var _ encoding.TextUnmarshaler = (*IdentityAlgorithm)(nil)

const (
	IdentityAlgorithmUnknown IdentityAlgorithm = ""
	IdentityAlgorithmX25519  IdentityAlgorithm = "x25519"
	IdentityAlgorithmHybrid  IdentityAlgorithm = "hybrid"
)

const hybridRecipientPrefix = "age1pq1"

type IdentityAlgorithm string

func (a IdentityAlgorithm) ParseRecipient(raw string) (age.Recipient, error) {
	switch a {
	case IdentityAlgorithmHybrid:
		return age.ParseHybridRecipient(raw)
	case IdentityAlgorithmX25519:
		return age.ParseX25519Recipient(raw)
	case IdentityAlgorithmUnknown:
		fallthrough
	default:
		if strings.HasPrefix(raw, hybridRecipientPrefix) {
			return age.ParseHybridRecipient(raw)
		}

		return age.ParseX25519Recipient(raw)
	}
}

// UnmarshalText implements [encoding.TextUnmarshaler].
func (a *IdentityAlgorithm) UnmarshalText(text []byte) error {
	switch string(text) {
	case "x25519":
		*a = IdentityAlgorithmX25519
	case "hybrid":
		*a = IdentityAlgorithmHybrid
	default:
		return fmt.Errorf("%w: %s", ErrUnknownAlgorithm, text)
	}
	return nil
}

func (a IdentityAlgorithm) Generate() (Identity, error) {
	switch a {
	case IdentityAlgorithmX25519:
		identity, err := age.GenerateX25519Identity()
		if err != nil {
			return nil, err
		}
		return &identityWrapper[*age.X25519Recipient]{identity: identity}, nil
	case IdentityAlgorithmHybrid, IdentityAlgorithmUnknown:
		fallthrough
	default:
		identity, err := age.GenerateHybridIdentity()
		if err != nil {
			return nil, err
		}
		return &identityWrapper[*age.HybridRecipient]{identity: identity}, nil
	}
}

type recipientType interface {
	Recipient
	*age.HybridRecipient | *age.X25519Recipient
}

type identityHack[R recipientType] interface {
	Recipient() R
	Unwrap(stanzas []*age.Stanza) ([]byte, error)
	String() string
}

var _ Identity = (*identityWrapper[*age.HybridRecipient])(nil)

type identityWrapper[R recipientType] struct {
	identity identityHack[R]
}

// String implements [Identity].
func (i *identityWrapper[R]) String() string {
	return i.identity.String()
}

// Unwrap implements [Identity].
func (i *identityWrapper[R]) Unwrap(stanzas []*age.Stanza) ([]byte, error) {
	return i.identity.Unwrap(stanzas)
}

// Recipient implements [Identity].
func (i *identityWrapper[R]) Recipient() Recipient {
	return i.identity.Recipient()
}
