package hostserve

import (
	"context"
	"io"
	"io/fs"
	"os"
)

// IHostServices is an interface that combines IHostFS and IHostEnv to provide file system and environment services.
type IHostServices interface {
	IHostFS
	IHostEnv
}

// IHostFS is an interface that defines methods to interact with the host file system.
type IHostFS interface {

	// ReadDir reads the contents of the directory specified by `path` and returns a slice of directory
	// entries or an error.
	ReadDir(ctx context.Context, path string) ([]fs.DirEntry, error)

	// ReadFile reads the specified file from the given directory and returns its contents as a byte slice or an error.
	ReadFile(ctx context.Context, path string) ([]byte, error)

	// WriteFile writes data to the specified file within the given directory, applying the provided file permissions.
	WriteFile(ctx context.Context, path string, data []byte, perm os.FileMode) error

	// Mkdir creates a new directory within the specified root directory with the given name and permissions.
	Mkdir(ctx context.Context, rootDir string, name string, perm os.FileMode) error

	// MkdirAll creates a directory specified by the path, including all necessary parent directories, with the given permissions.
	MkdirAll(ctx context.Context, rootDir string, path string, perm os.FileMode) error

	MkdirTemp(ctx context.Context, rootDir string, pattern string) (string, error)

	// FileCreate creates a new file at the specified path and returns a FileHandle for the file or an error if it fails.
	FileCreate(ctx context.Context, path string) (FileHandle, error)

	//FileCreateTemp(ctx context.Context, rootDir string, pattern string) (FileHandle, error)

	// FileOpen opens a file at the specified path with the given flags and permissions,
	// returning a handle, file size, and error.
	FileOpen(ctx context.Context, path string, flag int, perm os.FileMode) (FileHandle, uint64, error)

	// FileClose closes an open file identified by the provided FileHandle. Returns an error if the operation fails.
	FileClose(ctx context.Context, handle FileHandle) error

	// FileReader reads data from an open file, identified by the provided FileHandle, in chunks of the specified size.
	FileReader(ctx context.Context, handle FileHandle, chunkSize uint32) (io.Reader, error)

	FileWriter(ctx context.Context, handle FileHandle) (io.WriteCloser, error)
}

// IHostEnv defines a contract for interacting with environment variables in the host system.
type IHostEnv interface {

	// GetEnv fetches the value of an environment variable by its key and returns it as a string.
	GetEnv(ctx context.Context, key string) (string, error)
}
