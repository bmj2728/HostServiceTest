package hostserve

import (
	"os"
	"sync"
)

// RootHandle represents a unique identifier for a root resource within the system.
type RootHandle string

// String returns the string representation of the RootHandle.
func (rh RootHandle) String() string {
	return string(rh)
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

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// FileHandle represents a unique identifier for an open file within a specific client context.
type FileHandle string

// String converts the FileHandle to its underlying string representation.
func (fh FileHandle) String() string {
	return string(fh)
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
