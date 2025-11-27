package hostserve

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
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

	clientID := getClientIDFromContext(ctx)
	hclog.Default().Debug("Pseudocode cap check - GET_ENV", "clientID", clientID)
	if clientID == "" {
		return "", fmt.Errorf("client ID not found in context")
	}

	val := os.Getenv(key)
	if val == "" {
		return val, fmt.Errorf("environment variable %s not found", key)
	}
	return val, nil
}

func (he *HostEnv) TempDir(ctx context.Context) (string, error) {
	clientID := getClientIDFromContext(ctx)
	hclog.Default().Debug("Pseudocode cap check - TEMP_DIR", "clientID", clientID)
	if clientID == "" {
		return "", fmt.Errorf("client ID not found in context")
	}
	return os.TempDir(), nil
}
