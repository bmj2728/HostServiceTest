package hostserve

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
)

const (
	PermissionsMask     = fs.FileMode(0777)
	StandardPermissions = fs.FileMode(0644)
)

// ErrInvalidPath represents an error indicating the provided path is invalid or not a directory.
var (
	ErrInvalidPath = errors.New("invalid path")
)

// HostFS is a file system abstraction that provides methods to interact with a host's file system.
type HostFS struct {
	openFiles *OpenFiles
	openRoots *OpenRoots
}

// NewHostFS creates and returns a new instance of HostFS.
func NewHostFS() *HostFS {
	return &HostFS{
		openFiles: NewOpenFiles(),
		openRoots: NewOpenRoots(),
	}
}

// ReadDir reads the contents of the specified directory path and returns a slice of directory entries or an error.
func (hf *HostFS) ReadDir(ctx context.Context, path string) ([]fs.DirEntry, error) {
	r, err := getRoot(path)
	if err != nil {
		hclog.Default().Error("Failed to open root", "path", path, "err", err)
		return nil, err
	}
	defer closeRoot(r)
	entries, err := fs.ReadDir(r.FS(), ".")
	if err != nil {
		hclog.Default().Error("Failed to read directory", "path", path, "err", err)
		return nil, err
	}
	return entries, nil
}

// ReadFile reads the specified file from the given directory and returns its contents as a byte slice or an error.
func (hf *HostFS) ReadFile(ctx context.Context, path string) ([]byte, error) {
	dir, file := filepath.Split(path)
	r, err := getRoot(dir)
	if err != nil {
		hclog.Default().Error("Failed to open root", "path", dir, "err", err)
		return nil, err
	}
	defer closeRoot(r)
	data, err := fs.ReadFile(r.FS(), file)
	if err != nil {
		hclog.Default().Error("Failed to read file", "path", path, "err", err)
		return nil, err
	}
	return data, nil
}

// WriteFile writes the specified data to a file within the given directory using the provided permissions.
// If the provided permissions are zero, it defaults to StandardPermissions. Returns an error if the operation fails.
func (hf *HostFS) WriteFile(ctx context.Context, path string, data []byte, perm os.FileMode) error {
	if perm&PermissionsMask == 0 {
		perm = StandardPermissions
	}
	dir, file := filepath.Split(path)
	r, err := getRoot(dir)
	if err != nil {
		hclog.Default().Error("Failed to open root", "path", dir, "err", err)
		return err
	}
	defer closeRoot(r)
	err = r.WriteFile(file, data, perm)
	if err != nil {
		hclog.Default().Error("Failed to write file", "path", path, "err", err)
		return err
	}
	return nil
}

func (hf *HostFS) OpenFile(ctx context.Context, path string, flag int, perm os.FileMode) (FileHandle, uint64, error) {
	clientID := getClientIDFromContext(ctx)
	d, f := filepath.Split(path)
	r, err := getRoot(d)
	if err != nil {
		hclog.Default().Error("Failed to open root", "path", d, "err", err)
		return "", 0, err
	}
	defer closeRoot(r)
	file, err := r.OpenFile(f, flag, perm)
	if err != nil {
		hclog.Default().Error("Failed to open file", "path", path, "err", err)
		return "", 0, err
	}
	fh := newFileHandle()
	err = hf.openFiles.AddFile(clientID, fh, file)
	if err != nil {
		return "", 0, err
	}
	info, err := file.Stat()
	if err != nil {
		return fh, 0, err
	}
	return fh, uint64(info.Size()), nil
}
