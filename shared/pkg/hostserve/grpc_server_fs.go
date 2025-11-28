package hostserve

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

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

	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)
	if err != nil {
		hclog.Default().Info("ReadDir bad request from client",
			ctxClientIDKey, clientID,
			ctxClientOwner, owner,
			ctxHostRequestIDKey, reqID,
			"error", err,
		)
		return &hostservev1.ReadDirResponse{
			Entries: nil,
			Error:   proto.String(err.Error()),
		}, nil
	}

	ap, err := filepath.Abs(request.Path)
	if err != nil {
		ap = request.Path
	}

	hclog.Default().Info("ReadDir request from client",
		ctxClientIDKey, clientID,
		ctxClientOwner, owner,
		ctxHostRequestIDKey, reqID,
		"path", ap)

	entries, err := s.Impl.ReadDir(ctx, request.Path)
	if err != nil {
		return &hostservev1.ReadDirResponse{
			Entries: nil,
			Error:   proto.String(err.Error()),
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
}

// ReadFile handles a gRPC request to read a specific file from a specified directory and returns its contents
// or an error.
func (s *HostServiceGRPCServer) ReadFile(ctx context.Context,
	request *hostservev1.ReadFileRequest,
) (*hostservev1.ReadFileResponse, error) {

	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)
	if err != nil {
		hclog.Default().Info("ReadFile bad request from client",
			ctxClientIDKey, clientID,
			ctxClientOwner, owner,
			ctxHostRequestIDKey, reqID,
			"error", err,
		)
		return &hostservev1.ReadFileResponse{
			Contents: nil,
			Error:    proto.String(err.Error()),
		}, nil
	}

	ap, err := filepath.Abs(request.Path)
	if err != nil {
		ap = request.Path
	}
	hclog.Default().Info("ReadFile request from client",
		ctxClientIDKey, clientID,
		ctxClientOwner, owner,
		ctxHostRequestIDKey, reqID,
		"path", ap)

	bytes, err := s.Impl.ReadFile(ctx, request.Path)
	if err != nil {
		return &hostservev1.ReadFileResponse{
			Contents: nil,
			Error:    proto.String(err.Error()),
		}, nil
	}
	return &hostservev1.ReadFileResponse{
		Contents: bytes,
	}, nil
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

	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)
	if err != nil {
		hclog.Default().Info("WriteFile bad request from client",
			ctxClientIDKey, clientID,
			ctxClientOwner, owner,
			ctxHostRequestIDKey, reqID,
			"error", err,
		)
		return &hostservev1.WriteFileResponse{
			Error: proto.String(err.Error()),
		}, nil
	}

	ap, err := filepath.Abs(request.Path)
	if err != nil {
		ap = request.Path
	}
	hclog.Default().Info("WriteFile request from client",
		ctxClientIDKey, clientID,
		ctxClientOwner, owner,
		ctxHostRequestIDKey, reqID,
		"path", ap)

	err = s.Impl.WriteFile(ctx, request.Path, request.Data, os.FileMode(request.Perm))
	if err != nil {
		return &hostservev1.WriteFileResponse{Error: proto.String(err.Error())}, nil
	}
	return &hostservev1.WriteFileResponse{}, nil
}

func (s *HostServiceGRPCServer) Stat(ctx context.Context, request *hostservev1.StatRequest) (*hostservev1.StatResponse, error) {

	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)
	if err != nil {
		hclog.Default().Info("Stat bad request from client",
			ctxClientIDKey, clientID,
			ctxClientOwner, owner,
			ctxHostRequestIDKey, reqID,
			"error", err,
		)
		return &hostservev1.StatResponse{
			Error: proto.String(err.Error()),
		}, nil
	}

	hclog.Default().Info("Stat request from client",
		ctxClientIDKey, clientID,
		ctxClientOwner, owner,
		ctxHostRequestIDKey, reqID,
		"path", request.Path)

	info, err := s.Impl.Stat(ctx, request.Path)
	if err != nil {
		return &hostservev1.StatResponse{Error: proto.String(err.Error())}, nil
	}
	return &hostservev1.StatResponse{Info: fileInfoToProtoFileInfo(info)}, nil
}

