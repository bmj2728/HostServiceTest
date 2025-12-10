package hostserve

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/bmj2728/hst/shared/protogen/hostserve/v1"
	"github.com/hashicorp/go-hclog"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

var (
	ErrInvalidHostFS = errors.New("invalid host file system")
)

// GetHostFS retrieves the HostFS instance from the internal host services or returns an error if unavailable or invalid.
func (s *HostServiceGRPCServer) GetHostFS() (*HostFS, error) {
	hs, err := s.getHostServices()
	if err != nil {
		return nil, err
	}
	hfs, ok := hs.IHostFS.(*HostFS)
	if !ok {
		return nil, ErrInvalidHostFS
	}
	return hfs, nil
}

// ReadDir processes a gRPC request to read contents of a directory specified by the request path and returns
// the results.
func (s *HostServiceGRPCServer) ReadDir(ctx context.Context,
	request *hostservev1.ReadDirRequest,
) (*hostservev1.ReadDirResponse, error) {

	// Use the generic wrapper that handles logging, metrics, and context processing.
	// The handler lambda reads the directory and converts the entries to protobuf format.
	response, err := withRequestLoggingAndResponse(s, ctx, "ReadDir", request,
		func(ctx context.Context, req *hostservev1.ReadDirRequest) (*hostservev1.ReadDirResponse, error) {
			// Call the implementation to read the directory
			entries, implErr := s.Impl.ReadDir(ctx, req.RootDir, req.Path)
			if implErr != nil {
				return &hostservev1.ReadDirResponse{
					Entries: nil,
					Error:   proto.String(implErr.Error()),
				}, nil
			}

			// Convert fs.DirEntry to protobuf DirEntry
			var pbEntries []*hostservev1.DirEntry
			for _, entry := range entries {
				pbEntries = append(pbEntries, &hostservev1.DirEntry{
					Name:  entry.Name(),
					IsDir: entry.IsDir(),
				})
			}

			return &hostservev1.ReadDirResponse{
				Entries: pbEntries,
			}, nil
		})

	// Handle context validation errors
	if err != nil {
		return &hostservev1.ReadDirResponse{
			Entries: nil,
			Error:   proto.String(err.Error()),
		}, nil
	}

	return response, nil
}

// ReadFile handles a gRPC request to read a specific file from a specified directory and returns its contents
// or an error.
func (s *HostServiceGRPCServer) ReadFile(ctx context.Context,
	request *hostservev1.ReadFileRequest,
) (*hostservev1.ReadFileResponse, error) {

	response, err := withRequestLoggingAndResponse(s, ctx, "ReadFile", request,
		func(ctx context.Context, req *hostservev1.ReadFileRequest) (*hostservev1.ReadFileResponse, error) {
			bytes, implErr := s.Impl.ReadFile(ctx, req.RootDir, req.Path)
			if implErr != nil {
				return &hostservev1.ReadFileResponse{
					Contents: nil,
					Error:    proto.String(implErr.Error()),
				}, nil
			}
			return &hostservev1.ReadFileResponse{
				Contents: bytes,
			}, nil
		})

	if err != nil {
		return &hostservev1.ReadFileResponse{
			Contents: nil,
			Error:    proto.String(err.Error()),
		}, nil
	}
	return response, nil
}

