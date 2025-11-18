package hostserve

import (
	"io/fs"
	"time"
)

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