// Mkdir handles a gRPC request to create a new directory at the specified root directory with the given name and permissions.
func (s *HostServiceGRPCServer) Mkdir(ctx context.Context,
	request *hostservev1.MkdirRequest,
) (*hostservev1.MkdirResponse, error) {

	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)
	if err != nil {
		hclog.Default().Info("Mkdir bad request from client",
			ctxClientIDKey, clientID,
			ctxClientOwner, owner,
			ctxHostRequestIDKey, reqID,
			"error", err,
		)
		return &hostservev1.MkdirResponse{
			Error: proto.String(err.Error()),
		}, nil
	}

	hclog.Default().Info("Mkdir request from client",
		ctxClientIDKey, clientID,
		ctxClientOwner, owner,
		ctxHostRequestIDKey, reqID,
		"rootDir", request.RootDir,
		"name", request.Name,
		"perm", request.Perm)

	err = s.Impl.Mkdir(ctx, request.RootDir, request.Name, os.FileMode(request.Perm))
	if err != nil {
		return &hostservev1.MkdirResponse{
			Error: proto.String(err.Error()),
		}, nil
	}
	return &hostservev1.MkdirResponse{}, nil
}

// MkdirAll handles a request to create a directory hierarchy at the specified path with the given permissions.
func (s *HostServiceGRPCServer) MkdirAll(ctx context.Context, request *hostservev1.MkdirAllRequest) (*hostservev1.MkdirAllResponse, error) {
	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)
	if err != nil {
		hclog.Default().Info("MkdirAll bad request from client",
			ctxClientIDKey, clientID,
			ctxClientOwner, owner,
			ctxHostRequestIDKey, reqID,
			"error", err,
		)
		return &hostservev1.MkdirAllResponse{
			Error: proto.String(err.Error()),
		}, nil
	}

	hclog.Default().Info("MkdirAll request from client",
		ctxClientIDKey, clientID,
		ctxClientOwner, owner,
		ctxHostRequestIDKey, reqID,
		"rootDir", request.RootDir,
		"path", request.Path,
		"perm", request.Perm)

	err = s.Impl.MkdirAll(ctx, request.RootDir, request.Path, os.FileMode(request.Perm))
	if err != nil {
		return &hostservev1.MkdirAllResponse{
			Error: proto.String(err.Error()),
		}, nil
	}
	return &hostservev1.MkdirAllResponse{}, nil
}

// MkdirTemp creates a temporary directory using the provided root directory and pattern, returning its path or an error.
func (s *HostServiceGRPCServer) MkdirTemp(ctx context.Context, request *hostservev1.MkdirTempRequest) (*hostservev1.MkdirTempResponse, error) {
	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)
	if err != nil {
		hclog.Default().Info("MkdirTemp bad request from client",
			ctxClientIDKey, clientID,
			ctxClientOwner, owner,
			ctxHostRequestIDKey, reqID,
			"error", err,
		)
		return &hostservev1.MkdirTempResponse{
			Error: proto.String(err.Error()),
		}, nil
	}

	hclog.Default().Info("MkdirTemp request from client",
		ctxClientIDKey, clientID,
		ctxClientOwner, owner,
		ctxHostRequestIDKey, reqID,
		"rootDir", request.RootDir,
		"pattern", request.Pattern)

	dir, err := s.Impl.MkdirTemp(ctx, request.RootDir, request.Pattern)
	if err != nil {
		return &hostservev1.MkdirTempResponse{
			Error: proto.String(err.Error()),
		}, nil
	}
	return &hostservev1.MkdirTempResponse{Path: dir}, nil
}

