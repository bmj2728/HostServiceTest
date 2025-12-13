package hostserve

import (
	"context"
	"io"
	"io/fs"
	"os"
	"time"

	"github.com/bmj2728/hst/shared/protogen/hostserve/v1"
	"github.com/hashicorp/go-hclog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ReadDir retrieves a list of directory entries from the given path through a gRPC call to the host service.
// Returns a slice of fs.DirEntry or an error if the operation fails.
func (c *HostServiceGRPCClient) ReadDir(ctx context.Context, rootDir, path string) ([]fs.DirEntry, error) {

	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.ReadDir(ctx, &hostservev1.ReadDirRequest{
		RootDir: rootDir,
		Path:    path,
	})
	if err != nil {
		return nil, &HostServiceError{Message: err.Error()}
	}
	if resp.Error != nil {
		return nil, &HostServiceError{Message: *resp.Error}
	}

	// Convert protobuf DirEntry to fs.DirEntry
	var entries []fs.DirEntry
	for _, entry := range resp.Entries {
		entries = append(entries, &RemoteDirEntry{
			name:  entry.Name,
			isDir: entry.IsDir,
		})
	}

	return entries, nil
}

// ReadFile reads the specified file from the given directory and returns its contents as a byte slice.
// Returns an error if the file cannot be read or the service encounters an issue.
func (c *HostServiceGRPCClient) ReadFile(ctx context.Context, rootDir, path string) ([]byte, error) {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.ReadFile(ctx, &hostservev1.ReadFileRequest{
		RootDir: rootDir,
		Path:    path,
	})
	if err != nil {
		return nil, &HostServiceError{Message: err.Error()}
	}
	if resp.Error != nil {
		return nil, &HostServiceError{Message: *resp.Error}
	}
	return resp.Contents, nil
}

// WriteFile writes data to the specified file path on the server with the given permissions.
// The context is used for tracing and request cancellation.
// If permissions are set to 0, standardPermissions will be applied as the default.
// Returns an error if the operation fails or in case of a nil response from the server.
func (c *HostServiceGRPCClient) WriteFile(ctx context.Context, rootDir, path string, data []byte, perm os.FileMode) error {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	if perm == 0 {
		perm = standardPermissions
	}
	resp, err := c.client.WriteFile(ctx, &hostservev1.WriteFileRequest{
		RootDir: rootDir,
		Path:    path,
		Data:    data,
		Perm:    uint32(perm),
	})
	if err != nil {
		return &HostServiceError{Message: err.Error()}
	}
	// Defensive: handle unexpected nil resp
	if resp == nil {
		return &HostServiceError{Message: "nil response from WriteFile"}
	}
	if resp.Error != nil {
		return &HostServiceError{Message: *resp.Error}
	}
	return nil
}

// Mkdir creates a new directory with the specified name and permissions under the given root directory.
func (c *HostServiceGRPCClient) Mkdir(ctx context.Context, rootDir string, name string, perm os.FileMode) error {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	if perm == 0 {
		perm = standardPermissions
	}
	resp, err := c.client.Mkdir(ctx, &hostservev1.MkdirRequest{
		RootDir: rootDir,
		Name:    name,
		Perm:    uint32(perm),
	})
	if err != nil {
		return &HostServiceError{Message: err.Error()}
	}
	if resp == nil {
		return &HostServiceError{Message: "nil response from Mkdir"}
	}
	if resp.Error != nil {
		return &HostServiceError{Message: *resp.Error}
	}
	return nil
}

// MkdirAll creates all necessary directories along the specified path with the provided permissions.
// It operates relative to the given rootDir and uses the optional context for tracing and logging.
func (c *HostServiceGRPCClient) MkdirAll(ctx context.Context, rootDir string, path string, perm os.FileMode) error {

	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())

	if perm == 0 {
		perm = standardPermissions
	}
	resp, err := c.client.MkdirAll(ctx, &hostservev1.MkdirAllRequest{
		RootDir: rootDir,
		Path:    path,
		Perm:    uint32(perm),
	})
	if err != nil {
		return &HostServiceError{Message: err.Error()}
	}
	if resp == nil {
		return &HostServiceError{Message: "nil response from MkdirAll"}
	}
	if resp.Error != nil {
		return &HostServiceError{Message: *resp.Error}
	}
	return nil
}

