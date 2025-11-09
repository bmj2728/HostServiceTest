package hostserve

import "os"

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
func (he *HostEnv) GetEnv(key string) string {
	return os.Getenv(key)
}
