package hostserve

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/grpc/metadata"
)

// Test constants
const TestClientID = ClientID("test-client-id-12345")

// CreateTestContext creates a context with gRPC metadata containing a test client ID
// This simulates how the actual gRPC server adds client IDs to incoming contexts
func CreateTestContext() context.Context {
	md := metadata.New(map[string]string{
		ctxClientIDKey: TestClientID.String(),
	})
	return metadata.NewIncomingContext(context.Background(), md)
}

// CreateTempTestDir creates a temporary directory for testing and schedules cleanup
func CreateTempTestDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "hostserve-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	return dir
}

// CreateTestFile creates a file with specified content in the given directory
func CreateTestFile(t *testing.T, dir, filename, content string) string {
	t.Helper()
	path := filepath.Join(dir, filename)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file %s: %v", path, err)
	}
	return path
}

// CreateTestSubdir creates a subdirectory in the given directory
func CreateTestSubdir(t *testing.T, dir, subdirName string) string {
	t.Helper()
	path := filepath.Join(dir, subdirName)
	err := os.Mkdir(path, 0755)
	if err != nil {
		t.Fatalf("Failed to create test subdir %s: %v", path, err)
	}
	return path
}

// AssertFileExists verifies a file exists at the given path
func AssertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Expected file to exist at %s", path)
	}
}

// AssertFileNotExists verifies a file does NOT exist at the given path
func AssertFileNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("Expected file to NOT exist at %s", path)
	}
}

// AssertFileContent verifies file contains expected content
func AssertFileContent(t *testing.T, path, expected string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", path, err)
	}
	if string(content) != expected {
		t.Errorf("File content mismatch at %s.\nExpected: %q\nGot: %q", path, expected, string(content))
	}
}

// CreateTestFiles creates multiple numbered test files for concurrent testing
func CreateTestFiles(t *testing.T, dir string, count int, contentPrefix string) []string {
	t.Helper()
	var paths []string
	for i := 0; i < count; i++ {
		filename := fmt.Sprintf("file%d.txt", i)
		content := fmt.Sprintf("%s-%d", contentPrefix, i)
		path := CreateTestFile(t, dir, filename, content)
		paths = append(paths, path)
	}
	return paths
}
