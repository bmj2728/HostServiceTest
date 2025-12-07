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

// Getuid retrieves the user ID of the calling process as an integer.
// Returns an error if the client ID is not found in the provided context.
func (he *HostEnv) Getuid(ctx context.Context) (int32, error) {
	clientID := getClientIDFromContext(ctx)
	hclog.Default().Debug("Pseudocode cap check - GET_UID", "clientID", clientID)
	if clientID == "" {
		return 0, fmt.Errorf("client ID not found in context")
	}
	return int32(os.Getuid()), nil
}

// Getgid retrieves the group ID of the current user as an int32.
// Returns an error if the client ID is not found in the provided context.
func (he *HostEnv) Getgid(ctx context.Context) (int32, error) {
	clientID := getClientIDFromContext(ctx)
	hclog.Default().Debug("Pseudocode cap check - GET_GID", "clientID", clientID)
	if clientID == "" {
		return 0, fmt.Errorf("client ID not found in context")
	}
	return int32(os.Getgid()), nil
}

// Geteuid retrieves the effective user ID (EUID) of the calling process as an int32.
// Returns an error if the client ID is missing from the context.
func (he *HostEnv) Geteuid(ctx context.Context) (int32, error) {
	clientID := getClientIDFromContext(ctx)
	hclog.Default().Debug("Pseudocode cap check - GET_EUID", "clientID", clientID)
	if clientID == "" {
		return 0, fmt.Errorf("client ID not found in context")
	}
	return int32(os.Geteuid()), nil
}

// Getegid retrieves the effective group ID of the calling process as an integer. Returns an error if the client ID is not found.
func (he *HostEnv) Getegid(ctx context.Context) (int32, error) {
	clientID := getClientIDFromContext(ctx)
	hclog.Default().Debug("Pseudocode cap check - GET_EGID", "clientID", clientID)
	if clientID == "" {
		return 0, fmt.Errorf("client ID not found in context")
	}
	return int32(os.Getegid()), nil
}

// GetGroups retrieves the list of group IDs associated with the current process as a slice of int32 values.
// Returns an error if the client ID is not found in the context or if the group retrieval fails.
func (he *HostEnv) GetGroups(ctx context.Context) ([]int32, error) {
	clientID := getClientIDFromContext(ctx)
	hclog.Default().Debug("Pseudocode cap check - GET_GROUPS", "clientID", clientID)
	if clientID == "" {
		return nil, fmt.Errorf("client ID not found in context")
	}
	groups, err := os.Getgroups()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve groups for client ID %s: %v", clientID, err)
	}
	groupIDs := make([]int32, len(groups))
	for i, group := range groups {
		groupIDs[i] = int32(group)
	}
	return groupIDs, nil
}

// Getpid retrieves the process ID of the calling process as an int32. Returns an error if the client ID is missing in the context.
func (he *HostEnv) Getpid(ctx context.Context) (int32, error) {
	clientID := getClientIDFromContext(ctx)
	hclog.Default().Debug("Pseudocode cap check - GET_PID", "clientID", clientID)
	if clientID == "" {
		return 0, fmt.Errorf("client ID not found in context")
	}
	return int32(os.Getpid()), nil
}

// Getppid retrieves the parent process ID of the calling process as an int32. Returns an error if client ID is missing.
func (he *HostEnv) Getppid(ctx context.Context) (int32, error) {
	clientID := getClientIDFromContext(ctx)
	hclog.Default().Debug("Pseudocode cap check - GET_PPID", "clientID", clientID)
	if clientID == "" {
		return 0, fmt.Errorf("client ID not found in context")
	}
	return int32(os.Getppid()), nil
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

// TempDir retrieves the system's temporary directory for the current user and returns it as a string.
// Returns an error if the client ID is not found in the provided context.
func (he *HostEnv) TempDir(ctx context.Context) (string, error) {
	clientID := getClientIDFromContext(ctx)
	hclog.Default().Debug("Pseudocode cap check - TEMP_DIR", "clientID", clientID)
	if clientID == "" {
		return "", fmt.Errorf("client ID not found in context")
	}
	return os.TempDir(), nil
}

// UserCacheDir retrieves the user cache directory path for the current system environment.
// Returns an error if the client ID is missing in the context or the directory retrieval fails.
func (he *HostEnv) UserCacheDir(ctx context.Context) (string, error) {
	clientID := getClientIDFromContext(ctx)
	hclog.Default().Debug("Pseudocode cap check - USER_CACHE_DIR", "clientID", clientID)
	if clientID == "" {
		return "", fmt.Errorf("client ID not found in context")
	}
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user cache dir: %v", err)
	}
	return dir, nil
}

// UserConfigDir retrieves the user-specific configuration directory based on the context's client ID. Returns an error if the client ID is missing or the directory cannot be determined.
func (he *HostEnv) UserConfigDir(ctx context.Context) (string, error) {
	clientID := getClientIDFromContext(ctx)
	hclog.Default().Debug("Pseudocode cap check - USER_CONFIG_DIR", "clientID", clientID)
	if clientID == "" {
		return "", fmt.Errorf("client ID not found in context")
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config dir: %v", err)
	}
	return dir, nil
}

// UserHomeDir retrieves the home directory path of the current user from the environment.
// Returns an error if the client ID is not found in the context or if the home directory retrieval fails.
func (he *HostEnv) UserHomeDir(ctx context.Context) (string, error) {
	clientID := getClientIDFromContext(ctx)
	hclog.Default().Debug("Pseudocode cap check - USER_HOME_DIR", "clientID", clientID)
	if clientID == "" {
		return "", fmt.Errorf("client ID not found in context")
	}
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home dir: %v", err)
	}
	return dir, nil
}
