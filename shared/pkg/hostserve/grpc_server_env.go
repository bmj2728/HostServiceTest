package hostserve

import (
	"context"

	hostservev1 "github.com/bmj2728/hst/shared/protogen/hostserve/v1"
	"github.com/hashicorp/go-hclog"
	"google.golang.org/protobuf/proto"
)

// Getuid handles a gRPC request to retrieve a UID from the service.
// Returns a GetuidResponse containing the UID or an error if the request fails.
// Logs client request information during processing.
func (s *HostServiceGRPCServer) Getuid(ctx context.Context, _ *hostservev1.GetuidRequest) (*hostservev1.GetuidResponse, error) {
	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)
	if err != nil {
		hclog.Default().Info("Getuid bad request from client",
			ctxClientIDKey, clientID,
			ctxClientOwner, owner,
			ctxHostRequestIDKey, reqID,
			"error", err,
		)
		return &hostservev1.GetuidResponse{
			Error: proto.String(err.Error()),
		}, nil
	}

	// Log the request
	hclog.Default().Info("Getuid request from client",
		ctxClientIDKey, clientID,
		ctxClientOwner, owner,
		ctxHostRequestIDKey, reqID)

	uid, err := s.Impl.Getuid(ctx)
	if err != nil {
		return &hostservev1.GetuidResponse{
			Error: proto.String(err.Error()),
		}, nil
	}
	return &hostservev1.GetuidResponse{
		Uid: uid,
	}, nil
}

// Geteuid handles a gRPC request to fetch the effective user ID (euid) of the server and returns the result or an error.
func (s *HostServiceGRPCServer) Geteuid(ctx context.Context, _ *hostservev1.GeteuidRequest) (*hostservev1.GeteuidResponse, error) {
	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)
	if err != nil {
		hclog.Default().Info("Geteuid bad request from client",
			ctxClientIDKey, clientID,
			ctxClientOwner, owner,
			ctxHostRequestIDKey, reqID,
			"error", err,
		)
		return &hostservev1.GeteuidResponse{
			Error: proto.String(err.Error()),
		}, nil
	}

	// Log the request
	hclog.Default().Info("Geteuid request from client",
		ctxClientIDKey, clientID,
		ctxClientOwner, owner,
		ctxHostRequestIDKey, reqID)

	euid, err := s.Impl.Geteuid(ctx)
	if err != nil {
		return &hostservev1.GeteuidResponse{
			Error: proto.String(err.Error()),
		}, nil
	}
	return &hostservev1.GeteuidResponse{
		Euid: euid,
	}, nil
}

// Getgid processes a gRPC request to retrieve a GID, logs the request, and returns the response or an error.
func (s *HostServiceGRPCServer) Getgid(ctx context.Context, _ *hostservev1.GetgidRequest) (*hostservev1.GetgidResponse, error) {
	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)
	if err != nil {
		hclog.Default().Info("Getgid bad request from client",
			ctxClientIDKey, clientID,
			ctxClientOwner, owner,
			ctxHostRequestIDKey, reqID,
			"error", err,
		)
		return &hostservev1.GetgidResponse{
			Error: proto.String(err.Error()),
		}, nil
	}

	// Log the request
	hclog.Default().Info("Getgid request from client",
		ctxClientIDKey, clientID,
		ctxClientOwner, owner,
		ctxHostRequestIDKey, reqID)

	gid, err := s.Impl.Getgid(ctx)
	if err != nil {
		return &hostservev1.GetgidResponse{
			Error: proto.String(err.Error()),
		}, nil
	}
	return &hostservev1.GetgidResponse{
		Gid: gid,
	}, nil
}

