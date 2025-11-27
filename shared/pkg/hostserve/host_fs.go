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
	hf.GetTempPaths().CleanupAll()
}

// HostFS is a file system abstraction that provides methods to interact with a host's file system.
type HostFS struct {
	openFiles *OpenFiles
	tempPaths *TempPaths
}

// NewHostFS creates and returns a new instance of HostFS.
func NewHostFS() *HostFS {
	return &HostFS{
		openFiles: newOpenFiles(),
		tempPaths: newTempPaths(),
	}
}

// GetTempPaths returns a pointer to the TempPaths object, creating a new one if it has not been initialized.
func (hf *HostFS) GetTempPaths() *TempPaths {
	if hf.tempPaths == nil {
		hf.tempPaths = newTempPaths()
	}
	return hf.tempPaths
}

// GetOpenFiles returns a reference to the open files manager, initializing it if not already created.
func (hf *HostFS) GetOpenFiles() *OpenFiles {
	if hf.openFiles == nil {
		hf.openFiles = newOpenFiles()
	}
	return hf.openFiles
}

// ReadDir reads the contents of the specified directory path and returns a slice of directory entries or an error.
func (hf *HostFS) ReadDir(ctx context.Context, path string) ([]fs.DirEntry, error) {
	clientID := getClientIDFromContext(ctx)
	hclog.Default().Debug("Placeholder Pseudo Cap Check...", "clientID", clientID, "path", path)
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
	clientID := getClientIDFromContext(ctx)
	hclog.Default().Debug("Placeholder Pseudo Cap Check...", "clientID", clientID, "path", path)
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
	clientID := getClientIDFromContext(ctx)
	hclog.Default().Debug("Placeholder Pseudo Cap Check...", "clientID", clientID, "path", path)
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

func (hf *HostFS) Mkdir(ctx context.Context, rootDir string, name string, perm os.FileMode) error {
	clientID := getClientIDFromContext(ctx)
	hclog.Default().Debug("Placeholder Pseudo Cap Check...", "clientID", clientID, "path", filepath.Join(rootDir, name))
	if perm&permissionsMask == 0 {
		perm = standardPermissions
	}
	r, err := getRoot(rootDir)
	if err != nil {
		hclog.Default().Error("Failed to open root", "path", rootDir, "err", err)
		return err
	}
	defer closeRoot(r)
	err = r.Mkdir(name, perm)
	if err != nil {
		hclog.Default().Error("Failed to create directory", "path", filepath.Join(rootDir, name), "err", err)
		return err
	}
	return nil
}

func (hf *HostFS) MkdirAll(ctx context.Context, rootDir string, path string, perm os.FileMode) error {
	clientID := getClientIDFromContext(ctx)
	hclog.Default().Debug("Placeholder Pseudo Cap Check...", "clientID", clientID, "path", filepath.Join(rootDir, path))
	if perm&permissionsMask == 0 {
		perm = standardPermissions
	}
	r, err := getRoot(rootDir)
	if err != nil {
		hclog.Default().Error("Failed to open root", "path", rootDir, "err", err)
		return err
	}
	defer closeRoot(r)
	err = r.MkdirAll(path, perm)
	if err != nil {
		hclog.Default().Error("Failed to create directory", "path", filepath.Join(rootDir, path), "err", err)
		return err
	}
	return nil
}

func (hf *HostFS) FileCreate(ctx context.Context, path string) (FileHandle, error) {
	clientID := getClientIDFromContext(ctx)
	d, f := filepath.Split(path)
	r, err := getRoot(d)
	if err != nil {
		hclog.Default().Error("Failed to open root", "path", d, "err", err)
		return "", err
	}
	defer closeRoot(r)
	file, err := r.Create(f)
	if err != nil {
		hclog.Default().Error("Failed to create file", "path", path, "err", err)
		return "", err
	}
	fh := newFileHandle()
	err = hf.openFiles.AddFile(clientID, fh, file)
	if err != nil {
		return "", err
	}
	return fh, nil
}

// FileOpen opens a file at the specified path with the given flags and permissions, returning a FileHandle and file size.
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

// FileClose closes an open file associated with the given file handle for a specific client and removes its reference.
// Returns an error if the handle does not exist, the file cannot be closed, or the file reference cannot be removed.
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

// FileReader returns a reader for the specified file handle, typically os.File/fs.File.
// This implementation contains pseudocode for suggested security checks.
func (hf *HostFS) FileReader(ctx context.Context, handle FileHandle, chunkSize uint32) (io.Reader, error) {
	// Validate chunk size - return early if invalid
	if chunkSize < minChunkSize || chunkSize > maxChunkSize {
		return nil, fmt.Errorf("chunk size must be between %d and %d bytes", minChunkSize, maxChunkSize)
	}

	file, err := hf.retrieveOpenFile(ctx, handle)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// FileWriter returns a WriteCloser for the specified file handle, allowing write operations on the file.
func (hf *HostFS) FileWriter(ctx context.Context, handle FileHandle) (io.WriteCloser, error) {

	file, err := hf.retrieveOpenFile(ctx, handle)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (hf *HostFS) MkdirTemp(ctx context.Context, rootDir string, pattern string) (string, error) {
	//We'd check for access to the root directory here
	clientID := getClientIDFromContext(ctx)
	if rootDir == "" {
		rootDir = os.TempDir()
	}
	hclog.Default().Debug("Placeholder Pseudo Cap Check...", "clientID", clientID, "path", filepath.Join(rootDir, pattern))
	// We could mirror the logic of os.MkdirTemp here, but use Mkdir in a Root for enhanced security
	dir, err := os.MkdirTemp(rootDir, pattern)
	if err != nil {
		return "", err
	}
	hf.GetTempPaths().AddPath(clientID, dir)
	return dir, nil
}
