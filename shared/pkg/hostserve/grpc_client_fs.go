package hostserve

import (
	"context"
	"io/fs"
	"os"

	"github.com/bmj2728/hst/shared/protogen/hostserve/v1"
)

// NewHostServiceGRPCClient creates a new instance of HostServiceGRPCClient using the provided HostServiceClient.
func NewHostServiceGRPCClient(client hostservev1.HostServiceClient) *HostServiceGRPCClient {
	return &HostServiceGRPCClient{client: client}
}

// ReadDir retrieves a list of directory entries from the given path through a gRPC call to the host service.
// Returns a slice of fs.DirEntry or an error if the operation fails.
func (c *HostServiceGRPCClient) ReadDir(path string) ([]fs.DirEntry, error) {
	resp, err := c.client.ReadDir(context.Background(), &hostservev1.ReadDirRequest{
		Path: path,
	})
	if err != nil {
		return nil, err
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
func (c *HostServiceGRPCClient) ReadFile(dir, file string) ([]byte, error) {
	resp, err := c.client.ReadFile(context.Background(), &hostservev1.ReadFileRequest{
		Dir:  dir,
		File: file,
	})
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, &HostServiceError{Message: *resp.Error}
	}
	return resp.Contents, nil
}

func (c *HostServiceGRPCClient) WriteFile(dir, file string, data []byte, perm os.FileMode) error {
	if perm == 0 {
		perm = StandardPermissions
	}
	resp, err := c.client.WriteFile(context.Background(), &hostservev1.WriteFileRequest{
		Dir:  dir,
		File: file,
		Data: data,
		Perm: uint32(perm),
	})
	if err != nil {
		return &HostServiceError{Message: err.Error()}
	}
	return &HostServiceError{Message: *resp.Error}
}
