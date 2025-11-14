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

// ClientID represents a unique identifier for a client in a system or application.
type ClientID string

func newClientID() ClientID {
	return ClientID(uuid.New().String())
}

// String returns the ClientID as its underlying string representation.
func (cid ClientID) String() string {
	return string(cid)
}

// RequestID represents a unique identifier for a specific request, typically used for tracing and tracking purposes.
type RequestID string

func NewRequestID() RequestID {
	return RequestID(uuid.New().String())
}

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
