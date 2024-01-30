package ports

import "filippo.io/age"

type Identities interface {
	Generate(comment string) (pubKey string, err error)
	All() ([]age.Identity, error)
}
