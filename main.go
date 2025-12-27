package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/bmj2728/hst/shared/pkg/filelister"
	"github.com/bmj2728/hst/shared/pkg/hostconn"
	"github.com/bmj2728/hst/shared/pkg/hostdemo"
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
	"pl-plugin": &filelister.FileListerGRPCPlugin{},
	"hd-plugin": &hostdemo.HostDemoGRPCPlugin{},
}

func startPluginClient(pluginPath string, logger hclog.Logger) (string, *plugin.Client, error) {
	absPath, err := filepath.Abs(pluginPath)
	if err != nil {
		logger.Error("Failed to get absolute path", "err", err)
		return "", nil, fmt.Errorf("failed to get absolute path: %w", err)
	}
	_, bin := filepath.Split(absPath)
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  handshakeConfig,
		Plugins:          pluginMap,
		Cmd:              exec.Command(absPath),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger:           logger,
	})
	return bin, client, nil
}

func connectToPlugin(client *plugin.Client, dispenseName string, logger hclog.Logger) (interface{}, error) {
	rpcClient, err := client.Client()
	if err != nil {
		logger.Error("Failed to get RPC client", "err", err)
		return nil, err
	}
	raw, err := rpcClient.Dispense(dispenseName)
	if err != nil {
		logger.Error("Failed to get RPC client", "err", err)
		return nil, err
	}
	return raw, nil
}

func connectHost(plugin interface{}, plugBin string, hostServices *hostserve.HostServices, logger hclog.Logger) error {
	if plugin == nil {
		logger.Error("Plugin is nil")
		return fmt.Errorf("plugin is nil")
	}
	if hostServices == nil {
		logger.Error("Host services is nil")
		return fmt.Errorf("host services is nil")
	}
	cid, err := hostconn.EstablishHostServiceConnection(plugin, hostServices, logger)
	if err != nil {
		logger.Error("Failed to establish host services", "err", err)
		return fmt.Errorf("failed to establish host services: %w", err)
	}
	if cid != "" {
		err = hostServices.ActiveClients().AddClient(cid, plugBin)
		if err != nil {
			logger.Error("Failed to add client", "err", err)
			return fmt.Errorf("failed to add client: %w", err)
		}
		logger.Info("Host services established", "bin", plugBin, "cid", cid)
	}
	return nil
}

