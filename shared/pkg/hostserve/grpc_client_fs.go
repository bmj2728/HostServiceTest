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

func (c *HostServiceGRPCClient) FileWriter(ctx context.Context, handle FileHandle) (io.WriteCloser, error) {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	stream, err := c.client.FileWriter(ctx)
	if err != nil {
		return nil, &HostServiceError{Message: err.Error()}
	}
	return &grpcFileStreamWriter{stream: stream, handle: handle}, nil
}

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
