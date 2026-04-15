package ports

import (
	"filippo.io/age"
)

type Identity interface {
	Recipient() Recipient
	Unwrap(stanzas []*age.Stanza) ([]byte, error)
	String() string
}

type Recipient interface {
	String() string
	Wrap(fileKey []byte) ([]*age.Stanza, error)
}
