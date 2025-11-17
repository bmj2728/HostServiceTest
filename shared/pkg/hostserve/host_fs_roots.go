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

func newRootHandle() RootHandle {
	return RootHandle(newUUID().String())
}

// OpenRootMap is a nested map associating a ClientID with RootHandles and their corresponding os.Root instances.
type OpenRootMap map[ClientID]map[RootHandle]*OpenRoot

func newOpenRootMap() OpenRootMap {
	return make(OpenRootMap)
}

// OpenRoots manages a thread-safe collection of open root directories accessible by different clients.
type OpenRoots struct {
	roots OpenRootMap
	mu    sync.RWMutex
}

// NewOpenRoots creates and returns a new instance of OpenRoots with an empty OpenRootMap and an initialized mutex.
func newOpenRoots() *OpenRoots {
	return &OpenRoots{
		roots: newOpenRootMap(),
	}
}

type OpenRoot struct {
	root      *os.Root
	openFiles *OpenFiles
}

func newOpenRoot(root *os.Root) *OpenRoot {
	// Lazy init for open files
	return &OpenRoot{
		root: root,
	}
}
