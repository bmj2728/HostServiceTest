package hostserve

import (
	"context"

	hostservev1 "github.com/bmj2728/hst/shared/protogen/hostserve/v1"
	"github.com/hashicorp/go-hclog"
)

// GetEnv retrieves the value of the specified environment variable via a gRPC call to the host service.
// Returns an empty string if an error occurs.
func (c *HostServiceGRPCClient) GetEnv(ctx context.Context, key string) (string, error) {
	reqID := NewRequestID()
	hclog.Default().Info("GetEnv request ID", "request", reqID.String())
	ctx = addTracingIDsToContext(ctx, c.clientID, reqID)
	resp, err := c.client.GetEnv(ctx, &hostservev1.GetEnvRequest{
		Key: key,
	})
	if err != nil {
		return "", err
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
