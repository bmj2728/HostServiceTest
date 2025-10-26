package filelister

import (
	"context"

	filelisterv1 "github.com/bmj2728/hst/shared/protogen/filelister/v1"
)

type GRPCServer struct {
	Impl FileLister
	filelisterv1.UnimplementedFileListerServer
}

func (s *GRPCServer) List(ctx context.Context, request *filelisterv1.FileListRequest) (*filelisterv1.FileListResponse, error) {
	entries, err := s.Impl.ListFiles(request.Dir)
	if err != nil {
		errMsg := err.Error()
		return &filelisterv1.FileListResponse{
			Entry: nil,
			Error: &errMsg,
		}, nil
	}
	return &filelisterv1.FileListResponse{
		Entry: entries,
		Error: nil,
	}, nil
}

func (s *GRPCServer) EstablishHostServices(ctx context.Context, request *filelisterv1.HostServiceRequest) (*filelisterv1.Empty, error) {
	s.Impl.EstablishHostServices(request.HostService)
	return &filelisterv1.Empty{}, nil
}
