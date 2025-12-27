package tui

import (
	"github.com/bmj2728/hst/shared/pkg/filelister"
	"github.com/bmj2728/hst/shared/pkg/hostdemo"
	"github.com/hashicorp/go-plugin"
)

// ViewMode represents the current view state of the TUI
type ViewMode int

const (
	ViewMainMenu ViewMode = iota
	ViewPluginMenu
	ViewInputCollection
	ViewOutput
)

// PluginType identifies the type of plugin
type PluginType string

const (
	PluginFileLister PluginType = "filelister"
	PluginHostDemo   PluginType = "hostdemo"
)

// PluginInfo holds information about a loaded plugin
type PluginInfo struct {
	Name      string
	Type      PluginType
	Client    *plugin.Client
	Interface interface{}
	Functions []PluginFunction
}

// PluginFunction describes a callable function on a plugin
type PluginFunction struct {
	Name        string
	DisplayName string
	Inputs      []FunctionInput
}

// FunctionInput describes an input parameter for a plugin function
type FunctionInput struct {
	Name        string
	DisplayName string
	Value       string
}

// ExecutionResult holds the result of plugin execution
type ExecutionResult struct {
	PluginName string
	Function   string
	Output     string
	Error      error
}

// GetFileListerFunctions returns the available functions for file lister plugins
func GetFileListerFunctions() []PluginFunction {
	return []PluginFunction{
		{
			Name:        "ListFiles",
			DisplayName: "List Files",
			Inputs: []FunctionInput{
				{Name: "rootDir", DisplayName: "Root Directory", Value: ""},
				{Name: "path", DisplayName: "Path", Value: ""},
			},
		},
	}
}

// GetHostDemoFunctions returns the available functions for the host demo plugin
func GetHostDemoFunctions() []PluginFunction {
	return []PluginFunction{
		{
			Name:        "GetEnvDemo",
			DisplayName: "GetEnv Demo",
			Inputs: []FunctionInput{
				{Name: "key", DisplayName: "Environment Variable", Value: ""},
			},
		},
		{
			Name:        "EnvDemo",
			DisplayName: "Env Demo (System Info)",
			Inputs:      []FunctionInput{},
		},
		{
			Name:        "TempDemo",
			DisplayName: "Temp Demo",
			Inputs: []FunctionInput{
				{Name: "pattern", DisplayName: "Pattern", Value: "Host-Demo-*-Temp"},
				{Name: "textToWrite", DisplayName: "Text to Write", Value: "This is a temp file"},
			},
		},
	}
}

// ExecutePluginFunction executes the specified function on a plugin
func ExecutePluginFunction(plugin *PluginInfo, function PluginFunction) ExecutionResult {
	result := ExecutionResult{
		PluginName: plugin.Name,
		Function:   function.Name,
	}

	switch plugin.Type {
	case PluginFileLister:
		lister, ok := plugin.Interface.(filelister.FileLister)
		if !ok {
			result.Error = ErrInvalidPluginType
			return result
		}

		if function.Name == "ListFiles" {
			rootDir := function.Inputs[0].Value
			path := function.Inputs[1].Value
			entries, err := lister.ListFiles(rootDir, path)
			if err != nil {
				result.Error = err
				return result
			}
			output := ""
			for _, entry := range entries {
				output += entry + "\n"
			}
			result.Output = output
		}

	case PluginHostDemo:
		demo, ok := plugin.Interface.(hostdemo.HostDemo)
		if !ok {
			result.Error = ErrInvalidPluginType
			return result
		}

		switch function.Name {
		case "GetEnvDemo":
			key := function.Inputs[0].Value
			output, err := demo.GetEnvDemo(key)
			if err != nil {
				result.Error = err
				return result
			}
			result.Output = output

		case "EnvDemo":
			output, err := demo.EnvDemo()
			if err != nil {
				result.Error = err
				return result
			}
			result.Output = output

		case "TempDemo":
			pattern := function.Inputs[0].Value
			textToWrite := function.Inputs[1].Value
			output, err := demo.TempDemo(pattern, textToWrite)
			if err != nil {
				result.Error = err
				return result
			}
			result.Output = output
		}
	}

	return result
}
