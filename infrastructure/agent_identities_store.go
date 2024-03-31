package infrastructure

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"buf.build/gen/go/git-age/agent/connectrpc/go/agent/v1/agentv1connect"
	agentv1 "buf.build/gen/go/git-age/agent/protocolbuffers/go/agent/v1"
	"buf.build/gen/go/grpc/grpc/connectrpc/go/grpc/health/v1/healthv1connect"
	healthv1 "buf.build/gen/go/grpc/grpc/protocolbuffers/go/grpc/health/v1"
	"connectrpc.com/connect"
	"filippo.io/age"

	"github.com/prskr/git-age/core/dto"
	"github.com/prskr/git-age/core/ports"
)

var (
	_ ports.IdentitiesStore = (*AgentIdentitiesStore)(nil)
	_ identityStoreSource   = (*AgentIdentitiesStoreSource)(nil)
)

func NewAgentIdentitiesStoreSource() *AgentIdentitiesStoreSource {
	return &AgentIdentitiesStoreSource{
		BaseUrl: os.ExpandEnv(os.Getenv("GIT_AGE_AGENT_HOST")),
	}
}

type AgentIdentitiesStoreSource struct {
	BaseUrl string
	Client  connect.HTTPClient
}

func (a *AgentIdentitiesStoreSource) IsValid(ctx context.Context) (bool, error) {
	if a.BaseUrl == "" {
		return false, nil
	}

	if a.Client == nil {
		a.BaseUrl, a.Client = prepareClient(a.BaseUrl)
	}

	healthClient := healthv1connect.NewHealthClient(a.Client, a.BaseUrl)
	healthRequest := &healthv1.HealthCheckRequest{Service: agentv1connect.IdentitiesStoreServiceName}
	resp, err := healthClient.Check(ctx, connect.NewRequest(healthRequest))
	if err != nil {
		return false, err
	} else if resp.Msg.Status != healthv1.HealthCheckResponse_SERVING {
		slog.Info("agent health check failed", slog.String("status", resp.Msg.Status.String()))
		return false, nil
	}

	return true, nil
}

func (a *AgentIdentitiesStoreSource) GetStore() (ports.IdentitiesStore, error) {
	return &AgentIdentitiesStore{
		identitiesClient: agentv1connect.NewIdentitiesStoreServiceClient(a.Client, a.BaseUrl),
	}, nil
}

type AgentIdentitiesStore struct {
	identitiesClient agentv1connect.IdentitiesStoreServiceClient
}

func (a AgentIdentitiesStore) Generate(
	ctx context.Context,
	cmd dto.GenerateIdentityCommand,
) (publicKey string, err error) {
	newId, err := age.GenerateX25519Identity()
	if err != nil {
		return "", err
	}

	if cmd.Comment == "" {
		cmd.Comment = "Generated on " + time.Now().Format(time.RFC3339)
	}

	publicKey = newId.Recipient().String()

	req := &agentv1.StoreIdentityRequest{
		PublicKey:  publicKey,
		PrivateKey: newId.String(),
		Comment:    cmd.Comment,
		Remote:     cmd.Remote,
	}

	if _, err = a.identitiesClient.StoreIdentity(ctx, connect.NewRequest(req)); err != nil {
		return "", err
	}

	return publicKey, nil
}

func (a AgentIdentitiesStore) Identities(ctx context.Context, query dto.IdentitiesQuery) ([]age.Identity, error) {
	req := &agentv1.GetIdentitiesRequest{
		Remotes: query.Remotes,
	}

	resp, err := a.identitiesClient.GetIdentities(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}

	ids := make([]age.Identity, 0, len(resp.Msg.Keys))
	for _, raw := range resp.Msg.Keys {
		id, err := age.ParseX25519Identity(raw)
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}

func prepareClient(rawUrl string) (baseUrl string, client *http.Client) {
	const unixNetwork = "unix://"
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	if strings.HasPrefix(rawUrl, unixNetwork) {
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			path := rawUrl[len(unixNetwork):]
			unescaped, err := url.PathUnescape(path)
			if err != nil {
				return nil, err
			}
			return dialer.DialContext(ctx, "unix", unescaped)
		}

		return "http://localhost", &http.Client{
			Transport: transport,
		}
	} else {
		transport.DialContext = dialer.DialContext
	}

	return rawUrl, &http.Client{
		Transport: transport,
	}
}
