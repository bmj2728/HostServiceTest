package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/bmj2728/hst/shared/pkg/hostdemo"
	"github.com/bmj2728/hst/shared/pkg/hostserve"
	hostservev1 "github.com/bmj2728/hst/shared/protogen/hostserve/v1"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

type HostDemo struct {
	broker            *plugin.GRPCBroker
	hostServiceClient hostserve.IHostServices
	conn              *grpc.ClientConn
	connMutex         sync.Mutex
}

func (h *HostDemo) GetEnvDemo(ctx context.Context, env string) (string, error) {
	val, err := h.hostServiceClient.GetEnv(ctx, env)
	if err != nil {
		return "", fmt.Errorf("failed to get env: %w", err)
	}
	response := fmt.Sprintf("***GetEnv Demo***\n\nRequested Env Var: '%s'\nValue: '%s'", env, val)

	return response, nil
}

func (h *HostDemo) EnvDemo(ctx context.Context) (string, error) {
	uid, err := h.hostServiceClient.Getuid(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get uid: %w", err)
	}
	gid, err := h.hostServiceClient.Getgid(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get gid: %w", err)
	}
	euid, err := h.hostServiceClient.Geteuid(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get egid: %w", err)
	}
	egid, err := h.hostServiceClient.Getegid(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get egid: %w", err)
	}
	groups, err := h.hostServiceClient.GetGroups(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get groups: %w", err)
	}
	pid, err := h.hostServiceClient.Getpid(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get pid: %w", err)
	}
	ppid, err := h.hostServiceClient.Getppid(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get ppid: %w", err)
	}
	td, err := h.hostServiceClient.TempDir(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get temp dir: %w", err)
	}
	uCache, err := h.hostServiceClient.UserCacheDir(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get user cache: %w", err)
	}
	uConfig, err := h.hostServiceClient.UserConfigDir(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get user config dir: %w", err)
	}
	uHome, err := h.hostServiceClient.UserHomeDir(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get user home dir: %w", err)
	}
	groupStr := ""
	for i, group := range groups {
		if i == 0 {
			groupStr += "group"
		} else {
			groupStr += fmt.Sprintf(" |  %d", group)
		}
	}
	response := fmt.Sprintf("***Env Demo***\n\nUID: %d\nGID: %d\nEUID: %d\nEGID: %d\nGroups: %s\nPID: %d\nPPID: %dTemp Dir: %s\nUser Cache: %s\nUser Config: %s\nUser Home: %s\n",
		uid, gid, euid, egid, groupStr, pid, ppid, td, uCache, uConfig, uHome)
	return response, nil
}

func (h *HostDemo) EstablishHostServices(hostServiceID uint32) (hostserve.ClientID, error) {
	h.connMutex.Lock()
	defer h.connMutex.Unlock()

	conn, err := h.broker.Dial(hostServiceID)
	if err != nil {
		hclog.Default().Error("Failed to dial host service", "err", err)
		return "", fmt.Errorf("failed to dial broker: %w", err)
	}

	h.conn = conn
	client := hostserve.NewHostServiceGRPCClient(hostservev1.NewHostServiceClient(conn))
	h.hostServiceClient = client
	return client.ClientID(), nil
}

func (h *HostDemo) DisconnectHostServices() {
	h.connMutex.Lock()
	defer h.connMutex.Unlock()

	if h.conn != nil {
		if err := h.conn.Close(); err != nil {
			hclog.Default().Error("Failed to close connection", "err", err)
		}
		h.conn = nil
		h.hostServiceClient = nil
		hclog.Default().Info("Disconnected from host services")
	}
}

func (h *HostDemo) SetBroker(broker *plugin.GRPCBroker) {
	h.broker = broker
}

var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "TEST_KEY",
	MagicCookieValue: "TEST_VALUE",
}

func main() {
	hd := &HostDemo{}

	pluginMap := map[string]plugin.Plugin{
		"hd-plugin": &hostdemo.HostDemoGRPCPlugin{Impl: hd},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
		GRPCServer:      plugin.DefaultGRPCServer,
	})
}
