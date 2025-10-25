package hostserve

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
)

type IHostServices interface {
	ReadDir(path string) ([]fs.DirEntry, error)
}

type HostServices struct {
}

func (hs *HostServices) ReadDir(path string) ([]fs.DirEntry, error) {
	if !filepath.IsAbs(path) {
		p, err := filepath.Abs(path)
		if err != nil {
			hclog.Default().Error("Failed to get absolute path", "path", path, "err", err)
			return nil, err
		}
		path = p
	}
	info, err := os.Stat(path)
	if err != nil {
		hclog.Default().Error("Failed to get file info", "path", path, "err", err)
		return nil, err
	}
	if !info.IsDir() {
		hclog.Default().Error("Path is not a directory", "path", path)
		return nil, err
	}
	r, err := os.OpenRoot(path)
	if err != nil {
		hclog.Default().Error("Failed to open root", "path", path, "err", err)
		return nil, err
	}
	defer func(r *os.Root) {
		err := r.Close()
		if err != nil {
			hclog.Default().Error("Failed to close root", "path", path, "err", err)
		}
	}(r)
	entries, err := fs.ReadDir(r.FS(), ".")
	if err != nil {
		hclog.Default().Error("Failed to read directory", "path", path, "err", err)
		return nil, err
	}
	for _, entry := range entries {
		fmt.Println(entry.Name(), " ", entry.IsDir())
	}
	return entries, nil
}
