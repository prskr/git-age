package services

import (
	"context"
	"errors"
	"fmt"

	"filippo.io/age"

	"github.com/prskr/git-age/core/dto"
	"github.com/prskr/git-age/core/ports"
)

var _ ports.IdentitiesStore = (*IdentitiesStoreChain)(nil)

var ErrEmptyChain = errors.New("empty identities chain")

func NewIdentitiesStoreChain(stores ...ports.IdentitiesStore) IdentitiesStoreChain {
	return stores
}

type IdentitiesStoreChain []ports.IdentitiesStore

func (i IdentitiesStoreChain) Generate(
	ctx context.Context,
	cmd dto.GenerateIdentityCommand,
) (publicKey string, err error) {
	for _, store := range i {
		return store.Generate(ctx, cmd)
	}

	return "", fmt.Errorf("cannot generate identity: %w", ErrEmptyChain)
}

func (i IdentitiesStoreChain) Identities(
	ctx context.Context,
	query dto.IdentitiesQuery,
) (result []age.Identity, err error) {
	var (
		out  = make(chan []age.Identity)
		errs = make(chan error)
	)

	defer func() {
		close(out)
		close(errs)
	}()

	for _, store := range i {
		go func() {
			ids, err := store.Identities(ctx, query)
			if err != nil {
				errs <- err
				return
			}

			out <- ids
		}()
	}

	for range len(i) {
		select {
		case ids := <-out:
			result = append(result, ids...)
		case e := <-errs:
			err = errors.Join(err, e)
		}
	}

	return result, err
}
