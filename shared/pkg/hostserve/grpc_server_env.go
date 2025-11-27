package hostserve

import (
	"context"

	hostservev1 "github.com/bmj2728/hst/shared/protogen/hostserve/v1"
	"github.com/hashicorp/go-hclog"
	"google.golang.org/protobuf/proto"
)

// GetEnv handles a gRPC request to retrieve the value of an environment variable identified by the request key.
func (s *HostServiceGRPCServer) GetEnv(ctx context.Context,
	request *hostservev1.GetEnvRequest) (*hostservev1.GetEnvResponse, error) {

	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)
	if err != nil {
		hclog.Default().Info("GetEnv bad request from client",
			ctxClientIDKey, clientID,
			ctxClientOwner, owner,
			ctxHostRequestIDKey, reqID,
			"error", err,
		)
		return &hostservev1.GetEnvResponse{
			Val:   "",
			Error: proto.String(err.Error()),
		}, nil
	}

	// Log the request
	hclog.Default().Info("GetEnv request from client",
		ctxClientIDKey, clientID,
		ctxClientOwner, owner,
		ctxHostRequestIDKey, reqID,
		"key", request.Key)

	val, err := s.Impl.GetEnv(ctx, request.Key)
	if err != nil {
		return &hostservev1.GetEnvResponse{
			Val:   val,
			Error: proto.String(err.Error()),
		}, nil
	}
	return &hostservev1.GetEnvResponse{
		Val: val,
	}, nil
}

func (s *HostServiceGRPCServer) TempDir(ctx context.Context, _ *hostservev1.TempDirRequest) (*hostservev1.TempDirResponse, error) {

	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)
	if err != nil {
		hclog.Default().Info("TempDir bad request from client",
			ctxClientIDKey, clientID,
			ctxClientOwner, owner,
			ctxHostRequestIDKey, reqID,
			"error", err,
		)
		return &hostservev1.TempDirResponse{
			Error: proto.String(err.Error()),
		}, nil
	}

	hclog.Default().Info("TempDir request from client",
		ctxClientIDKey, clientID,
		ctxClientOwner, owner,
		ctxHostRequestIDKey, reqID,
	)

	dir, err := s.Impl.TempDir(ctx)
	if err != nil {
		return &hostservev1.TempDirResponse{
			Error: proto.String(err.Error()),
		}, nil
	}
	return &hostservev1.TempDirResponse{
		Dir: dir,
	}, nil
}
