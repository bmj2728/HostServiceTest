package main

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/bmj2728/hst/shared/pkg/filelister"
	"github.com/bmj2728/hst/shared/pkg/hostserve"
	hostservev1 "github.com/bmj2728/hst/shared/protogen/hostserve/v1"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

type FileLister struct {
	broker            *plugin.GRPCBroker
	hostServiceClient hostserve.IHostServices
	conn              *grpc.ClientConn
	connMutex         sync.Mutex
}

func (f *FileLister) ListFiles(dir string) ([]string, error) {
	ctx := context.Background()
	home := f.hostServiceClient.GetEnv(ctx, "HOME")
	dirEntries, err := f.hostServiceClient.ReadDir(ctx, dir)
	if err != nil {
		hclog.Default().Error("Failed to read directory via host service", "dir", dir, "err", err)
		return nil, err
	}

	var entries []string
	var buf bytes.Buffer
	entries = append(entries, home)
	for _, entry := range dirEntries {
		if entry.IsDir() {
			entries = append(entries, entry.Name())
			buf.WriteString(entry.Name())
		} else {
			entries = append(entries, entry.Name())
			buf.WriteString(entry.Name())
		}
	}

	err = f.hostServiceClient.WriteFile(ctx, filepath.Join(dir, "listed_files.txt"), buf.Bytes(), 0644)
	if err != nil {
		hclog.Default().Error("Failed to write file via host service", "dir", dir, "err", err)
	}
	return entries, nil
}

func (f *FileLister) EstablishHostServices(hostServiceID uint32) (string, error) {
	f.connMutex.Lock()
	defer f.connMutex.Unlock()

	conn, err := f.broker.Dial(hostServiceID)
	if err != nil {
		hclog.Default().Error("Failed to dial host service", "err", err)
		return "", fmt.Errorf("failed to dial broker: %w", err)
	}

	f.conn = conn
	client := hostserve.NewHostServiceGRPCClient(hostservev1.NewHostServiceClient(conn))
	f.hostServiceClient = client
	hclog.Default().Info("Established host services", "id", hostServiceID, "clientID", client.ClientID())
	return client.ClientID().String(), nil
}

func (f *FileLister) DisconnectHostServices() {
	f.connMutex.Lock()
	defer f.connMutex.Unlock()

	if f.conn != nil {
		if err := f.conn.Close(); err != nil {
			hclog.Default().Error("Failed to close connection", "err", err)
		}
		f.conn = nil
		f.hostServiceClient = nil
		hclog.Default().Info("Disconnected from host services")
	}
}

func (f *FileLister) SetBroker(broker *plugin.GRPCBroker) {
	f.broker = broker
}

var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "TEST_KEY",
	MagicCookieValue: "TEST_VALUE",
}

func main() {
	fl := &FileLister{}

	pluginMap := map[string]plugin.Plugin{
		"fl-plugin": &filelister.FileListerGRPCPlugin{Impl: fl},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
		GRPCServer:      plugin.DefaultGRPCServer,
	})
}
