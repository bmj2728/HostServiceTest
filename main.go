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
}

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "host",
		Output: os.Stdout,
		Level:  hclog.Debug,
	})

	// Create the plugin client
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  handshakeConfig,
		Plugins:          pluginMap,
		Cmd:              exec.Command("./plugins/filelister/filelister"),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger:           logger,
	})
	defer client.Kill()

	// Connect via gRPC
	rpcClient, err := client.Client()
	if err != nil {
		logger.Error("Failed to get RPC client", "err", err)
		os.Exit(1)
	}

	// Request the FileLister plugin
	raw, err := rpcClient.Dispense("fl-plugin")
	if err != nil {
		logger.Error("Failed to dispense plugin", "err", err)
		os.Exit(1)
	}

	// Get the FileLister interface (which is actually a *GRPCClient)
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

	// Test the plugin by listing files in the current directory
	entries, err := fileLister.ListFiles(".", hostServiceID)
	if err != nil {
		logger.Error("Failed to list files", "err", err)
		os.Exit(1)
	}

	logger.Info("Successfully listed files")
	for _, entry := range entries {
		fmt.Println(entry)
	}

	// Clean shutdown - disconnect from host services
	logger.Info("Shutting down plugin")
	fileLister.DisconnectHostServices()
}
