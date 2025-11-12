package hostserve

import (
	"context"

	hostservev1 "github.com/bmj2728/hst/shared/protogen/hostserve/v1"
	"github.com/hashicorp/go-hclog"
)

// GetEnv handles a gRPC request to retrieve the value of an environment variable identified by the request key.
func (s *HostServiceGRPCServer) GetEnv(ctx context.Context,
	request *hostservev1.GetEnvRequest) (*hostservev1.GetEnvResponse, error) {

	clientID := getClientIDFromContext(ctx)
	reqID := getRequestIDFromContext(ctx)
	hclog.Default().Info("GetEnv request from client",
		ctxClientIDKey, clientID,
		ctxHostRequestIDKey, reqID,
		"key", request.Key)

	val := s.Impl.GetEnv(ctx, request.Key)
	return &hostservev1.GetEnvResponse{Val: val}, nil
}
