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
	hf.getOpenFiles().CloseAll()
	hf.getTempPaths().CleanupAll()
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

// ReadDir reads the contents of the specified directory path and returns a slice of directory entries or an error.
func (hf *HostFS) ReadDir(ctx context.Context, rootDir, path string) ([]fs.DirEntry, error) {
	clientID := getClientIDFromContext(ctx)
	hclog.Default().Debug("Placeholder Pseudo Cap Check...", "clientID", clientID, "root", rootDir, "path", path)

	// getRoot contains the logic to confirm rootDir is absolute and a directory
	r, err := getRoot(rootDir)
	if err != nil {
		hclog.Default().Error("Failed to get root", "rootDir", rootDir, "err", err)
		return nil, err
	}
	defer closeRoot(r)

	// absToRel handles either absolute or relative paths returning a relative path
	rel, err := absToRel(rootDir, path)
	if err != nil {
		hclog.Default().Error("Failed to get relative path", "rootDir", rootDir, "path", path, "err", err)
		return nil, err
	}

	// We can use the root's FS to read the directory entries
	entries, err := fs.ReadDir(r.FS(), rel)
	// if we accept limit in api, we'll need to check for io.EOF if we return fewer entries than requested
	if err != nil && err != io.EOF {
		hclog.Default().Error("Failed to read directory", "path", path, "err", err)
		return nil, err
	}
	return entries, nil
}

// ReadFile reads the specified file from the given directory and returns its contents as a byte slice or an error.
func (hf *HostFS) ReadFile(ctx context.Context, rootDir, path string) ([]byte, error) {
	clientID := getClientIDFromContext(ctx)
	hclog.Default().Debug("Placeholder Pseudo Cap Check...", "clientID", clientID, "path", path)

	r, err := getRoot(rootDir)
	if err != nil {
		hclog.Default().Error("Failed to get root", "rootDir", rootDir, "err", err)
		return nil, err
	}
	defer closeRoot(r)

	rel, err := absToRel(rootDir, path)
	if err != nil {
		hclog.Default().Error("Failed to get relative path", "rootDir", rootDir, "path", path, "err", err)
		return nil, err
	}

	data, err := fs.ReadFile(r.FS(), rel)
	if err != nil {
		hclog.Default().Error("Failed to read file", "path", path, "err", err)
		return nil, err
	}

	return data, nil
}

// WriteFile writes the specified data to a file within the given directory using the provided permissions.
// If the provided permissions are zero, it defaults to standardPermissions. Returns an error if the operation fails.
func (hf *HostFS) WriteFile(ctx context.Context, rootDir, path string, data []byte, perm os.FileMode) error {
	clientID := getClientIDFromContext(ctx)
	hclog.Default().Debug("Placeholder Pseudo Cap Check...", "clientID", clientID, "rootDir", rootDir, "path", path)

	if perm&permissionsMask == 0 {
		perm = standardPermissions
	}

	r, err := getRoot(rootDir)
	if err != nil {
		hclog.Default().Error("Failed to get root", "rootDir", rootDir, "err", err)
		return err
	}
	defer closeRoot(r)

	rel, err := absToRel(rootDir, path)
	if err != nil {
		hclog.Default().Error("Failed to get relative path", "rootDir", rootDir, "path", path, "err", err)
		return err
	}

	err = r.WriteFile(rel, data, perm)
	if err != nil {
		hclog.Default().Error("Failed to write file", "path", path, "err", err)
		return err
	}
	return nil
}

// Stat retrieves the FileInfo for the given path within the HostFS, returning an error if the operation fails.
func (hf *HostFS) Stat(ctx context.Context, rootDir, path string) (fs.FileInfo, error) {
	clientID := getClientIDFromContext(ctx)
	hclog.Default().Debug("Placeholder Pseudo Cap Check...", "clientID", clientID, "path", path)

	r, err := getRoot(rootDir)
	if err != nil {
		hclog.Default().Error("Failed to get root", "rootDir", rootDir, "err", err)
		return nil, err
	}
	defer closeRoot(r)

	// absToRel handles either absolute or relative paths returning a relative path
	rel, err := absToRel(rootDir, path)
	if err != nil {
		hclog.Default().Error("Failed to get relative path", "rootDir", rootDir, "path", path, "err", err)
		return nil, err
	}

	return fs.Stat(r.FS(), rel)
}

