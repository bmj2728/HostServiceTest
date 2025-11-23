package hostserve

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/bmj2728/hst/shared/protogen/hostserve/v1"
	"github.com/hashicorp/go-hclog"
	"google.golang.org/protobuf/proto"
)

var (
	ErrInvalidHostFS = errors.New("invalid host file system")
)

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

func (s *HostServiceGRPCServer) WriteFile(ctx context.Context,
	request *hostservev1.WriteFileRequest,
) (*hostservev1.WriteFileResponse, error) {

	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)
	if err != nil {
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

func (s *HostServiceGRPCServer) FileOpen(ctx context.Context,
	request *hostservev1.FileOpenRequest,
) (*hostservev1.FileOpenResponse, error) {

	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)
	if err != nil {
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

func (s *HostServiceGRPCServer) FileClose(ctx context.Context,
	request *hostservev1.FileCloseRequest,
) (*hostservev1.FileCloseResponse, error) {

	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)
	if err != nil {
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
