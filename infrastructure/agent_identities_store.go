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

	"github.com/prskr/git-age/core/ports"
)

var (
	_ ports.IdentitiesStore = (*AgentIdentitiesStore)(nil)
	_ identityStoreSource   = (*AgentIdentitiesStoreSource)(nil)
)

func NewAgentIdentitiesStoreSource(env ports.OSEnv) *AgentIdentitiesStoreSource {
	return &AgentIdentitiesStoreSource{
		BaseURL: os.ExpandEnv(env.Get("GIT_AGE_AGENT_HOST")),
	}
}

type AgentIdentitiesStoreSource struct {
	BaseURL string
	Client  connect.HTTPClient
}

func (a *AgentIdentitiesStoreSource) IsValid(ctx context.Context) (isValid bool, err error) {
	if a.BaseURL == "" {
		slog.DebugContext(ctx, "Skipping agent because url is empty")
		return false, nil
	}

	if a.Client == nil {
		a.BaseURL, a.Client, err = prepareClient(a.BaseURL)
		if err != nil {
			return false, err
		}
	}

	healthClient := healthv1connect.NewHealthClient(a.Client, a.BaseURL)
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
		IdentitiesClient: agentv1connect.NewIdentitiesStoreServiceClient(a.Client, a.BaseURL),
	}, nil
}

type AgentIdentitiesStore struct {
	IdentitiesClient agentv1connect.IdentitiesStoreServiceClient
}

func (a AgentIdentitiesStore) Generate(
	ctx context.Context,
	cmd ports.GenerateIdentityCommand,
) (publicKey string, err error) {
	newID, err := cmd.Algorithm.Generate()
	if err != nil {
		return "", err
	}

	if cmd.Comment == "" {
		cmd.Comment = "Generated on " + time.Now().Format(time.RFC3339)
	}

	publicKey = newID.Recipient().String()

	req := &agentv1.StoreIdentityRequest{
		PublicKey:  publicKey,
		PrivateKey: newID.String(),
		Comment:    cmd.Comment,
		Remote:     cmd.Remote,
	}

	if _, err = a.IdentitiesClient.StoreIdentity(ctx, connect.NewRequest(req)); err != nil {
		return "", err
	}

	return publicKey, nil
}

func (a AgentIdentitiesStore) Identities(ctx context.Context, query ports.IdentitiesQuery) ([]age.Identity, error) {
	req := &agentv1.GetIdentitiesRequest{
		Remotes: query.Remotes,
	}

	slog.DebugContext(
		ctx,
		"Fetching identities from remote agent",
		slog.String("remotes", strings.Join(query.Remotes, ",")),
	)

	resp, err := a.IdentitiesClient.GetIdentities(ctx, connect.NewRequest(req))
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

func prepareClient(rawUrl string) (baseUrl string, client *http.Client, err error) {
	const unixScheme = "unix"
	parsed, err := url.Parse(rawUrl)
	if err != nil {
		return "", nil, err
	}

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

	if parsed.Scheme == unixScheme {
		slog.Debug("Trying to connect to unix socket", slog.String("path", parsed.Path))

		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			unescaped, err := url.PathUnescape(parsed.Path)
			if err != nil {
				return nil, err
			}
			return dialer.DialContext(ctx, unixScheme, unescaped)
		}

		return "http://localhost", &http.Client{
			Transport: transport,
		}, nil
	} else {
		transport.DialContext = dialer.DialContext
	}

	return rawUrl, &http.Client{
		Transport: transport,
	}, nil
}