// WriteFile handles a client request to write data to a file at a specified path with given permissions.
//
// Parameters:
// - ctx: The context for managing request deadlines and cancellations.
// - request: Contains the file path, data to be written, and file permissions.
//
// Returns:
// - A WriteFileResponse containing an error message if the operation fails, or an empty response on success.
func (s *HostServiceGRPCServer) WriteFile(ctx context.Context,
	request *hostservev1.WriteFileRequest,
) (*hostservev1.WriteFileResponse, error) {

	response, err := withRequestLoggingAndResponse(s, ctx, "WriteFile", request,
		func(ctx context.Context, req *hostservev1.WriteFileRequest) (*hostservev1.WriteFileResponse, error) {
			implErr := s.Impl.WriteFile(ctx, req.RootDir, req.Path, req.Data, os.FileMode(req.Perm))
			if implErr != nil {
				return &hostservev1.WriteFileResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.WriteFileResponse{}, nil
		})

	if err != nil {
		return &hostservev1.WriteFileResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}

// Stat handles a request to retrieve file information for a specified path and root directory from the host system.
// Returns a StatResponse containing file information or an error message. Accepts context and StatRequest as parameters.
func (s *HostServiceGRPCServer) Stat(ctx context.Context, request *hostservev1.StatRequest) (*hostservev1.StatResponse, error) {

	response, err := withRequestLoggingAndResponse(s, ctx, "Stat", request,
		func(ctx context.Context, req *hostservev1.StatRequest) (*hostservev1.StatResponse, error) {
			info, implErr := s.Impl.Stat(ctx, req.RootDir, req.Path)
			if implErr != nil {
				return &hostservev1.StatResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.StatResponse{Info: fileInfoToProtoFileInfo(info)}, nil
		})

	if err != nil {
		return &hostservev1.StatResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}

// Rename handles renaming of a resource by processing the context and forwarding the request to the underlying implementation.
func (s *HostServiceGRPCServer) Rename(ctx context.Context, request *hostservev1.RenameRequest) (*hostservev1.RenameResponse, error) {
	response, err := withRequestLoggingAndResponse(s, ctx, "Rename", request,
		func(ctx context.Context, req *hostservev1.RenameRequest) (*hostservev1.RenameResponse, error) {
			implErr := s.Impl.Rename(ctx, req.RootDir, req.OldName, req.NewName)
			if implErr != nil {
				return &hostservev1.RenameResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.RenameResponse{}, nil
		})

	if err != nil {
		return &hostservev1.RenameResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}

// Remove handles a client request to remove a specified path under a root directory and returns a response or an error.
func (s *HostServiceGRPCServer) Remove(ctx context.Context, request *hostservev1.RemoveRequest) (*hostservev1.RemoveResponse, error) {
	response, err := withRequestLoggingAndResponse(s, ctx, "Remove", request,
		func(ctx context.Context, req *hostservev1.RemoveRequest) (*hostservev1.RemoveResponse, error) {
			implErr := s.Impl.Remove(ctx, req.RootDir, req.Path)
			if implErr != nil {
				return &hostservev1.RemoveResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.RemoveResponse{}, nil
		})

	if err != nil {
		return &hostservev1.RemoveResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}

// RemoveAll removes all resources specified by the RootDir and Path fields in the given request.
// Returns a response with an error message if the operation fails.
func (s *HostServiceGRPCServer) RemoveAll(ctx context.Context, request *hostservev1.RemoveAllRequest) (*hostservev1.RemoveAllResponse, error) {
	response, err := withRequestLoggingAndResponse(s, ctx, "RemoveAll", request,
		func(ctx context.Context, req *hostservev1.RemoveAllRequest) (*hostservev1.RemoveAllResponse, error) {
			implErr := s.Impl.RemoveAll(ctx, req.RootDir, req.Path)
			if implErr != nil {
				return &hostservev1.RemoveAllResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.RemoveAllResponse{}, nil
		})

	if err != nil {
		return &hostservev1.RemoveAllResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}

// Mkdir handles a gRPC request to create a new directory at the specified root directory with the given name and permissions.
func (s *HostServiceGRPCServer) Mkdir(ctx context.Context,
	request *hostservev1.MkdirRequest,
) (*hostservev1.MkdirResponse, error) {

	response, err := withRequestLoggingAndResponse(s, ctx, "Mkdir", request,
		func(ctx context.Context, req *hostservev1.MkdirRequest) (*hostservev1.MkdirResponse, error) {
			implErr := s.Impl.Mkdir(ctx, req.RootDir, req.Name, os.FileMode(req.Perm))
			if implErr != nil {
				return &hostservev1.MkdirResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.MkdirResponse{}, nil
		})

	if err != nil {
		return &hostservev1.MkdirResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}

// MkdirAll handles a request to create a directory hierarchy at the specified path with the given permissions.
func (s *HostServiceGRPCServer) MkdirAll(ctx context.Context, request *hostservev1.MkdirAllRequest) (*hostservev1.MkdirAllResponse, error) {
	response, err := withRequestLoggingAndResponse(s, ctx, "MkdirAll", request,
		func(ctx context.Context, req *hostservev1.MkdirAllRequest) (*hostservev1.MkdirAllResponse, error) {
			implErr := s.Impl.MkdirAll(ctx, req.RootDir, req.Path, os.FileMode(req.Perm))
			if implErr != nil {
				return &hostservev1.MkdirAllResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.MkdirAllResponse{}, nil
		})

	if err != nil {
		return &hostservev1.MkdirAllResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}

// MkdirTemp creates a temporary directory using the provided root directory and pattern, returning its path or an error.
func (s *HostServiceGRPCServer) MkdirTemp(ctx context.Context, request *hostservev1.MkdirTempRequest) (*hostservev1.MkdirTempResponse, error) {
	response, err := withRequestLoggingAndResponse(s, ctx, "MkdirTemp", request,
		func(ctx context.Context, req *hostservev1.MkdirTempRequest) (*hostservev1.MkdirTempResponse, error) {
			dir, implErr := s.Impl.MkdirTemp(ctx, req.RootDir, req.Pattern)
			if implErr != nil {
				return &hostservev1.MkdirTempResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.MkdirTempResponse{Path: dir}, nil
		})

	if err != nil {
		return &hostservev1.MkdirTempResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}

// Chmod modifies the permissions of a file or directory specified in the request using the provided mode.
func (s *HostServiceGRPCServer) Chmod(ctx context.Context, request *hostservev1.ChmodRequest) (*hostservev1.ChmodResponse, error) {
	response, err := withRequestLoggingAndResponse(s, ctx, "Chmod", request,
		func(ctx context.Context, req *hostservev1.ChmodRequest) (*hostservev1.ChmodResponse, error) {
			implErr := s.Impl.Chmod(ctx, req.RootDir, req.Path, os.FileMode(req.Mode))
			if implErr != nil {
				return &hostservev1.ChmodResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.ChmodResponse{}, nil
		})

	if err != nil {
		return &hostservev1.ChmodResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}

// Chown changes the owner and group of a specified file or directory to the provided UID and GID.
// It processes request context information such as client ID and request ID for logging and validation.
// Returns a ChownResponse with an error message in case of failure or an empty response on success.
func (s *HostServiceGRPCServer) Chown(ctx context.Context, request *hostservev1.ChownRequest) (*hostservev1.ChownResponse, error) {
	// Use the generic wrapper that handles logging, metrics, and context processing.
	// The handler lambda performs the actual ownership change operation.
	response, err := withRequestLoggingAndResponse(s, ctx, "Chown", request,
		func(ctx context.Context, req *hostservev1.ChownRequest) (*hostservev1.ChownResponse, error) {
			// Execute the ownership change operation
			implErr := s.Impl.Chown(ctx, req.RootDir, req.Path, int(req.Uid), int(req.Gid))
			if implErr != nil {
				return &hostservev1.ChownResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.ChownResponse{}, nil
		})

	// Handle context validation errors (bad clientID, missing requestID, etc.)
	if err != nil {
		return &hostservev1.ChownResponse{Error: proto.String(err.Error())}, nil
	}

	// Return the response from the handler
	return response, nil
}

// Chtimes handles a request to change the access and modification times of a file at the given path within the specified root directory.
// It takes a context, a request containing the root directory, file path, and new timestamps, and returns a response or an error.
// Logs details about the incoming request including client ID, owner, request ID, and timestamps.
// Processes the request using the underlying implementation and returns a response indicating success or the encountered error.
func (s *HostServiceGRPCServer) Chtimes(ctx context.Context, request *hostservev1.ChtimesRequest) (*hostservev1.ChtimesResponse, error) {

	response, err := withRequestLoggingAndResponse(s, ctx, "Chtimes", request,
		func(ctx context.Context, req *hostservev1.ChtimesRequest) (*hostservev1.ChtimesResponse, error) {
			implErr := s.Impl.Chtimes(ctx, req.RootDir, req.Path, req.Atime.AsTime(), req.Mtime.AsTime())
			if implErr != nil {
				return &hostservev1.ChtimesResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.ChtimesResponse{}, nil
		})

	if err != nil {
		return &hostservev1.ChtimesResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}

func (s *HostServiceGRPCServer) Lchown(ctx context.Context, request *hostservev1.LchownRequest) (*hostservev1.LchownResponse, error) {
	response, err := withRequestLoggingAndResponse(s, ctx, "Lchown", request,
		func(ctx context.Context, req *hostservev1.LchownRequest) (*hostservev1.LchownResponse, error) {
			implErr := s.Impl.Lchown(ctx, req.RootDir, req.Path, int(req.Uid), int(req.Gid))
			if implErr != nil {
				return &hostservev1.LchownResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.LchownResponse{}, nil
		})

	if err != nil {
		return &hostservev1.LchownResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}

// Lstat handles a request to retrieve file or directory metadata based on a specified root directory and path.
// It validates the context, processes the request parameters, and invokes the underlying implementation of Lstat.
// Returns metadata in an LstatResponse or an error message if the operation fails.
func (s *HostServiceGRPCServer) Lstat(ctx context.Context, request *hostservev1.LstatRequest) (*hostservev1.LstatResponse, error) {

	response, err := withRequestLoggingAndResponse(s, ctx, "Lstat", request,
		func(ctx context.Context, req *hostservev1.LstatRequest) (*hostservev1.LstatResponse, error) {
			info, implErr := s.Impl.Lstat(ctx, req.RootDir, req.Path)
			if implErr != nil {
				return &hostservev1.LstatResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.LstatResponse{Info: fileInfoToProtoFileInfo(info)}, nil
		})

	if err != nil {
		return &hostservev1.LstatResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}

// Readlink processes a request to resolve a symbolic link at the specified path relative to the given root directory.
// Returns the resolved destination of the symbolic link or an error if the operation fails.
func (s *HostServiceGRPCServer) Readlink(ctx context.Context, request *hostservev1.ReadlinkRequest) (*hostservev1.ReadlinkResponse, error) {
	response, err := withRequestLoggingAndResponse(s, ctx, "Readlink", request,
		func(ctx context.Context, req *hostservev1.ReadlinkRequest) (*hostservev1.ReadlinkResponse, error) {
			link, implErr := s.Impl.Readlink(ctx, req.RootDir, req.Path)
			if implErr != nil {
				return &hostservev1.ReadlinkResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.ReadlinkResponse{Destination: link}, nil
		})

	if err != nil {
		return &hostservev1.ReadlinkResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}

// Link processes a request to create a link between an old and a new path within a specified root directory.
// It handles the request context, logs relevant data, and delegates the operation to the underlying service implementation.
// Returns a LinkResponse indicating success or an error message if the operation fails.
func (s *HostServiceGRPCServer) Link(ctx context.Context, request *hostservev1.LinkRequest) (*hostservev1.LinkResponse, error) {
	response, err := withRequestLoggingAndResponse(s, ctx, "Link", request,
		func(ctx context.Context, req *hostservev1.LinkRequest) (*hostservev1.LinkResponse, error) {
			implErr := s.Impl.Link(ctx, req.RootDir, req.OldPath, req.NewPath)
			if implErr != nil {
				return &hostservev1.LinkResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.LinkResponse{}, nil
		})

	if err != nil {
		return &hostservev1.LinkResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}

// Symlink handles a request to create a symbolic link from OldPath to NewPath within the specified RootDir.
// Returns a response containing an error message if the operation fails.
func (s *HostServiceGRPCServer) Symlink(ctx context.Context, request *hostservev1.SymlinkRequest) (*hostservev1.SymlinkResponse, error) {
	response, err := withRequestLoggingAndResponse(s, ctx, "Symlink", request,
		func(ctx context.Context, req *hostservev1.SymlinkRequest) (*hostservev1.SymlinkResponse, error) {
			implErr := s.Impl.Symlink(ctx, req.RootDir, req.OldPath, req.NewPath)
			if implErr != nil {
				return &hostservev1.SymlinkResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.SymlinkResponse{}, nil
		})

	if err != nil {
		return &hostservev1.SymlinkResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}

// FileCreate is a method that processes a request to create a new file with the specified path and returns its handle.
func (s *HostServiceGRPCServer) FileCreate(ctx context.Context,
	request *hostservev1.FileCreateRequest) (response *hostservev1.FileCreateResponse, err error) {

	res, validationErr := withRequestLoggingAndResponse(s, ctx, "FileCreate", request,
		func(ctx context.Context, req *hostservev1.FileCreateRequest) (*hostservev1.FileCreateResponse, error) {
			fh, implErr := s.Impl.FileCreate(ctx, req.RootDir, req.Path)
			if implErr != nil {
				return &hostservev1.FileCreateResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.FileCreateResponse{Handle: fh.String()}, nil
		})

	if validationErr != nil {
		return &hostservev1.FileCreateResponse{Error: proto.String(validationErr.Error())}, nil
	}
	return res, nil
}

// FileCreateTemp creates a temporary file in the specified root directory with the given pattern.
// Returns a file handle on success or an error message on failure.
// Logs request details and errors for monitoring and debugging purposes.
func (s *HostServiceGRPCServer) FileCreateTemp(ctx context.Context,
	request *hostservev1.FileCreateTempRequest) (response *hostservev1.FileCreateTempResponse, err error) {

	res, validationErr := withRequestLoggingAndResponse(s, ctx, "FileCreateTemp", request,
		func(ctx context.Context, req *hostservev1.FileCreateTempRequest) (*hostservev1.FileCreateTempResponse, error) {
			fh, implErr := s.Impl.FileCreateTemp(ctx, req.RootDir, req.Pattern)
			if implErr != nil {
				return &hostservev1.FileCreateTempResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.FileCreateTempResponse{Handle: fh.String()}, nil
		})

	if validationErr != nil {
		return &hostservev1.FileCreateTempResponse{Error: proto.String(validationErr.Error())}, nil
	}
	return res, nil
}

// FileOpen handles a request to open a file on the server with specified path, mode, and permissions.
// It returns a file handle and size upon success or an error if the operation fails.
func (s *HostServiceGRPCServer) FileOpen(ctx context.Context,
	request *hostservev1.FileOpenRequest,
) (*hostservev1.FileOpenResponse, error) {

	response, err := withRequestLoggingAndResponse(s, ctx, "FileOpen", request,
		func(ctx context.Context, req *hostservev1.FileOpenRequest) (*hostservev1.FileOpenResponse, error) {
			fh, size, implErr := s.Impl.FileOpen(ctx, req.RootDir, req.Path, fromOpenFileFLags(req.Flags), os.FileMode(req.Perm))
			if implErr != nil {
				return &hostservev1.FileOpenResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.FileOpenResponse{Handle: fh.String(), Size: size}, nil
		})

	if err != nil {
		return &hostservev1.FileOpenResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}

// FileStat handles a request to retrieve metadata for a file identified by a handle and responds with file information.
func (s *HostServiceGRPCServer) FileStat(ctx context.Context, request *hostservev1.FileStatRequest) (*hostservev1.FileStatResponse, error) {

	fh := FileHandle(request.Handle)

	response, err := withRequestLoggingAndResponse(s, ctx, "FileStat", request,
		func(ctx context.Context, req *hostservev1.FileStatRequest) (*hostservev1.FileStatResponse, error) {
			info, implErr := s.Impl.FileStat(ctx, fh)
			if implErr != nil {
				return &hostservev1.FileStatResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.FileStatResponse{Info: fileInfoToProtoFileInfo(info)}, nil
		})

	if err != nil {
		return &hostservev1.FileStatResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}

// FileSeek handles a file seek operation by delegating the request to the implementation and returns the new offset.
func (s *HostServiceGRPCServer) FileSeek(ctx context.Context, request *hostservev1.FileSeekRequest) (*hostservev1.FileSeekResponse, error) {
	fh := FileHandle(request.Handle)

	response, err := withRequestLoggingAndResponse(s, ctx, "FileSeek", request,
		func(ctx context.Context, req *hostservev1.FileSeekRequest) (*hostservev1.FileSeekResponse, error) {
			newOffset, implErr := s.Impl.FileSeek(ctx, fh, int64(req.Offset), int(req.Whence))
			if implErr != nil {
				return &hostservev1.FileSeekResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.FileSeekResponse{NewOffset: uint64(newOffset)}, nil
		})

	if err != nil {
		return &hostservev1.FileSeekResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}

// FileSync synchronizes a file based on the given file handle in the request and returns the operation result.
func (s *HostServiceGRPCServer) FileSync(ctx context.Context, request *hostservev1.FileSyncRequest) (*hostservev1.FileSyncResponse, error) {

	fh := FileHandle(request.Handle)

	response, err := withRequestLoggingAndResponse(s, ctx, "FileSync", request,
		func(ctx context.Context, req *hostservev1.FileSyncRequest) (*hostservev1.FileSyncResponse, error) {
			implErr := s.Impl.FileSync(ctx, fh)
			if implErr != nil {
				return &hostservev1.FileSyncResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.FileSyncResponse{}, nil
		})

	if err != nil {
		return &hostservev1.FileSyncResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}

// FileClose handles the request to close a file identified by its handle sent by the client over gRPC.
// It verifies the request context, extracts relevant metadata, and invokes the implementation's FileClose method.
// Returns a FileCloseResponse containing an error message if the operation fails or an empty response on success.
func (s *HostServiceGRPCServer) FileClose(ctx context.Context,
	request *hostservev1.FileCloseRequest,
) (*hostservev1.FileCloseResponse, error) {

	fh := FileHandle(request.Handle)

	response, err := withRequestLoggingAndResponse(s, ctx, "FileClose", request,
		func(ctx context.Context, req *hostservev1.FileCloseRequest) (*hostservev1.FileCloseResponse, error) {
			implErr := s.Impl.FileClose(ctx, fh)
			if implErr != nil {
				return &hostservev1.FileCloseResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.FileCloseResponse{}, nil
		})

	if err != nil {
		return &hostservev1.FileCloseResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}

// FileTruncate processes a request to truncate a file to the specified size and returns an error if the operation fails.
func (s *HostServiceGRPCServer) FileTruncate(ctx context.Context, request *hostservev1.FileTruncateRequest) (*hostservev1.FileTruncateResponse, error) {

	fh := FileHandle(request.Handle)

	response, err := withRequestLoggingAndResponse(s, ctx, "FileTruncate", request,
		func(ctx context.Context, req *hostservev1.FileTruncateRequest) (*hostservev1.FileTruncateResponse, error) {
			implErr := s.Impl.FileTruncate(ctx, fh, req.Size)
			if implErr != nil {
				return &hostservev1.FileTruncateResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.FileTruncateResponse{}, nil
		})

	if err != nil {
		return &hostservev1.FileTruncateResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}

// FileChmod processes a file mode change request by applying the specified permissions to the identified file handle.
func (s *HostServiceGRPCServer) FileChmod(ctx context.Context, request *hostservev1.FileChmodRequest) (*hostservev1.FileChmodResponse, error) {

	fh := FileHandle(request.Handle)

	response, err := withRequestLoggingAndResponse(s, ctx, "FileChmod", request,
		func(ctx context.Context, req *hostservev1.FileChmodRequest) (*hostservev1.FileChmodResponse, error) {
			implErr := s.Impl.FileChmod(ctx, fh, os.FileMode(req.Mode))
			if implErr != nil {
				return &hostservev1.FileChmodResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.FileChmodResponse{}, nil
		})

	if err != nil {
		return &hostservev1.FileChmodResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}

// FileChown handles a request to change the ownership of a file identified by its handle, using supplied UID and GID.
func (s *HostServiceGRPCServer) FileChown(ctx context.Context, request *hostservev1.FileChownRequest) (*hostservev1.FileChownResponse, error) {

	fh := FileHandle(request.Handle)

	response, err := withRequestLoggingAndResponse(s, ctx, "FileChown", request,
		func(ctx context.Context, req *hostservev1.FileChownRequest) (*hostservev1.FileChownResponse, error) {
			implErr := s.Impl.FileChown(ctx, fh, int(req.Uid), int(req.Gid))
			if implErr != nil {
				return &hostservev1.FileChownResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.FileChownResponse{}, nil
		})

	if err != nil {
		return &hostservev1.FileChownResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}

// FileReader streams chunks of a file to the client based on the given file handle and chunk size in the request.
// The method validates client metadata, fetches the file, reads it in chunks, and streams the data until EOF or error.
// It sends an error response back to the client if any issues occur during validation, reading, or streaming.
func (s *HostServiceGRPCServer) FileReader(request *hostservev1.FileReadRequest,
	stream grpc.ServerStreamingServer[hostservev1.FileReadResponse],
) error {

	// Get file handle
	fh := FileHandle(request.Handle)

	// Use the streaming wrapper to handle context processing, logging, and metrics
	return withServerStreamLogging(s, stream, "FileReader", request,
		func(ctx context.Context, clientID ClientID, reqID RequestID, owner string, logFields []interface{}) error {
			// Get the file from the open files map
			reader, err := s.Impl.FileReader(ctx, fh, request.ChunkSize)
			if err != nil {
				return stream.Send(&hostservev1.FileReadResponse{Error: proto.String(err.Error())})
			}

			// A buffer to read data into
			buffer := make([]byte, request.ChunkSize)
			// Offset is the number of bytes read from the file so far
			var offset uint64

			// Loop by chunk size until EOF
			for {
				// Read from the file into the buffer
				n, err := reader.Read(buffer)
				// Check the bytes read
				if n > 0 {
					// set isFinal to true if we've reached EOF
					isFinal := err == io.EOF
					// Send the chunk
					if sendErr := stream.Send(&hostservev1.FileReadResponse{
						Chunk: &hostservev1.FileChunk{
							Data:    buffer[:n],
							Offset:  offset,
							IsFinal: isFinal,
						},
					}); sendErr != nil {
						// break if we have an error sending the chunk
						return sendErr
					}
					// update the offset
					offset += uint64(n)
				}
				// break if we have reached EOF
				if err == io.EOF {
					return nil
				}
				// break if we have an error reading from the file - propagating to the client
				if err != nil {
					return stream.Send(&hostservev1.FileReadResponse{
						Error: proto.String(err.Error()),
					})
				}
			}
		})
}

// FileWriter handles client stream requests to write a file, validating chunks and writing data incrementally to storage.
func (s *HostServiceGRPCServer) FileWriter(stream grpc.ClientStreamingServer[hostservev1.FileWriteRequest, hostservev1.FileWriteResponse]) error {

	// Use the client streaming wrapper to handle context processing, logging, and metrics
	return withClientStreamLogging(s, stream, "FileWriter",
		func(ctx context.Context, clientID ClientID, reqID RequestID, owner string, logFields []interface{}, firstReq *hostservev1.FileWriteRequest, st grpc.ClientStreamingServer[hostservev1.FileWriteRequest, hostservev1.FileWriteResponse]) error {
			totalBytes := uint32(0)

			// Get handle from first request
			handle := FileHandle(firstReq.Handle)

			// Get the file writer
			file, err := s.Impl.FileWriter(ctx, handle)
			if err != nil {
				return st.SendAndClose(&hostservev1.FileWriteResponse{Error: proto.String(err.Error())})
			}
			info, err := file.(*os.File).Stat()
			if err != nil {
				return st.SendAndClose(&hostservev1.FileWriteResponse{Error: proto.String(err.Error())})
			}

			// Log additional file info
			hclog.Default().Debug("FileWriter processing",
				append(logFields, "handle", handle, "file", info.Name())...)

			// Write the first chunk
			if firstReq.Chunk != nil && len(firstReq.Chunk.Data) > 0 {
				n, err := file.Write(firstReq.Chunk.Data)
				if err != nil {
					return st.SendAndClose(&hostservev1.FileWriteResponse{
						Error: proto.String(err.Error()),
					})
				}
				hclog.Default().Debug("Wrote chunk", "bytes", n, "handle", handle)
				totalBytes += uint32(n)
			}

			// Now we're grooving - loop until EOF
			for {
				req, err := st.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					return st.SendAndClose(&hostservev1.FileWriteResponse{
						Error: proto.String(err.Error()),
					})
				}

				// Validate handle consistency
				if FileHandle(req.Handle) != handle {
					return st.SendAndClose(&hostservev1.FileWriteResponse{
						Error: proto.String(fmt.Sprintf("handle mismatch: expected %s, got %s", handle, req.Handle)),
					})
				}

				// Write chunk
				if req.Chunk != nil && len(req.Chunk.Data) > 0 {
					n, err := file.Write(req.Chunk.Data)
					if err != nil {
						return st.SendAndClose(&hostservev1.FileWriteResponse{
							Error: proto.String(err.Error()),
						})
					}
					hclog.Default().Debug("Wrote chunk", "bytes", n, "handle", handle)
					totalBytes += uint32(n)
				}

				// Check if final
				if req.Chunk != nil && req.Chunk.IsFinal {
					break
				}
			}

			// Sync the file
			err = file.(*os.File).Sync()
			if err != nil {
				return st.SendAndClose(&hostservev1.FileWriteResponse{
					Error: proto.String(fmt.Sprintf("failed to sync file: %v", err)),
				})
			}

			return st.SendAndClose(&hostservev1.FileWriteResponse{
				BytesWritten: totalBytes,
			})
		})
}
