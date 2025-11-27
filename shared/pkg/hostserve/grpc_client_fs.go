package hostserve

import (
	"context"
	"io"
	"io/fs"
	"os"

	"github.com/bmj2728/hst/shared/protogen/hostserve/v1"
)

// ReadDir retrieves a list of directory entries from the given path through a gRPC call to the host service.
// Returns a slice of fs.DirEntry or an error if the operation fails.
func (c *HostServiceGRPCClient) ReadDir(ctx context.Context, path string) ([]fs.DirEntry, error) {

	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.ReadDir(ctx, &hostservev1.ReadDirRequest{
		Path: path,
	})
	if err != nil {
		return nil, &HostServiceError{Message: err.Error()}
	}
	if resp.Error != nil {
		return nil, &HostServiceError{Message: *resp.Error}
	}

	// Convert protobuf DirEntry to fs.DirEntry
	var entries []fs.DirEntry
	for _, entry := range resp.Entries {
		entries = append(entries, &RemoteDirEntry{
			name:  entry.Name,
			isDir: entry.IsDir,
		})
	}

	return entries, nil
}

// ReadFile reads the specified file from the given directory and returns its contents as a byte slice.
// Returns an error if the file cannot be read or the service encounters an issue.
func (c *HostServiceGRPCClient) ReadFile(ctx context.Context, path string) ([]byte, error) {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.ReadFile(ctx, &hostservev1.ReadFileRequest{
		Path: path,
	})
	if err != nil {
		return nil, &HostServiceError{Message: err.Error()}
	}
	if resp.Error != nil {
		return nil, &HostServiceError{Message: *resp.Error}
	}
	return resp.Contents, nil
}

// WriteFile writes data to the specified file path on the server with the given permissions.
// The context is used for tracing and request cancellation.
// If permissions are set to 0, standardPermissions will be applied as the default.
// Returns an error if the operation fails or in case of a nil response from the server.
func (c *HostServiceGRPCClient) WriteFile(ctx context.Context, path string, data []byte, perm os.FileMode) error {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	if perm == 0 {
		perm = standardPermissions
	}
	resp, err := c.client.WriteFile(ctx, &hostservev1.WriteFileRequest{
		Path: path,
		Data: data,
		Perm: uint32(perm),
	})
	if err != nil {
		return &HostServiceError{Message: err.Error()}
	}
	// Defensive: handle unexpected nil resp
	if resp == nil {
		return &HostServiceError{Message: "nil response from WriteFile"}
	}
	if resp.Error != nil {
		return &HostServiceError{Message: *resp.Error}
	}
	return nil
}

// Mkdir creates a new directory with the specified name and permissions under the given root directory.
func (c *HostServiceGRPCClient) Mkdir(ctx context.Context, rootDir string, name string, perm os.FileMode) error {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	if perm == 0 {
		perm = standardPermissions
	}
	resp, err := c.client.Mkdir(ctx, &hostservev1.MkdirRequest{
		RootDir: rootDir,
		Name:    name,
		Perm:    uint32(perm),
	})
	if err != nil {
		return &HostServiceError{Message: err.Error()}
	}
	if resp == nil {
		return &HostServiceError{Message: "nil response from Mkdir"}
	}
	if resp.Error != nil {
		return &HostServiceError{Message: *resp.Error}
	}
	return nil
}

// MkdirAll creates all necessary directories along the specified path with the provided permissions.
// It operates relative to the given rootDir and uses the optional context for tracing and logging.
func (c *HostServiceGRPCClient) MkdirAll(ctx context.Context, rootDir string, path string, perm os.FileMode) error {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	if perm == 0 {
		perm = standardPermissions
	}
	resp, err := c.client.MkdirAll(ctx, &hostservev1.MkdirAllRequest{
		RootDir: rootDir,
		Path:    path,
		Perm:    uint32(perm),
	})
	if err != nil {
		return &HostServiceError{Message: err.Error()}
	}
	if resp == nil {
		return &HostServiceError{Message: "nil response from MkdirAll"}
	}
	if resp.Error != nil {
		return &HostServiceError{Message: *resp.Error}
	}
	return nil
}

// FileCreate creates a new file at the specified path and returns a handle for the created file or an error if any occurs.
func (c *HostServiceGRPCClient) FileCreate(ctx context.Context, path string) (FileHandle, error) {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.FileCreate(ctx, &hostservev1.FileCreateRequest{
		Path: path,
	})
	if err != nil {
		return "", &HostServiceError{Message: err.Error()}
	}
	if resp == nil {
		return "", &HostServiceError{Message: "nil response from FileCreate"}
	}
	if resp.Error != nil {
		return "", &HostServiceError{Message: *resp.Error}
	}
	return FileHandle(resp.Handle), nil
}