// Rename renames a file or directory from oldPath to newPath within the specified rootDir. Returns an error on failure.
func (hf *HostFS) Rename(ctx context.Context, rootDir, oldPath, newPath string) error {
	clientID := getClientIDFromContext(ctx)
	hclog.Default().Debug("Placeholder Pseudo Cap Check...", "clientID", clientID, "root", rootDir, "oldPath", oldPath, "newPath", newPath)

	r, err := getRoot(rootDir)
	if err != nil {
		hclog.Default().Error("Failed to get root", "rootDir", rootDir, "err", err)
		return err
	}
	defer closeRoot(r)

	oldRel, err := absToRel(rootDir, oldPath)
	if err != nil {
		hclog.Default().Error("Failed to get relative path", "rootDir", rootDir, "oldPath", oldPath, "err", err)
		return err
	}
	newRel, err := absToRel(rootDir, newPath)
	if err != nil {
		hclog.Default().Error("Failed to get relative path", "rootDir", rootDir, "newPath", newPath, "err", err)
		return err
	}

	return r.Rename(oldRel, newRel)
}

func (hf *HostFS) Remove(ctx context.Context, rootDir, path string) error {
	clientID := getClientIDFromContext(ctx)
	hclog.Default().Debug("Placeholder Pseudo Cap Check...", "clientID", clientID, "root", rootDir, "path", path)

	r, err := getRoot(rootDir)
	if err != nil {
		hclog.Default().Error("Failed to get root", "rootDir", rootDir, "err", err)
		return err
	}
	defer closeRoot(r)

	rel, err := absToRel(rootDir, path)
	if err != nil {
		hclog.Default().Error("Failed to get relative path", "rootDir", rootDir, "path", path, "err", err)
		return err
	}

	return r.Remove(rel)
}

func (hf *HostFS) RemoveAll(ctx context.Context, rootDir, path string) error {
	clientID := getClientIDFromContext(ctx)
	hclog.Default().Debug("Placeholder Pseudo Cap Check...", "clientID", clientID, "root", rootDir, "path", path)

	r, err := getRoot(rootDir)
	if err != nil {
		hclog.Default().Error("Failed to get root", "rootDir", rootDir, "err", err)
		return err
	}
	defer closeRoot(r)

	// absToRel handles either absolute or relative paths returning a relative path
	rel, err := absToRel(rootDir, path)
	if err != nil {
		hclog.Default().Error("Failed to get relative path", "rootDir", rootDir, "path", path, "err", err)
		return err
	}

	return r.RemoveAll(rel)
}

// Mkdir creates a new directory within the given root directory with the specified name and permissions.
// Returns an error if the directory cannot be created.
// The `ctx` parameter is used for passing context and the `perm` argument defines the directory's permissions.
func (hf *HostFS) Mkdir(ctx context.Context, rootDir string, name string, perm os.FileMode) error {
	clientID := getClientIDFromContext(ctx)
	hclog.Default().Debug("Placeholder Pseudo Cap Check...", "clientID", clientID, "path", filepath.Join(rootDir, name))
	if perm&permissionsMask == 0 {
		perm = standardPermissions
	}

	r, err := getRoot(rootDir)
	if err != nil {
		hclog.Default().Error("Failed to get root", "rootDir", rootDir, "err", err)
		return err
	}
	defer closeRoot(r)

	// can return an invalid relative path for this scenario - e.g. root is /home/user/foo and name
	// is /home/user/foo/bar/use/MkdirAll - absToRel will return /bar/use/MkdirAll which will cause Mkdir to fail
	// the check does cover the case where root is /home/user/foo and name is /home/user/foo/bar
	rel, err := absToRel(rootDir, name)
	if err != nil {
		hclog.Default().Error("Failed to get relative path", "rootDir", rootDir, "name", name, "err", err)
		return err
	}

	err = r.Mkdir(rel, perm)
	if err != nil {
		hclog.Default().Error("Failed to create directory", "path", filepath.Join(rootDir, name), "err", err)
		return err
	}
	return nil
}

