package tui

import "errors"

var (
	ErrInvalidPluginType = errors.New("invalid plugin type")
	ErrPluginNotFound    = errors.New("plugin not found")
	ErrNoInputs          = errors.New("no inputs provided")
)
