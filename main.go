package main

import (
	"fmt"
	"os"

	"github.com/bmj2728/hst/internal/tui"
	"github.com/bmj2728/hst/internal/utils"
	"github.com/bmj2728/hst/shared/pkg/filelister"
	"github.com/bmj2728/hst/shared/pkg/hostdemo"
	"github.com/bmj2728/hst/shared/pkg/hostserve"
	"github.com/hashicorp/go-hclog"
)

func main() {

	//SETUP

	// Cleanup prior runs
	err := utils.ResetDemoState()
	if err != nil {
		hclog.Default().Error("Failed to reset state", "err", err)
	}

	// Set up logging

	logFile, err := os.OpenFile("hst-demo.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		return
	}
	defer func(logFile *os.File) {
		err := logFile.Close()
		if err != nil {
			fmt.Printf("Failed to close log file: %v\n", err)
		}
	}(logFile)

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "host",
		Output: logFile,
		Level:  hclog.Info,
		Color:  hclog.ForceColor,
	})
	hclog.SetDefault(logger)

	// Set up host services
	hostServices := hostserve.NewHostServices(hostserve.NewHostFS(), hostserve.NewHostEnv())

	//Start File Lister
	flRaw, flClient, err := utils.LaunchPluginClient("./plugins/filelister/filelister", "file-lister", hostServices, logger)
	if err != nil {
		logger.Error("Failed to start plugin", "err", err)
		return
	}
	defer flClient.Kill()
	fileLister := flRaw.(filelister.FileLister)

	//Start Color Lister
	clRaw, clClient, err := utils.LaunchPluginClient("./plugins/colorlister/colorlister", "color-lister", hostServices, logger)
	if err != nil {
		logger.Error("Failed to start plugin", "err", err)
		return
	}
	defer clClient.Kill()
	colorLister := clRaw.(filelister.FileLister)

	//Start Python Lister

	pyRaw, pyClient, err := utils.LaunchPluginClient("./plugins/pylelister/dist/pylelister", "py-lister", hostServices, logger)
	if err != nil {
		logger.Error("Failed to start plugin", "err", err)
		return
	}
	defer pyClient.Kill()
	pyLister := pyRaw.(filelister.FileLister)

	//Start Host Demo

	hdRaw, hdClient, err := utils.LaunchPluginClient("./plugins/hostdemo/hostdemo", "host-demo", hostServices, logger)
	if err != nil {
		logger.Error("Failed to start plugin", "err", err)
		return
	}
	defer hdClient.Kill()
	demo := hdRaw.(hostdemo.HostDemo)

	//END SETUP

	// Build plugin info for TUI
	plugins := []tui.PluginInfo{
		{
			Name:      "File Lister (Go)",
			Type:      tui.PluginFileLister,
			Client:    flClient,
			Interface: fileLister,
			Functions: tui.GetFileListerFunctions(),
		},
		{
			Name:      "Color Lister (Go)",
			Type:      tui.PluginColorLister,
			Client:    clClient,
			Interface: colorLister,
			Functions: tui.GetFileListerFunctions(),
		},
		{
			Name:      "Python Lister",
			Type:      tui.PluginPyLister,
			Client:    pyClient,
			Interface: pyLister,
			Functions: tui.GetFileListerFunctions(),
		},
		{
			Name:      "Host Demo",
			Type:      tui.PluginHostDemo,
			Client:    hdClient,
			Interface: demo,
			Functions: tui.GetHostDemoFunctions(),
		},
	}

	// Launch the TUI
	err = tui.Launch(plugins, hostServices, logger)
	if err != nil {
		logger.Error("TUI error", "err", err)
		os.Exit(1)
	}

}
