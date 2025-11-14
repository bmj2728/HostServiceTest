package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/bmj2728/hst/shared/pkg/filelister"
	"github.com/bmj2728/hst/shared/pkg/hostconn"
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
		Level:  hclog.Info,
	})

	// Set up host services - create the implementation
	// HostServices is a struct that embeds the HostFS and HostEnv interfaces
	hostServices := hostserve.NewHostServices(hostserve.NewHostFS(), hostserve.NewHostEnv())
	//Start plugin 1
	flAbspath, err := filepath.Abs("./plugins/filelister/filelister")
	if err != nil {
		logger.Error("Failed to get absolute path", "err", err)
		flAbspath = "./plugins/filelister/filelister"
	}
	flDir, flBin := filepath.Split(flAbspath)
	logger.Info("Starting plugin", "dir", flDir, "bin", flBin)
	// Create the plugin client - plumbing
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  handshakeConfig,
		Plugins:          pluginMap,
		Cmd:              exec.Command(flAbspath),
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

	// Setup host services for the plugin (if supported)
	cid, err := hostconn.EstablishHostServiceConnection(raw, hostServices, logger)
	if err != nil {
		logger.Error("Failed to establish host services", "err", err)
		os.Exit(1)
	}
	if cid != "" {
		logger.Info("Host services established", "bin", flBin, "cid", cid)
	}

	// End plugin 1

	////Start plugin 2
	clAbspath, err := filepath.Abs("./plugins/colorlister/colorlister")
	if err != nil {
		logger.Error("Failed to get absolute path", "err", err)
		clAbspath = "./plugins/colorlister/colorlister"
	}
	clDir, clBin := filepath.Split(clAbspath)
	logger.Info("Starting plugin", "dir", clDir, "bin", clBin)
	color := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  handshakeConfig,
		Plugins:          pluginMap,
		Cmd:              exec.Command(clAbspath),
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

	// Setup host services for the plugin (if supported)
	cdi2, err := hostconn.EstablishHostServiceConnection(rawColor, hostServices, logger)
	if err != nil {
		logger.Error("Failed to establish host services", "err", err)
		os.Exit(1)
	}
	if cdi2 != "" {
		logger.Info("Host services established", "bin", clBin, "cid", cdi2)
	}

	// End plugin 2

	// Test the plugin by listing files in the current directory
	entries, err := fileLister.ListFiles(".")
	if err != nil {
		logger.Error("Failed to list files", "err", err)
		os.Exit(1)
	}

	colorEntries, err := colorlister.ListFiles(".")
	if err != nil {
		logger.Error("Failed to list files", "err", err)
		os.Exit(1)
	}

	logger.Info("Successfully listed files - no color")
	for _, entry := range entries {
		fmt.Println(entry)
	}

	logger.Info("Successfully listed files - with color")
	for _, entry := range colorEntries {
		fmt.Println(entry)
	}

	// Clean shutdown - disconnect from host services
	logger.Info("Shutting down plugins")
	hostconn.DisconnectHostServices(raw, logger)
	hostconn.DisconnectHostServices(rawColor, logger)
	os.Exit(0)
}
