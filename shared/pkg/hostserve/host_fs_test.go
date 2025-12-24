package hostserve

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewHostFS(t *testing.T) {
	hfs := NewHostFS()
	if hfs == nil {
		t.Fatal("Expected non-nil HostFS")
	}
	if hfs.openFiles == nil {
		t.Error("Expected initialized openFiles")
	}
	if hfs.tempPaths == nil {
		t.Error("Expected initialized tempPaths")
	}
}

func TestHostFS_ReadDir(t *testing.T) {
	hfs := NewHostFS()
	ctx := CreateTestContext()
	tempDir := CreateTempTestDir(t)

	// Create test files and subdirs
	CreateTestFile(t, tempDir, "file1.txt", "content1")
	CreateTestFile(t, tempDir, "file2.txt", "content2")
	CreateTestSubdir(t, tempDir, "subdir1")

	t.Run("read directory with files", func(t *testing.T) {
		entries, err := hfs.ReadDir(ctx, tempDir, tempDir)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(entries) != 3 {
			t.Errorf("Expected 3 entries, got %d", len(entries))
		}
	})

	t.Run("read with relative path", func(t *testing.T) {
		entries, err := hfs.ReadDir(ctx, tempDir, ".")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(entries) != 3 {
			t.Errorf("Expected 3 entries, got %d", len(entries))
		}
	})

	t.Run("read non-existent directory", func(t *testing.T) {
		_, err := hfs.ReadDir(ctx, tempDir, filepath.Join(tempDir, "nonexistent"))
		if err == nil {
			t.Error("Expected error for non-existent directory")
		}
	})
}

func TestHostFS_ReadFile(t *testing.T) {
	hfs := NewHostFS()
	ctx := CreateTestContext()
	tempDir := CreateTempTestDir(t)

	testContent := "test file content"
	testFile := CreateTestFile(t, tempDir, "test.txt", testContent)

	t.Run("read existing file", func(t *testing.T) {
		data, err := hfs.ReadFile(ctx, tempDir, testFile)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if string(data) != testContent {
			t.Errorf("Expected %q, got %q", testContent, string(data))
		}
	})

	t.Run("read non-existent file", func(t *testing.T) {
		_, err := hfs.ReadFile(ctx, tempDir, filepath.Join(tempDir, "nonexistent.txt"))
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})
}

func TestHostFS_WriteFile(t *testing.T) {
	hfs := NewHostFS()
	ctx := CreateTestContext()
	tempDir := CreateTempTestDir(t)

	testContent := "new file content"
	testPath := filepath.Join(tempDir, "newfile.txt")

	t.Run("write new file", func(t *testing.T) {
		err := hfs.WriteFile(ctx, tempDir, testPath, []byte(testContent), 0644)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		AssertFileExists(t, testPath)
		AssertFileContent(t, testPath, testContent)
	})

	t.Run("overwrite existing file", func(t *testing.T) {
		newContent := "overwritten"
		err := hfs.WriteFile(ctx, tempDir, testPath, []byte(newContent), 0644)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		AssertFileContent(t, testPath, newContent)
	})

	t.Run("write with zero permissions defaults", func(t *testing.T) {
		path2 := filepath.Join(tempDir, "defaultperm.txt")
		err := hfs.WriteFile(ctx, tempDir, path2, []byte("content"), 0)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		AssertFileExists(t, path2)
	})
}

func TestHostFS_Stat(t *testing.T) {
	hfs := NewHostFS()
	ctx := CreateTestContext()
	tempDir := CreateTempTestDir(t)

	testFile := CreateTestFile(t, tempDir, "test.txt", "content")

	t.Run("stat existing file", func(t *testing.T) {
		info, err := hfs.Stat(ctx, tempDir, testFile)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if info.Name() != "test.txt" {
			t.Errorf("Expected name 'test.txt', got %q", info.Name())
		}
		if info.IsDir() {
			t.Error("Expected file, not directory")
		}
	})

	t.Run("stat non-existent file", func(t *testing.T) {
		_, err := hfs.Stat(ctx, tempDir, filepath.Join(tempDir, "nonexistent.txt"))
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})
}

func TestHostFS_Rename(t *testing.T) {
	hfs := NewHostFS()
	ctx := CreateTestContext()
	tempDir := CreateTempTestDir(t)

	oldPath := CreateTestFile(t, tempDir, "old.txt", "content")
	newPath := filepath.Join(tempDir, "new.txt")

	t.Run("rename file", func(t *testing.T) {
		err := hfs.Rename(ctx, tempDir, oldPath, newPath)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		AssertFileNotExists(t, oldPath)
		AssertFileExists(t, newPath)
	})
}

func TestHostFS_Remove(t *testing.T) {
	hfs := NewHostFS()
	ctx := CreateTestContext()
	tempDir := CreateTempTestDir(t)

	testFile := CreateTestFile(t, tempDir, "remove.txt", "content")

	t.Run("remove file", func(t *testing.T) {
		err := hfs.Remove(ctx, tempDir, testFile)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		AssertFileNotExists(t, testFile)
	})
}

