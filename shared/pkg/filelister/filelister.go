package filelister

import (
	"context"
	"sync/atomic"

	"github.com/bmj2728/hst/shared/pkg/hostserve"
	filelisterv1 "github.com/bmj2728/hst/shared/protogen/filelister/v1"
	hostservev1 "github.com/bmj2728/hst/shared/protogen/hostserve/v1"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

var brokerIDCounter uint32 = 0

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

type GRPCClient struct {
	client        filelisterv1.FileListerClient
	broker        *plugin.GRPCBroker
	hostServiceID uint32
}

func (c *GRPCClient) EstablishHostServices(hostServiceID uint32) {
	c.hostServiceID = hostServiceID
	// Also call the gRPC method to notify the plugin
	_, _ = c.client.EstablishHostServices(context.Background(), &filelisterv1.HostServiceRequest{
		HostService: hostServiceID,
	})
}

func (c *GRPCClient) DisconnectHostServices() {
	// The host manages its own server lifecycle
	// This is called during plugin shutdown to do any cleanup
	// Currently no cleanup needed on the client side
}

func (c *GRPCClient) ListFiles(dir string) ([]string, error) {
	resp, err := c.client.List(context.Background(), &filelisterv1.FileListRequest{
		Dir:         dir,
		HostService: c.hostServiceID,
	})
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, &FileListerError{Message: *resp.Error}
	}
	return resp.Entry, nil
}

// SetupHostService registers a host service with the broker and returns its service ID.
// This allows plugins to dial back to host services for bidirectional communication.
func (c *GRPCClient) SetupHostService(hostServices hostserve.IHostServices) (uint32, error) {
	// Allocate a unique ID for this service using atomic counter
	serviceID := atomic.AddUint32(&brokerIDCounter, 1)

	// Start a gRPC server for the host service via the broker at the allocated ID
	go c.broker.AcceptAndServe(serviceID, func(opts []grpc.ServerOption) *grpc.Server {
		server := grpc.NewServer(opts...)
		hostservev1.RegisterHostServiceServer(server, &hostserve.HostServiceGRPCServer{
			Impl: hostServices,
		})
		return server
	})

	return serviceID, nil
}

type FileListerError struct {
	Message string
}

func (e *FileListerError) Error() string {
	return e.Message
}
