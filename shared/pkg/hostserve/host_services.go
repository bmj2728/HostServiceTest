package hostserve

// HostServices provides functionalities for interacting with the host file system and environment variables.
type HostServices struct {
	ActiveClients *ActiveClients
	IHostFS
	IHostEnv
}

// NewHostServices creates a new HostServices instance using the provided file system and environment abstractions.
func NewHostServices(fs IHostFS, env IHostEnv) *HostServices {
	return &HostServices{
		ActiveClients: NewActiveClients(),
		IHostFS:       fs,
		IHostEnv:      env,
	}
}
