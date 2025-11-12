package hostserve

import (
	"os"
	"path/filepath"

	"github.com/bmj2728/hst/shared/protogen/hostserve/v1"
	"github.com/hashicorp/go-hclog"
)

// openFileModeToFlags converts the OpenFileMode enum to appropriate file flags for use with os package operations.
func openFileModeToFlags(mode hostservev1.OpenFileMode) int {
	switch mode {
	case hostservev1.OpenFileMode_READ_ONLY:
		return os.O_RDONLY
	case hostservev1.OpenFileMode_WRITE_TRUNCATE:
		return os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	case hostservev1.OpenFileMode_WRITE_APPEND:
		return os.O_WRONLY | os.O_CREATE | os.O_APPEND
	case hostservev1.OpenFileMode_WRITE_EXCLUSIVE:
		return os.O_WRONLY | os.O_CREATE | os.O_EXCL
	case hostservev1.OpenFileMode_READ_WRITE:
		return os.O_RDWR
	case hostservev1.OpenFileMode_READ_WRITE_CREATE:
		return os.O_RDWR | os.O_CREATE
	case hostservev1.OpenFileMode_READ_WRITE_TRUNCATE:
		return os.O_RDWR | os.O_CREATE | os.O_TRUNC
	case hostservev1.OpenFileMode_READ_WRITE_APPEND:
		return os.O_RDWR | os.O_CREATE | os.O_APPEND
	default:
		return os.O_RDONLY
	}
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
	r, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	// TODO add to the map of open roots
	return r, nil
}

// closeRoot ensures the provided root is closed and logs an error if the operation fails.
// It handles logging the root's name and the corresponding error details.
func closeRoot(r *os.Root) {
	err := r.Close()
	if err != nil {
		hclog.Default().Error("Failed to close root", "path", r.Name(), "err", err)
	}
	// TODO remove from the map of open roots
}
