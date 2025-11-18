// Package hostserve contains functionality for interacting with the code host system.
// Allowing fine-grained control over operations performed by the plugin process.
// Adding functionality to the host service:
// 1. Update the hostserve.proto file ***breaking changes require additional steps for backwards compatibility**
// 2. Regenerate the gRPC code using buf (buf generate)
// 3. Add a new method to the appropriate interface(e.g. IHostFS)
// 4. Implement the method in the appropriate struct(e.g. HostFS)
// 5. Implement the gRPC client & server methods. (grpc_client_fs.go/grpc_server_fs.go)
// 6. Rebuild plugins to implement updated host services sdk
//   - Rebuild is not required for plugins that don't use the new methods
package hostserve

// HostServices provides functionalities for interacting with the host file system and environment variables.
type HostServices struct {
	activeClients *ActiveClients
	IHostFS
	IHostEnv
}

// NewHostServices creates a new HostServices instance using the provided file system and environment abstractions.
func NewHostServices(fs IHostFS, env IHostEnv) *HostServices {
	return &HostServices{
		activeClients: newActiveClients(),
		IHostFS:       fs,
		IHostEnv:      env,
	}
}

func (hs *HostServices) ActiveClients() *ActiveClients {
	if hs.activeClients == nil {
		return newActiveClients()
	}
	return hs.activeClients
}
