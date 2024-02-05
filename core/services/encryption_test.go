package services_test

import (
	"testing"

	"github.com/prskr/git-age/internal/testx"

	"filippo.io/age"
	"github.com/prskr/git-age/core/services"
)

func TestAgeSealer_AddRecipients(t *testing.T) {
	t.Parallel()
	type args struct {
		r func(tb testing.TB) []age.Recipient
	}
	tests := []struct {
		name        string
		args        args
		wantCanSeal bool
	}{
		{
			name:        "Empty recipients",
			wantCanSeal: false,
		},
		{
			name: "Non-empty recipients",
			args: args{
				r: func(tb testing.TB) []age.Recipient {
					tb.Helper()

					return []age.Recipient{
						testx.ResultOf(t, age.GenerateX25519Identity).Recipient(),
					}
				},
			},
			wantCanSeal: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h, err := services.NewAgeSealer()
			if err != nil {
				t.Errorf("NewAgeSealer() error = %v", err)
				return
			}

			var recipientsToAdd []age.Recipient

			if tt.args.r != nil {
				recipientsToAdd = tt.args.r(t)
			}

			h.AddRecipients(recipientsToAdd...)

			if got := h.CanSeal(); got != tt.wantCanSeal {
				t.Errorf("AgeSealer.CanSeal() = %v, want %v", got, tt.wantCanSeal)
			}
		})
	}
}

func TestAgeSealer_AddIdentities(t *testing.T) {
	t.Parallel()
	type args struct {
		identities func(tb testing.TB) []age.Identity
	}
	tests := []struct {
		name        string
		args        args
		wantCanOpen bool
	}{
		{
			name:        "Empty identities",
			wantCanOpen: false,
		},
		{
			name: "Non-empty identities",
			args: args{
				identities: func(tb testing.TB) []age.Identity {
					tb.Helper()

					return []age.Identity{
						testx.ResultOf(t, age.GenerateX25519Identity),
					}
				},
			},
			wantCanOpen: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h, err := services.NewAgeSealer()
			if err != nil {
				t.Errorf("NewAgeSealer() error = %v", err)
				return
			}

			var identitiesToAdd []age.Identity

			if tt.args.identities != nil {
				identitiesToAdd = tt.args.identities(t)
			}

			h.AddIdentities(identitiesToAdd...)

			if got := h.CanOpen(); got != tt.wantCanOpen {
				t.Errorf("AgeSealer.CanOpen() = %v, want %v", got, tt.wantCanOpen)
			}
		})
	}
}
