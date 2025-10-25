package main

import (
	"sync"

	"github.com/bmj2728/hst/shared/pkg/filelister"
	"github.com/bmj2728/hst/shared/pkg/hostserve"
	hostservev1 "github.com/bmj2728/hst/shared/protogen/hostserve/v1"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

type FileLister struct {
	broker       *plugin.GRPCBroker
	hostServices uint32
	connections  []*grpc.ClientConn // Track connections for cleanup
	connMutex    sync.Mutex
}

func (f *FileLister) ListFiles(dir string, hostService uint32) ([]string, error) {
	// Connect to the host service using the broker
	conn, err := f.broker.Dial(hostService)
	if err != nil {
		hclog.Default().Error("Failed to dial host service", "err", err)
		return nil, err
	}

	// Track this connection for cleanup
	f.connMutex.Lock()
	f.connections = append(f.connections, conn)
	f.connMutex.Unlock()

	// Note: Not using defer conn.Close() here because we're tracking connections
	// They'll be closed in DisconnectHostServices()

	// Create host service client
	hsClient := hostserve.NewHostServiceGRPCClient(hostservev1.NewHostServiceClient(conn))

	// Call ReadDir via host service - returns []fs.DirEntry with name AND isDir
	dirEntries, err := hsClient.ReadDir(dir)
	if err != nil {
		hclog.Default().Error("Failed to read directory via host service", "dir", dir, "err", err)
		return nil, err
	}

	// Convert DirEntry to string names (plugin could filter/process based on isDir here)
	var entries []string
	for _, entry := range dirEntries {
		if entry.IsDir() {
			entries = append(entries, entry.Name()+"-d")
		} else {
			entries = append(entries, entry.Name()+"-f")
		}
	}

	return entries, nil
}

func (f *FileLister) EstablishHostServices(hostServiceID uint32) {
	// Store the host service ID provided by the host
	f.hostServices = hostServiceID
	hclog.Default().Info("Established host services", "id", hostServiceID)
}

func (f *FileLister) DisconnectHostServices() {
	// Close all connections to host services
	f.connMutex.Lock()
	defer f.connMutex.Unlock()

	hclog.Default().Info("Disconnecting from host services", "connection_count", len(f.connections))
	for _, conn := range f.connections {
		if err := conn.Close(); err != nil {
			hclog.Default().Error("Failed to close connection", "err", err)
		}
	}
	f.connections = nil
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
