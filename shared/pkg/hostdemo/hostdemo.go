package hostdemo

import (
	"context"

	"github.com/bmj2728/hst/shared/pkg/hostconn"
	hostdemov1 "github.com/bmj2728/hst/shared/protogen/hostdemo/v1"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

type HostDemo interface {
	GetEnvDemo(key string) (string, error)
	EnvDemo() (string, error)
	TempDemo(pattern, textToWrite string) (string, error)
}

type HostDemoGRPCPlugin struct {
	plugin.Plugin
	Impl HostDemo
}

func (hd *HostDemoGRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	// If the plugin's implementation implements HostConnection, set the broker
	if hostConn, ok := hd.Impl.(hostconn.HostConnection); ok {
		hostConn.SetBroker(broker)
	}
	hostdemov1.RegisterHostDemoServer(s, &GRPCServer{Impl: hd.Impl})
	return nil
}

func (hd *HostDemoGRPCPlugin) GRPCClient(ctx context.Context,
	broker *plugin.GRPCBroker,
	c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{
		client: hostdemov1.NewHostDemoClient(c),
		broker: broker,
	}, nil
}
