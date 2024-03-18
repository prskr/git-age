package infrastructure

import (
	"bytes"
	"io"
	"net/url"

	"filippo.io/age"
	"github.com/99designs/keyring"

	"github.com/prskr/git-age/core/ports"
)

var _ ports.KeysStore = (*KeyRingKeysStore)(nil)

type KeyRingKeysStore url.URL

func (k KeyRingKeysStore) Reader() (io.ReadCloser, error) {
	keyRingCfg := keyring.Config{
		ServiceName: k.Host,
	}
	ring, err := keyring.Open(keyRingCfg)
	if err != nil {
		return nil, err
	}

	keys, err := ring.Keys()
	if err != nil {
		return nil, err
	}

	keysBuf := bytes.NewBuffer(nil)

	for _, k := range keys {
		item, err := ring.Get(k)
		if err != nil {
			return nil, err
		}

		_, _ = keysBuf.Write(item.Data)
		_, _ = keysBuf.WriteRune('\n')
	}

	return io.NopCloser(keysBuf), nil
}

func (k KeyRingKeysStore) Write(id *age.X25519Identity, comment string) (err error) {
	keyRingCfg := keyring.Config{
		ServiceName: k.Host,
	}

	ring, err := keyring.Open(keyRingCfg)
	if err != nil {
		return err
	}

	it := keyring.Item{
		Key:         id.Recipient().String(),
		Data:        []byte(id.String()),
		Description: comment,
	}

	return ring.Set(it)
}

func (k KeyRingKeysStore) Clear() error {
	keyRingCfg := keyring.Config{
		ServiceName: k.Host,
	}

	ring, err := keyring.Open(keyRingCfg)
	if err != nil {
		return err
	}

	allKeys, err := ring.Keys()
	if err != nil {
		return err
	}

	for _, k := range allKeys {
		if err := ring.Remove(k); err != nil {
			return err
		}
	}

	return nil
}
