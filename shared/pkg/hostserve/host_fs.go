package hostserve

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	hostservev1 "github.com/bmj2728/hst/shared/protogen/hostserve/v1"
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

// openFileModeToFlags converts the OpenFileMode enum to appropriate file flags for use with os package operations.
func openFileModeToFlags(mode hostservev1.OpenFileMode) int {
	switch mode {
	case hostservev1.OpenFileMode_READ_ONLY:
		return os.O_RDONLY
	case hostservev1.OpenFileMode_WRITE_TRUNCATE:
		return os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	case hostservev1.OpenFileMode_WRITE_APPEND:
		return os.O_WRONLY | os.O_CREATE | os.O_APPEND
	case hostservev1.OpenFileMode_WRITE_EXCLUSIVE:
		return os.O_WRONLY | os.O_CREATE | os.O_EXCL
	case hostservev1.OpenFileMode_READ_WRITE:
		return os.O_RDWR
	case hostservev1.OpenFileMode_READ_WRITE_CREATE:
		return os.O_RDWR | os.O_CREATE
	case hostservev1.OpenFileMode_READ_WRITE_TRUNCATE:
		return os.O_RDWR | os.O_CREATE | os.O_TRUNC
	case hostservev1.OpenFileMode_READ_WRITE_APPEND:
		return os.O_RDWR | os.O_CREATE | os.O_APPEND
	default:
		return os.O_RDONLY
	}
}

// ClientID represents a unique identifier for a client in a system or application.
type ClientID string

// RootHandle represents a unique identifier for a root resource within the system.
type RootHandle string

// OpenRootMap is a nested map associating a ClientID with RootHandles and their corresponding os.Root instances.
type OpenRootMap map[ClientID]map[RootHandle]*os.Root

// OpenRoots manages a thread-safe collection of open root directories accessible by different clients.
type OpenRoots struct {
	roots OpenRootMap
	mu    sync.RWMutex
}

// NewOpenRoots creates and returns a new instance of OpenRoots with an empty OpenRootMap and an initialized mutex.
func NewOpenRoots() *OpenRoots {
	return &OpenRoots{
		roots: make(OpenRootMap),
	}
}

// FileHandle represents a unique identifier for an open file within a specific client context.
type FileHandle string

// OpenFileMap represents a mapping of ClientIDs to their associated FileHandles and open file pointers.
type OpenFileMap map[ClientID]map[FileHandle]*os.File

// OpenFiles manages a thread-safe collection of open file references, grouped by client and file handle.
type OpenFiles struct {
	files OpenFileMap
	mu    sync.RWMutex
}

// NewOpenFiles initializes and returns a new instance of OpenFiles with an empty OpenFileMap and a RWMutex.
func NewOpenFiles() *OpenFiles {
	return &OpenFiles{
		files: make(OpenFileMap),
	}
}

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

// getRoot resolves the absolute path of the given directory and validates if it is a directory
// before returning an Root object for it.
func getRoot(dir string) (*os.Root, error) {
	if !filepath.IsAbs(dir) {
		p, err := filepath.Abs(dir)
		if err != nil {
			return nil, err
		}
		dir = p
	}
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, ErrInvalidPath
	}
	return os.OpenRoot(dir)
}

// closeRoot ensures the provided root is closed and logs an error if the operation fails.
// It handles logging the root's name and the corresponding error details.
func closeRoot(r *os.Root) {
	err := r.Close()
	if err != nil {
		hclog.Default().Error("Failed to close root", "path", r.Name(), "err", err)
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
	}
	return err
}
