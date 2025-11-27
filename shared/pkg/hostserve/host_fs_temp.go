package hostserve

import (
	"os"
	"sync"

	"github.com/hashicorp/go-hclog"
)

// TempPathMap maps client IDs to a list of temporary files/dir paths
type TempPathMap map[ClientID][]string

// TempPaths is a thread-safe structure for managing temporary file or directory paths for specific clients.
type TempPaths struct {
	paths TempPathMap
	mu    sync.RWMutex
}

// newTempPaths initializes and returns a new instance of TempPaths with an empty TempPathMap.
func newTempPaths() *TempPaths {
	return &TempPaths{
		paths: make(TempPathMap),
	}
}

// Paths returns the slice of temporary file/directory paths associated with the given ClientID.
// If no paths exist for the ClientID, an empty slice is initialized and returned.
func (tp *TempPaths) Paths(clientID ClientID) []string {
	tp.mu.RLock()
	defer tp.mu.RUnlock()
	_, exists := tp.paths[clientID]
	if !exists {
		tp.paths[clientID] = make([]string, 0)
		return tp.paths[clientID]
	}
	return tp.paths[clientID]
}

// AddPath associates a new temporary path with a given client ID, creating an entry if the client ID does not exist.
func (tp *TempPaths) AddPath(clientID ClientID, path string) {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	_, exists := tp.paths[clientID]
	if !exists {
		tp.paths[clientID] = make([]string, 0)
	}
	tp.paths[clientID] = append(tp.paths[clientID], path)
}

// Cleanup removes all temporary paths associated with the given clientID and deletes the entry from the map.
func (tp *TempPaths) Cleanup(clientID ClientID) {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	paths, exists := tp.paths[clientID]
	if !exists {
		return
	}
	for _, path := range paths {
		err := os.RemoveAll(path)
		if err != nil {
			hclog.Default().Error("Failed to remove temp path", "path", path, "error", err)
			continue
		}
		hclog.Default().Debug("Removed temp path", "path", path)
	}
	delete(tp.paths, clientID)
}

// CleanupAll removes all temporary paths for all clients and clears the internal storage. Uses Cleanup for each client.
func (tp *TempPaths) CleanupAll() {
	for clientID := range tp.paths {
		tp.Cleanup(clientID)
	}
}
