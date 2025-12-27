package hostserve

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestNewTempPaths(t *testing.T) {
	tp := newTempPaths()
	if tp == nil {
		t.Fatal("Expected non-nil TempPaths")
	}
	if tp.paths == nil {
		t.Error("Expected initialized paths map")
	}
	if len(tp.paths) != 0 {
		t.Errorf("Expected empty map, got %d entries", len(tp.paths))
	}
}

func TestTempPaths_AddPath(t *testing.T) {
	tp := newTempPaths()
	clientID := ClientID("client-1")

	t.Run("add first path", func(t *testing.T) {
		tp.AddPath(clientID, "/tmp/path1")
		paths := tp.Paths(clientID)
		if len(paths) != 1 {
			t.Errorf("Expected 1 path, got %d", len(paths))
		}
		if paths[0] != "/tmp/path1" {
			t.Errorf("Expected '/tmp/path1', got %q", paths[0])
		}
	})

	t.Run("add second path for same client", func(t *testing.T) {
		tp.AddPath(clientID, "/tmp/path2")
		paths := tp.Paths(clientID)
		if len(paths) != 2 {
			t.Errorf("Expected 2 paths, got %d", len(paths))
		}
	})

	t.Run("add path for different client", func(t *testing.T) {
		tp.AddPath(ClientID("client-2"), "/tmp/path3")
		paths := tp.Paths(ClientID("client-2"))
		if len(paths) != 1 {
			t.Errorf("Expected 1 path for client-2, got %d", len(paths))
		}
	})
}

func TestTempPaths_Paths(t *testing.T) {
	tp := newTempPaths()
	clientID := ClientID("client-1")

	t.Run("get paths for client with no paths", func(t *testing.T) {
		paths := tp.Paths(clientID)
		if paths == nil {
			t.Error("Expected non-nil slice")
		}
		if len(paths) != 0 {
			t.Errorf("Expected empty slice, got %d entries", len(paths))
		}
	})

	t.Run("get paths after adding", func(t *testing.T) {
		tp.AddPath(clientID, "/tmp/path1")
		tp.AddPath(clientID, "/tmp/path2")
		paths := tp.Paths(clientID)
		if len(paths) != 2 {
			t.Errorf("Expected 2 paths, got %d", len(paths))
		}
	})
}

func TestTempPaths_Cleanup(t *testing.T) {
	tp := newTempPaths()
	clientID := ClientID("client-1")

	t.Run("cleanup existing paths", func(t *testing.T) {
		// Create actual temp directory
		tempDir, err := os.MkdirTemp("", "temppath-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}

		tp.AddPath(clientID, tempDir)

		// Verify directory exists
		AssertFileExists(t, tempDir)

		// Cleanup
		tp.Cleanup(clientID)

		// Verify directory removed
		AssertFileNotExists(t, tempDir)

		// Verify client entry removed
		if _, exists := tp.paths[clientID]; exists {
			t.Error("Expected client entry to be removed")
		}
	})

	t.Run("cleanup non-existent client", func(t *testing.T) {
		// Should not panic
		tp.Cleanup(ClientID("non-existent"))
	})

	t.Run("cleanup non-existent path", func(t *testing.T) {
		client := ClientID("client-3")
		tp.AddPath(client, "/tmp/non-existent-path-12345")
		// Should handle gracefully
		tp.Cleanup(client)
		if _, exists := tp.paths[client]; exists {
			t.Error("Expected client entry to be removed even with non-existent path")
		}
	})
}

func TestTempPaths_CleanupAll(t *testing.T) {
	tp := newTempPaths()

	// Create temp directories for multiple clients
	tempDir1, _ := os.MkdirTemp("", "temppath-1-*")
	tempDir2, _ := os.MkdirTemp("", "temppath-2-*")

	tp.AddPath(ClientID("client-1"), tempDir1)
	tp.AddPath(ClientID("client-2"), tempDir2)

	// Verify directories exist
	AssertFileExists(t, tempDir1)
	AssertFileExists(t, tempDir2)

	// Cleanup all
	tp.CleanupAll()

	// Verify both removed
	AssertFileNotExists(t, tempDir1)
	AssertFileNotExists(t, tempDir2)

	// Verify paths map empty
	if len(tp.paths) != 0 {
		t.Errorf("Expected empty paths map, got %d entries", len(tp.paths))
	}
}

func TestTempPaths_CleanupMultiplePaths(t *testing.T) {
	tp := newTempPaths()
	clientID := ClientID("client-1")

	// Create multiple temp directories
	dirs := make([]string, 3)
	for i := 0; i < 3; i++ {
		dir, _ := os.MkdirTemp("", fmt.Sprintf("temppath-%d-*", i))
		dirs[i] = dir
		tp.AddPath(clientID, dir)
	}

	// Verify all exist
	for _, dir := range dirs {
		AssertFileExists(t, dir)
	}

	// Cleanup
	tp.Cleanup(clientID)

	// Verify all removed
	for _, dir := range dirs {
		AssertFileNotExists(t, dir)
	}
}

func TestTempPaths_CleanupNestedDirectories(t *testing.T) {
	tp := newTempPaths()
	clientID := ClientID("client-1")

	// Create temp directory with nested structure
	tempDir, _ := os.MkdirTemp("", "temppath-nested-*")
	nestedDir := filepath.Join(tempDir, "nested", "deep")
	os.MkdirAll(nestedDir, 0755)
	os.WriteFile(filepath.Join(nestedDir, "file.txt"), []byte("content"), 0644)

	tp.AddPath(clientID, tempDir)
	tp.Cleanup(clientID)

	// Verify entire tree removed
	AssertFileNotExists(t, tempDir)
}

func TestTempPaths_ConcurrentAccess(t *testing.T) {
	tp := newTempPaths()
	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrent adds
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			clientID := ClientID(fmt.Sprintf("client-%d", id))
			path := fmt.Sprintf("/tmp/path-%d", id)
			tp.AddPath(clientID, path)
		}(i)
	}
	wg.Wait()

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			clientID := ClientID(fmt.Sprintf("client-%d", id))
			_ = tp.Paths(clientID)
		}(i)
	}
	wg.Wait()

	if len(tp.paths) == 0 {
		t.Error("Expected paths to be added")
	}
}