func main() {

	//clean up
	err := os.RemoveAll("./created_dir")
	if err != nil {
		hclog.Default().Error("Failed to remove temp dir", "err", err)
	}

	err = os.RemoveAll("./nested")
	if err != nil {
		hclog.Default().Error("Failed to remove temp dir", "err", err)
	}

	err = os.Rename("/home/brian/GolandProjects/HostServiceTest/rename_works.md",
		"/home/brian/GolandProjects/HostServiceTest/renameme.md")
	if err != nil {
		hclog.Default().Error("Failed to rename file", "err", err)
	}

	_, err = os.Create("./deleteme.txt")

	// Set up logging
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "host",
		Output: os.Stdout,
		Level:  hclog.Info,
		Color:  hclog.ForceColor,
	})
	hclog.SetDefault(logger)

	// Set up host services
	hostServices := hostserve.NewHostServices(hostserve.NewHostFS(), hostserve.NewHostEnv())

	//Start plugin 1

	flBin, flClient, err := startPluginClient("./plugins/filelister/filelister", logger)
	if err != nil {
		logger.Error("Failed to start plugin", "err", err)
	}
	raw, err := connectToPlugin(flClient, "fl-plugin", logger)
	if err != nil {
		logger.Error("Failed to connect to plugin", "err", err)
	}
	// Coerce the raw interface to the FileLister type
	fileLister := raw.(filelister.FileLister)
	err = connectHost(raw, flBin, hostServices, logger)
	if err != nil {
		logger.Error("Failed to connect host", "err", err)
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
	cid2, err := hostconn.EstablishHostServiceConnection(rawColor, hostServices, logger)
	if err != nil {
		logger.Error("Failed to establish host services", "err", err)
		os.Exit(1)
	}
	if cid2 != "" {
		err = hostServices.ActiveClients().AddClient(cid2, clBin)
		if err != nil {
			logger.Error("Failed to add client", "err", err)
			os.Exit(1)
		}
		logger.Info("Host services established", "bin", clBin, "cid", cid2)
	}

	// End plugin 2

	////Start plugin 3
	plAbspath, err := filepath.Abs("./plugins/pylelister/dist/pylelister")
	if err != nil {
		logger.Error("Failed to get absolute path", "err", err)
		plAbspath = "./plugins/pylelister/dist/pylelister"
	}
	plDir, plBin := filepath.Split(plAbspath)
	logger.Info("Starting plugin", "dir", plDir, "bin", plBin)
	python := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  handshakeConfig,
		Plugins:          pluginMap,
		Cmd:              exec.Command(plAbspath),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger:           logger,
	})
	defer python.Kill()

	// Connect via gRPC - porcelain
	rpcClientPython, err := python.Client()
	if err != nil {
		logger.Error("Failed to get RPC client", "err", err)
		os.Exit(1)
	}

	// Request the FileLister plugin - the raw interface
	rawPython, err := rpcClientPython.Dispense("pl-plugin")
	if err != nil {
		logger.Error("Failed to dispense plugin", "err", err)
		os.Exit(1)
	}

	// Coerce the raw interface to the FileLister type
	pythonlister := rawPython.(filelister.FileLister)

	// Setup host services for the plugin (if supported)
	cid3, err := hostconn.EstablishHostServiceConnection(rawPython, hostServices, logger)
	if err != nil {
		logger.Error("Failed to establish host services", "err", err)
		os.Exit(1)
	}
	if cid3 != "" {
		err = hostServices.ActiveClients().AddClient(cid3, plBin)
		if err != nil {
			logger.Error("Failed to add client", "err", err)
			os.Exit(1)
		}
		logger.Info("Host services established", "bin", plBin, "cid", cid3)
	}

	// End plugin 3

	////Start plugin Host Demo
	hdAbspath, err := filepath.Abs("./plugins/hostdemo/hostdemo")
	if err != nil {
		logger.Error("Failed to get absolute path", "err", err)
		hdAbspath = "./plugins/hd/hd"
	}
	hdDir, hdBin := filepath.Split(hdAbspath)
	logger.Info("Starting plugin", "dir", hdDir, "bin", hdBin)
	hd := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  handshakeConfig,
		Plugins:          pluginMap,
		Cmd:              exec.Command(hdAbspath),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger:           logger,
	})
	defer hd.Kill()

	// Connect via gRPC - porcelain
	rpcClientHostDemo, err := hd.Client()
	if err != nil {
		logger.Error("Failed to get RPC client", "err", err)
		os.Exit(1)
	}

	// Request the FileLister plugin - the raw interface
	rawHD, err := rpcClientHostDemo.Dispense("hd-plugin")
	if err != nil {
		logger.Error("Failed to dispense plugin", "err", err)
		os.Exit(1)
	}

	demo := rawHD.(hostdemo.HostDemo)

	// Setup host services for the plugin (if supported)
	cid4, err := hostconn.EstablishHostServiceConnection(rawHD, hostServices, logger)
	if err != nil {
		logger.Error("Failed to establish host services", "err", err)
		os.Exit(1)
	}
	if cid4 != "" {
		err = hostServices.ActiveClients().AddClient(cid4, hdBin)
		if err != nil {
			logger.Error("Failed to add client", "err", err)
			os.Exit(1)
		}
		logger.Info("Host services established", "bin", hdBin, "cid", cid4)
	}

	// End plugin 4

	// Run some demos

	go func() {
		// Test the plugin by listing files in the current directory
		entries, err := fileLister.ListFiles("/home/brian/GolandProjects/HostServiceTest", "plugins/filelister")
		if err != nil {
			logger.Error("Failed to list files", "err", err)
			os.Exit(1)
		}
		logger.Info("Successfully listed files - no color")
		for _, entry := range entries {
			fmt.Println(entry)
		}
	}()

	go func() {
		colorEntries, err := colorlister.ListFiles("/home/brian/GolandProjects/HostServiceTest", "/home/brian/GolandProjects/HostServiceTest")
		if err != nil {
			logger.Error("Failed to list files", "err", err)
			os.Exit(1)
		}
		logger.Info("Successfully listed files - with color")
		for _, entry := range colorEntries {
			fmt.Println(entry)
		}
	}()

	go func() {
		pythonEntries, err := pythonlister.ListFiles("/home/brian/GolandProjects/HostServiceTest", "/home/brian/GolandProjects/HostServiceTest")
		if err != nil {
			logger.Error("Failed to list files", "err", err)
			os.Exit(1)
		}
		logger.Info("Successfully listed files - with color")
		for _, entry := range pythonEntries {
			fmt.Println(entry)
		}
	}()

	go func() {
		val, err := demo.GetEnvDemo(context.Background(), "GOPATH")
		if err != nil {
			logger.Error("Failed get env demo", "err", err)
		}
		fmt.Println(val)
	}()

	go func() {
		envDat, err := demo.EnvDemo(context.Background())
		if err != nil {
			logger.Error("Failed env demo", "err", err)
		}
		fmt.Println(envDat)
	}()

	go func() {
		tempDemo, err := demo.TempDemo(context.Background(), "Host-Demo-*-Temp", "This is a temp file")
		if err != nil {
			logger.Error("Failed temp demo", "err", err)
		}
		fmt.Println(tempDemo)
	}()

	time.Sleep(100 * time.Millisecond)

	// Clean shutdown - disconnect from host services
	logger.Info("Shutting down plugins")
	hostconn.DisconnectHostServices(raw, logger)
	hostconn.DisconnectHostServices(rawColor, logger)
	hfs, ok := hostServices.IHostFS.(*hostserve.HostFS)
	if !ok {
		logger.Error("Failed to cast host services to HostFS")
		os.Exit(1)
	}
	hfs.Cleanup()
	hostServices.ActiveClients().Clear()

	plugin.CleanupClients() //make sure we actually shutdown the plugins

	os.Exit(0)
}