// Chmod updates the file permissions for a specific path within the root directory using the provided mode.
func (c *HostServiceGRPCClient) Chmod(ctx context.Context, rootDir, path string, mode os.FileMode) error {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	if mode == 0 {
		return &HostServiceError{Message: "invalid mode provided"}
	}
	resp, err := c.client.Chmod(ctx, &hostservev1.ChmodRequest{
		RootDir: rootDir,
		Path:    path,
		Mode:    uint32(mode),
	})
	if err != nil {
		return &HostServiceError{Message: err.Error()}
	}
	if resp == nil {
		return &HostServiceError{Message: "nil response from Chmod"}
	}
	if resp.Error != nil {
		return &HostServiceError{Message: *resp.Error}
	}
	return nil
}

// Chown changes the ownership of a file or directory at the specified path to the provided UID and GID.
func (c *HostServiceGRPCClient) Chown(ctx context.Context, rootDir, path string, uid, gid int) error {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.Chown(ctx, &hostservev1.ChownRequest{
		RootDir: rootDir,
		Path:    path,
		Uid:     int32(uid),
		Gid:     int32(gid),
	})
	if err != nil {
		return &HostServiceError{Message: err.Error()}
	}
	if resp == nil {
		return &HostServiceError{Message: "nil response from Chown"}
	}
	if resp.Error != nil {
		return &HostServiceError{Message: *resp.Error}
	}
	return nil
}

// Chtimes updates the access and modification timestamps of a given file or directory within a specified root directory.
// It sends a gRPC request using the provided context, rootDir, path, atime, and mtime parameters and processes the response.
// Returns an error if the operation fails or the response is invalid.
func (c *HostServiceGRPCClient) Chtimes(ctx context.Context, rootDir, path string, atime, mtime time.Time) error {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.Chtimes(ctx, &hostservev1.ChtimesRequest{
		RootDir: rootDir,
		Path:    path,
		Atime:   timestamppb.New(atime),
		Mtime:   timestamppb.New(mtime),
	})
	if err != nil {
		return &HostServiceError{Message: err.Error()}
	}
	if resp == nil {
		return &HostServiceError{Message: "nil response from Chtimes"}
	}
	if resp.Error != nil {
		return &HostServiceError{Message: *resp.Error}
	}
	return nil
}

// Lchown changes the ownership of a file at the specified path within the given root directory to the provided UID and GID.
// It performs the operation through a gRPC client, handling tracing metadata and errors.
// Returns an error if the operation fails or the response is invalid.
func (c *HostServiceGRPCClient) Lchown(ctx context.Context, rootDir, path string, uid, gid int) error {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.Lchown(ctx, &hostservev1.LchownRequest{
		RootDir: rootDir,
		Path:    path,
		Uid:     int32(uid),
		Gid:     int32(gid),
	})
	if err != nil {
		return &HostServiceError{Message: err.Error()}
	}
	if resp == nil {
		return &HostServiceError{Message: "nil response from Lchown"}
	}
	if resp.Error != nil {
		return &HostServiceError{Message: *resp.Error}
	}
	return nil
}

// Lstat retrieves file or directory information from a remote host at the specified path within the provided root directory.
func (c *HostServiceGRPCClient) Lstat(ctx context.Context, rootDir, path string) (fs.FileInfo, error) {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.Lstat(ctx, &hostservev1.LstatRequest{
		RootDir: rootDir,
		Path:    path,
	})
	if err != nil {
		return nil, &HostServiceError{Message: err.Error()}
	}
	if resp == nil {
		return nil, &HostServiceError{Message: "nil response from Lstat"}
	}
	if resp.Error != nil {
		return nil, &HostServiceError{Message: *resp.Error}
	}
	return protoFileInfoToRemoteFileInfo(resp.Info), nil
}

// Readlink resolves the target of a symbolic link identified by rootDir and path, returning the destination or an error.
func (c *HostServiceGRPCClient) Readlink(ctx context.Context, rootDir, path string) (string, error) {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.Readlink(ctx, &hostservev1.ReadlinkRequest{
		RootDir: rootDir,
		Path:    path,
	})
	if err != nil {
		return "", &HostServiceError{Message: err.Error()}
	}
	if resp == nil {
		return "", &HostServiceError{Message: "nil response from Readlink"}
	}
	if resp.Error != nil {
		return "", &HostServiceError{Message: *resp.Error}
	}
	return resp.Destination, nil
}

