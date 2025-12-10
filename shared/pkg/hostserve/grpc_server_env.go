package hostserve

import (
	"context"

	hostservev1 "github.com/bmj2728/hst/shared/protogen/hostserve/v1"
	//"github.com/hashicorp/go-hclog"
	"google.golang.org/protobuf/proto"
)

// Getuid handles a gRPC request to retrieve a UID from the service.
// Returns a GetuidResponse containing the UID or an error if the request fails.
// Logs client request information during processing.
func (s *HostServiceGRPCServer) Getuid(ctx context.Context, request *hostservev1.GetuidRequest) (*hostservev1.GetuidResponse, error) {
	response, err := withRequestLoggingAndResponse(s, ctx, "Getuid", request,
		func(ctx context.Context, req *hostservev1.GetuidRequest) (*hostservev1.GetuidResponse, error) {
			uid, implErr := s.Impl.Getuid(ctx)
			if implErr != nil {
				return &hostservev1.GetuidResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.GetuidResponse{Uid: uid}, nil
		})

	if err != nil {
		return &hostservev1.GetuidResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}

// Geteuid handles a gRPC request to fetch the effective user ID (euid) of the server and returns the result or an error.
func (s *HostServiceGRPCServer) Geteuid(ctx context.Context, request *hostservev1.GeteuidRequest) (*hostservev1.GeteuidResponse, error) {
	response, err := withRequestLoggingAndResponse(s, ctx, "Geteuid", request,
		func(ctx context.Context, req *hostservev1.GeteuidRequest) (*hostservev1.GeteuidResponse, error) {
			euid, implErr := s.Impl.Geteuid(ctx)
			if implErr != nil {
				return &hostservev1.GeteuidResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.GeteuidResponse{Euid: euid}, nil
		})

	if err != nil {
		return &hostservev1.GeteuidResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}

// Getgid processes a gRPC request to retrieve a GID, logs the request, and returns the response or an error.
func (s *HostServiceGRPCServer) Getgid(ctx context.Context, request *hostservev1.GetgidRequest) (*hostservev1.GetgidResponse, error) {
	response, err := withRequestLoggingAndResponse(s, ctx, "Getgid", request,
		func(ctx context.Context, req *hostservev1.GetgidRequest) (*hostservev1.GetgidResponse, error) {
			gid, implErr := s.Impl.Getgid(ctx)
			if implErr != nil {
				return &hostservev1.GetgidResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.GetgidResponse{Gid: gid}, nil
		})

	if err != nil {
		return &hostservev1.GetgidResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}

// Getegid retrieves the effective group ID (EGID) for a given host context and returns a response with the value or an error message.
func (s *HostServiceGRPCServer) Getegid(ctx context.Context, request *hostservev1.GetegidRequest) (*hostservev1.GetegidResponse, error) {
	response, err := withRequestLoggingAndResponse(s, ctx, "Getegid", request,
		func(ctx context.Context, req *hostservev1.GetegidRequest) (*hostservev1.GetegidResponse, error) {
			egid, implErr := s.Impl.Getegid(ctx)
			if implErr != nil {
				return &hostservev1.GetegidResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.GetegidResponse{Egid: egid}, nil
		})

	if err != nil {
		return &hostservev1.GetegidResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}

// GetGroups handles a gRPC request to retrieve a list of groups associated with a client.
// It validates the request context, logs the operation, and fetches group data via the implementation layer.
// Returns a GetGroupsResponse containing the list of group IDs or an error message if the operation fails.
func (s *HostServiceGRPCServer) GetGroups(ctx context.Context, request *hostservev1.GetGroupsRequest) (*hostservev1.GetGroupsResponse, error) {
	response, err := withRequestLoggingAndResponse(s, ctx, "GetGroups", request,
		func(ctx context.Context, req *hostservev1.GetGroupsRequest) (*hostservev1.GetGroupsResponse, error) {
			groups, implErr := s.Impl.GetGroups(ctx)
			if implErr != nil {
				return &hostservev1.GetGroupsResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.GetGroupsResponse{Groups: groups}, nil
		})

	if err != nil {
		return &hostservev1.GetGroupsResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}

// Getpid handles a request to retrieve the process ID of the server and returns it in a GetpidResponse.
func (s *HostServiceGRPCServer) Getpid(ctx context.Context, request *hostservev1.GetpidRequest) (*hostservev1.GetpidResponse, error) {
	response, err := withRequestLoggingAndResponse(s, ctx, "Getpid", request,
		func(ctx context.Context, req *hostservev1.GetpidRequest) (*hostservev1.GetpidResponse, error) {
			pid, implErr := s.Impl.Getpid(ctx)
			if implErr != nil {
				return &hostservev1.GetpidResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.GetpidResponse{Pid: pid}, nil
		})

	if err != nil {
		return &hostservev1.GetpidResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}

// Getppid handles a request to retrieve the parent process ID (PPID) using information from the server's implementation.
func (s *HostServiceGRPCServer) Getppid(ctx context.Context, request *hostservev1.GetppidRequest) (*hostservev1.GetppidResponse, error) {
	response, err := withRequestLoggingAndResponse(s, ctx, "Getppid", request,
		func(ctx context.Context, req *hostservev1.GetppidRequest) (*hostservev1.GetppidResponse, error) {
			ppid, implErr := s.Impl.Getppid(ctx)
			if implErr != nil {
				return &hostservev1.GetppidResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.GetppidResponse{Ppid: ppid}, nil
		})

	if err != nil {
		return &hostservev1.GetppidResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}

// GetEnv handles a gRPC request to retrieve the value of an environment variable identified by the request key.
func (s *HostServiceGRPCServer) GetEnv(ctx context.Context,
	request *hostservev1.GetEnvRequest) (*hostservev1.GetEnvResponse, error) {

	response, err := withRequestLoggingAndResponse(s, ctx, "GetEnv", request,
		func(ctx context.Context, req *hostservev1.GetEnvRequest) (*hostservev1.GetEnvResponse, error) {
			val, implErr := s.Impl.GetEnv(ctx, req.Key)
			if implErr != nil {
				return &hostservev1.GetEnvResponse{
					Val:   val,
					Error: proto.String(implErr.Error()),
				}, nil
			}
			return &hostservev1.GetEnvResponse{Val: val}, nil
		})

	if err != nil {
		return &hostservev1.GetEnvResponse{
			Val:   "",
			Error: proto.String(err.Error()),
		}, nil
	}
	return response, nil
}

// TempDir handles a client request to retrieve the temporary directory path on the server. Returns the path or an error.
func (s *HostServiceGRPCServer) TempDir(ctx context.Context, request *hostservev1.TempDirRequest) (*hostservev1.TempDirResponse, error) {

	response, err := withRequestLoggingAndResponse(s, ctx, "TempDir", request,
		func(ctx context.Context, req *hostservev1.TempDirRequest) (*hostservev1.TempDirResponse, error) {
			dir, implErr := s.Impl.TempDir(ctx)
			if implErr != nil {
				return &hostservev1.TempDirResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.TempDirResponse{Dir: dir}, nil
		})

	if err != nil {
		return &hostservev1.TempDirResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}

// UserCacheDir handles client requests to fetch the user's cache directory, returning its path or an error.
func (s *HostServiceGRPCServer) UserCacheDir(ctx context.Context, request *hostservev1.UserCacheDirRequest) (*hostservev1.UserCacheDirResponse, error) {

	response, err := withRequestLoggingAndResponse(s, ctx, "UserCacheDir", request,
		func(ctx context.Context, req *hostservev1.UserCacheDirRequest) (*hostservev1.UserCacheDirResponse, error) {
			dir, implErr := s.Impl.UserCacheDir(ctx)
			if implErr != nil {
				return &hostservev1.UserCacheDirResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.UserCacheDirResponse{Dir: dir}, nil
		})

	if err != nil {
		return &hostservev1.UserCacheDirResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}

// UserConfigDir handles the request to retrieve the user-specific configuration directory path and returns it to the caller.
func (s *HostServiceGRPCServer) UserConfigDir(ctx context.Context, request *hostservev1.UserConfigDirRequest) (*hostservev1.UserConfigDirResponse, error) {

	response, err := withRequestLoggingAndResponse(s, ctx, "UserConfigDir", request,
		func(ctx context.Context, req *hostservev1.UserConfigDirRequest) (*hostservev1.UserConfigDirResponse, error) {
			dir, implErr := s.Impl.UserConfigDir(ctx)
			if implErr != nil {
				return &hostservev1.UserConfigDirResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.UserConfigDirResponse{Dir: dir}, nil
		})

	if err != nil {
		return &hostservev1.UserConfigDirResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}

// UserHomeDir handles a request to retrieve the user's home directory path and returns it in a response.
// It processes the request context for logging and error handling, and invokes the underlying implementation.
func (s *HostServiceGRPCServer) UserHomeDir(ctx context.Context, request *hostservev1.UserHomeDirRequest) (*hostservev1.UserHomeDirResponse, error) {

	response, err := withRequestLoggingAndResponse(s, ctx, "UserHomeDir", request,
		func(ctx context.Context, req *hostservev1.UserHomeDirRequest) (*hostservev1.UserHomeDirResponse, error) {
			dir, implErr := s.Impl.UserHomeDir(ctx)
			if implErr != nil {
				return &hostservev1.UserHomeDirResponse{Error: proto.String(implErr.Error())}, nil
			}
			return &hostservev1.UserHomeDirResponse{Dir: dir}, nil
		})

	if err != nil {
		return &hostservev1.UserHomeDirResponse{Error: proto.String(err.Error())}, nil
	}
	return response, nil
}
