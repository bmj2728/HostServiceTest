package hostserve

import (
	"context"
	"io/fs"

	"github.com/bmj2728/hst/shared/protogen/hostserve/v1"
)

func NewHostServiceGRPCClient(client hostservev1.HostServiceClient) *HostServiceGRPCClient {
	return &HostServiceGRPCClient{client: client}
}

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
	// This is
	var entries []fs.DirEntry
	for _, entry := range resp.Entries {
		entries = append(entries, &RemoteDirEntry{
			name:  entry.Name,
			isDir: entry.IsDir,
		})
	}

	return entries, nil
}

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
