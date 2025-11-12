package hostserve

import (
	"context"
	"io/fs"
	"time"

	"github.com/bmj2728/hst/shared/protogen/hostserve/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
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

// ctxClientIDKey is the context key used to store the client identifier in a context for outgoing requests.
const ctxClientIDKey = "client"

// addClientIDToContext attaches the specified clientID to the outgoing context metadata for gRPC requests.
func addClientIDToContext(ctx context.Context, clientID ClientID) context.Context {
	return metadata.AppendToOutgoingContext(ctx, ctxClientIDKey, clientID.String())
}

// getClientIDFromContext extracts the client ID from the provided gRPC context and returns it as a string.
// Returns an empty string if no client ID is found or the metadata is unavailable.
func getClientIDFromContext(ctx context.Context) ClientID {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	clientID := md.Get(ctxClientIDKey)
	if len(clientID) == 0 {
		return ""
	}
	return ClientID(clientID[0])
}

// RemoteDirEntry implements fs.DirEntry, this wrapper allows conversion from protobuf to fs.DirEntry
type RemoteDirEntry struct {
	name  string
	isDir bool
}

// Name returns the name of the directory entry as a string.
func (e *RemoteDirEntry) Name() string {
	return e.name
}

// IsDir reports whether the given RemoteDirEntry represents a directory.
func (e *RemoteDirEntry) IsDir() bool {
	return e.isDir
}

// Type returns the fs.FileMode for the remote directory entry, indicating if it represents a directory or a file.
func (e *RemoteDirEntry) Type() fs.FileMode {
	if e.isDir {
		return fs.ModeDir
	}
	return 0
}

// Info returns a fs.FileInfo for the remote directory entry. As full FileInfo is unavailable, it provides limited data.
func (e *RemoteDirEntry) Info() (fs.FileInfo, error) {
	// Remote entries don't have full FileInfo available
	return &RemoteFileInfo{
		name:  e.name,
		isDir: e.isDir,
	}, nil
}

// RemoteFileInfo implements fs.FileInfo for remote directory entries
type RemoteFileInfo struct {
	name  string
	isDir bool
}

// Name returns the base name of the directory entry.
func (i *RemoteFileInfo) Name() string { return i.name }

// Size returns the length in bytes for the file represented by RemoteFileInfo. Always returns 0 for remote entries.
func (i *RemoteFileInfo) Size() int64 { return 0 }

// Mode returns the file mode for the remote file or directory. Directories are identified with fs.ModeDir flag.
func (i *RemoteFileInfo) Mode() fs.FileMode {
	if i.isDir {
		return fs.ModeDir | 0755
	}
	return 0644
}

// ModTime returns the modification time of the file represented by RemoteFileInfo. It defaults to the zero
// value of time.Time.
func (i *RemoteFileInfo) ModTime() time.Time { return time.Time{} }

// IsDir reports whether the file info describes a directory.
func (i *RemoteFileInfo) IsDir() bool { return i.isDir }

// Sys returns underlying data source (can be nil) for the RemoteFileInfo, typically used in os.FileInfo
// implementations.
func (i *RemoteFileInfo) Sys() interface{} { return nil }

// HostServiceError represents an error returned by the host service.
// Message is a description of the error.
type HostServiceError struct {
	Message string
}

// Error returns the error message stored in the HostServiceError as a string.
func (e *HostServiceError) Error() string {
	return e.Message
}