// Link creates a new link from oldpath to newpath within the specified rootDir using a gRPC service call.
// It attaches tracing identifiers to the context and returns an error if the operation fails.
func (c *HostServiceGRPCClient) Link(ctx context.Context, rootDir, oldpath, newpath string) error {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.Link(ctx, &hostservev1.LinkRequest{
		RootDir: rootDir,
		OldPath: oldpath,
		NewPath: newpath,
	})
	if err != nil {
		return &HostServiceError{Message: err.Error()}
	}
	if resp == nil {
		return &HostServiceError{Message: "nil response from Link"}
	}
	if resp.Error != nil {
		return &HostServiceError{Message: *resp.Error}
	}
	return nil
}

// Symlink creates a symbolic link from oldpath to newpath within the specified rootDir and returns an error if any occurs.
func (c *HostServiceGRPCClient) Symlink(ctx context.Context, rootDir, oldpath, newpath string) error {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.Symlink(ctx, &hostservev1.SymlinkRequest{
		RootDir: rootDir,
		OldPath: oldpath,
		NewPath: newpath,
	})
	if err != nil {
		return &HostServiceError{Message: err.Error()}
	}
	if resp == nil {
		return &HostServiceError{Message: "nil response from Symlink"}
	}
	if resp.Error != nil {
		return &HostServiceError{Message: *resp.Error}
	}
	return nil
}

// Stat retrieves file or directory information from the remote host based on the provided path.
// Returns fs.FileInfo for the requested path or an error if the operation fails.
// Uses gRPC to communicate with the remote service and adds tracing metadata to the context.
func (c *HostServiceGRPCClient) Stat(ctx context.Context, rootDir, path string) (fs.FileInfo, error) {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.Stat(ctx, &hostservev1.StatRequest{
		RootDir: rootDir,
		Path:    path,
	})
	if err != nil {
		return nil, &HostServiceError{Message: err.Error()}
	}
	if resp == nil {
		return nil, &HostServiceError{Message: "nil response from Stat"}
	}
	if resp.Error != nil {
		return nil, &HostServiceError{Message: *resp.Error}
	}
	return protoFileInfoToRemoteFileInfo(resp.Info), nil
}

// Rename renames a file or directory from oldPath to newPath within the specified rootDir using the gRPC client.
func (c *HostServiceGRPCClient) Rename(ctx context.Context, rootDir, oldPath, newPath string) error {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.Rename(ctx, &hostservev1.RenameRequest{
		RootDir: rootDir,
		OldName: oldPath,
		NewName: newPath,
	})
	if err != nil {
		return &HostServiceError{Message: err.Error()}
	}
	if resp == nil {
		return &HostServiceError{Message: "nil response from Rename"}
	}
	if resp.Error != nil {
		return &HostServiceError{Message: *resp.Error}
	}
	return nil
}

// Remove removes a specified path within a root directory by making a gRPC call to the host service.
// Returns an error if the operation fails or if the response contains an error.
// Context is enhanced with tracing IDs before sending the request.
func (c *HostServiceGRPCClient) Remove(ctx context.Context, rootDir, path string) error {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.Remove(ctx, &hostservev1.RemoveRequest{
		RootDir: rootDir,
		Path:    path,
	})
	if err != nil {
		return &HostServiceError{Message: err.Error()}
	}
	if resp == nil {
		return &HostServiceError{Message: "nil response from Remove"}
	}
	if resp.Error != nil {
		return &HostServiceError{Message: *resp.Error}
	}
	return nil
}

// RemoveAll removes all files and directories within the specified path under the provided root directory via gRPC.
// It expects a context with tracing metadata and returns an error if the operation fails or the response is nil or contains errors.
func (c *HostServiceGRPCClient) RemoveAll(ctx context.Context, rootDir, path string) error {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.RemoveAll(ctx, &hostservev1.RemoveAllRequest{
		RootDir: rootDir,
		Path:    path,
	})
	if err != nil {
		return &HostServiceError{Message: err.Error()}
	}
	if resp == nil {
		return &HostServiceError{Message: "nil response from RemoveAll"}
	}
	if resp.Error != nil {
		return &HostServiceError{Message: *resp.Error}
	}
	return nil
}

// FileCreate creates a new file at the specified path and returns a handle for the created file or an error if any occurs.
func (c *HostServiceGRPCClient) FileCreate(ctx context.Context, rootDir, path string) (FileHandle, error) {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.FileCreate(ctx, &hostservev1.FileCreateRequest{
		RootDir: rootDir,
		Path:    path,
	})
	if err != nil {
		return "", &HostServiceError{Message: err.Error()}
	}
	if resp == nil {
		return "", &HostServiceError{Message: "nil response from FileCreate"}
	}
	if resp.Error != nil {
		return "", &HostServiceError{Message: *resp.Error}
	}
	return FileHandle(resp.Handle), nil
}

