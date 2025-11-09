package hostserve

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
)

var (
	ErrInvalidPath = errors.New("invalid path")
)

type HostFS struct {
	//TBD fields
}

func NewHostFS() *HostFS {
	return &HostFS{}
}

// getRoot resolves the absolute path of the given directory and validates if it is a directory
// before returning an Root object for it.
func getRoot(dir string) (*os.Root, error) {
	if !filepath.IsAbs(dir) {
		p, err := filepath.Abs(dir)
		if err != nil {
			return nil, err
		}
		dir = p
	}
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, ErrInvalidPath
	}
	return os.OpenRoot(dir)
}

func closeRoot(r *os.Root) {
	err := r.Close()
	if err != nil {
		hclog.Default().Error("Failed to close root", "path", r.Name(), "err", err)
	}
}

func (hf *HostFS) ReadDir(path string) ([]fs.DirEntry, error) {
	r, err := getRoot(path)
	if err != nil {
		hclog.Default().Error("Failed to open root", "path", path, "err", err)
		return nil, err
	}
	defer closeRoot(r)
	entries, err := fs.ReadDir(r.FS(), ".")
	if err != nil {
		hclog.Default().Error("Failed to read directory", "path", path, "err", err)
		return nil, err
	}

	return entries, nil
}

func (hf *HostFS) ReadFile(dir, file string) ([]byte, error) {
	r, err := getRoot(dir)
	if err != nil {
		hclog.Default().Error("Failed to open root", "path", dir, "err", err)
		return nil, err
	}
	defer closeRoot(r)
	data, err := fs.ReadFile(r.FS(), file)
	if err != nil {
		hclog.Default().Error("Failed to read file", "path", file, "err", err)
		return nil, err
	}
	return data, nil
}
