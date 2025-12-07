package hostserve

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/bmj2728/hst/shared/protogen/hostserve/v1"
	"github.com/hashicorp/go-hclog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// fromOpenFileFLags converts the OpenFileFlags enum to appropriate file flags for use with os package operations.
func fromOpenFileFLags(mode hostservev1.OpenFileFlags) int {
	switch mode {
	case hostservev1.OpenFileFlags_READ_ONLY:
		return os.O_RDONLY
	case hostservev1.OpenFileFlags_WRITE_TRUNCATE:
		return os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	case hostservev1.OpenFileFlags_WRITE_APPEND:
		return os.O_WRONLY | os.O_CREATE | os.O_APPEND
	case hostservev1.OpenFileFlags_WRITE_EXCLUSIVE:
		return os.O_WRONLY | os.O_CREATE | os.O_EXCL
	case hostservev1.OpenFileFlags_READ_WRITE:
		return os.O_RDWR
	case hostservev1.OpenFileFlags_READ_WRITE_CREATE:
		return os.O_RDWR | os.O_CREATE
	case hostservev1.OpenFileFlags_READ_WRITE_TRUNCATE:
		return os.O_RDWR | os.O_CREATE | os.O_TRUNC
	case hostservev1.OpenFileFlags_READ_WRITE_APPEND:
		return os.O_RDWR | os.O_CREATE | os.O_APPEND
	default:
		return os.O_RDONLY
	}
}

// toOpenFileFlags converts integer file open flags into hostservev1.OpenFileFlags enumeration values.
func toOpenFileFlags(flags int) hostservev1.OpenFileFlags {
	switch flags {
	case os.O_RDONLY:
		return hostservev1.OpenFileFlags_READ_ONLY
	case os.O_WRONLY | os.O_CREATE | os.O_TRUNC:
		return hostservev1.OpenFileFlags_WRITE_TRUNCATE
	case os.O_WRONLY | os.O_CREATE | os.O_APPEND:
		return hostservev1.OpenFileFlags_WRITE_APPEND
	case os.O_WRONLY | os.O_CREATE | os.O_EXCL:
		return hostservev1.OpenFileFlags_WRITE_EXCLUSIVE
	case os.O_RDWR:
		return hostservev1.OpenFileFlags_READ_WRITE
	case os.O_RDWR | os.O_CREATE:
		return hostservev1.OpenFileFlags_READ_WRITE_CREATE
	case os.O_RDWR | os.O_CREATE | os.O_TRUNC:
		return hostservev1.OpenFileFlags_READ_WRITE_TRUNCATE
	case os.O_RDWR | os.O_CREATE | os.O_APPEND:
		return hostservev1.OpenFileFlags_READ_WRITE_APPEND
	default:
		return hostservev1.OpenFileFlags_READ_ONLY
	}
}

// getRoot resolves the absolute path of the given directory and validates if it is a directory
// before returning an Root object for it.
func getRoot(rootDir string) (*os.Root, error) {
	if !filepath.IsAbs(rootDir) {
		return nil, fmt.Errorf("directory path must be absolute: %s", rootDir)
	}
	info, err := os.Stat(rootDir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", rootDir)
	}
	r, err := os.OpenRoot(rootDir)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// closeRoot ensures the provided root is closed and logs an error if the operation fails.
// It handles logging the root's name and the corresponding error details.
func closeRoot(r *os.Root) {
	err := r.Close()
	if err != nil {
		hclog.Default().Error("Failed to close root", "path", r.Name(), "err", err)
	}
}

// retrieveOpenFile retrieves an open file associated with a given client ID and file handle from the HostFS.
// Returns the *os.File instance if found or an error if the file handle is invalid or the client ID is missing.
func (hf *HostFS) retrieveOpenFile(ctx context.Context, handle FileHandle) (*os.File, error) {

	clientId := getClientIDFromContext(ctx)
	if clientId == "" {
		return nil, fmt.Errorf("client ID not found in context")
	}
	file, err := hf.getOpenFiles().GetFile(clientId, handle)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// protoFileInfoToRemoteFileInfo converts a hostservev1.FileInfo to a RemoteFileInfo instance.
// Returns nil if the input is nil. Transfers basic file information such as name, size, mode, modTime, and isDir.
func protoFileInfoToRemoteFileInfo(pi *hostservev1.FileInfo) *RemoteFileInfo {
	if pi == nil {
		return nil
	}
	rfi := &RemoteFileInfo{
		name:    pi.Name,
		size:    pi.Size,
		mode:    os.FileMode(pi.Mode),
		modTime: pi.ModTime.AsTime(),
		isDir:   pi.IsDir,
	}
	return rfi
}

// fileInfoToProtoFileInfo converts an fs.FileInfo object to a hostservev1.FileInfo protobuf message.
// Returns nil if the input fs.FileInfo is nil.
func fileInfoToProtoFileInfo(fi fs.FileInfo) *hostservev1.FileInfo {
	if fi == nil {
		return nil
	}
	return &hostservev1.FileInfo{
		Name:    fi.Name(),
		Size:    fi.Size(),
		Mode:    uint32(fi.Mode()),
		ModTime: timestamppb.New(fi.ModTime()),
		IsDir:   fi.IsDir(),
	}
}

// getTempPaths returns a pointer to the TempPaths object, creating a new one if it has not been initialized.
func (hf *HostFS) getTempPaths() *TempPaths {
	if hf.tempPaths == nil {
		hf.tempPaths = newTempPaths()
	}
	return hf.tempPaths
}

// getOpenFiles returns a reference to the open files manager, initializing it if not already created.
func (hf *HostFS) getOpenFiles() *OpenFiles {
	if hf.openFiles == nil {
		hf.openFiles = newOpenFiles()
	}
	return hf.openFiles
}

func absToRel(root, path string) (string, error) {
	if filepath.IsAbs(path) {
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return "", fmt.Errorf("failed to get relative path: %w", err)
		}
		return filepath.Clean(rel), nil
	}
	return filepath.Clean(path), nil
}