// FileCreate is a method that processes a request to create a new file with the specified path and returns its handle.
func (s *HostServiceGRPCServer) FileCreate(ctx context.Context,
	request *hostservev1.FileCreateRequest) (response *hostservev1.FileCreateResponse, err error) {
	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)
	if err != nil {
		hclog.Default().Info("FileCreate bad request from client",
			ctxClientIDKey, clientID,
			ctxClientOwner, owner,
			ctxHostRequestIDKey, reqID,
			"error", err,
		)
		return &hostservev1.FileCreateResponse{
			Error: proto.String(err.Error()),
		}, nil
	}

	hclog.Default().Info("FileCreate request from client",
		ctxClientIDKey, clientID,
		ctxClientOwner, owner,
		ctxHostRequestIDKey, reqID,
		"path", request.Path)

	fh, err := s.Impl.FileCreate(ctx, request.Path)
	if err != nil {
		return &hostservev1.FileCreateResponse{
			Error: proto.String(err.Error()),
		}, nil
	}
	return &hostservev1.FileCreateResponse{Handle: fh.String()}, nil
}

// FileCreateTemp creates a temporary file in the specified root directory with the given pattern.
// Returns a file handle on success or an error message on failure.
// Logs request details and errors for monitoring and debugging purposes.
func (s *HostServiceGRPCServer) FileCreateTemp(ctx context.Context,
	request *hostservev1.FileCreateTempRequest) (response *hostservev1.FileCreateTempResponse, err error) {

	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)
	if err != nil {
		hclog.Default().Info("FileCreateTemp bad request from client",
			ctxClientIDKey, clientID,
			ctxClientOwner, owner,
			ctxHostRequestIDKey, reqID,
			"error", err,
		)
		return &hostservev1.FileCreateTempResponse{
			Error: proto.String(err.Error()),
		}, nil
	}

	hclog.Default().Info("FileCreateTemp request from client",
		ctxClientIDKey, clientID,
		ctxClientOwner, owner,
		ctxHostRequestIDKey, reqID,
		"rootDir", request.RootDir,
		"pattern", request.Pattern)

	fh, err := s.Impl.FileCreateTemp(ctx, request.RootDir, request.Pattern)
	if err != nil {
		return &hostservev1.FileCreateTempResponse{
			Error: proto.String(err.Error()),
		}, nil
	}

	return &hostservev1.FileCreateTempResponse{Handle: fh.String()}, nil
}

// FileOpen handles a request to open a file on the server with specified path, mode, and permissions.
// It returns a file handle and size upon success or an error if the operation fails.
func (s *HostServiceGRPCServer) FileOpen(ctx context.Context,
	request *hostservev1.FileOpenRequest,
) (*hostservev1.FileOpenResponse, error) {

	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)
	if err != nil {
		hclog.Default().Info("FileOpen bad request from client",
			ctxClientIDKey, clientID,
			ctxClientOwner, owner,
			ctxHostRequestIDKey, reqID,
			"error", err,
		)
		return &hostservev1.FileOpenResponse{
			Error: proto.String(err.Error()),
		}, nil
	}

	ap, err := filepath.Abs(request.Path)
	if err != nil {
		ap = request.Path
	}

	hclog.Default().Info("FileOpen request from client",
		ctxClientIDKey, clientID,
		ctxClientOwner, owner,
		ctxHostRequestIDKey, reqID,
		"path", ap)

	fh, size, err := s.Impl.FileOpen(ctx, request.Path, openFileModeToFlags(request.Mode), os.FileMode(request.Perm))
	if err != nil {
		return &hostservev1.FileOpenResponse{Error: proto.String(err.Error())}, nil
	}

	hclog.Default().Info("File opened successfully",
		ctxClientIDKey, clientID,
		ctxClientOwner, owner,
		ctxHostRequestIDKey, reqID,
		"path", ap,
		"handle", fh,
		"size", size)
	return &hostservev1.FileOpenResponse{Handle: fh.String(), Size: size}, nil
}

func (s *HostServiceGRPCServer) FileStat(ctx context.Context, request *hostservev1.FileStatRequest) (*hostservev1.FileStatResponse, error) {
	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)
	if err != nil {
		hclog.Default().Info("FileSeek bad request from client",
			ctxClientIDKey, clientID,
			ctxClientOwner, owner,
			ctxHostRequestIDKey, reqID,
			"error", err,
		)
		return &hostservev1.FileStatResponse{
			Error: proto.String(err.Error()),
		}, nil
	}

	fh := FileHandle(request.Handle)

	hclog.Default().Info("FileSeek request from client",
		ctxClientIDKey, clientID,
		ctxClientOwner, owner,
		ctxHostRequestIDKey, reqID,
		"handle", fh,
	)

	info, err := s.Impl.FileStat(ctx, fh)
	if err != nil {
		return &hostservev1.FileStatResponse{Error: proto.String(err.Error())}, nil
	}

	return &hostservev1.FileStatResponse{Info: fileInfoToProtoFileInfo(info)}, nil

}

