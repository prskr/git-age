package ports

import (
	"context"

	"filippo.io/age"

	"github.com/prskr/git-age/core/dto"
)

type IdentitiesStore interface {
	Generate(ctx context.Context, cmd dto.GenerateIdentityCommand) (publicKey string, err error)
	Identities(ctx context.Context, query dto.IdentitiesQuery) ([]age.Identity, error)
}
