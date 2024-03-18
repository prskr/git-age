//go:build integration

package infrastructure_test

import (
	"fmt"
	"testing"

	"filippo.io/age"

	"github.com/prskr/git-age/infrastructure"
)

func TestKeyRingKeysStore_Write(t *testing.T) {
	t.Parallel()

	id, err := age.GenerateX25519Identity()
	if err != nil {
		t.Errorf("failed to generate identity: %v", err)
		return
	}

	store := (*infrastructure.KeyRingKeysStore)(mustParseUrl(t, fmt.Sprintf("keychain://git-age-%s", t.Name())))
	if err := store.Write(id, t.Name()); err != nil {
		t.Errorf("failed to write identity: %v", err)
	}

	t.Cleanup(func() {
		if err := store.Clear(); err != nil {
			t.Errorf("failed to clear identities: %v", err)
		}
	})

	r, err := store.Reader()
	if err != nil {
		t.Errorf("failed to read identities: %v", err)
		return
	}

	ids, err := age.ParseIdentities(r)
	if err != nil {
		t.Errorf("failed to parse identities: %v", err)
		return
	}

	if len(ids) < 1 {
		t.Errorf("no identities read")
	}
}