// FileSeek handles a file seek operation by delegating the request to the implementation and returns the new offset.
func (s *HostServiceGRPCServer) FileSeek(ctx context.Context, request *hostservev1.FileSeekRequest) (*hostservev1.FileSeekResponse, error) {

	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)
	if err != nil {
		hclog.Default().Info("FileSeek bad request from client",
			ctxClientIDKey, clientID,
			ctxClientOwner, owner,
			ctxHostRequestIDKey, reqID,
			"error", err,
		)
		return &hostservev1.FileSeekResponse{
			Error: proto.String(err.Error()),
		}, nil
	}

	fh := FileHandle(request.Handle)

	hclog.Default().Info("FileSeek request from client",
		ctxClientIDKey, clientID,
		ctxClientOwner, owner,
		ctxHostRequestIDKey, reqID,
		"handle", fh,
		"offset", request.Offset,
		"whence", request.Whence,
	)

	newOffset, err := s.Impl.FileSeek(ctx, fh, int64(request.Offset), int(request.Whence))
	if err != nil {
		return &hostservev1.FileSeekResponse{
			Error: proto.String(err.Error()),
		}, nil
	}

	return &hostservev1.FileSeekResponse{NewOffset: uint64(newOffset)}, nil
}

func (s *HostServiceGRPCServer) FileSync(ctx context.Context, request *hostservev1.FileSyncRequest) (*hostservev1.FileSyncResponse, error) {
	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)
	if err != nil {
		hclog.Default().Info("FileSync bad request from client",
			ctxClientIDKey, clientID,
			ctxClientOwner, owner,
			ctxHostRequestIDKey, reqID,
			"error", err,
		)
		return &hostservev1.FileSyncResponse{
			Error: proto.String(err.Error()),
		}, nil
	}

	fh := FileHandle(request.Handle)
	hclog.Default().Info("FileSync request from client",
		ctxClientIDKey, clientID,
		ctxClientOwner, owner,
		ctxHostRequestIDKey, reqID,
		"handle", fh)

	err = s.Impl.FileSync(ctx, fh)
	if err != nil {
		return &hostservev1.FileSyncResponse{Error: proto.String(err.Error())}, nil
	}
	return &hostservev1.FileSyncResponse{}, nil
}

// FileClose handles the request to close a file identified by its handle sent by the client over gRPC.
// It verifies the request context, extracts relevant metadata, and invokes the implementation's FileClose method.
// Returns a FileCloseResponse containing an error message if the operation fails or an empty response on success.
func (s *HostServiceGRPCServer) FileClose(ctx context.Context,
	request *hostservev1.FileCloseRequest,
) (*hostservev1.FileCloseResponse, error) {

	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)
	if err != nil {
		hclog.Default().Info("FileClose bad request from client",
			ctxClientIDKey, clientID,
			ctxClientOwner, owner,
			ctxHostRequestIDKey, reqID,
			"error", err,
		)
		return &hostservev1.FileCloseResponse{
			Error: proto.String(err.Error()),
		}, nil
	}

	fh := FileHandle(request.Handle)
	hclog.Default().Info("FileClose request from client",
		ctxClientIDKey, clientID,
		ctxClientOwner, owner,
		ctxHostRequestIDKey, reqID,
		"handle", fh)
	err = s.Impl.FileClose(ctx, fh)
	if err != nil {
		return &hostservev1.FileCloseResponse{Error: proto.String(err.Error())}, nil
	}
	return &hostservev1.FileCloseResponse{}, nil
}

