package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

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

func (h *HostDemo) GetEnvDemo(key string) (string, error) {
	start := time.Now()
	ctx := context.Background()
	val, err := h.hostServiceClient.GetEnv(ctx, key)
	if err != nil {
		return "", fmt.Errorf("failed to get env: %w", err)
	}
	duration := time.Since(start)
	response := fmt.Sprintf("***GetEnv Demo***\n\nRequested Env Var: '%s'\nValue: '%s'\nDuration: %v\n", key, val, duration)

	return response, nil
}

func (h *HostDemo) EnvDemo() (string, error) {
	ctx := context.Background()
	start := time.Now()
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
			groupStr += strconv.Itoa(int(group))
		} else {
			groupStr += fmt.Sprintf(" |  %d", group)
		}
	}
	duration := time.Since(start)
	response := fmt.Sprintf("***Env Demo***\n\nUID: %d\nGID: %d\nEUID: %d\nEGID: %d\nGroups: %s\nPID: %d\nPPID: %d\nTemp Dir: %s\nUser Cache: %s\nUser Config: %s\nUser Home: %s\nDuration: %v\n",
		uid, gid, euid, egid, groupStr, pid, ppid, td, uCache, uConfig, uHome, duration)
	return response, nil
}

func (h *HostDemo) TempDemo(pattern, textToWrite string) (string, error) {
	ctx := context.Background()
	var sb strings.Builder
	sb.WriteString("***Temp Demo***\n\n")
	sb.WriteString("Pattern: " + pattern + "\n")
	start := time.Now()
	rd, err := h.hostServiceClient.TempDir(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get temp dir: %w", err)
	}
	td, err := h.hostServiceClient.MkdirTemp(ctx, rd, pattern)
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	mdtStr := fmt.Sprintf("%v Created Temp Dir: %s\n", time.Since(start), td)
	sb.WriteString(mdtStr)

	tf, err := h.hostServiceClient.FileCreateTemp(ctx, td, pattern+".txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	rfi, err := h.hostServiceClient.FileStat(ctx, tf)
	if err != nil {
		return "", fmt.Errorf("failed to stat temp file: %w", err)
	}
	mdfStr := fmt.Sprintf("%v Created Temp File: %s - %s\n", time.Since(start), tf, rfi.Name())
	sb.WriteString(mdfStr)

	writer, err := h.hostServiceClient.FileWriter(ctx, tf)
	if err != nil {
		return "", fmt.Errorf("failed to create file writer: %w", err)
	}
	b, err := writer.Write([]byte(textToWrite))
	if err != nil {
		return "", fmt.Errorf("failed to write to temp file: %w", err)
	}
	err = writer.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close file writer: %w", err)
	}
	err = h.hostServiceClient.FileSync(ctx, tf)
	if err != nil {
		return "", fmt.Errorf("failed to sync temp file: %w", err)
	}
	rfi2, err := h.hostServiceClient.FileStat(ctx, tf)
	if err != nil {
		return "", fmt.Errorf("failed to stat temp file: %w", err)
	}

	wStr := fmt.Sprintf("%v Wrote %d bytes to File: %s - %s\nCurrent Size: %d\n", time.Since(start), b, tf, textToWrite, rfi2.Size())
	sb.WriteString(wStr)

	err = h.hostServiceClient.FileClose(ctx, tf)
	if err != nil {
		return "", fmt.Errorf("failed to close temp file: %w", err)
	}
	r, p := filepath.Split(td)
	err = h.hostServiceClient.RemoveAll(ctx, r, p)
	if err != nil {
		return "", fmt.Errorf("failed to remove temp dir: %w", err)
	}

	duration := time.Since(start)
	fStr := fmt.Sprintf("Cleaned up temp files.\nDuration: %v\n", duration)
	sb.WriteString(fStr)
	return sb.String(), nil
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
