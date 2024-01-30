package ports

import "filippo.io/age"

type Recipients interface {
	All() ([]age.Recipient, error)
	Append(pubKey string, comment string) ([]age.Recipient, error)
}
