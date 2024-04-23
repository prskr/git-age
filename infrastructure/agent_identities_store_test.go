package infrastructure_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"buf.build/gen/go/git-age/agent/connectrpc/go/agent/v1/agentv1connect"
	agentv1 "buf.build/gen/go/git-age/agent/protocolbuffers/go/agent/v1"
	"connectrpc.com/connect"
	"connectrpc.com/grpchealth"
	"filippo.io/age"

	"github.com/prskr/git-age/core/dto"
	"github.com/prskr/git-age/infrastructure"
	"github.com/prskr/git-age/internal/testx"
)

func TestAgentIdentitiesStore_Generate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		onGenerate func(
			ctx context.Context,
			c *connect.Request[agentv1.StoreIdentityRequest],
		) (*connect.Response[agentv1.StoreIdentityResponse], error)
		wantErr bool
	}{
		{
			name: "Success",
			onGenerate: func(
				context.Context,
				*connect.Request[agentv1.StoreIdentityRequest],
			) (*connect.Response[agentv1.StoreIdentityResponse], error) {
				return connect.NewResponse(new(agentv1.StoreIdentityResponse)), nil
			},
		},
		{
			name: "Fail",
			onGenerate: func(
				context.Context,
				*connect.Request[agentv1.StoreIdentityRequest],
			) (*connect.Response[agentv1.StoreIdentityResponse], error) {
				return nil, errors.New("fail")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mux := http.NewServeMux()
			mux.Handle(grpchealth.NewHandler(grpchealth.NewStaticChecker(agentv1connect.IdentitiesStoreServiceName)))
			storeMock := &AgentStoreMock{
				OnGenerate: tt.onGenerate,
			}

			mux.Handle(agentv1connect.NewIdentitiesStoreServiceHandler(storeMock))
			server := httptest.NewServer(mux)
			t.Cleanup(server.Close)

			storeSource := infrastructure.AgentIdentitiesStoreSource{
				BaseUrl: server.URL,
				Client:  server.Client(),
			}

			if valid, err := storeSource.IsValid(testx.Context(t)); err != nil {
				t.Errorf("failed to validate identities store source: %v", err)
				return
			} else if !valid {
				t.Errorf("identities store source is not valid")
				return
			}

			identitiesStore, err := storeSource.GetStore()
			if err != nil {
				t.Errorf("failed to generate identities store: %v", err)
				return
			}

			pubKey, err := identitiesStore.Generate(testx.Context(t), dto.GenerateIdentityCommand{})
			if (err != nil) != tt.wantErr {
				t.Errorf("failed to generate identity store pubkey: %v", err)
			}

			if err == nil && pubKey == "" {
				t.Errorf("expected non-empty pubkey")
			}
		})
	}
}

func TestAgentIdentitiesStore_Identities(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args struct {
			remotes []string
		}
		onGetIdentities func(
			ctx context.Context,
			c *connect.Request[agentv1.GetIdentitiesRequest],
		) (*connect.Response[agentv1.GetIdentitiesResponse], error)
		wantIdentities int
		wantErr        bool
	}{
		{
			name: "Empty keystore",
			onGetIdentities: func(
				context.Context,
				*connect.Request[agentv1.GetIdentitiesRequest],
			) (*connect.Response[agentv1.GetIdentitiesResponse], error) {
				return connect.NewResponse(new(agentv1.GetIdentitiesResponse)), nil
			},
		},
		{
			name: "Invalid identity",
			onGetIdentities: func(
				context.Context,
				*connect.Request[agentv1.GetIdentitiesRequest],
			) (*connect.Response[agentv1.GetIdentitiesResponse], error) {
				return connect.NewResponse(&agentv1.GetIdentitiesResponse{
					Keys: []string{"hello, world!"},
				}), nil
			},
			wantErr: true,
		},
		{
			name: "Valid identity",
			onGetIdentities: func(
				context.Context,
				*connect.Request[agentv1.GetIdentitiesRequest],
			) (*connect.Response[agentv1.GetIdentitiesResponse], error) {
				id, err := age.GenerateX25519Identity()
				if err != nil {
					return nil, err
				}
				return connect.NewResponse(&agentv1.GetIdentitiesResponse{
					Keys: []string{id.String()},
				}), nil
			},
			wantIdentities: 1,
		},
		{
			name: "Fail",
			onGetIdentities: func(
				context.Context,
				*connect.Request[agentv1.GetIdentitiesRequest],
			) (*connect.Response[agentv1.GetIdentitiesResponse], error) {
				return nil, errors.New("some error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mux := http.NewServeMux()
			mux.Handle(grpchealth.NewHandler(grpchealth.NewStaticChecker(agentv1connect.IdentitiesStoreServiceName)))
			storeMock := &AgentStoreMock{
				OnGetIdentities: tt.onGetIdentities,
			}

			mux.Handle(agentv1connect.NewIdentitiesStoreServiceHandler(storeMock))
			server := httptest.NewServer(mux)
			t.Cleanup(server.Close)

			storeSource := infrastructure.AgentIdentitiesStoreSource{
				BaseUrl: server.URL,
				Client:  server.Client(),
			}

			if valid, err := storeSource.IsValid(testx.Context(t)); err != nil {
				t.Errorf("failed to validate identities store source: %v", err)
				return
			} else if !valid {
				t.Errorf("identities store source is not valid")
				return
			}

			identitiesStore, err := storeSource.GetStore()
			if err != nil {
				t.Errorf("failed to generate identities store: %v", err)
				return
			}

			query := dto.IdentitiesQuery{Remotes: tt.args.remotes}
			identities, err := identitiesStore.Identities(testx.Context(t), query)
			if (err != nil) != tt.wantErr {
				t.Errorf("failed to generate identity store pubkey: %v", err)
			}

			if len(identities) != tt.wantIdentities {
				t.Logf("expected %d identities, got %d", tt.wantIdentities, len(identities))
			}
		})
	}
}

var _ agentv1connect.IdentitiesStoreServiceHandler = (*AgentStoreMock)(nil)

//nolint:lll // doesn't make sense to break type in struct
type AgentStoreMock struct {
	OnGenerate      func(ctx context.Context, c *connect.Request[agentv1.StoreIdentityRequest]) (*connect.Response[agentv1.StoreIdentityResponse], error)
	OnGetIdentities func(ctx context.Context, c *connect.Request[agentv1.GetIdentitiesRequest]) (*connect.Response[agentv1.GetIdentitiesResponse], error)
}

func (a AgentStoreMock) GetIdentities(
	ctx context.Context,
	c *connect.Request[agentv1.GetIdentitiesRequest],
) (*connect.Response[agentv1.GetIdentitiesResponse], error) {
	if a.OnGetIdentities != nil {
		return a.OnGetIdentities(ctx, c)
	}
	return nil, connect.NewError(connect.CodeInternal, errors.New("no mock configured"))
}

func (a AgentStoreMock) StoreIdentity(
	ctx context.Context,
	c *connect.Request[agentv1.StoreIdentityRequest],
) (*connect.Response[agentv1.StoreIdentityResponse], error) {
	if a.OnGenerate != nil {
		return a.OnGenerate(ctx, c)
	}
	return nil, connect.NewError(connect.CodeInternal, errors.New("no mock configured"))
}
