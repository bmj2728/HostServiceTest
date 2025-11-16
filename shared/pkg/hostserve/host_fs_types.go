package hostserve

import (
	"fmt"
	"os"
	"sync"
)

// RootHandle represents a unique identifier for a root resource within the system.
type RootHandle string

// String returns the string representation of the RootHandle.
func (rh RootHandle) String() string {
	return string(rh)
}

func NewRootHandle() RootHandle {
	return RootHandle(NewUUID().String())
}

// OpenRootMap is a nested map associating a ClientID with RootHandles and their corresponding os.Root instances.
type OpenRootMap map[ClientID]map[RootHandle]*os.Root

// OpenRoots manages a thread-safe collection of open root directories accessible by different clients.
type OpenRoots struct {
	roots OpenRootMap
	mu    sync.RWMutex
}

// NewOpenRoots creates and returns a new instance of OpenRoots with an empty OpenRootMap and an initialized mutex.
func NewOpenRoots() *OpenRoots {
	return &OpenRoots{
		roots: make(OpenRootMap),
	}
}

func (or *OpenRoots) AddRoot(clientID ClientID, rootHandle RootHandle, root *os.Root) error {
	or.mu.Lock()
	defer or.mu.Unlock()
	if or.roots[clientID] == nil {
		or.roots[clientID] = make(map[RootHandle]*os.Root)
	}
	if _, exists := or.roots[clientID][rootHandle]; exists {
		return fmt.Errorf("root handle %s already exists for client %s", rootHandle, clientID)
	}
	or.roots[clientID][rootHandle] = root
	return nil
}

func (or *OpenRoots) RemoveRoot(clientID ClientID, rootHandle RootHandle) error {
	or.mu.Lock()
	defer or.mu.Unlock()
	if _, exists := or.roots[clientID]; !exists {
		return fmt.Errorf("client %s has no open roots", clientID)
	}
	if _, exists := or.roots[clientID][rootHandle]; !exists {
		return fmt.Errorf("root handle %s does not exist for client %s", rootHandle, clientID)
	}
	root := or.roots[clientID][rootHandle]
	err := root.Close()
	if err != nil {
		delete(or.roots[clientID], rootHandle)
		return err
	}
	delete(or.roots[clientID], rootHandle)
	return nil
}

func (or *OpenRoots) GetRootsByClient(clientID ClientID) (map[RootHandle]*os.Root, error) {
	or.mu.RLock()
	defer or.mu.RUnlock()
	if _, exists := or.roots[clientID]; !exists {
		return nil, fmt.Errorf("client %s has no open roots", clientID)
	}
	return or.roots[clientID], nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// FileHandle represents a unique identifier for an open file within a specific client context.
type FileHandle string

// String converts the FileHandle to its underlying string representation.
func (fh FileHandle) String() string {
	return string(fh)
}

func NewFileHandle() FileHandle {
	return FileHandle(NewUUID().String())
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