func TestHostFS_RemoveAll(t *testing.T) {
	hfs := NewHostFS()
	ctx := CreateTestContext()
	tempDir := CreateTempTestDir(t)

	// Create nested structure
	subdir := CreateTestSubdir(t, tempDir, "subdir")
	CreateTestFile(t, subdir, "file.txt", "content")
	nestedDir := CreateTestSubdir(t, subdir, "nested")
	CreateTestFile(t, nestedDir, "nested.txt", "nested")

	t.Run("remove directory tree", func(t *testing.T) {
		err := hfs.RemoveAll(ctx, tempDir, subdir)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		AssertFileNotExists(t, subdir)
	})
}

func TestHostFS_Mkdir(t *testing.T) {
	hfs := NewHostFS()
	ctx := CreateTestContext()
	tempDir := CreateTempTestDir(t)

	t.Run("create directory", func(t *testing.T) {
		newDir := filepath.Join(tempDir, "newdir")
		err := hfs.Mkdir(ctx, tempDir, "newdir", 0755)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		AssertFileExists(t, newDir)
	})

	t.Run("create existing directory fails", func(t *testing.T) {
		existingDir := CreateTestSubdir(t, tempDir, "existing")
		err := hfs.Mkdir(ctx, tempDir, "existing", 0755)
		if err == nil {
			t.Error("Expected error for existing directory")
		}
		AssertFileExists(t, existingDir)
	})
}

func TestHostFS_MkdirAll(t *testing.T) {
	hfs := NewHostFS()
	ctx := CreateTestContext()
	tempDir := CreateTempTestDir(t)

	t.Run("create nested directories", func(t *testing.T) {
		nestedPath := filepath.Join(tempDir, "a", "b", "c")
		err := hfs.MkdirAll(ctx, tempDir, nestedPath, 0755)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		AssertFileExists(t, nestedPath)
	})

	t.Run("create with relative path", func(t *testing.T) {
		err := hfs.MkdirAll(ctx, tempDir, "rel/nested/path", 0755)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		AssertFileExists(t, filepath.Join(tempDir, "rel/nested/path"))
	})
}

func TestHostFS_MkdirTemp(t *testing.T) {
	hfs := NewHostFS()
	ctx := CreateTestContext()
	tempDir := CreateTempTestDir(t)

	t.Run("create temp directory", func(t *testing.T) {
		tmpDir, err := hfs.MkdirTemp(ctx, tempDir, "test-*")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		AssertFileExists(t, tmpDir)
	})

	t.Run("empty root uses system temp", func(t *testing.T) {
		tmpDir, err := hfs.MkdirTemp(ctx, "", "test-*")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		defer os.RemoveAll(tmpDir)
		AssertFileExists(t, tmpDir)
	})
}

func TestHostFS_Chmod(t *testing.T) {
	hfs := NewHostFS()
	ctx := CreateTestContext()
	tempDir := CreateTempTestDir(t)

	testFile := CreateTestFile(t, tempDir, "chmod.txt", "content")

	t.Run("change file permissions", func(t *testing.T) {
		err := hfs.Chmod(ctx, tempDir, testFile, 0600)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		info, _ := os.Stat(testFile)
		if info.Mode().Perm() != 0600 {
			t.Errorf("Expected 0600, got %o", info.Mode().Perm())
		}
	})

	t.Run("invalid permissions", func(t *testing.T) {
		err := hfs.Chmod(ctx, tempDir, testFile, 0)
		if err == nil {
			t.Error("Expected error for invalid permissions")
		}
	})
}

func TestHostFS_Chtimes(t *testing.T) {
	hfs := NewHostFS()
	ctx := CreateTestContext()
	tempDir := CreateTempTestDir(t)

	testFile := CreateTestFile(t, tempDir, "chtimes.txt", "content")
	atime := time.Now().Add(-1 * time.Hour)
	mtime := time.Now().Add(-2 * time.Hour)

	t.Run("change file times", func(t *testing.T) {
		err := hfs.Chtimes(ctx, tempDir, testFile, atime, mtime)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		info, _ := os.Stat(testFile)
		if info.ModTime().Truncate(time.Second) != mtime.Truncate(time.Second) {
			t.Errorf("Expected mtime %v, got %v", mtime, info.ModTime())
		}
	})
}

func TestHostFS_FileCreate(t *testing.T) {
	hfs := NewHostFS()
	ctx := CreateTestContext()
	tempDir := CreateTempTestDir(t)

	newFile := filepath.Join(tempDir, "created.txt")

	t.Run("create new file", func(t *testing.T) {
		handle, err := hfs.FileCreate(ctx, tempDir, newFile)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if handle == "" {
			t.Error("Expected non-empty file handle")
		}
		AssertFileExists(t, newFile)
		hfs.FileClose(ctx, handle)
	})
}