// Getegid retrieves the effective group ID (EGID) for a given host context and returns a response with the value or an error message.
func (s *HostServiceGRPCServer) Getegid(ctx context.Context, _ *hostservev1.GetegidRequest) (*hostservev1.GetegidResponse, error) {
	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)
	if err != nil {
		hclog.Default().Info("Getegid bad request from client",
			ctxClientIDKey, clientID,
			ctxClientOwner, owner,
			ctxHostRequestIDKey, reqID,
			"error", err,
		)
		return &hostservev1.GetegidResponse{
			Error: proto.String(err.Error()),
		}, nil
	}

	// Log the request
	hclog.Default().Info("Getegid request from client",
		ctxClientIDKey, clientID,
		ctxClientOwner, owner,
		ctxHostRequestIDKey, reqID)

	egid, err := s.Impl.Getegid(ctx)
	if err != nil {
		return &hostservev1.GetegidResponse{
			Error: proto.String(err.Error()),
		}, nil
	}
	return &hostservev1.GetegidResponse{
		Egid: egid,
	}, nil
}

// GetGroups handles a gRPC request to retrieve a list of groups associated with a client.
// It validates the request context, logs the operation, and fetches group data via the implementation layer.
// Returns a GetGroupsResponse containing the list of group IDs or an error message if the operation fails.
func (s *HostServiceGRPCServer) GetGroups(ctx context.Context, _ *hostservev1.GetGroupsRequest) (*hostservev1.GetGroupsResponse, error) {
	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)
	if err != nil {
		hclog.Default().Info("GetGroups bad request from client",
			ctxClientIDKey, clientID,
			ctxClientOwner, owner,
			ctxHostRequestIDKey, reqID,
			"error", err,
		)
		return &hostservev1.GetGroupsResponse{
			Error: proto.String(err.Error()),
		}, nil
	}

	// Log the request
	hclog.Default().Info("GetGroups request from client",
		ctxClientIDKey, clientID,
		ctxClientOwner, owner,
		ctxHostRequestIDKey, reqID)

	groups, err := s.Impl.GetGroups(ctx)
	if err != nil {
		return &hostservev1.GetGroupsResponse{
			Error: proto.String(err.Error()),
		}, nil
	}
	return &hostservev1.GetGroupsResponse{
		Groups: groups,
	}, nil
}

// Getpid handles a request to retrieve the process ID of the server and returns it in a GetpidResponse.
func (s *HostServiceGRPCServer) Getpid(ctx context.Context, _ *hostservev1.GetpidRequest) (*hostservev1.GetpidResponse, error) {
	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)
	if err != nil {
		hclog.Default().Info("Getpid bad request from client",
			ctxClientIDKey, clientID,
			ctxClientOwner, owner,
			ctxHostRequestIDKey, reqID,
			"error", err,
		)
		return &hostservev1.GetpidResponse{
			Error: proto.String(err.Error()),
		}, nil
	}

	// Log the request
	hclog.Default().Info("Getpid request from client",
		ctxClientIDKey, clientID,
		ctxClientOwner, owner,
		ctxHostRequestIDKey, reqID)

	pid, err := s.Impl.Getpid(ctx)
	if err != nil {
		return &hostservev1.GetpidResponse{
			Error: proto.String(err.Error()),
		}, nil
	}
	return &hostservev1.GetpidResponse{
		Pid: pid,
	}, nil
}

// Getppid handles a request to retrieve the parent process ID (PPID) using information from the server's implementation.
func (s *HostServiceGRPCServer) Getppid(ctx context.Context, _ *hostservev1.GetppidRequest) (*hostservev1.GetppidResponse, error) {
	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)
	if err != nil {
		hclog.Default().Info("Getppid bad request from client",
			ctxClientIDKey, clientID,
			ctxClientOwner, owner,
			ctxHostRequestIDKey, reqID,
			"error", err,
		)
		return &hostservev1.GetppidResponse{
			Error: proto.String(err.Error()),
		}, nil
	}

	// Log the request
	hclog.Default().Info("Getppid request from client",
		ctxClientIDKey, clientID,
		ctxClientOwner, owner,
		ctxHostRequestIDKey, reqID)

	ppid, err := s.Impl.Getppid(ctx)
	if err != nil {
		return &hostservev1.GetppidResponse{
			Error: proto.String(err.Error()),
		}, nil
	}
	return &hostservev1.GetppidResponse{
		Ppid: ppid,
	}, nil
}

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

