package hostserve

import (
	"context"

	hostservev1 "github.com/bmj2728/hst/shared/protogen/hostserve/v1"
)

// GetEnv retrieves the value of the specified environment variable via a gRPC call to the host service.
// Returns an empty string if an error occurs.
func (c *HostServiceGRPCClient) GetEnv(key string) string {
	resp, err := c.client.GetEnv(context.Background(), &hostservev1.GetEnvRequest{
		Key: key,
	})
	if err != nil {
		return ""
	}
	return resp.Val
}
