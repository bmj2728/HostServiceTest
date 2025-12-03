package hostserve

import (
	"context"

	hostservev1 "github.com/bmj2728/hst/shared/protogen/hostserve/v1"
)

// Getuid retrieves the unique user ID from the host service using gRPC and includes tracing metadata in the request context.
func (c *HostServiceGRPCClient) Getuid(ctx context.Context) (int32, error) {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.Getuid(ctx, &hostservev1.GetuidRequest{})
	if err != nil {
		return 0, err
	}
	if resp == nil {
		return 0, &HostServiceError{Message: "nil response from Getuid"}
	}
	if resp.GetError() != "" {
		return 0, &HostServiceError{Message: resp.GetError()}
	}

	return resp.Uid, nil
}

// Getgid retrieves the GID by performing a gRPC request, ensuring tracing metadata is appended to the context.
func (c *HostServiceGRPCClient) Getgid(ctx context.Context) (int32, error) {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.Getgid(ctx, &hostservev1.GetgidRequest{})
	if err != nil {
		return 0, err
	}
	if resp == nil {
		return 0, &HostServiceError{Message: "nil response from Getgid"}
	}
	if resp.GetError() != "" {
		return 0, &HostServiceError{Message: resp.GetError()}
	}
	return resp.Gid, nil
}

// Geteuid retrieves the effective user ID (EUID) by making a gRPC call and returns it as an integer. Returns an error if unsuccessful.
func (c *HostServiceGRPCClient) Geteuid(ctx context.Context) (int32, error) {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.Geteuid(ctx, &hostservev1.GeteuidRequest{})
	if err != nil {
		return 0, err
	}
	if resp == nil {
		return 0, &HostServiceError{Message: "nil response from Geteuid"}
	}
	if resp.GetError() != "" {
		return 0, &HostServiceError{Message: resp.GetError()}
	}
	return resp.Euid, nil
}

// Getegid retrieves the effective group ID (EGID) using a gRPC client context and returns it as an int32 value.
func (c *HostServiceGRPCClient) Getegid(ctx context.Context) (int32, error) {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.Getegid(ctx, &hostservev1.GetegidRequest{})
	if err != nil {
		return 0, err
	}
	if resp == nil {
		return 0, &HostServiceError{Message: "nil response from Getegid"}
	}
	if resp.GetError() != "" {
		return 0, &HostServiceError{Message: resp.GetError()}
	}
	return resp.Egid, nil
}

// GetGroups retrieves a list of group IDs from the host service via a gRPC client and returns them or an error encountered.
func (c *HostServiceGRPCClient) GetGroups(ctx context.Context) ([]int32, error) {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.GetGroups(ctx, &hostservev1.GetGroupsRequest{})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, &HostServiceError{Message: "nil response from GetGroups"}
	}
	if resp.GetError() != "" {
		return nil, &HostServiceError{Message: resp.GetError()}
	}

	return resp.Groups, nil
}

// GetEnv retrieves the value of the specified environment variable via a gRPC call to the host service.
// Returns an empty string if an error occurs.
func (c *HostServiceGRPCClient) GetEnv(ctx context.Context, key string) (string, error) {

	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.GetEnv(ctx, &hostservev1.GetEnvRequest{
		Key: key,
	})
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", &HostServiceError{Message: "nil response from GetEnv"}
	}
	if resp.GetError() != "" {
		return "", &HostServiceError{Message: resp.GetError()}
	}
	return resp.Val, nil
}

func (c *HostServiceGRPCClient) TempDir(ctx context.Context) (string, error) {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.TempDir(ctx, &hostservev1.TempDirRequest{})
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", &HostServiceError{Message: "nil response from TempDir"}
	}
	if resp.GetError() != "" {
		return "", &HostServiceError{Message: resp.GetError()}
	}
	return resp.Dir, nil
}

func (c *HostServiceGRPCClient) UserCacheDir(ctx context.Context) (string, error) {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.UserCacheDir(ctx, &hostservev1.UserCacheDirRequest{})
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", &HostServiceError{Message: "nil response from UserCacheDir"}
	}
	if resp.GetError() != "" {
		return "", &HostServiceError{Message: resp.GetError()}
	}
	return resp.Dir, nil
}

func (c *HostServiceGRPCClient) UserConfigDir(ctx context.Context) (string, error) {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.UserConfigDir(ctx, &hostservev1.UserConfigDirRequest{})
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", &HostServiceError{Message: "nil response from UserConfigDir"}
	}
	if resp.GetError() != "" {
		return "", &HostServiceError{Message: resp.GetError()}
	}
	return resp.Dir, nil
}

func (c *HostServiceGRPCClient) UserHomeDir(ctx context.Context) (string, error) {
	ctx = addTracingIDsToContext(ctx, c.clientID, NewRequestID())
	resp, err := c.client.UserHomeDir(ctx, &hostservev1.UserHomeDirRequest{})
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", &HostServiceError{Message: "nil response from UserHomeDir"}
	}
	if resp.GetError() != "" {
		return "", &HostServiceError{Message: resp.GetError()}
	}
	return resp.Dir, nil
}
