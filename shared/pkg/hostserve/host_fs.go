package hostserve

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
)

const (
	permissionsMask     = fs.FileMode(0777)
	standardPermissions = fs.FileMode(0644)

	// minChunkSize is used to avoid excessive overhead
	// 8K is used as it aligns with common buffer and block sizes
	minChunkSize = 8 * 1024

	// maxChunkSize gRPC allows a maximum of 4MB(4194304) per message.
	// Setting to 4,000,000(3.81MB) to allow room for message overhead.
	maxChunkSize = 4000000
)

// ErrInvalidPath represents an error indicating the provided path is invalid or not a directory.
var (
	ErrInvalidPath = errors.New("invalid path")
)

func (hf *HostFS) Cleanup() {
	hclog.Default().Info("Cleaning up HostFS resources")
	hf.GetOpenFiles().CloseAll()
}

// HostFS is a file system abstraction that provides methods to interact with a host's file system.
type HostFS struct {
	openFiles *OpenFiles
}

// NewHostFS creates and returns a new instance of HostFS.
func NewHostFS() *HostFS {
	return &HostFS{
		openFiles: newOpenFiles(),
	}
}

func (hf *HostFS) GetOpenFiles() *OpenFiles {
	if hf.openFiles == nil {
		hf.openFiles = newOpenFiles()
	}
	return hf.openFiles
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
// If the provided permissions are zero, it defaults to standardPermissions. Returns an error if the operation fails.
func (hf *HostFS) WriteFile(ctx context.Context, path string, data []byte, perm os.FileMode) error {
	if perm&permissionsMask == 0 {
		perm = standardPermissions
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

func (hf *HostFS) FileOpen(ctx context.Context, path string, flag int, perm os.FileMode) (FileHandle, uint64, error) {
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

func (hf *HostFS) FileClose(ctx context.Context, handle FileHandle) error {
	clientID := getClientIDFromContext(ctx)
	files, err := hf.GetOpenFiles().GetFilesByClient(clientID)
	if err != nil {
		return err
	}
	file, exists := files[handle]
	if !exists {
		return fmt.Errorf("file handle %s does not exist for client %s", handle, clientID)
	}
	err = file.Close()
	if err != nil {
		return err
	}
	err = hf.GetOpenFiles().RemoveFile(clientID, handle)
	if err != nil {
		return err
	}
	return nil
}

// FileRead returns a reader for the specified file handle, typically os.File/fs.File.
// This implementation contains pseudocode for suggested security checks.
func (hf *HostFS) FileRead(ctx context.Context, handle FileHandle, chunkSize uint32) (io.Reader, error) {
	// Validate chunk size - return early if invalid
	if chunkSize < minChunkSize || chunkSize > maxChunkSize {
		return nil, fmt.Errorf("chunk size must be between %d and %d bytes", minChunkSize, maxChunkSize)
	}
	// Extract client ID from context
	clientId := getClientIDFromContext(ctx)

	file, err := hf.GetOpenFiles().GetFile(clientId, handle)
	if err != nil {
		return nil, err
	}

	return file, nil
}
