package hostserve

import (
	"context"

	hostservev1 "github.com/bmj2728/hst/shared/protogen/hostserve/v1"
)

func (s *HostServiceGRPCServer) GetEnv(ctx context.Context, request *hostservev1.GetEnvRequest) (*hostservev1.GetEnvResponse, error) {
	val := s.Impl.GetEnv(request.Key)
	return &hostservev1.GetEnvResponse{Val: val}, nil
}
