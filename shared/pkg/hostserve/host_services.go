package hostserve

// HostServices provides functionalities for interacting with the host file system and environment variables.
type HostServices struct {
	IHostFS
	IHostEnv
}

// NewHostServices creates a new HostServices instance using the provided file system and environment abstractions.
func NewHostServices(fs IHostFS, env IHostEnv) *HostServices {
	return &HostServices{
		IHostFS:  fs,
		IHostEnv: env,
	}
}
