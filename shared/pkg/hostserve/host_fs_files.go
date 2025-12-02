package hostserve

import (
	"fmt"
	"os"
	"sync"

	"github.com/hashicorp/go-hclog"
)

//**************************HANDLE****************************************

// FileHandle represents a unique identifier for an open file in the system.
type FileHandle string

// String returns the string representation of the FileHandle.
func (fh FileHandle) String() string {
	return string(fh)
}

// newFileHandle generates and returns a new unique FileHandle using a UUID.
func newFileHandle() FileHandle {
	return FileHandle(newUUID().String())
}

//**************************Map Structure****************************************

// FileMap represents a mapping of FileHandle identifiers to their corresponding *os.File instances.
type FileMap map[FileHandle]*os.File

func newFileMap() FileMap {
	return make(FileMap)
}

// ClientFileMap is a nested map structure, where each ClientID maps to another map of FileHandle to *os.File.
type ClientFileMap map[ClientID]FileMap

// newClientFileMap initializes and returns a new instance of ClientFileMap.
// It creates an empty map structure to store file handles organized by client IDs.
func newClientFileMap() ClientFileMap {
	return make(ClientFileMap)
}

//**************************Thread-safe Struct****************************************

// OpenFiles is a thread-safe struct for managing a collection of open files organized by client IDs and file handles.
type OpenFiles struct {
	files ClientFileMap
	mu    sync.RWMutex
}

// newOpenFiles initializes and returns a new instance of OpenFiles with an empty ClientFileMap.
func newOpenFiles() *OpenFiles {
	return &OpenFiles{
		files: newClientFileMap(),
	}
}

//**************************Functions****************************************

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
	files, exists := of.files[clientID]
	if !exists {
		return fmt.Errorf("client %s has no open files", clientID)
	}
	file, exists := files[fileHandle]
	if !exists {
		return fmt.Errorf("file handle %s does not exist for client %s", fileHandle, clientID)
	}
	err := file.Close()
	if err != nil {
		hclog.Default().Error("Failed to close file", "file", fileHandle, "err", err)
		return err
	}
	delete(files, fileHandle)
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

func (of *OpenFiles) GetFiles() ClientFileMap {
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
			} else {
				hclog.Default().Debug("Closed file", "file", h)
			}
			delete(of.files[c], h)
		}
	}
}