func TestHostFS_FileOpen(t *testing.T) {
	hfs := NewHostFS()
	ctx := CreateTestContext()
	tempDir := CreateTempTestDir(t)

	testContent := "test content"
	testFile := CreateTestFile(t, tempDir, "open.txt", testContent)

	t.Run("open existing file", func(t *testing.T) {
		handle, size, err := hfs.FileOpen(ctx, tempDir, testFile, os.O_RDONLY, 0)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if handle == "" {
			t.Error("Expected non-empty handle")
		}
		if size != uint64(len(testContent)) {
			t.Errorf("Expected size %d, got %d", len(testContent), size)
		}
		hfs.FileClose(ctx, handle)
	})

	t.Run("open non-existent file", func(t *testing.T) {
		_, _, err := hfs.FileOpen(ctx, tempDir, filepath.Join(tempDir, "nonexistent.txt"), os.O_RDONLY, 0)
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})
}

func TestHostFS_FileReadSection(t *testing.T) {
	hfs := NewHostFS()
	ctx := CreateTestContext()
	tempDir := CreateTempTestDir(t)

	testContent := "0123456789abcdef"
	testFile := CreateTestFile(t, tempDir, "read.txt", testContent)

	handle, _, _ := hfs.FileOpen(ctx, tempDir, testFile, os.O_RDONLY, 0)
	defer hfs.FileClose(ctx, handle)

	t.Run("read section from offset", func(t *testing.T) {
		data, err := hfs.FileReadSection(ctx, handle, 5, 8)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		expected := "56789abc"
		if string(data) != expected {
			t.Errorf("Expected %q, got %q", expected, string(data))
		}
	})

	t.Run("read from beginning", func(t *testing.T) {
		data, err := hfs.FileReadSection(ctx, handle, 0, 5)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		expected := "01234"
		if string(data) != expected {
			t.Errorf("Expected %q, got %q", expected, string(data))
		}
	})
}

func TestHostFS_FileWriteSection(t *testing.T) {
	hfs := NewHostFS()
	ctx := CreateTestContext()
	tempDir := CreateTempTestDir(t)

	testFile := filepath.Join(tempDir, "write.txt")
	handle, _ := hfs.FileCreate(ctx, tempDir, testFile)
	defer hfs.FileClose(ctx, handle)

	t.Run("write section", func(t *testing.T) {
		data := []byte("hello world")
		n, err := hfs.FileWriteSection(ctx, handle, 0, 100, data)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if n != len(data) {
			t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
		}
	})

	t.Run("write exceeds max length", func(t *testing.T) {
		data := []byte("too long")
		_, err := hfs.FileWriteSection(ctx, handle, 0, 5, data)
		if err == nil {
			t.Error("Expected error when data exceeds max length")
		}
	})
}

func TestHostFS_FileReader(t *testing.T) {
	hfs := NewHostFS()
	ctx := CreateTestContext()
	tempDir := CreateTempTestDir(t)

	testContent := "test content for reader"
	testFile := CreateTestFile(t, tempDir, "reader.txt", testContent)

	handle, _, _ := hfs.FileOpen(ctx, tempDir, testFile, os.O_RDONLY, 0)
	defer hfs.FileClose(ctx, handle)

	t.Run("get file reader", func(t *testing.T) {
		reader, err := hfs.FileReader(ctx, handle, minChunkSize)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if reader == nil {
			t.Error("Expected non-nil reader")
		}
		data, _ := io.ReadAll(reader)
		if string(data) != testContent {
			t.Errorf("Expected %q, got %q", testContent, string(data))
		}
	})

	t.Run("invalid chunk size too small", func(t *testing.T) {
		_, err := hfs.FileReader(ctx, handle, minChunkSize-1)
		if err == nil {
			t.Error("Expected error for small chunk size")
		}
	})

	t.Run("invalid chunk size too large", func(t *testing.T) {
		_, err := hfs.FileReader(ctx, handle, maxChunkSize+1)
		if err == nil {
			t.Error("Expected error for large chunk size")
		}
	})
}

func TestHostFS_FileWriter(t *testing.T) {
	hfs := NewHostFS()
	ctx := CreateTestContext()
	tempDir := CreateTempTestDir(t)

	testFile := filepath.Join(tempDir, "writer.txt")
	handle, _ := hfs.FileCreate(ctx, tempDir, testFile)
	defer hfs.FileClose(ctx, handle)

	t.Run("get file writer", func(t *testing.T) {
		writer, err := hfs.FileWriter(ctx, handle)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if writer == nil {
			t.Error("Expected non-nil writer")
		}
		testData := []byte("written data")
		n, err := writer.Write(testData)
		if err != nil {
			t.Errorf("Error writing: %v", err)
		}
		if n != len(testData) {
			t.Errorf("Expected to write %d bytes, wrote %d", len(testData), n)
		}
	})
}

func TestHostFS_Cleanup(t *testing.T) {
	hfs := NewHostFS()
	ctx := CreateTestContext()
	tempDir := CreateTempTestDir(t)

	// Open a file
	testFile := CreateTestFile(t, tempDir, "cleanup.txt", "content")
	_, _, _ = hfs.FileOpen(ctx, tempDir, testFile, os.O_RDONLY, 0)

	// Create a temp directory
	tmpDir, _ := hfs.MkdirTemp(ctx, tempDir, "cleanup-*")

	// Cleanup
	hfs.Cleanup()

	// Verify temp directory removed
	AssertFileNotExists(t, tmpDir)
}
