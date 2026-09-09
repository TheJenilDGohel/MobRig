package android

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectADB_CustomPath(t *testing.T) {
	// Create a dummy executable in temp dir
	tmp := t.TempDir()
	dummyADB := filepath.Join(tmp, "adb.exe")
	if err := os.WriteFile(dummyADB, []byte("fake binary"), 0o755); err != nil {
		t.Fatalf("failed to create dummy file: %v", err)
	}

	found, err := DetectADB(dummyADB)
	if err != nil {
		t.Fatalf("expected custom path to be detected, got: %v", err)
	}
	if filepath.Clean(found) != filepath.Clean(dummyADB) {
		t.Errorf("expected %s, got %s", dummyADB, found)
	}
}

func TestDetectADB_SystemDetection(t *testing.T) {
	// DetectADB should find the real ADB installed on this machine
	path, err := DetectADB("")
	if err != nil {
		t.Logf("system adb not found: %v", err)
		return
	}
	if path == "" {
		t.Error("expected non-empty adb path")
	}
	t.Logf("Found ADB at: %s", path)
}

func TestDetectAndroidCLI_NotFound(t *testing.T) {
	// On a path that doesn't exist
	_, found := DetectAndroidCLI("/nonexistent/path/to/android")
	if found {
		t.Error("expected false for nonexistent path")
	}
}
