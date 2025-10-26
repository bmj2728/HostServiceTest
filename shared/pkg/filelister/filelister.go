package filelister

import (
	"context"

	"github.com/bmj2728/hst/shared/pkg/hostconn"
	filelisterv1 "github.com/bmj2728/hst/shared/protogen/filelister/v1"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// FileLister is the business interface for file listing plugins.
// This interface contains only the core business logic methods.
type FileLister interface {
	ListFiles(dir string) ([]string, error)
}

type FileListerGRPCPlugin struct {
	plugin.Plugin
	Impl FileLister
}

func (fl *FileListerGRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	// If the plugin's implementation implements HostConnection, set the broker
	if hostConn, ok := fl.Impl.(hostconn.HostConnection); ok {
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
