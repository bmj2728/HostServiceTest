package hostserve

import (
	"fmt"
	"os"
	"sync"
	"testing"
)

func TestNewFileHandle(t *testing.T) {
	fh1 := newFileHandle()
	fh2 := newFileHandle()

	if fh1 == "" {
		t.Error("Expected non-empty file handle")
	}
	if fh2 == "" {
		t.Error("Expected non-empty file handle")
	}
	if fh1 == fh2 {
		t.Error("Expected unique file handles")
	}
	if len(fh1.String()) != 36 {
		t.Errorf("Expected UUID format (36 chars), got %d", len(fh1.String()))
	}
}

func TestFileHandle_String(t *testing.T) {
	fh := FileHandle("test-handle")
	if fh.String() != "test-handle" {
		t.Errorf("Expected 'test-handle', got %q", fh.String())
	}
}

func TestNewOpenFiles(t *testing.T) {
	of := newOpenFiles()
	if of == nil {
		t.Fatal("Expected non-nil OpenFiles")
	}
	if of.files == nil {
		t.Error("Expected initialized files map")
	}
	if len(of.files) != 0 {
		t.Errorf("Expected empty map, got %d entries", len(of.files))
	}
}

func TestOpenFiles_AddFile(t *testing.T) {
	of := newOpenFiles()
	tempDir := CreateTempTestDir(t)
	testFile := CreateTestFile(t, tempDir, "test.txt", "content")

	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	clientID := ClientID("client-1")
	fh := newFileHandle()

	t.Run("add new file", func(t *testing.T) {
		err := of.AddFile(clientID, fh, file)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(of.files[clientID]) != 1 {
			t.Errorf("Expected 1 file, got %d", len(of.files[clientID]))
		}
	})

	t.Run("add duplicate handle", func(t *testing.T) {
		err := of.AddFile(clientID, fh, file)
		if err == nil {
			t.Error("Expected error for duplicate handle")
		}
	})

	t.Run("add file for new client", func(t *testing.T) {
		newClient := ClientID("client-2")
		newHandle := newFileHandle()
		err := of.AddFile(newClient, newHandle, file)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(of.files) != 2 {
			t.Errorf("Expected 2 clients, got %d", len(of.files))
		}
	})
}

func TestOpenFiles_RemoveFile(t *testing.T) {
	of := newOpenFiles()
	tempDir := CreateTempTestDir(t)
	testFile := CreateTestFile(t, tempDir, "test.txt", "content")

	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	clientID := ClientID("client-1")
	fh := newFileHandle()
	of.AddFile(clientID, fh, file)

	t.Run("remove existing file", func(t *testing.T) {
		err := of.RemoveFile(clientID, fh)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(of.files[clientID]) != 0 {
			t.Errorf("Expected 0 files, got %d", len(of.files[clientID]))
		}
	})

	t.Run("remove non-existent file", func(t *testing.T) {
		err := of.RemoveFile(clientID, FileHandle("non-existent"))
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	t.Run("remove file for non-existent client", func(t *testing.T) {
		err := of.RemoveFile(ClientID("non-existent"), fh)
		if err == nil {
			t.Error("Expected error for non-existent client")
		}
	})
}

func TestOpenFiles_GetFile(t *testing.T) {
	of := newOpenFiles()
	tempDir := CreateTempTestDir(t)
	testFile := CreateTestFile(t, tempDir, "test.txt", "content")

	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	clientID := ClientID("client-1")
	fh := newFileHandle()
	of.AddFile(clientID, fh, file)

	t.Run("get existing file", func(t *testing.T) {
		gotFile, err := of.GetFile(clientID, fh)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if gotFile != file {
			t.Error("Expected same file pointer")
		}
	})

	t.Run("get non-existent file", func(t *testing.T) {
		_, err := of.GetFile(clientID, FileHandle("non-existent"))
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	t.Run("get file for non-existent client", func(t *testing.T) {
		_, err := of.GetFile(ClientID("non-existent"), fh)
		if err == nil {
			t.Error("Expected error for non-existent client")
		}
	})
}

func TestOpenFiles_GetFilesByClient(t *testing.T) {
	of := newOpenFiles()
	tempDir := CreateTempTestDir(t)

	files := CreateTestFiles(t, tempDir, 2, "content")
	f1, _ := os.Open(files[0])
	f2, _ := os.Open(files[1])
	defer f1.Close()
	defer f2.Close()

	clientID := ClientID("client-1")
	fh1 := newFileHandle()
	fh2 := newFileHandle()
	of.AddFile(clientID, fh1, f1)
	of.AddFile(clientID, fh2, f2)

	t.Run("get files for existing client", func(t *testing.T) {
		clientFiles, err := of.GetFilesByClient(clientID)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(clientFiles) != 2 {
			t.Errorf("Expected 2 files, got %d", len(clientFiles))
		}
	})

	t.Run("get files for non-existent client", func(t *testing.T) {
		_, err := of.GetFilesByClient(ClientID("non-existent"))
		if err == nil {
			t.Error("Expected error for non-existent client")
		}
	})
}

func TestOpenFiles_Len(t *testing.T) {
	of := newOpenFiles()

	if of.Len() != 0 {
		t.Errorf("Expected len 0, got %d", of.Len())
	}

	tempDir := CreateTempTestDir(t)
	files := CreateTestFiles(t, tempDir, 2, "content")
	f1, _ := os.Open(files[0])
	f2, _ := os.Open(files[1])
	defer f1.Close()
	defer f2.Close()

	of.AddFile(ClientID("client-1"), newFileHandle(), f1)
	of.AddFile(ClientID("client-2"), newFileHandle(), f2)

	if of.Len() != 2 {
		t.Errorf("Expected len 2, got %d", of.Len())
	}
}

func TestOpenFiles_CloseAll(t *testing.T) {
	of := newOpenFiles()
	tempDir := CreateTempTestDir(t)

	files := CreateTestFiles(t, tempDir, 2, "content")
	f1, _ := os.Open(files[0])
	f2, _ := os.Open(files[1])

	of.AddFile(ClientID("client-1"), newFileHandle(), f1)
	of.AddFile(ClientID("client-2"), newFileHandle(), f2)

	of.CloseAll()

	// Verify files are closed (read should fail)
	buf := make([]byte, 1)
	_, err := f1.Read(buf)
	if err == nil {
		t.Error("Expected error reading from closed file")
	}
}

func TestOpenFiles_ConcurrentAccess(t *testing.T) {
	of := newOpenFiles()
	tempDir := CreateTempTestDir(t)
	var wg sync.WaitGroup
	numGoroutines := 50

	files := CreateTestFiles(t, tempDir, numGoroutines, "content")
	var openFiles []*os.File
	for _, path := range files {
		f, _ := os.Open(path)
		defer f.Close()
		openFiles = append(openFiles, f)
	}

	// Concurrent adds
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			clientID := ClientID(fmt.Sprintf("client-%d", idx))
			fh := newFileHandle()
			of.AddFile(clientID, fh, openFiles[idx])
		}(i)
	}
	wg.Wait()

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = of.GetFiles()
			_ = of.Len()
		}()
	}
	wg.Wait()

	if of.Len() == 0 {
		t.Error("Expected files to be added")
	}
}
