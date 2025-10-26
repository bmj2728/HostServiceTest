package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/bmj2728/hst/shared/pkg/filelister"
	"github.com/bmj2728/hst/shared/pkg/hostserve"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "TEST_KEY",
	MagicCookieValue: "TEST_VALUE",
}

var pluginMap = map[string]plugin.Plugin{
	"fl-plugin": &filelister.FileListerGRPCPlugin{},
	"cl-plugin": &filelister.FileListerGRPCPlugin{},
}

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "host",
		Output: os.Stdout,
		Level:  hclog.Debug,
	})

	//Start plugin 1

	// Create the plugin client - plumbing
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  handshakeConfig,
		Plugins:          pluginMap,
		Cmd:              exec.Command("./plugins/filelister/filelister"),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger:           logger,
	})
	defer client.Kill()

	// Connect via gRPC - porcelain
	rpcClient, err := client.Client()
	if err != nil {
		logger.Error("Failed to get RPC client", "err", err)
		os.Exit(1)
	}

	// Request the FileLister plugin - the raw interface
	raw, err := rpcClient.Dispense("fl-plugin")
	if err != nil {
		logger.Error("Failed to dispense plugin", "err", err)
		os.Exit(1)
	}

	// Coerce the raw interface to the FileLister type
	fileLister := raw.(filelister.FileLister)

	// Set up host services - create the implementation
	hostServices := &hostserve.HostServices{}

	// We need to access the GRPCClient to set up the host service via the broker
	grpcClientImpl, ok := raw.(*filelister.GRPCClient)
	if !ok {
		logger.Error("Failed to cast to GRPCClient")
		os.Exit(1)
	}

	// Set up the host service server via the broker and get the allocated service ID
	hostServiceID, err := grpcClientImpl.SetupHostService(hostServices)
	if err != nil {
		logger.Error("Failed to setup host service", "err", err)
		os.Exit(1)
	}
	logger.Info("Host service registered with broker", "id", hostServiceID)

	// Tell the plugin about the host service ID so it can dial back
	fileLister.EstablishHostServices(hostServiceID)

	// End plugin 1

	//Start plugin 2
	color := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  handshakeConfig,
		Plugins:          pluginMap,
		Cmd:              exec.Command("./plugins/colorlister/colorlister"),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger:           logger,
	})
	defer color.Kill()

	// Connect via gRPC - porcelain
	rpcClientColor, err := color.Client()
	if err != nil {
		logger.Error("Failed to get RPC client", "err", err)
		os.Exit(1)
	}

	// Request the FileLister plugin - the raw interface
	rawColor, err := rpcClientColor.Dispense("cl-plugin")
	if err != nil {
		logger.Error("Failed to dispense plugin", "err", err)
		os.Exit(1)
	}

	// Coerce the raw interface to the FileLister type
	colorlister := rawColor.(filelister.FileLister)

	// Set up host services - create the implementation
	hostServicesColor := &hostserve.HostServices{}

	// We need to access the GRPCClient to set up the host service via the broker
	grpcClientImplColor, ok := rawColor.(*filelister.GRPCClient)
	if !ok {
		logger.Error("Failed to cast to GRPCClient")
		os.Exit(1)
	}

	// Set up the host service server via the broker and get the allocated service ID
	hostServiceIDColor, err := grpcClientImplColor.SetupHostService(hostServicesColor)
	if err != nil {
		logger.Error("Failed to setup host service", "err", err)
		os.Exit(1)
	}
	logger.Info("Host service registered with broker", "id", hostServiceIDColor)

	// Tell the plugin about the host service ID so it can dial back
	colorlister.EstablishHostServices(hostServiceIDColor)
	// End plugin 2

	// Test the plugin by listing files in the current directory
	entries, err := fileLister.ListFiles(".")
	if err != nil {
		logger.Error("Failed to list files", "err", err)
		os.Exit(1)
	}

	logger.Info("Successfully listed files - no color")
	for _, entry := range entries {
		fmt.Println(entry)
	}

	colorEntries, err := colorlister.ListFiles(".")
	if err != nil {
		logger.Error("Failed to list files", "err", err)
		os.Exit(1)
	}

	logger.Info("Successfully listed files - with color")
	for _, entry := range colorEntries {
		fmt.Println(entry)
	}

	// Clean shutdown - disconnect from host services
	logger.Info("Shutting down plugins")
	fileLister.DisconnectHostServices()
	colorlister.DisconnectHostServices()
	os.Exit(0)
}
