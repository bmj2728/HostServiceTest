package hostserve

import (
	"context"

	hostservev1 "github.com/bmj2728/hst/shared/protogen/hostserve/v1"
)

// GetEnv handles a gRPC request to retrieve the value of an environment variable identified by the request key.
func (s *HostServiceGRPCServer) GetEnv(ctx context.Context,
	request *hostservev1.GetEnvRequest) (*hostservev1.GetEnvResponse, error) {
	val := s.Impl.GetEnv(ctx, request.Key)
	return &hostservev1.GetEnvResponse{Val: val}, nil
}
