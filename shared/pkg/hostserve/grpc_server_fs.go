package hostserve

import (
	"context"

	"github.com/bmj2728/hst/shared/protogen/hostserve/v1"
)

// ReadDir processes a gRPC request to read contents of a directory specified by the request path and returns
// the results.
func (s *HostServiceGRPCServer) ReadDir(ctx context.Context,
	request *hostservev1.ReadDirRequest,
) (*hostservev1.ReadDirResponse, error) {

	entries, err := s.Impl.ReadDir(request.Path)
	if err != nil {
		errMsg := err.Error()
		return &hostservev1.ReadDirResponse{
			Entries: nil,
			Error:   &errMsg,
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
		Error:   nil,
	}, nil
}

// ReadFile handles a gRPC request to read a specific file from a specified directory and returns its contents
// or an error.
func (s *HostServiceGRPCServer) ReadFile(ctx context.Context,
	request *hostservev1.ReadFileRequest,
) (*hostservev1.ReadFileResponse, error) {

	bytes, err := s.Impl.ReadFile(request.Dir, request.File)
	if err != nil {
		errMsg := err.Error()
		return &hostservev1.ReadFileResponse{Contents: nil, Error: &errMsg}, err
	}
	return &hostservev1.ReadFileResponse{
		Contents: bytes,
		Error:    nil,
	}, nil
}
