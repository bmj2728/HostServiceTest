package hostserve

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	hostservev1 "github.com/bmj2728/hst/shared/protogen/hostserve/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestAbsToRel(t *testing.T) {
	tests := []struct {
		name        string
		root        string
		path        string
		expected    string
		expectError bool
	}{
		{
			name:        "absolute path within root",
			root:        "/home/user",
			path:        "/home/user/docs/file.txt",
			expected:    "docs/file.txt",
			expectError: false,
		},
		{
			name:        "relative path",
			root:        "/home/user",
			path:        "docs/file.txt",
			expected:    "docs/file.txt",
			expectError: false,
		},
		{
			name:        "path with dot segments",
			root:        "/home/user",
			path:        "./docs/../docs/file.txt",
			expected:    "docs/file.txt",
			expectError: false,
		},
		{
			name:        "root path itself",
			root:        "/home/user",
			path:        "/home/user",
			expected:    ".",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := absToRel(tt.root, tt.path)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %q, got %q", tt.expected, result)
				}
			}
		})
	}
}

func TestGetRoot(t *testing.T) {
	tempDir := CreateTempTestDir(t)

	tests := []struct {
		name        string
		setupFunc   func() string
		expectError bool
	}{
		{
			name: "valid directory",
			setupFunc: func() string {
				return tempDir
			},
			expectError: false,
		},
		{
			name: "non-existent directory",
			setupFunc: func() string {
				return filepath.Join(tempDir, "nonexistent")
			},
			expectError: true,
		},
		{
			name: "file instead of directory",
			setupFunc: func() string {
				return CreateTestFile(t, tempDir, "notadir.txt", "content")
			},
			expectError: true,
		},
		{
			name: "relative path",
			setupFunc: func() string {
				return "relative/path"
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setupFunc()
			root, err := getRoot(path)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if root != nil {
					t.Error("Expected nil root on error")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if root == nil {
					t.Error("Expected valid root")
				} else {
					closeRoot(root)
				}
			}
		})
	}
}