// FileOpen opens a file on the host, given the path, flags, and permissions, and returns a file handle, size, and error if any.
func (c *HostServiceGRPCClient) FileOpen(ctx context.Context, rootDir, path string, flag int, perm os.FileMode) (FileHandle, uint64, error) {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.FileOpen(ctx, &hostservev1.FileOpenRequest{
		RootDir: rootDir,
		Path:    path,
		Flags:   toOpenFileFlags(flag),
		Perm:    uint32(perm),
	})
	if err != nil {
		return "", 0, &HostServiceError{Message: err.Error()}
	}
	if resp == nil {
		return "", 0, &HostServiceError{Message: "nil response from FileOpen"}
	}
	if resp.Error != nil {
		return "", 0, &HostServiceError{Message: *resp.Error}
	}
	return FileHandle(resp.Handle), resp.Size, nil
}

// FileStat retrieves file information for a given handle from the remote host and returns it as fs.FileInfo.
func (c *HostServiceGRPCClient) FileStat(ctx context.Context, handle FileHandle) (fs.FileInfo, error) {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.FileStat(ctx, &hostservev1.FileStatRequest{
		Handle: string(handle),
	})
	if err != nil {
		return nil, &HostServiceError{Message: err.Error()}
	}
	if resp == nil {
		return nil, &HostServiceError{Message: "nil response from FileStat"}
	}
	if resp.Error != nil {
		return nil, &HostServiceError{Message: *resp.Error}
	}

	return protoFileInfoToRemoteFileInfo(resp.Info), nil
}

// FileSeek adjusts the file offset for a file identified by handle, based on the specified offset and whence parameters.
// Returns the new file offset or an error if the operation fails.
func (c *HostServiceGRPCClient) FileSeek(ctx context.Context, handle FileHandle, offset int64, whence int) (int64, error) {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.FileSeek(ctx, &hostservev1.FileSeekRequest{
		Handle: string(handle),
		Offset: uint64(offset),
		Whence: uint32(whence),
	})
	if err != nil {
		return 0, &HostServiceError{Message: err.Error()}
	}
	if resp == nil {
		return 0, &HostServiceError{Message: "nil response from FileSeek"}
	}
	if resp.Error != nil {
		return 0, &HostServiceError{Message: *resp.Error}
	}
	return int64(resp.NewOffset), nil
}

// FileSync synchronizes the state of a given file handle with the host service over a gRPC connection.
func (c *HostServiceGRPCClient) FileSync(ctx context.Context, handle FileHandle) error {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.FileSync(ctx, &hostservev1.FileSyncRequest{
		Handle: string(handle),
	})
	if err != nil {
		return &HostServiceError{Message: err.Error()}
	}
	if resp == nil {
		return &HostServiceError{Message: "nil response from FileSync"}
	}
	if resp.Error != nil {
		return &HostServiceError{Message: *resp.Error}
	}
	return nil
}

// FileClose closes a file associated with the provided handle and releases any resources tied to it.
func (c *HostServiceGRPCClient) FileClose(ctx context.Context, handle FileHandle) error {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.FileClose(ctx, &hostservev1.FileCloseRequest{
		Handle: string(handle),
	})
	if err != nil {
		hclog.Default().Error("Failed to close file via host service", "err", err)
		return &HostServiceError{Message: err.Error()}
	}
	if resp == nil {
		hclog.Default().Error("Received nil response from host service when closing file", "handle", handle)
		return &HostServiceError{Message: "nil response from FileClose"}
	}
	if resp.Error != nil {
		hclog.Default().Error("Received error response from host service when closing file", "handle", handle, "err", *resp.Error)
		return &HostServiceError{Message: *resp.Error}
	}
	return nil
}

// FileTruncate truncates a remote file identified by the provided FileHandle to the specified size in bytes.
func (c *HostServiceGRPCClient) FileTruncate(ctx context.Context, handle FileHandle, size int64) error {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.FileTruncate(ctx, &hostservev1.FileTruncateRequest{
		Handle: string(handle),
		Size:   size,
	})
	if err != nil {
		return &HostServiceError{Message: err.Error()}
	}
	if resp == nil {
		return &HostServiceError{Message: "nil response from FileTruncate"}
	}
	if resp.Error != nil {
		return &HostServiceError{Message: *resp.Error}
	}
	return nil
}

