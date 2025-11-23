package hostserve

import (
	"errors"

	"github.com/bmj2728/hst/shared/protogen/hostserve/v1"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////

var (
	ErrInvalidHostServices = errors.New("invalid host services")
)

// HostServiceGRPCServer provides a gRPC server implementation for host services using the IHostServices interface.
type HostServiceGRPCServer struct {
	Impl IHostServices
	hostservev1.UnimplementedHostServiceServer
}

// HostServiceGRPCClient wraps the filesystemv1.HostServiceClient to provide higher-level client methods.
type HostServiceGRPCClient struct {
	client   hostservev1.HostServiceClient
	clientID ClientID
}

// NewHostServiceGRPCClient creates a new instance of HostServiceGRPCClient wrapping the provided gRPC client.
func NewHostServiceGRPCClient(client hostservev1.HostServiceClient) *HostServiceGRPCClient {
	clientID := newClientID()
	return &HostServiceGRPCClient{
		client:   client,
		clientID: clientID,
	}
}

func (c *HostServiceGRPCClient) ClientID() ClientID {
	if c == nil {
		return ""
	}
	return c.clientID
}

/////////////////////////////////////////////////////////////////////////////////////////////////////

// HostServiceError represents an error returned by the host service.
// Message is a description of the error.
type HostServiceError struct {
	Message string
}

// Error returns the error message stored in the HostServiceError as a string.
func (e *HostServiceError) Error() string {
	return e.Message
}