func TestFromOpenFileFlags(t *testing.T) {
	tests := []struct {
		name     string
		flag     hostservev1.OpenFileFlags
		expected int
	}{
		{
			name:     "READ_ONLY",
			flag:     hostservev1.OpenFileFlags_READ_ONLY,
			expected: os.O_RDONLY,
		},
		{
			name:     "WRITE_TRUNCATE",
			flag:     hostservev1.OpenFileFlags_WRITE_TRUNCATE,
			expected: os.O_WRONLY | os.O_CREATE | os.O_TRUNC,
		},
		{
			name:     "WRITE_APPEND",
			flag:     hostservev1.OpenFileFlags_WRITE_APPEND,
			expected: os.O_WRONLY | os.O_CREATE | os.O_APPEND,
		},
		{
			name:     "READ_WRITE",
			flag:     hostservev1.OpenFileFlags_READ_WRITE,
			expected: os.O_RDWR,
		},
		{
			name:     "READ_WRITE_CREATE",
			flag:     hostservev1.OpenFileFlags_READ_WRITE_CREATE,
			expected: os.O_RDWR | os.O_CREATE,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fromOpenFileFLags(tt.flag)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestToOpenFileFlags(t *testing.T) {
	tests := []struct {
		name     string
		flags    int
		expected hostservev1.OpenFileFlags
	}{
		{
			name:     "O_RDONLY",
			flags:    os.O_RDONLY,
			expected: hostservev1.OpenFileFlags_READ_ONLY,
		},
		{
			name:     "WRITE_TRUNCATE",
			flags:    os.O_WRONLY | os.O_CREATE | os.O_TRUNC,
			expected: hostservev1.OpenFileFlags_WRITE_TRUNCATE,
		},
		{
			name:     "READ_WRITE",
			flags:    os.O_RDWR,
			expected: hostservev1.OpenFileFlags_READ_WRITE,
		},
		{
			name:     "unknown flags default to READ_ONLY",
			flags:    999,
			expected: hostservev1.OpenFileFlags_READ_ONLY,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toOpenFileFlags(tt.flags)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFileInfoToProtoFileInfo(t *testing.T) {
	tempDir := CreateTempTestDir(t)
	testFile := CreateTestFile(t, tempDir, "test.txt", "content")

	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	t.Run("valid FileInfo", func(t *testing.T) {
		proto := fileInfoToProtoFileInfo(info)
		if proto == nil {
			t.Fatal("Expected non-nil proto FileInfo")
		}
		if proto.Name != info.Name() {
			t.Errorf("Name mismatch: expected %q, got %q", info.Name(), proto.Name)
		}
		if proto.Size != info.Size() {
			t.Errorf("Size mismatch: expected %d, got %d", info.Size(), proto.Size)
		}
		if proto.Mode != uint32(info.Mode()) {
			t.Errorf("Mode mismatch: expected %d, got %d", uint32(info.Mode()), proto.Mode)
		}
		if proto.IsDir != info.IsDir() {
			t.Errorf("IsDir mismatch: expected %v, got %v", info.IsDir(), proto.IsDir)
		}
	})

	t.Run("nil FileInfo", func(t *testing.T) {
		proto := fileInfoToProtoFileInfo(nil)
		if proto != nil {
			t.Error("Expected nil for nil input")
		}
	})
}

func TestProtoFileInfoToRemoteFileInfo(t *testing.T) {
	modTime := time.Now().Truncate(time.Second)

	t.Run("valid proto FileInfo", func(t *testing.T) {
		proto := &hostservev1.FileInfo{
			Name:    "test.txt",
			Size:    1234,
			Mode:    uint32(0644),
			ModTime: timestamppb.New(modTime),
			IsDir:   false,
		}

		remote := protoFileInfoToRemoteFileInfo(proto)
		if remote == nil {
			t.Fatal("Expected non-nil remote FileInfo")
		}
		if remote.Name() != proto.Name {
			t.Errorf("Name mismatch: expected %q, got %q", proto.Name, remote.Name())
		}
		if remote.Size() != proto.Size {
			t.Errorf("Size mismatch: expected %d, got %d", proto.Size, remote.Size())
		}
		if remote.Mode() != fs.FileMode(proto.Mode) {
			t.Errorf("Mode mismatch: expected %v, got %v", fs.FileMode(proto.Mode), remote.Mode())
		}
		if !remote.ModTime().Equal(modTime) {
			t.Errorf("ModTime mismatch: expected %v, got %v", modTime, remote.ModTime())
		}
		if remote.IsDir() != proto.IsDir {
			t.Errorf("IsDir mismatch: expected %v, got %v", proto.IsDir, remote.IsDir())
		}
	})

	t.Run("nil proto FileInfo", func(t *testing.T) {
		remote := protoFileInfoToRemoteFileInfo(nil)
		if remote != nil {
			t.Error("Expected nil for nil input")
		}
	})
}

func TestRoundTripFileInfoConversion(t *testing.T) {
	tempDir := CreateTempTestDir(t)
	testFile := CreateTestFile(t, tempDir, "roundtrip.txt", "test content")

	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	// Convert to proto and back
	proto := fileInfoToProtoFileInfo(info)
	remote := protoFileInfoToRemoteFileInfo(proto)

	if remote.Name() != info.Name() {
		t.Errorf("Name mismatch after round trip")
	}
	if remote.Size() != info.Size() {
		t.Errorf("Size mismatch after round trip")
	}
	if remote.Mode() != info.Mode() {
		t.Errorf("Mode mismatch after round trip")
	}
	if remote.IsDir() != info.IsDir() {
		t.Errorf("IsDir mismatch after round trip")
	}
	// ModTime has precision loss, check it's close
	if !remote.ModTime().Truncate(time.Second).Equal(info.ModTime().Truncate(time.Second)) {
		t.Errorf("ModTime mismatch after round trip")
	}
}
