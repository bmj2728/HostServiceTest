package hostserve

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
)

type IHostServices interface {
	ReadDir(path string) []fs.DirEntry
}

type HostServices struct {
}

func (hs *HostServices) ReadDir(path string) []fs.DirEntry {
	if !filepath.IsAbs(path) {
		p, err := filepath.Abs(path)
		if err != nil {
			hclog.Default().Error("Failed to get absolute path", "path", path, "err", err)
			return nil
		}
		path = p
	}
	info, err := os.Stat(path)
	if err != nil {
		hclog.Default().Error("Failed to get file info", "path", path, "err", err)
		return nil
	}
	if !info.IsDir() {
		hclog.Default().Error("Path is not a directory", "path", path)
		return nil
	}
	r, err := os.OpenRoot(path)
	if err != nil {
		hclog.Default().Error("Failed to open root", "path", path, "err", err)
		return nil
	}
	defer func(r *os.Root) {
		err := r.Close()
		if err != nil {
			hclog.Default().Error("Failed to close root", "path", path, "err", err)
		}
	}(r)
	entries, err := fs.ReadDir(r.FS(), path)
	if err != nil {
		hclog.Default().Error("Failed to read directory", "path", path, "err", err)
		return nil
	}
	return entries
}
