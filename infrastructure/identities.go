package infrastructure

import (
	"context"

	"github.com/prskr/git-age/core/ports"
	"github.com/prskr/git-age/core/services"
)

type identityStoreSource interface {
	IsValid(ctx context.Context) (bool, error)
	GetStore() (ports.IdentitiesStore, error)
}

func IdentitiesStore(ctx context.Context, sources ...identityStoreSource) (ports.IdentitiesStore, error) {
	var stores []ports.IdentitiesStore

	for _, src := range sources {
		if isValid, err := src.IsValid(ctx); err != nil {
			return nil, err
		} else if isValid {
			store, err := src.GetStore()
			if err != nil {
				return nil, err
			}

			stores = append(stores, store)
		}
	}

	return services.NewIdentitiesStoreChain(stores...), nil
}