// FileChmod changes the permissions of a file identified by the given FileHandle to the specified mode.
func (c *HostServiceGRPCClient) FileChmod(ctx context.Context, handle FileHandle, mode os.FileMode) error {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.FileChmod(ctx, &hostservev1.FileChmodRequest{
		Handle: string(handle),
		Mode:   uint32(mode),
	})
	if err != nil {
		return &HostServiceError{Message: err.Error()}
	}
	if resp == nil {
		return &HostServiceError{Message: "nil response from FileChmod"}
	}
	if resp.Error != nil {
		return &HostServiceError{Message: *resp.Error}
	}
	return nil
}

// FileChown updates the ownership of a file identified by the handle to the given user ID (uid) and group ID (gid).
// It communicates with the host service via gRPC and returns an error if the operation fails.
func (c *HostServiceGRPCClient) FileChown(ctx context.Context, handle FileHandle, uid, gid int) error {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.FileChown(ctx, &hostservev1.FileChownRequest{
		Handle: string(handle),
		Uid:    int32(uid),
		Gid:    int32(gid),
	})
	if err != nil {
		return &HostServiceError{Message: err.Error()}
	}
	if resp == nil {
		return &HostServiceError{Message: "nil response from FileChown"}
	}
	if resp.Error != nil {
		return &HostServiceError{Message: *resp.Error}
	}
	return nil
}

// FileReadSection reads a specified section of an open file identified by the handle, starting at an offset for a given length.
func (c *HostServiceGRPCClient) FileReadSection(ctx context.Context, handle FileHandle, offset int64, length int32) ([]byte, error) {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.FileReadSection(ctx, &hostservev1.FileReadSectionRequest{
		Handle: string(handle),
		Offset: uint64(offset),
		Length: uint32(length),
	})
	if err != nil {
		return nil, &HostServiceError{Message: err.Error()}
	}
	if resp == nil {
		return nil, &HostServiceError{Message: "nil response from FileReadSection"}
	}
	if resp.Error != nil {
		return nil, &HostServiceError{Message: *resp.Error}
	}
	return resp.Contents, nil
}

// FileReader provides a gRPC client-side implementation to read files in chunks via a streaming connection.
func (c *HostServiceGRPCClient) FileReader(ctx context.Context, handle FileHandle, chunkSize uint32) (io.Reader, error) {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	stream, err := c.client.FileReader(ctx, &hostservev1.FileReadRequest{
		Handle:    string(handle),
		ChunkSize: chunkSize,
	})
	if err != nil {
		return nil, &HostServiceError{Message: err.Error()}
	}
	return &grpcFileStreamReader{stream: stream}, nil
}

// FileWriter opens a stream to write file data to a remote host, returning an io.WriteCloser for handling the stream.
// It accepts a context for request lifecycle management and a file handle for identifying the target file.
// Returns an io.WriteCloser on success or an error if the stream initialization fails.
func (c *HostServiceGRPCClient) FileWriter(ctx context.Context, handle FileHandle) (io.WriteCloser, error) {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	stream, err := c.client.FileWriter(ctx)
	if err != nil {
		return nil, &HostServiceError{Message: err.Error()}
	}
	return &grpcFileStreamWriter{stream: stream, handle: handle}, nil
}

// MkdirTemp creates a new temporary directory in the specified root directory with a given pattern and returns its path.
// Returns an error if the operation fails or if there is an issue with the response from the server.
func (c *HostServiceGRPCClient) MkdirTemp(ctx context.Context, rootDir string, pattern string) (string, error) {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.MkdirTemp(ctx, &hostservev1.MkdirTempRequest{
		RootDir: rootDir,
		Pattern: pattern,
	})
	if err != nil {
		return "", &HostServiceError{Message: err.Error()}
	}
	if resp == nil {
		return "", &HostServiceError{Message: "nil response from MkdirTemp"}
	}
	if resp.Error != nil {
		return "", &HostServiceError{Message: *resp.Error}
	}
	return resp.GetPath(), nil
}

// FileCreateTemp creates a temporary file with the specified root directory and pattern using the gRPC client.
// Returns a FileHandle for the created file or an error if the operation fails.
func (c *HostServiceGRPCClient) FileCreateTemp(ctx context.Context, rootDir string, pattern string) (FileHandle, error) {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.FileCreateTemp(ctx, &hostservev1.FileCreateTempRequest{
		RootDir: rootDir,
		Pattern: pattern,
	})
	if err != nil {
		return "", &HostServiceError{Message: err.Error()}
	}
	if resp == nil {
		return "", &HostServiceError{Message: "nil response from FileCreateTemp"}
	}
	if resp.Error != nil {
		return "", &HostServiceError{Message: *resp.Error}
	}
	return FileHandle(resp.Handle), nil
}
