package filelister

import (
	"context"
	"fmt"

	"github.com/bmj2728/hst/shared/pkg/hostconn"
	"github.com/bmj2728/hst/shared/pkg/hostserve"
	filelisterv1 "github.com/bmj2728/hst/shared/protogen/filelister/v1"
	"github.com/bmj2728/hst/shared/protogen/hostserve/v1"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// GRPCServer implements the FileLister gRPC server and bridges the interface with gRPC request handlers.
type GRPCServer struct {
	Impl FileLister
	filelisterv1.UnimplementedFileListerServer
}

// List retrieves a list of file entries from a specified directory and returns them in a FileListResponse.
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

// EstablishHostServices sets up communication with host services if the plugin implements the HostConnection interface.
// If the plugin does not require host services, this method silently succeeds without taking further action.
// The method accepts a context and a HostServiceRequest containing the host service ID.
// Returns an Empty response and error if any issues occur during execution.
func (s *GRPCServer) EstablishHostServices(ctx context.Context,
	request *filelisterv1.HostServiceRequest) (*filelisterv1.HostServiceResponse, error) {

	if hostConn, ok := s.Impl.(hostconn.HostConnection); ok {
		clientID, err := hostConn.EstablishHostServices(request.HostService)
		if err != nil {
			return nil, fmt.Errorf("plugin failed to establish host services: %w", err)
		}
		return &filelisterv1.HostServiceResponse{ClientId: clientID}, nil
	}

	// Plugin doesn't implement HostConnection - not an error
	return &filelisterv1.HostServiceResponse{}, nil
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
func (c *GRPCClient) EstablishHostServices(hostServiceID uint32) (string, error) {
	c.hostServiceID = hostServiceID

	resp, err := c.client.EstablishHostServices(context.Background(),
		&filelisterv1.HostServiceRequest{
			HostService: hostServiceID,
		})
	if err != nil {
		return "", fmt.Errorf("gRPC call failed: %w", err) // CHANGED - return error with context
	}

	return resp.ClientId, nil
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
