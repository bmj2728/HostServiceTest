package filelister

import (
	"context"

	filelisterv1 "github.com/bmj2728/hst/shared/protogen/filelister/v1"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// FileLister is the business interface for file listing plugins.
// This interface contains only the core business logic methods.
type FileLister interface {
	ListFiles(dir string) ([]string, error)
}

// HostConnection handles bidirectional communication with host services.
// Plugin implementations should implement this interface if they need
// to call back to the host for privileged operations or shared services.
//
// This interface separates infrastructure concerns (connection management)
// from business logic (FileLister interface), making it optional for
// plugins that don't require host services.
type HostConnection interface {
	// SetBroker provides the gRPC broker for bidirectional communication
	SetBroker(broker *plugin.GRPCBroker)

	// EstablishHostServices receives the broker service ID for host services
	EstablishHostServices(hostServiceID uint32)

	// DisconnectHostServices cleans up connections to host services
	DisconnectHostServices()
}

type FileListerGRPCPlugin struct {
	plugin.Plugin
	Impl FileLister
}

func (fl *FileListerGRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	// If the implementation needs host services, provide the broker
	if hostConn, ok := fl.Impl.(HostConnection); ok {
		hostConn.SetBroker(broker)
	}
	filelisterv1.RegisterFileListerServer(s, &GRPCServer{Impl: fl.Impl})
	return nil
}

func (fl *FileListerGRPCPlugin) GRPCClient(ctx context.Context,
	broker *plugin.GRPCBroker,
	c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{
		client: filelisterv1.NewFileListerClient(c),
		broker: broker,
	}, nil
}
