package hostconn

import (
	"fmt"

	"github.com/bmj2728/hst/shared/pkg/hostserve"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

// HostConnection handles bidirectional communication with host services.
// Plugin implementations should implement this interface if they need
// to call back to the host for privileged operations or shared services.
//
// This interface separates infrastructure concerns (connection management)
// from business logic, making it optional for plugins that don't require
// host services.
type HostConnection interface {
	// SetBroker provides the gRPC broker for bidirectional communication
	SetBroker(broker *plugin.GRPCBroker)

	// EstablishHostServices receives the broker service ID for host services
	EstablishHostServices(hostServiceID uint32) (clientID string, err error)

	// DisconnectHostServices cleans up connections to host services
	DisconnectHostServices()
}

// HostServiceRegistrar allows plugin clients to register host services with the broker.
// This is typically implemented by the host-side plugin client wrapper (e.g., GRPCClient).
type HostServiceRegistrar interface {
	// RegisterHostService registers a host service implementation with the broker
	// and returns the allocated service ID that plugins can use to dial back.
	RegisterHostService(hostServices hostserve.IHostServices) (uint32, error)
}

// EstablishHostServices handles the complete setup flow for connecting a plugin to host services.
// It encapsulates the following steps:
// 1. Checks if plugin supports host service registration (via HostServiceRegistrar)
// 2. Registers the host service with the broker and gets a service ID
// 3. Notifies the plugin of the service ID (via HostConnection.EstablishHostServices)
//
// This function gracefully handles plugins that don't support host services.
//
// Parameters:
//   - pluginClient: The dispensed plugin client (typically from rpcClient.Dispense())
//   - hostServices: The host service implementation to expose to the plugin
//   - logger: Logger for status messages
//
// Returns an error if registration fails. Returns nil if plugin doesn't support
// host services (this is not considered an error).
func EstablishHostServices(
	pluginClient interface{},
	hostServices hostserve.IHostServices,
	logger hclog.Logger,
) (string, error) {
	// Check if plugin supports host service registration
	registrar, ok := pluginClient.(HostServiceRegistrar)
	if !ok {
		logger.Debug("Plugin doesn't support host services (no HostServiceRegistrar)")
		return "", nil // Not an error - plugin simply doesn't need host services
	}

	// Register host service with broker and get service ID
	serviceID, err := registrar.RegisterHostService(hostServices)
	if err != nil {
		return "", fmt.Errorf("failed to register host service: %w", err)
	}
	logger.Info("Host service registered with broker", "id", serviceID)

	// Notify plugin of the service ID so it can dial back
	if hostConn, ok := pluginClient.(HostConnection); ok {
		clientID, err := hostConn.EstablishHostServices(serviceID)
		if err != nil {
			return "", fmt.Errorf("failed to establish connection: %w", err)
		}
		return clientID, nil
	}
	logger.Warn("Plugin supports registration but not connection")
	return "", nil
}

// DisconnectHostServices cleanly disconnects a plugin from host services.
// It checks if the plugin implements HostConnection and calls DisconnectHostServices if so.
//
// This function gracefully handles plugins that don't implement HostConnection.
func DisconnectHostServices(pluginClient interface{}, logger hclog.Logger) {
	if hostConn, ok := pluginClient.(HostConnection); ok {
		hostConn.DisconnectHostServices()
		logger.Debug("Disconnected from host services")
	}
}
