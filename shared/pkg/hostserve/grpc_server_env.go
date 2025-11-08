package hostserve

import (
	"context"

	hostservev1 "github.com/bmj2728/hst/shared/protogen/hostserve/v1"
)

func (s *HostServiceGRPCServer) GetEnv(context.Context, *hostservev1.GetEnvRequest) (*hostservev1.GetEnvResponse, error) {
	panic("implement me")
}
