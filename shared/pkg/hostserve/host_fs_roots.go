package hostserve

import (
	"fmt"
	"os"
	"sync"
)

//**************************HANDLE****************************************

// RootHandle represents a unique identifier for a root resource within the system.
type RootHandle string

// String returns the string representation of the RootHandle.
func (rh RootHandle) String() string {
	return string(rh)
}

func newRootHandle() RootHandle {
	return RootHandle(newUUID().String())
}

//**************************Root Wrapper*************************************

// OpenRoot represents a structure managing a root resource and its associated open files.
type OpenRoot struct {
	root      *os.Root
	openFiles *OpenFiles
}

// newOpenRoot initializes a new OpenRoot instance with a given os.Root reference.
// It is used to manage filesystem operations related to the provided root.
func newOpenRoot(root *os.Root) *OpenRoot {
	return &OpenRoot{
		root: root,
	}
}

// Root returns the root resource managed by the OpenRoot instance.
func (or *OpenRoot) Root() *os.Root {
	if or == nil {
		return nil
	}
	return or.root
}

// OpenFiles ensures the OpenFiles instance is initialized and returns it, creating a new one if it is nil.
func (or *OpenRoot) OpenFiles() *OpenFiles {
	if or.openFiles == nil {
		or.openFiles = NewOpenFiles()
	}
	return or.openFiles
}

//**************************Map Structure*************************************

// RootMap represents a mapping of RootHandle identifiers to their corresponding OpenRoot instances.
type RootMap map[RootHandle]*OpenRoot

// newRootsMap initializes and returns a new, empty RootsMap instance.
func newRootsMap() RootMap {
	return make(RootMap)
}

// ClientRootMap is a nested map associating a ClientID with RootHandles and their corresponding os.Root instances.
type ClientRootMap map[ClientID]RootMap

func newOpenRootMap() ClientRootMap {
	return make(ClientRootMap)
}

// Roots retrieves the RootMap associated with the given clientID. Returns nil if clientID is empty or not found.
func (orm ClientRootMap) Roots(clientID ClientID) RootMap {
	if clientID == "" {
		return nil
	}
	return orm[clientID]
}

// Len returns the number of RootHandles associated with the provided clientID in the ClientRootMap.
func (orm ClientRootMap) Len(clientID ClientID) int {
	return len(orm.Roots(clientID))
}

//**************************Host FS Level Collection*************************************

// OpenRoots manages a thread-safe collection of open root directories accessible by different clients.
type OpenRoots struct {
	roots ClientRootMap
	mu    sync.RWMutex
}

// NewOpenRoots creates and returns a new instance of OpenRoots with an empty ClientRootMap and an initialized mutex.
func newOpenRoots() *OpenRoots {
	return &OpenRoots{
		roots: newOpenRootMap(),
	}
}

func (or *OpenRoots) Roots(clientID ClientID) RootMap {
	if or.roots == nil {
		or.roots = newOpenRootMap()
	}
	or.mu.RLock()
	defer or.mu.RUnlock()
	roots := or.roots[clientID]
	if roots == nil {
		return nil
	}
	return roots
}

// AddOpenRoot adds a new root resource for the given clientID and associates it with a generated RootHandle.
// If clientID has no associated roots, a new map is created for them.
// Returns the generated RootHandle or an error if the operation fails.
func (or *OpenRoots) AddOpenRoot(clientID ClientID, root *os.Root) (RootHandle, error) {
	or.mu.Lock()
	defer or.mu.Unlock()
	// Check if the client already has an open root map
	_, exists := or.roots[clientID]
	if !exists {
		// If not, create a new one
		or.roots[clientID] = make(map[RootHandle]*OpenRoot)
	}
	// Generate a new root handle
	rh := newRootHandle()

	// Wrap the root in an OpenRoot instance
	openRoot := newOpenRoot(root)

	// Add the root to the map
	or.roots[clientID][rh] = openRoot

	// Return the generated root handle
	return rh, nil
}

func (or *OpenRoots) GetOpenRoot(clientID ClientID, rh RootHandle) (*OpenRoot, bool) {
	// Exit early if the map is nil
	if or.roots == nil {
		return nil, false
	}
	or.mu.RLock()
	defer or.mu.RUnlock()
	roots := or.Roots(clientID)
	// Check if the client has an open root map
	openRoot, exists := roots[rh]
	if !exists {
		// If not, return nil
		return nil, false
	}
	// Return the root associated with the provided handle
	return openRoot, exists
}

func (or *OpenRoots) RemoveOpenRoot(clientID ClientID, rh RootHandle) error {
	or.mu.Lock()
	defer or.mu.Unlock()
	r, exists := or.roots[clientID][rh]
	if !exists {
		return fmt.Errorf("root not found for client %s and handle %s", clientID, rh)
	}
	delete(or.roots[clientID], rh)
	err := r.Root().Close()
	if err != nil {
		return err
	}
	return nil
}
