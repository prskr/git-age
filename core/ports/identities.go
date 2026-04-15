package ports

import (
	"context"

	"filippo.io/age"
)

type GenerateIdentityCommand struct {
	Comment   string
	Remote    string
	Algorithm IdentityAlgorithm
}

type IdentitiesQuery struct {
	Remotes []string
}

type IdentitiesStore interface {
	Generate(ctx context.Context, cmd GenerateIdentityCommand) (publicKey string, err error)
	Identities(ctx context.Context, query IdentitiesQuery) ([]age.Identity, error)
}
