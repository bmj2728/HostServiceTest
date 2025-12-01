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

	// ReadDir reads the contents of a directory specified by rootDir and path, returning a list of fs.DirEntry or an error.
	ReadDir(ctx context.Context, rootDir, path string) ([]fs.DirEntry, error)

	// ReadFile reads the contents of a file specified by the given path in the root directory and returns the data or an error.
	ReadFile(ctx context.Context, rootDir, path string) ([]byte, error)

	// WriteFile writes the specified data to a file at the given path in the root directory, with the provided permissions.
	// Parameters:
	// - ctx: Context for request management, such as timeouts or cancellations.
	// - rootDir: Root directory where the file operation is executed.
	// - path: Path of the file relative to the root directory.
	// - data: Byte slice containing the file's content to be written.
	// - perm: File permissions to set on the newly written file.
	// Returns an error if the file write operation fails.
	WriteFile(ctx context.Context, rootDir, path string, data []byte, perm os.FileMode) error

	// Stat retrieves information about the specified file or directory at the given path relative to the rootDir.
	Stat(ctx context.Context, rootDir, path string) (fs.FileInfo, error)

	// Rename renames a file or directory from oldPath to newPath within the specified rootDir. Returns an error if the operation fails.
	Rename(ctx context.Context, rootDir, oldPath, newPath string) error

	// Mkdir creates a new directory with the specified name and permissions at the given root directory.
	Mkdir(ctx context.Context, rootDir string, name string, perm os.FileMode) error

	// MkdirAll creates a directory hierarchy at the specified path with the given permissions.
	MkdirAll(ctx context.Context, rootDir string, path string, perm os.FileMode) error

	// MkdirTemp creates a temporary directory named according to the specified pattern under the given root directory.
	// It returns the full path of the created temporary directory or an error if the operation fails.
	MkdirTemp(ctx context.Context, rootDir string, pattern string) (string, error)

	// FileCreate creates a new file at the specified path within the root directory and returns a unique FileHandle.
	FileCreate(ctx context.Context, rootDir, path string) (FileHandle, error)

	// FileCreateTemp creates a temporary file in the specified root directory with the provided pattern.
	// Returns a file handle for the new temporary file or an error if the operation fails.
	FileCreateTemp(ctx context.Context, rootDir string, pattern string) (FileHandle, error)

	// FileOpen opens a file at the specified path with given flags and permissions, returning a file handle, its size, or an error.
	FileOpen(ctx context.Context, rootDir, path string, flag int, perm os.FileMode) (FileHandle, uint64, error)

	// FileStat retrieves the detailed information of a file associated with the provided file handle.
	// Returns fs.FileInfo and an error if the operation fails.
	FileStat(ctx context.Context, handle FileHandle) (fs.FileInfo, error)

	// FileSeek adjusts the file offset for the given file handle based on the specified offset and origin (whence).
	// It returns the new file offset relative to the file's start and any error encountered during the operation.
	FileSeek(ctx context.Context, handle FileHandle, offset int64, whence int) (int64, error)

	// FileSync flushes all buffered data for the given file handle, ensuring it is written to stable storage.
	FileSync(ctx context.Context, handle FileHandle) error

	// FileClose closes the file associated with the given handle and releases any system resources used by it.
	FileClose(ctx context.Context, handle FileHandle) error

	// FileTruncate truncates the file identified by the given FileHandle to the specified size in bytes. Returns an error if fails.
	FileTruncate(ctx context.Context, handle FileHandle, size int64) error

	// FileReader returns an io.Reader for sequentially reading data from an open file handle in chunks of the specified size.
	FileReader(ctx context.Context, handle FileHandle, chunkSize uint32) (io.Reader, error)

	// FileWriter returns an io.WriteCloser to write data to the file identified by the given FileHandle.
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
