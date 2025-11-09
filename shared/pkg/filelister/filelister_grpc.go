package filelister

import (
	"context"

	"github.com/bmj2728/hst/shared/pkg/hostconn"
	"github.com/bmj2728/hst/shared/pkg/hostserve"
	filelisterv1 "github.com/bmj2728/hst/shared/protogen/filelister/v1"
	"github.com/bmj2728/hst/shared/protogen/hostserve/v1"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

type GRPCServer struct {
	Impl FileLister
	filelisterv1.UnimplementedFileListerServer
}

func (s *GRPCServer) List(ctx context.Context,
	request *filelisterv1.FileListRequest) (*filelisterv1.FileListResponse, error) {

	entries, err := s.Impl.ListFiles(request.Dir)
	if err != nil {
		errMsg := err.Error()
		return &filelisterv1.FileListResponse{
			Entry: nil,
			Error: &errMsg,
		}, nil
	}
	return &filelisterv1.FileListResponse{
		Entry: entries,
		Error: nil,
	}, nil
}

func (s *GRPCServer) EstablishHostServices(ctx context.Context,
	request *filelisterv1.HostServiceRequest) (*filelisterv1.Empty, error) {

	// Only call EstablishHostServices if the plugin implements HostConnection
	if hostConn, ok := s.Impl.(hostconn.HostConnection); ok {
		hostConn.EstablishHostServices(request.HostService)
	}
	// If plugin doesn't implement HostConnection, silently succeed
	// (plugin doesn't need host services)
	return &filelisterv1.Empty{}, nil
}

// GRPCClient is the client side of the plugin.
// It implements plugin.GRPCPlugin so the plugin framework can communicate with it.
type GRPCClient struct {
	client        filelisterv1.FileListerClient
	broker        *plugin.GRPCBroker
	hostServiceID uint32
}

// SetBroker sets the gRPC broker for the client.
// This method is a no-op on the host side since the broker is already initialized in the constructor.
// It is implemented to fulfill the HostConnection interface requirements.
func (c *GRPCClient) SetBroker(broker *plugin.GRPCBroker) {
	// No-op on the host side - the broker is already set during construction
	// This method exists to satisfy the HostConnection interface
	c.broker = broker
}

// EstablishHostServices sets the host service ID and notifies the plugin via gRPC to establish the host service.
func (c *GRPCClient) EstablishHostServices(hostServiceID uint32) {
	c.hostServiceID = hostServiceID
	// Also call the gRPC method to notify the plugin
	_, _ = c.client.EstablishHostServices(context.Background(), &filelisterv1.HostServiceRequest{
		HostService: hostServiceID,
	})
}

// DisconnectHostServices performs cleanup actions during plugin shutdown, though no client-side cleanup is
// needed currently.
func (c *GRPCClient) DisconnectHostServices() {
	// The host manages its own server lifecycle
	// This is called during plugin shutdown to do any cleanup
	// Currently no cleanup needed on the client side
}

// ListFiles retrieves the list of files in the specified directory on the remote host using the gRPC client.
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

// RegisterHostService registers a host service with the broker and returns its service ID.
// This allows plugins to dial back to host services for bidirectional communication.
// Implements hostconn.HostServiceRegistrar interface.
func (c *GRPCClient) RegisterHostService(hostServices hostserve.IHostServices) (uint32, error) {
	// Allocate a unique ID for this service using the broker's built-in ID allocator
	serviceID := c.broker.NextId()

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

// FileListerError represents an error returned by the file listing service.
// It contains a message describing the error.
type FileListerError struct {
	Message string
}

// Error returns the error message stored in the FileListerError instance.
func (e *FileListerError) Error() string {
	return e.Message
}
