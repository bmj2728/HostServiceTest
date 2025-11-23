package hostserve

import (
	"context"
	"fmt"

	hostservev1 "github.com/bmj2728/hst/shared/protogen/hostserve/v1"
	"github.com/hashicorp/go-hclog"
)

// GetEnv handles a gRPC request to retrieve the value of an environment variable identified by the request key.
func (s *HostServiceGRPCServer) GetEnv(ctx context.Context,
	request *hostservev1.GetEnvRequest) (*hostservev1.GetEnvResponse, error) {

	clientID := getClientIDFromContext(ctx)
	reqID := getRequestIDFromContext(ctx)

	owner, err := s.clientIdToOwner(clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client owner: %w", err)
	}
	ctx = addClientOwnerToContext(ctx, owner)

	hclog.Default().Info("GetEnv request from client",
		ctxClientIDKey, clientID,
		ctxClientOwner, owner,
		ctxHostRequestIDKey, reqID,
		"key", request.Key)

	val, err := s.Impl.GetEnv(ctx, request.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to get environment variable: %w", err)
	}
	return &hostservev1.GetEnvResponse{Val: val}, nil
}
