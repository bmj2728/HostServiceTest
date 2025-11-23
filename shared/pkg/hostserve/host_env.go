package hostserve

import (
	"context"
	"fmt"
	"os"
)

// HostEnv represents the environment configuration or context for a host system.
// It provides methods to interact with and retrieve system environment variables.
type HostEnv struct {
	//TBD fields
}

// NewHostEnv creates and returns a new instance of HostEnv.
func NewHostEnv() *HostEnv {
	return &HostEnv{}
}

// GetEnv retrieves the environment variable value associated with the provided key.
func (he *HostEnv) GetEnv(ctx context.Context, key string) (string, error) {

	// Future State - Capability Check - we may need to update service to return an error
	// get the owner from the context
	// owner := getClientOwnerFromContext(ctx)
	// if owner == "" {
	// 	return ""
	// }
	// canRead := capabilities.CapabilityCheck(owner, []string{capabilities.CAP_ENV_READ})
	// if !canRead {
	// 	return ""
	//	//return errors.New("insufficient permissions to read environment variables")}

	val := os.Getenv(key)
	if val == "" {
		return val, fmt.Errorf("environment variable %s not found", key)
	}
	return val, nil
}
