package hostserve

import (
	"fmt"
	"os"
	"sync"

	"github.com/hashicorp/go-hclog"
)

// FileHandle represents a unique identifier for an open file within a specific client context.
type FileHandle string

// String converts the FileHandle to its underlying string representation.
func (fh FileHandle) String() string {
	return string(fh)
}

func newFileHandle() FileHandle {
	return FileHandle(newUUID().String())
}

// OpenFileMap represents a mapping of ClientIDs to their associated FileHandles and open file pointers.
type OpenFileMap map[ClientID]map[FileHandle]*os.File

// OpenFiles manages a thread-safe collection of open file references, grouped by client and file handle.
type OpenFiles struct {
	files OpenFileMap
	mu    sync.RWMutex
}

// NewOpenFiles initializes and returns a new instance of OpenFiles with an empty OpenFileMap and a RWMutex.
func NewOpenFiles() *OpenFiles {
	return &OpenFiles{
		files: make(OpenFileMap),
	}
}

func (of *OpenFiles) AddFile(clientID ClientID, fileHandle FileHandle, file *os.File) error {
	of.mu.Lock()
	defer of.mu.Unlock()
	if _, exists := of.files[clientID]; !exists {
		of.files[clientID] = make(map[FileHandle]*os.File)
	}
	if _, exists := of.files[clientID][fileHandle]; exists {
		return fmt.Errorf("file handle %s already exists for client %s", fileHandle, clientID)
	}
	of.files[clientID][fileHandle] = file
	return nil
}

func (of *OpenFiles) RemoveFile(clientID ClientID, fileHandle FileHandle) error {
	of.mu.Lock()
	defer of.mu.Unlock()
	if _, exists := of.files[clientID]; !exists {
		return fmt.Errorf("client %s has no open files", clientID)
	}
	if _, exists := of.files[clientID][fileHandle]; !exists {
		return fmt.Errorf("file handle %s does not exist for client %s", fileHandle, clientID)
	}
	err := of.files[clientID][fileHandle].Close()
	if err != nil {
		delete(of.files[clientID], fileHandle)
		return err
	}
	delete(of.files[clientID], fileHandle)
	return nil
}

func (of *OpenFiles) GetFilesByClient(clientID ClientID) (map[FileHandle]*os.File, error) {
	of.mu.RLock()
	defer of.mu.RUnlock()
	if _, exists := of.files[clientID]; !exists {
		return nil, fmt.Errorf("client %s has no open files", clientID)
	}
	return of.files[clientID], nil
}

func (of *OpenFiles) GetFile(clientID ClientID, fileHandle FileHandle) (*os.File, error) {
	files, err := of.GetFilesByClient(clientID)
	if err != nil {
		return nil, err
	}
	if _, exists := files[fileHandle]; !exists {
		return nil, fmt.Errorf("file handle %s does not exist for client %s", fileHandle, clientID)
	}
	return files[fileHandle], nil
}

func (of *OpenFiles) GetFiles() OpenFileMap {
	of.mu.RLock()
	defer of.mu.RUnlock()
	return of.files
}

func (of *OpenFiles) Len() int {
	of.mu.RLock()
	defer of.mu.RUnlock()
	length := 0
	for _, files := range of.files {
		length += len(files)
	}
	return length
}

func (of *OpenFiles) CloseAll() {
	of.mu.Lock()
	defer of.mu.Unlock()
	for c, files := range of.files {
		hclog.Default().Debug("Closing files for client", "client", c)
		for h, file := range files {
			err := file.Close()
			if err != nil {
				hclog.Default().Error("Failed to close file", "file", h, "err", err)
			}
			hclog.Default().Debug("Closing file", "file", h)
			delete(of.files[c], h)
		}
	}
}