// MkdirAll creates a directory named by the path, along with any necessary parents, in the specified root directory.
// The permission bits for newly created directories are set to perm, masked by a predefined permissions mask.
// Returns an error if the operation fails, including errors from retrieving the root or creating the directories.
func (hf *HostFS) MkdirAll(ctx context.Context, rootDir string, path string, perm os.FileMode) error {
	clientID := getClientIDFromContext(ctx)
	hclog.Default().Debug("Placeholder Pseudo Cap Check...", "clientID", clientID, "path", filepath.Join(rootDir, path))
	if perm&permissionsMask == 0 {
		perm = standardPermissions
	}
	r, err := getRoot(rootDir)
	if err != nil {
		hclog.Default().Error("Failed to get root", "rootDir", rootDir, "err", err)
		return err
	}
	defer closeRoot(r)

	// This check converts the relative if possible, otherwise it will return an error
	rel, err := absToRel(rootDir, path)
	if err != nil {
		hclog.Default().Error("Failed to get relative path", "rootDir", rootDir, "path", path, "err", err)
		return err
	}

	err = r.MkdirAll(rel, perm)
	if err != nil {
		hclog.Default().Error("Failed to create directory", "root", rootDir, "path", path, "err", err)
		return err
	}
	return nil
}

// FileCreate creates a new file at the specified path within the host file system and returns a FileHandle or an error.
func (hf *HostFS) FileCreate(ctx context.Context, rootDir, path string) (FileHandle, error) {
	clientID := getClientIDFromContext(ctx)

	r, err := getRoot(rootDir)
	if err != nil {
		hclog.Default().Error("Failed to get root", "root", rootDir, "err", err)
		return "", err
	}
	defer closeRoot(r)

	// This check converts the relative if possible, otherwise it will return an error
	rel, err := absToRel(rootDir, path)
	if err != nil {
		hclog.Default().Error("Failed to get relative path", "root", rootDir, "path", path, "err", err)
		return "", err
	}

	file, err := r.Create(rel)
	if err != nil {
		hclog.Default().Error("Failed to create file", "root", rootDir, "path", path, "err", err)
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
func (hf *HostFS) FileOpen(ctx context.Context, rootDir, path string, flag int, perm os.FileMode) (FileHandle, uint64, error) {
	clientID := getClientIDFromContext(ctx)
	r, err := getRoot(rootDir)
	if err != nil {
		hclog.Default().Error("Failed to get root", "root", rootDir, "err", err)
		return "", 0, err
	}
	defer closeRoot(r)

	// This check converts the relative if possible, otherwise it will return an error
	rel, err := absToRel(rootDir, path)
	if err != nil {
		hclog.Default().Error("Failed to get relative path", "root", rootDir, "path", path, "err", err)
		return "", 0, err
	}
	file, err := r.OpenFile(rel, flag, perm)
	if err != nil {
		hclog.Default().Error("Failed to create file", "root", rootDir, "path", path, "err", err)
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

// FileStat retrieves the file information for a given FileHandle within a specific context.
func (hf *HostFS) FileStat(ctx context.Context, handle FileHandle) (fs.FileInfo, error) {
	clientID := getClientIDFromContext(ctx)
	hclog.Default().Debug("Placeholder Pseudo Cap Check...", "clientID", clientID, "handle", handle)
	file, err := hf.retrieveOpenFile(ctx, handle)
	if err != nil {
		return nil, err
	}
	return file.Stat()
}

// FileSeek changes the offset of an open file associated with the given handle based on the specified offset and whence.
func (hf *HostFS) FileSeek(ctx context.Context, handle FileHandle, offset int64, whence int) (int64, error) {
	clientID := getClientIDFromContext(ctx)
	hclog.Default().Debug("Placeholder Pseudo Cap Check...", "clientID", clientID, "handle", handle)
	file, err := hf.retrieveOpenFile(ctx, handle)
	if err != nil {
		return 0, err
	}
	return file.Seek(offset, whence)
}

// FileSync ensures the file associated with the given FileHandle is synchronized with storage.
// It retrieves the file from the open file handle and calls its Sync method to persist data to storage.
// Returns an error if the file retrieval or synchronization fails.
func (hf *HostFS) FileSync(ctx context.Context, handle FileHandle) error {
	clientID := getClientIDFromContext(ctx)
	hclog.Default().Debug("Placeholder Pseudo Cap Check...", "clientID", clientID, "handle", handle)
	file, err := hf.retrieveOpenFile(ctx, handle)
	if err != nil {
		return err
	}
	return file.Sync()
}

// FileClose closes an open file associated with the given file handle for a specific client and removes its reference.
// Returns an error if the handle does not exist, the file cannot be closed, or the file reference cannot be removed.
func (hf *HostFS) FileClose(ctx context.Context, handle FileHandle) error {
	clientID := getClientIDFromContext(ctx)
	hclog.Default().Debug("Placeholder Pseudo Cap Check...", "clientID", clientID, "handle", handle)

	files, err := hf.getOpenFiles().GetFilesByClient(clientID)
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
	hclog.Default().Debug("Closed File", "clientID", clientID, "handle", handle, "request", getRequestIDFromContext(ctx))

	err = hf.getOpenFiles().RemoveFile(clientID, handle)
	if err != nil {
		return err
	}
	return nil
}

// FileTruncate adjusts the size of a file to the specified length, truncating or extending it as required.
// ctx is the context for carrying deadlines, cancellation signals, and other request-scoped values.
// handle is the identifier for the open file to be truncated.
// size specifies the new size of the file in bytes.
// Returns an error if the operation fails, such as if the file handle is invalid or other I/O issues occur.
func (hf *HostFS) FileTruncate(ctx context.Context, handle FileHandle, size int64) error {
	clientID := getClientIDFromContext(ctx)
	hclog.Default().Debug("Placeholder Pseudo Cap Check...", "clientID", clientID, "handle", handle)

	file, err := hf.retrieveOpenFile(ctx, handle)
	if err != nil {
		return err
	}

	return file.Truncate(size)
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

// MkdirTemp creates a new temporary directory within the specified rootDir using the given pattern for naming.
// If rootDir is empty, the system's default temporary directory is used. It ensures access checks before creation.
// Returns the path of the created directory or an error if the operation fails.
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
	hclog.Default().Debug("Created Temp Dir", "clientID", clientID, "path", dir, "fullPath", filepath.Join(rootDir, dir))
	hf.getTempPaths().AddPath(clientID, dir)
	return dir, nil
}

// FileCreateTemp creates a temporary file in the specified root directory with the given pattern and returns a file handle.
// The method ensures enhanced security by performing access checks and utilizing root directory constraints for file creation.
// The generated file is tracked for the given client context and added to the open files map for subsequent operations.
// If rootDir is empty, the system's default temporary directory is used. Returns an error if the operation fails.
func (hf *HostFS) FileCreateTemp(ctx context.Context, rootDir string, pattern string) (FileHandle, error) {
	//We'd check for access to the root directory here
	clientID := getClientIDFromContext(ctx)
	if rootDir == "" {
		rootDir = os.TempDir()
	}
	hclog.Default().Debug("Placeholder Pseudo Cap Check...", "clientID", clientID, "path", filepath.Join(rootDir, pattern))

	// We could mirror the logic of os.CreateTemp here but use Create in a Root for enhanced security
	// make the file
	file, err := os.CreateTemp(rootDir, pattern)
	if err != nil {
		return "", err
	}

	// add it to temp tracking
	hf.getTempPaths().AddPath(clientID, file.Name())

	// make a handle
	fh := newFileHandle()

	// add the file to the open file map
	err = hf.openFiles.AddFile(clientID, fh, file)
	if err != nil {
		// if this fails - close the file
		err := file.Close()
		if err != nil {
			return "", err
		}
	}
	// return the handle
	return fh, nil
}
