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

// getTempPaths returns a pointer to the TempPaths object, creating a new one if it has not been initialized.
func (hf *HostFS) getTempPaths() *TempPaths {
	if hf.tempPaths == nil {
		hf.tempPaths = newTempPaths()
	}
	return hf.tempPaths
}

// getOpenFiles returns a reference to the open files manager, initializing it if not already created.
func (hf *HostFS) getOpenFiles() *OpenFiles {
	if hf.openFiles == nil {
		hf.openFiles = newOpenFiles()
	}
	return hf.openFiles
}

func absToRel(root, path string) (string, error) {
	if filepath.IsAbs(path) {
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return "", fmt.Errorf("failed to get relative path: %w", err)
		}
		return filepath.Clean(rel), nil
	}
	return filepath.Clean(path), nil
}

// ReadDir reads the contents of the specified directory path and returns a slice of directory entries or an error.
func (hf *HostFS) ReadDir(ctx context.Context, rootDir, path string) ([]fs.DirEntry, error) {
	clientID := getClientIDFromContext(ctx)
	hclog.Default().Debug("Placeholder Pseudo Cap Check...", "clientID", clientID, "root", rootDir, "path", path)

	if !filepath.IsAbs(rootDir) {
		// we will need logic here to check if the rootDir is absolute and identify the rootDir based on the context or
		// a configuration file. For now, we will return an error.
		hclog.Default().Warn("rootDir is not absolute", "rootDir", rootDir)
		return nil, fmt.Errorf("rootDir is not absolute: %s", rootDir)
	}

	rel, err := absToRel(rootDir, path)
	if err != nil {
		hclog.Default().Error("Failed to get relative path", "rootDir", rootDir, "path", path, "err", err)
		return nil, err
	}

	// Since we want to read a directory, we can open the path in the root to retrieve and os.File object
	f, err := os.OpenInRoot(rootDir, rel)
	if err != nil {
		hclog.Default().Error("Failed to open directory", "path", path, "err", err)
		return nil, err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			hclog.Default().Error("Failed to close file", "path", path, "err", err)
		}
	}(f)

	// We can then use the os.File object to read the directory entries
	// potential future enhancement to accept an entry count parameter
	// for now return all entries
	entries, err := f.ReadDir(0)
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

	if !filepath.IsAbs(rootDir) {
		// we will need logic here to check if the rootDir is absolute and identify the rootDir based on the context or
		// a configuration file. For now, we will return an error.
		hclog.Default().Warn("rootDir is not absolute", "rootDir", rootDir)
		return nil, fmt.Errorf("rootDir is not absolute: %s", rootDir)
	}

	rel, err := absToRel(rootDir, path)
	if err != nil {
		hclog.Default().Error("Failed to get relative path", "rootDir", rootDir, "path", path, "err", err)
		return nil, err
	}

	r, err := os.OpenRoot(rootDir)
	if err != nil {
		hclog.Default().Error("Failed to open file", "path", path, "err", err)
		return nil, err
	}
	defer func(f *os.Root) {
		err := f.Close()
		if err != nil {
			hclog.Default().Error("Failed to close file", "path", path, "err", err)
		}
	}(r)

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

	if !filepath.IsAbs(rootDir) {
		// we will need logic here to check if the rootDir is absolute and identify the rootDir based on the context or
		// a configuration file. For now, we will return an error.
		hclog.Default().Warn("rootDir is not absolute", "rootDir", rootDir)
		return fmt.Errorf("rootDir is not absolute: %s", rootDir)
	}

	rel, err := absToRel(rootDir, path)
	if err != nil {
		hclog.Default().Error("Failed to get relative path", "rootDir", rootDir, "path", path, "err", err)
		return err
	}

	r, err := os.OpenRoot(rootDir)
	if err != nil {
		hclog.Default().Error("Failed to open file", "path", path, "err", err)
		return err
	}
	defer func(f *os.Root) {
		err := f.Close()
		if err != nil {
			hclog.Default().Error("Failed to close file", "path", path, "err", err)
		}
	}(r)

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

	if !filepath.IsAbs(rootDir) {
		// we will need logic here to check if the rootDir is absolute and identify the rootDir based on the context or
		// a configuration file. For now, we will return an error.
		hclog.Default().Warn("rootDir is not absolute", "rootDir", rootDir)
		return nil, fmt.Errorf("rootDir is not absolute: %s", rootDir)
	}
	rel, err := absToRel(rootDir, path)
	if err != nil {
		hclog.Default().Error("Failed to get relative path", "rootDir", rootDir, "path", path, "err", err)
		return nil, err
	}

	file, err := os.OpenInRoot(rootDir, rel)
	if err != nil {
		fmt.Println(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(file)

	return file.Stat()
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

// FileCreate creates a new file at the specified path within the host file system and returns a FileHandle or an error.
func (hf *HostFS) FileCreate(ctx context.Context, rootDir, path string) (FileHandle, error) {
	clientID := getClientIDFromContext(ctx)

	if !filepath.IsAbs(rootDir) {
		// we will need logic here to check if the rootDir is absolute and identify the rootDir based on the context or
		// a configuration file. For now, we will return an error.
		hclog.Default().Warn("rootDir is not absolute", "rootDir", rootDir)
		return "", fmt.Errorf("rootDir is not absolute: %s", rootDir)
	}
	rel, err := absToRel(rootDir, path)
	if err != nil {
		hclog.Default().Error("Failed to get relative path", "rootDir", rootDir, "path", path, "err", err)
		return "", err
	}

	r, err := os.OpenRoot(rootDir)
	if err != nil {
		hclog.Default().Error("Failed to open file", "rootDir", rootDir, "path", path, "err", err)
		return "", err
	}
	defer func(r *os.Root) {
		err := r.Close()
		if err != nil {
			hclog.Default().Error("Failed to close file", "rootDir", rootDir, "path", path, "err", err)
		}
	}(r)
	file, err := r.Create(rel)
	if err != nil {
		hclog.Default().Error("Failed to create file", "rootDir", rootDir, "path", path, "err", err)
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
	if !filepath.IsAbs(rootDir) {
		// we will need logic here to check if the rootDir is absolute and identify the rootDir based on the context or
		// a configuration file. For now, we will return an error.
		hclog.Default().Warn("rootDir is not absolute", "rootDir", rootDir)
		return "", 0, fmt.Errorf("rootDir is not absolute: %s", rootDir)
	}

	rel, err := absToRel(rootDir, path)
	if err != nil {
		hclog.Default().Error("Failed to get relative path", "rootDir", rootDir, "path", path, "err", err)
		return "", 0, err
	}

	r, err := os.OpenRoot(rootDir)
	if err != nil {
		hclog.Default().Error("Failed to open file", "rootDir", rootDir, "path", path, "err", err)
		return "", 0, err
	}
	defer func(r *os.Root) {
		err := r.Close()
		if err != nil {
			hclog.Default().Error("Failed to close file", "rootDir", rootDir, "path", path, "err", err)
		}
	}(r)
	file, err := r.OpenFile(rel, flag, perm)
	if err != nil {
		hclog.Default().Error("Failed to create file", "rootDir", rootDir, "path", path, "err", err)
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
	fp, err := filepath.Abs(file.Name())
	if err != nil {
		fp = file.Name()
	}
	hf.getTempPaths().AddPath(clientID, fp)

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
