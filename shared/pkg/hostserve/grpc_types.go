package hostserve

import (
	"io/fs"
	"time"

	"github.com/bmj2728/hst/shared/protogen/hostserve/v1"
)

///////////////////////////////////////////////////////////////////////////////////////////////////////

// HostServiceGRPCServer provides a gRPC server implementation for host services using the IHostServices interface.
type HostServiceGRPCServer struct {
	Impl IHostServices
	hostservev1.UnimplementedHostServiceServer
}

// HostServiceGRPCClient wraps the filesystemv1.HostServiceClient to provide higher-level client methods.
type HostServiceGRPCClient struct {
	client hostservev1.HostServiceClient
}

/////////////////////////////////////////////////////////////////////////////////////////////////////

// RemoteDirEntry implements fs.DirEntry, this wrapper allows conversion from protobuf to fs.DirEntry
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
