package hostserve

import (
	"context"
	"os"
	"path/filepath"

	"github.com/bmj2728/hst/shared/protogen/hostserve/v1"
	"github.com/hashicorp/go-hclog"
	"google.golang.org/protobuf/proto"
)

// ReadDir processes a gRPC request to read contents of a directory specified by the request path and returns
// the results.
func (s *HostServiceGRPCServer) ReadDir(ctx context.Context,
	request *hostservev1.ReadDirRequest,
) (*hostservev1.ReadDirResponse, error) {

	clientID := getClientIDFromContext(ctx)
	ap, err := filepath.Abs(request.Path)
	if err != nil {
		ap = request.Path
	}
	hclog.Default().Info("ReadDir request from client", "clientID", clientID, "path", ap)

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

	clientID := getClientIDFromContext(ctx)
	ap, err := filepath.Abs(request.Path)
	if err != nil {
		ap = request.Path
	}
	hclog.Default().Info("ReadFile request from client", "clientID", clientID, "path", ap)

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

	clientID := getClientIDFromContext(ctx)
	ap, err := filepath.Abs(request.Path)
	if err != nil {
		ap = request.Path
	}
	hclog.Default().Info("WriteFile request from client", "clientID", clientID, "path", ap)

	err = s.Impl.WriteFile(ctx, request.Path, request.Data, os.FileMode(request.Perm))
	if err != nil {
		return &hostservev1.WriteFileResponse{Error: proto.String(err.Error())}, nil
	}
	return &hostservev1.WriteFileResponse{}, nil
}
