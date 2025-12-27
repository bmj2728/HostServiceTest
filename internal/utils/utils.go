package utils

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

func LaunchPluginClient(pluginPath string, dispenseName string, hostServices *hostserve.HostServices, logger hclog.Logger) (interface{}, *plugin.Client, error) {
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

func ResetDemoState() error {
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
	info, err := os.Stat("./hst-demo.log")
	if err != nil {
		return fmt.Errorf("failed to stat log file: %w", err)
	}
	if info.Size() > 1024*1024*2 {
		err = os.Truncate("./hst-demo.log", 0)
		if err != nil {
			return fmt.Errorf("failed to truncate log file: %w", err)
		}
	}
	return err
}

func Shutdown(delay time.Duration, hostServices *hostserve.HostServices, logger hclog.Logger) {
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
