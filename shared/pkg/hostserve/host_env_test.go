package hostserve

import (
	"context"
	"os"
	"testing"
)

func TestNewHostEnv(t *testing.T) {
	he := NewHostEnv()
	if he == nil {
		t.Fatal("Expected non-nil HostEnv")
	}
}

func TestHostEnv_Getuid(t *testing.T) {
	he := NewHostEnv()
	ctx := CreateTestContext()

	t.Run("with valid context", func(t *testing.T) {
		uid, err := he.Getuid(ctx)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		expectedUID := int32(os.Getuid())
		if uid != expectedUID {
			t.Errorf("Expected UID %d, got %d", expectedUID, uid)
		}
	})

	t.Run("without client ID in context", func(t *testing.T) {
		emptyCtx := context.Background()
		_, err := he.Getuid(emptyCtx)
		if err == nil {
			t.Error("Expected error when client ID missing from context")
		}
	})
}

func TestHostEnv_Getgid(t *testing.T) {
	he := NewHostEnv()
	ctx := CreateTestContext()

	uid, err := he.Getgid(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expectedGID := int32(os.Getgid())
	if uid != expectedGID {
		t.Errorf("Expected GID %d, got %d", expectedGID, uid)
	}
}

func TestHostEnv_Geteuid(t *testing.T) {
	he := NewHostEnv()
	ctx := CreateTestContext()

	euid, err := he.Geteuid(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expectedEUID := int32(os.Geteuid())
	if euid != expectedEUID {
		t.Errorf("Expected EUID %d, got %d", expectedEUID, euid)
	}
}

func TestHostEnv_Getegid(t *testing.T) {
	he := NewHostEnv()
	ctx := CreateTestContext()

	egid, err := he.Getegid(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expectedEGID := int32(os.Getegid())
	if egid != expectedEGID {
		t.Errorf("Expected EGID %d, got %d", expectedEGID, egid)
	}
}

func TestHostEnv_GetGroups(t *testing.T) {
	he := NewHostEnv()
	ctx := CreateTestContext()

	groups, err := he.GetGroups(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expectedGroups, _ := os.Getgroups()
	if len(groups) != len(expectedGroups) {
		t.Errorf("Expected %d groups, got %d", len(expectedGroups), len(groups))
	}
	for i, group := range groups {
		if int(group) != expectedGroups[i] {
			t.Errorf("Group %d mismatch: expected %d, got %d", i, expectedGroups[i], group)
		}
	}
}

func TestHostEnv_Getpid(t *testing.T) {
	he := NewHostEnv()
	ctx := CreateTestContext()

	pid, err := he.Getpid(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expectedPID := int32(os.Getpid())
	if pid != expectedPID {
		t.Errorf("Expected PID %d, got %d", expectedPID, pid)
	}
}

func TestHostEnv_Getppid(t *testing.T) {
	he := NewHostEnv()
	ctx := CreateTestContext()

	ppid, err := he.Getppid(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expectedPPID := int32(os.Getppid())
	if ppid != expectedPPID {
		t.Errorf("Expected PPID %d, got %d", expectedPPID, ppid)
	}
}

func TestHostEnv_GetEnv(t *testing.T) {
	he := NewHostEnv()
	ctx := CreateTestContext()

	// Set test environment variable
	testKey := "TEST_HOST_ENV_VAR"
	testValue := "test-value-123"
	os.Setenv(testKey, testValue)
	defer os.Unsetenv(testKey)

	t.Run("get existing env var", func(t *testing.T) {
		value, err := he.GetEnv(ctx, testKey)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if value != testValue {
			t.Errorf("Expected %q, got %q", testValue, value)
		}
	})

	t.Run("get non-existent env var", func(t *testing.T) {
		_, err := he.GetEnv(ctx, "NON_EXISTENT_VAR_12345")
		if err == nil {
			t.Error("Expected error for non-existent env var")
		}
	})
}

func TestHostEnv_TempDir(t *testing.T) {
	he := NewHostEnv()
	ctx := CreateTestContext()

	tempDir, err := he.TempDir(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expectedTempDir := os.TempDir()
	if tempDir != expectedTempDir {
		t.Errorf("Expected %q, got %q", expectedTempDir, tempDir)
	}
}

func TestHostEnv_UserCacheDir(t *testing.T) {
	he := NewHostEnv()
	ctx := CreateTestContext()

	cacheDir, err := he.UserCacheDir(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expectedCacheDir, _ := os.UserCacheDir()
	if cacheDir != expectedCacheDir {
		t.Errorf("Expected %q, got %q", expectedCacheDir, cacheDir)
	}
}

func TestHostEnv_UserConfigDir(t *testing.T) {
	he := NewHostEnv()
	ctx := CreateTestContext()

	configDir, err := he.UserConfigDir(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expectedConfigDir, _ := os.UserConfigDir()
	if configDir != expectedConfigDir {
		t.Errorf("Expected %q, got %q", expectedConfigDir, configDir)
	}
}

func TestHostEnv_UserHomeDir(t *testing.T) {
	he := NewHostEnv()
	ctx := CreateTestContext()

	homeDir, err := he.UserHomeDir(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	expectedHomeDir, _ := os.UserHomeDir()
	if homeDir != expectedHomeDir {
		t.Errorf("Expected %q, got %q", expectedHomeDir, homeDir)
	}
}