// FileReader streams chunks of a file to the client based on the given file handle and chunk size in the request.
// The method validates client metadata, fetches the file, reads it in chunks, and streams the data until EOF or error.
// It sends an error response back to the client if any issues occur during validation, reading, or streaming.
func (s *HostServiceGRPCServer) FileReader(request *hostservev1.FileReadRequest,
	stream grpc.ServerStreamingServer[hostservev1.FileReadResponse],
) error {

	// Get context
	ctx := stream.Context()

	// Get file handle
	fh := FileHandle(request.Handle)

	// Get request metadata and add the client owner
	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)
	if err != nil {
		// Log error and return if we can't validate this request
		hclog.Default().Error("FileReader request with Error from client",
			ctxClientIDKey, clientID,
			ctxClientOwner, owner,
			ctxHostRequestIDKey, reqID,
			"handle", fh,
			"chunkSize", request.ChunkSize,
			"error", err)

		return stream.Send(&hostservev1.FileReadResponse{
			Error: proto.String(err.Error()),
		})
	}
	// Log request
	hclog.Default().Info("FileReader request from client",
		ctxClientIDKey, clientID,
		ctxClientOwner, owner,
		ctxHostRequestIDKey, reqID,
		"handle", fh,
		"chunkSize", request.ChunkSize)

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
}

// FileWriter handles client stream requests to write a file, validating chunks and writing data incrementally to storage.
func (s *HostServiceGRPCServer) FileWriter(stream grpc.ClientStreamingServer[hostservev1.FileWriteRequest, hostservev1.FileWriteResponse]) error {
	ctx := stream.Context()
	totalBytes := uint32(0)

	// Receive first message to extract metadata and get file
	req, err := stream.Recv()
	if err != nil {
		return err
	}

	// Process context
	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)
	if err != nil {
		hclog.Default().Info("FileWriter bad request from client",
			ctxClientIDKey, clientID,
			ctxClientOwner, owner,
			ctxHostRequestIDKey, reqID,
			"error", err,
		)
		return stream.SendAndClose(&hostservev1.FileWriteResponse{
			Error: proto.String(err.Error()),
		})
	}
	handle := FileHandle(req.Handle)
	file, err := s.Impl.FileWriter(ctx, handle)
	if err != nil {
		return stream.SendAndClose(&hostservev1.FileWriteResponse{Error: proto.String(err.Error())})
	}
	info, err := file.(*os.File).Stat()
	if err != nil {
		return stream.SendAndClose(&hostservev1.FileWriteResponse{Error: proto.String(err.Error())})
	}

	hclog.Default().Info("FileWriter request from client",
		ctxClientIDKey, clientID,
		ctxClientOwner, owner,
		ctxHostRequestIDKey, reqID,
		"handle", handle,
		"file", info.Name())

	// Write the first chunk
	if req.Chunk != nil && len(req.Chunk.Data) > 0 {
		n, err := file.Write(req.Chunk.Data)
		if err != nil {
			return stream.SendAndClose(&hostservev1.FileWriteResponse{
				Error: proto.String(err.Error()),
			})
		}
		hclog.Default().Debug("Wrote chunk", "bytes", n, "handle", handle)
		totalBytes += uint32(n)
	}

	// Now we're grooving - loop until EOF
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return stream.SendAndClose(&hostservev1.FileWriteResponse{
				Error: proto.String(err.Error()),
			})
		}

		// Validate handle consistency
		if FileHandle(req.Handle) != handle {
			return stream.SendAndClose(&hostservev1.FileWriteResponse{
				Error: proto.String(fmt.Sprintf("handle mismatch: expected %s, got %s", handle, req.Handle)),
			})
		}

		// Write chunk
		if req.Chunk != nil && len(req.Chunk.Data) > 0 {
			n, err := file.Write(req.Chunk.Data)
			if err != nil {
				return stream.SendAndClose(&hostservev1.FileWriteResponse{
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
	err = file.(*os.File).Sync()
	if err != nil {
		return stream.SendAndClose(&hostservev1.FileWriteResponse{
			Error: proto.String(fmt.Sprintf("failed to sync file: %v", err)),
		})
	}
	return stream.SendAndClose(&hostservev1.FileWriteResponse{
		BytesWritten: totalBytes,
	})
}
