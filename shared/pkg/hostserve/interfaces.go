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

	// Stat retrieves the file information for the specified path in the host file system.
	Stat(ctx context.Context, path string) (fs.FileInfo, error)

	// Mkdir creates a new directory within the specified root directory with the given name and permissions.
	Mkdir(ctx context.Context, rootDir string, name string, perm os.FileMode) error

	// MkdirAll creates a directory specified by the path, including all necessary parent directories, with the given permissions.
	MkdirAll(ctx context.Context, rootDir string, path string, perm os.FileMode) error

	// MkdirTemp creates a new temporary directory in the specified rootDir with a name formed using the given pattern.
	MkdirTemp(ctx context.Context, rootDir string, pattern string) (string, error)

	// FileCreate creates a new file at the specified path and returns a FileHandle for the file or an error if it fails.
	FileCreate(ctx context.Context, path string) (FileHandle, error)

	// FileCreateTemp creates a new temporary file in the specified root directory using the given pattern.
	FileCreateTemp(ctx context.Context, rootDir string, pattern string) (FileHandle, error)

	// FileOpen opens a file at the specified path with the given flags and permissions,
	// returning a handle, file size, and error.
	FileOpen(ctx context.Context, path string, flag int, perm os.FileMode) (FileHandle, uint64, error)

	// FileStat retrieves metadata about a file represented by the provided FileHandle. It returns file information or an error.
	FileStat(ctx context.Context, handle FileHandle) (fs.FileInfo, error)

	// FileSeek moves the file cursor to a new position as specified by offset and whence for the provided file handle.
	// Returns the new cursor offset from the start of the file or an error if the operation fails.
	FileSeek(ctx context.Context, handle FileHandle, offset int64, whence int) (int64, error)

	// FileSync synchronizes the state of a file associated with the given handle to the storage device.
	FileSync(ctx context.Context, handle FileHandle) error

	// FileClose closes an open file identified by the provided FileHandle. Returns an error if the operation fails.
	FileClose(ctx context.Context, handle FileHandle) error

	// FileTruncate truncates the file represented by the given handle to the specified size in bytes.
	// Returns an error if truncation fails, the size is invalid, or the handle is not valid.
	FileTruncate(ctx context.Context, handle FileHandle, size int64) error

	// FileReader reads data from an open file, identified by the provided FileHandle, in chunks of the specified size.
	FileReader(ctx context.Context, handle FileHandle, chunkSize uint32) (io.Reader, error)

	// FileWriter returns an io.WriteCloser for writing to the specified FileHandle.
	FileWriter(ctx context.Context, handle FileHandle) (io.WriteCloser, error)
}

// IHostEnv defines a contract for interacting with environment variables in the host system.
type IHostEnv interface {

	// GetEnv fetches the value of an environment variable by its key and returns it as a string.
	GetEnv(ctx context.Context, key string) (string, error)

	// TempDir retrieves the path to the temporary directory for the current user or system.
	TempDir(ctx context.Context) (string, error)

	// UserCacheDir retrieves the path to the user-specific cache directory on the host system.
	UserCacheDir(ctx context.Context) (string, error)

	// UserConfigDir retrieves the path to the user's configuration directory as a string within the current system context.
	UserConfigDir(ctx context.Context) (string, error)

	// UserHomeDir retrieves the current user's home directory from the host environment.
	UserHomeDir(ctx context.Context) (string, error)
}
