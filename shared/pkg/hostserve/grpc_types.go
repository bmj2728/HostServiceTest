package hostserve

import (
	"github.com/bmj2728/hst/shared/protogen/hostserve/v1"
	"github.com/google/uuid"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////

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
	// Generate a unique client ID for this connection - we'll improve this later
	clientUUID, err := uuid.NewV7()
	if err != nil {
		return nil
	}
	return &HostServiceGRPCClient{
		client:   client,
		clientID: ClientID(clientUUID.String()),
	}
}

/////////////////////////////////////////////////////////////////////////////////////////////////////

// ClientID represents a unique identifier for a client in a system or application.
type ClientID string

// String returns the ClientID as its underlying string representation.
func (cid ClientID) String() string {
	return string(cid)
}

// RequestID represents a unique identifier for a specific request, typically used for tracing and tracking purposes.
type RequestID string

// String converts the RequestID value to its string representation.
func (rid RequestID) String() string {
	return string(rid)
}

// HostServiceError represents an error returned by the host service.
// Message is a description of the error.
type HostServiceError struct {
	Message string
}

// Error returns the error message stored in the HostServiceError as a string.
func (e *HostServiceError) Error() string {
	return e.Message
}
