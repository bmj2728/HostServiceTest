package hostserve

import (
	"context"
	"io/fs"
	"time"

	hostservev1 "github.com/bmj2728/hst/shared/protogen/hostserve/v1"
)

type HostServiceGRPCServer struct {
	Impl IHostServices
	hostservev1.UnimplementedHostServiceServer
}

func (s *HostServiceGRPCServer) ReadDir(ctx context.Context, request *hostservev1.ReadDirRequest) (*hostservev1.ReadDirResponse, error) {
	entries, err := s.Impl.ReadDir(request.Path)
	if err != nil {
		errMsg := err.Error()
		return &hostservev1.ReadDirResponse{
			Entries: nil,
			Error:   &errMsg,
		}, nil
	}

	// Convert fs.DirEntry to protobuf DirEntry
	var pbEntries []*hostservev1.DirEntry
	for _, entry := range entries {
		pbEntries = append(pbEntries, &hostservev1.DirEntry{
			Name:  entry.Name(),
			IsDir: entry.IsDir(),
		})
	}

	return &hostservev1.ReadDirResponse{
		Entries: pbEntries,
		Error:   nil,
	}, nil
}

type HostServiceGRPCClient struct {
	client hostservev1.HostServiceClient
}

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
	var entries []fs.DirEntry
	for _, entry := range resp.Entries {
		entries = append(entries, &RemoteDirEntry{
			name:  entry.Name,
			isDir: entry.IsDir,
		})
	}

	return entries, nil
}

// RemoteDirEntry implements fs.DirEntry for directory entries received from the host
type RemoteDirEntry struct {
	name  string
	isDir bool
}

func (e *RemoteDirEntry) Name() string {
	return e.name
}

func (e *RemoteDirEntry) IsDir() bool {
	return e.isDir
}

func (e *RemoteDirEntry) Type() fs.FileMode {
	if e.isDir {
		return fs.ModeDir
	}
	return 0
}

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

func (i *RemoteFileInfo) Name() string { return i.name }
func (i *RemoteFileInfo) Size() int64  { return 0 }
func (i *RemoteFileInfo) Mode() fs.FileMode {
	if i.isDir {
		return fs.ModeDir | 0755
	}
	return 0644
}
func (i *RemoteFileInfo) ModTime() time.Time { return time.Time{} }
func (i *RemoteFileInfo) IsDir() bool        { return i.isDir }
func (i *RemoteFileInfo) Sys() interface{}   { return nil }

type HostServiceError struct {
	Message string
}

func (e *HostServiceError) Error() string {
	return e.Message
}