// FileOpen opens a file on the host, given the path, flag, and permissions, and returns a file handle, size, and error if any.
func (c *HostServiceGRPCClient) FileOpen(ctx context.Context, path string, flag int, perm os.FileMode) (FileHandle, uint64, error) {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.FileOpen(ctx, &hostservev1.FileOpenRequest{
		Path: path,
		Mode: flagsToOpenFileMode(flag),
		Perm: uint32(perm),
	})
	if err != nil {
		return "", 0, &HostServiceError{Message: err.Error()}
	}
	if resp == nil {
		return "", 0, &HostServiceError{Message: "nil response from FileOpen"}
	}
	if resp.Error != nil {
		return "", 0, &HostServiceError{Message: *resp.Error}
	}
	return FileHandle(resp.Handle), resp.Size, nil
}

// FileSeek adjusts the file offset for a file identified by handle, based on the specified offset and whence parameters.
// Returns the new file offset or an error if the operation fails.
func (c *HostServiceGRPCClient) FileSeek(ctx context.Context, handle FileHandle, offset int64, whence int) (int64, error) {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.FileSeek(ctx, &hostservev1.FileSeekRequest{
		Handle: string(handle),
		Offset: uint64(offset),
		Whence: uint32(whence),
	})
	if err != nil {
		return 0, &HostServiceError{Message: err.Error()}
	}
	if resp == nil {
		return 0, &HostServiceError{Message: "nil response from FileSeek"}
	}
	if resp.Error != nil {
		return 0, &HostServiceError{Message: *resp.Error}
	}
	return int64(resp.NewOffset), nil
}

// FileClose closes a file associated with the provided handle and releases any resources tied to it.
func (c *HostServiceGRPCClient) FileClose(ctx context.Context, handle FileHandle) error {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	_, err := c.client.FileClose(ctx, &hostservev1.FileCloseRequest{
		Handle: string(handle),
	})
	if err != nil {
		return &HostServiceError{Message: err.Error()}
	}
	return nil
}

// FileReader provides a gRPC client-side implementation to read files in chunks via a streaming connection.
func (c *HostServiceGRPCClient) FileReader(ctx context.Context, handle FileHandle, chunkSize uint32) (io.Reader, error) {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	stream, err := c.client.FileReader(ctx, &hostservev1.FileReadRequest{
		Handle:    string(handle),
		ChunkSize: chunkSize,
	})
	if err != nil {
		return nil, &HostServiceError{Message: err.Error()}
	}
	return &grpcFileStreamReader{stream: stream}, nil
}

// FileWriter opens a stream to write file data to a remote host, returning an io.WriteCloser for handling the stream.
// It accepts a context for request lifecycle management and a file handle for identifying the target file.
// Returns an io.WriteCloser on success or an error if the stream initialization fails.
func (c *HostServiceGRPCClient) FileWriter(ctx context.Context, handle FileHandle) (io.WriteCloser, error) {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	stream, err := c.client.FileWriter(ctx)
	if err != nil {
		return nil, &HostServiceError{Message: err.Error()}
	}
	return &grpcFileStreamWriter{stream: stream, handle: handle}, nil
}

// MkdirTemp creates a new temporary directory in the specified root directory with a given pattern and returns its path.
// Returns an error if the operation fails or if there is an issue with the response from the server.
func (c *HostServiceGRPCClient) MkdirTemp(ctx context.Context, rootDir string, pattern string) (string, error) {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.MkdirTemp(ctx, &hostservev1.MkdirTempRequest{
		RootDir: rootDir,
		Pattern: pattern,
	})
	if err != nil {
		return "", &HostServiceError{Message: err.Error()}
	}
	if resp == nil {
		return "", &HostServiceError{Message: "nil response from MkdirTemp"}
	}
	if resp.Error != nil {
		return "", &HostServiceError{Message: *resp.Error}
	}
	return resp.GetPath(), nil
}

// FileCreateTemp creates a temporary file with the specified root directory and pattern using the gRPC client.
// Returns a FileHandle for the created file or an error if the operation fails.
func (c *HostServiceGRPCClient) FileCreateTemp(ctx context.Context, rootDir string, pattern string) (FileHandle, error) {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.FileCreateTemp(ctx, &hostservev1.FileCreateTempRequest{
		RootDir: rootDir,
		Pattern: pattern,
	})
	if err != nil {
		return "", &HostServiceError{Message: err.Error()}
	}
	if resp == nil {
		return "", &HostServiceError{Message: "nil response from FileCreateTemp"}
	}
	if resp.Error != nil {
		return "", &HostServiceError{Message: *resp.Error}
	}
	return FileHandle(resp.Handle), nil
}
