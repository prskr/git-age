package infrastructure

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/prskr/git-age/core/ports"
)

var ErrNotSupportedKeysSource = errors.New("not supported keys source scheme")

func KeysStoreFor(u *url.URL) (ports.KeysStore, error) {
	switch u.Scheme {
	case "", "file":
		return (*FileKeysStore)(u), nil
	case "keychain":
		return (*KeyRingKeysStore)(u), nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrNotSupportedKeysSource, u.Scheme)
	}
}