// TempDir handles a client request to retrieve the temporary directory path on the server. Returns the path or an error.
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

// UserCacheDir handles client requests to fetch the user's cache directory, returning its path or an error.
func (s *HostServiceGRPCServer) UserCacheDir(ctx context.Context, _ *hostservev1.UserCacheDirRequest) (*hostservev1.UserCacheDirResponse, error) {

	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)
	if err != nil {
		hclog.Default().Info("UserCacheDir bad request from client",
			ctxClientIDKey, clientID,
			ctxClientOwner, owner,
			ctxHostRequestIDKey, reqID,
			"error", err,
		)
		return &hostservev1.UserCacheDirResponse{
			Error: proto.String(err.Error()),
		}, nil
	}

	hclog.Default().Info("UserCacheDir request from client",
		ctxClientIDKey, clientID,
		ctxClientOwner, owner,
		ctxHostRequestIDKey, reqID,
	)

	dir, err := s.Impl.UserCacheDir(ctx)
	if err != nil {
		return &hostservev1.UserCacheDirResponse{
			Error: proto.String(err.Error()),
		}, nil
	}
	return &hostservev1.UserCacheDirResponse{
		Dir: dir,
	}, nil
}

// UserConfigDir handles the request to retrieve the user-specific configuration directory path and returns it to the caller.
func (s *HostServiceGRPCServer) UserConfigDir(ctx context.Context, _ *hostservev1.UserConfigDirRequest) (*hostservev1.UserConfigDirResponse, error) {

	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)
	if err != nil {
		hclog.Default().Info("UserConfigDir bad request from client",
			ctxClientIDKey, clientID,
			ctxClientOwner, owner,
			ctxHostRequestIDKey, reqID,
			"error", err,
		)
		return &hostservev1.UserConfigDirResponse{
			Error: proto.String(err.Error()),
		}, nil
	}

	hclog.Default().Info("UserConfigDir request from client",
		ctxClientIDKey, clientID,
		ctxClientOwner, owner,
		ctxHostRequestIDKey, reqID,
	)

	dir, err := s.Impl.UserConfigDir(ctx)
	if err != nil {
		return &hostservev1.UserConfigDirResponse{
			Error: proto.String(err.Error()),
		}, nil
	}
	return &hostservev1.UserConfigDirResponse{
		Dir: dir,
	}, nil
}

// UserHomeDir handles a request to retrieve the user's home directory path and returns it in a response.
// It processes the request context for logging and error handling, and invokes the underlying implementation.
func (s *HostServiceGRPCServer) UserHomeDir(ctx context.Context, _ *hostservev1.UserHomeDirRequest) (*hostservev1.UserHomeDirResponse, error) {

	ctx, clientID, reqID, owner, err := s.processRequestContext(ctx)
	if err != nil {
		hclog.Default().Info("UserHomeDir bad request from client",
			ctxClientIDKey, clientID,
			ctxClientOwner, owner,
			ctxHostRequestIDKey, reqID,
			"error", err,
		)
		return &hostservev1.UserHomeDirResponse{
			Error: proto.String(err.Error()),
		}, nil
	}

	hclog.Default().Info("UserHomeDir request from client",
		ctxClientIDKey, clientID,
		ctxClientOwner, owner,
		ctxHostRequestIDKey, reqID,
	)

	dir, err := s.Impl.UserHomeDir(ctx)
	if err != nil {
		return &hostservev1.UserHomeDirResponse{
			Error: proto.String(err.Error()),
		}, nil
	}
	return &hostservev1.UserHomeDirResponse{
		Dir: dir,
	}, nil
}
