package main

import (
	"fmt"
	"os"
	"time"

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

	// Run some demos - to be replaced with interactive TUI
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

	// Clean up and shutdown
	utils.Shutdown(200*time.Millisecond, hostServices, logger)

}
