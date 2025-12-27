package main

import (
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
	"file-lister":  &filelister.FileListerGRPCPlugin{},
	"color-lister": &filelister.FileListerGRPCPlugin{},
	"py-lister":    &filelister.FileListerGRPCPlugin{},
	"host-demo":    &hostdemo.HostDemoGRPCPlugin{},
}

func startPluginClient(pluginPath string, dispenseName string, hostServices *hostserve.HostServices, logger hclog.Logger) (interface{}, *plugin.Client, error) {
	absPath, err := filepath.Abs(pluginPath)
	if err != nil {
		logger.Error("Failed to get absolute path", "err", err)
		return nil, nil, fmt.Errorf("failed to get absolute path: %w", err)
	}
	_, bin := filepath.Split(absPath)
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  handshakeConfig,
		Plugins:          pluginMap,
		Cmd:              exec.Command(absPath),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger:           logger,
	})
	rpcClient, err := client.Client()
	if err != nil {
		logger.Error("Failed to get RPC client", "err", err)
		return nil, client, err
	}
	raw, err := rpcClient.Dispense(dispenseName)
	if err != nil {
		logger.Error("Failed to get RPC client", "err", err)
		return nil, client, err
	}
	cid, err := hostconn.EstablishHostServiceConnection(raw, hostServices, logger)
	if err != nil {
		logger.Error("Failed to establish host services", "err", err)
		return nil, client, fmt.Errorf("failed to establish host services: %w", err)
	}
	if cid != "" {
		err = hostServices.ActiveClients().AddClient(cid, bin)
		if err != nil {
			logger.Error("Failed to add client", "err", err)
			return nil, client, fmt.Errorf("failed to add client: %w", err)
		}
		logger.Info("Host services established", "bin", bin, "cid", cid)
		hostServices.AddRawPlugin(raw)
	}
	return raw, client, nil
}

func resetState() error {
	err := os.RemoveAll("./created_dir")
	if err != nil {
		return fmt.Errorf("failed to remove created_dir: %w", err)
	}
	err = os.RemoveAll("./nested")
	if err != nil {
		return fmt.Errorf("failed to remove nested: %w", err)
	}
	err = os.Rename("./rename_works.md",
		"./renameme.md")
	if err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}
	_, err = os.Create("./deleteme.txt")
	return err
}

func shutdown(delay time.Duration, hostServices *hostserve.HostServices, logger hclog.Logger) {
	time.Sleep(delay)
	logger.Info("Shutting down plugins")
	for _, rawPlugin := range hostServices.RawPlugins() {
		hostconn.DisconnectHostServices(rawPlugin, logger)
	}

	hfs, ok := hostServices.IHostFS.(*hostserve.HostFS)
	if !ok {
		logger.Error("Failed to cast host services to HostFS")
		return
	}
	hfs.Cleanup()
	hostServices.ActiveClients().Clear()

	plugin.CleanupClients() //make sure we actually shutdown the plugins

	os.Exit(0)
}

func main() {

	// Cleanup prior runs
	err := resetState()
	if err != nil {
		hclog.Default().Error("Failed to reset state", "err", err)
		return
	}

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

	//Start File Lister
	flRaw, flClient, err := startPluginClient("./plugins/filelister/filelister", "file-lister", hostServices, logger)
	if err != nil {
		logger.Error("Failed to start plugin", "err", err)
		return
	}
	defer flClient.Kill()
	fileLister := flRaw.(filelister.FileLister)

	//Start Color Lister
	clRaw, clClient, err := startPluginClient("./plugins/colorlister/colorlister", "color-lister", hostServices, logger)
	if err != nil {
		logger.Error("Failed to start plugin", "err", err)
		return
	}
	defer clClient.Kill()
	colorLister := clRaw.(filelister.FileLister)

	////Start Python Lister

	pyRaw, pyClient, err := startPluginClient("./plugins/pylelister/dist/pylelister", "py-lister", hostServices, logger)
	if err != nil {
		logger.Error("Failed to start plugin", "err", err)
		return
	}
	defer pyClient.Kill()
	pyLister := pyRaw.(filelister.FileLister)

	//Start Host Demo

	hdRaw, hdClient, err := startPluginClient("./plugins/hostdemo/hostdemo", "host-demo", hostServices, logger)
	if err != nil {
		logger.Error("Failed to start plugin", "err", err)
		return
	}
	defer hdClient.Kill()
	demo := hdRaw.(hostdemo.HostDemo)

	// Run some demos
	cwd, err := os.Getwd()

	go func() {

		entries, err := fileLister.ListFiles(cwd, "plugins/filelister")
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
		colorEntries, err := colorLister.ListFiles(cwd, cwd)
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
		pythonEntries, err := pyLister.ListFiles(cwd, cwd)
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
		val, err := demo.GetEnvDemo("GOPATH")
		if err != nil {
			logger.Error("Failed get env demo", "err", err)
		}
		fmt.Println(val)
	}()

	go func() {
		envDat, err := demo.EnvDemo()
		if err != nil {
			logger.Error("Failed env demo", "err", err)
		}
		fmt.Println(envDat)
	}()

	go func() {
		tempDemo, err := demo.TempDemo("Host-Demo-*-Temp", "This is a temp file")
		if err != nil {
			logger.Error("Failed temp demo", "err", err)
		}
		fmt.Println(tempDemo)
	}()

	shutdown(200*time.Millisecond, hostServices, logger)

}
