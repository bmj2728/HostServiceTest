package filelister

import (
	"context"

	filelisterv1 "github.com/bmj2728/hst/shared/protogen/filelister/v1"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// FileLister is the business interface for file listing plugins
type FileLister interface {
	EstablishHostServices(hostServiceID uint32)
	DisconnectHostServices()
	ListFiles(dir string) ([]string, error)
}

// BrokerAware is an optional interface that plugin implementations can satisfy
// if they need access to the gRPC broker for bidirectional communication
type BrokerAware interface {
	SetBroker(broker *plugin.GRPCBroker)
}

type FileListerGRPCPlugin struct {
	plugin.Plugin
	Impl FileLister
}

func (fl *FileListerGRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	// If the implementation needs the broker for bidirectional communication, provide it
	if brokerAware, ok := fl.Impl.(BrokerAware); ok {
		brokerAware.SetBroker(broker)
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
