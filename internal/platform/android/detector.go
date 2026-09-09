package android

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// DetectADB attempts to locate the adb executable on the system.
// Priority order:
// 1. Explicit customPath (if provided and valid)
// 2. PATH lookup (via exec.LookPath)
// 3. ANDROID_HOME / ANDROID_SDK_ROOT environment variables
// 4. Standard platform-specific SDK installation directories
func DetectADB(customPath string) (string, error) {
	if customPath != "" {
		if path, err := verifyExecutable(customPath); err == nil {
			return path, nil
		}
	}

	// 1. Check system PATH
	exeName := "adb"
	if runtime.GOOS == "windows" {
		exeName = "adb.exe"
	}
	if path, err := exec.LookPath(exeName); err == nil {
		return filepath.Abs(path)
	}

	// 2. Check ANDROID_HOME / ANDROID_SDK_ROOT
	for _, envVar := range []string{"ANDROID_HOME", "ANDROID_SDK_ROOT"} {
		if val := os.Getenv(envVar); val != "" {
			candidate := filepath.Join(val, "platform-tools", exeName)
			if path, err := verifyExecutable(candidate); err == nil {
				return path, nil
			}
		}
	}

	// 3. Check well-known platform directories
	candidates := standardADBCandidates(exeName)
	for _, candidate := range candidates {
		if path, err := verifyExecutable(candidate); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("adb executable not found on PATH or standard SDK locations")
}

// DetectAndroidCLI looks for Google's official android CLI executable.
// Returns the resolved path and true if found and functional, or ("", false) if not found.
func DetectAndroidCLI(customPath string) (string, bool) {
	if customPath != "" {
		if path, err := verifyExecutable(customPath); err == nil {
			if verifyAndroidCLI(path) {
				return path, true
			}
		}
	}

	exeName := "android"
	if runtime.GOOS == "windows" {
		exeName = "android.exe"
	}

	// Search PATH
	if path, err := exec.LookPath(exeName); err == nil {
		if abs, err := filepath.Abs(path); err == nil {
			if verifyAndroidCLI(abs) {
				return abs, true
			}
		}
	}

	return "", false
}

// verifyAndroidCLI runs `android --version` with a short timeout to confirm it's the expected CLI tool.
func verifyAndroidCLI(path string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, "--version")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	// Google's android CLI output contains "android" or version numbers
	return len(strings.TrimSpace(string(out))) > 0
}

func verifyExecutable(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("%s is a directory", path)
	}
	return filepath.Abs(path)
}

func standardADBCandidates(exeName string) []string {
	var list []string

	home, _ := os.UserHomeDir()

	switch runtime.GOOS {
	case "windows":
		if local := os.Getenv("LOCALAPPDATA"); local != "" {
			list = append(list, filepath.Join(local, "Android", "Sdk", "platform-tools", exeName))
		}
		if home != "" {
			list = append(list, filepath.Join(home, "AppData", "Local", "Android", "Sdk", "platform-tools", exeName))
		}
		list = append(list, filepath.Join("C:", "Android", "platform-tools", exeName))
	case "darwin":
		if home != "" {
			list = append(list, filepath.Join(home, "Library", "Android", "sdk", "platform-tools", exeName))
		}
		list = append(list, filepath.Join("/opt", "homebrew", "bin", exeName))
		list = append(list, filepath.Join("/usr", "local", "bin", exeName))
	case "linux":
		if home != "" {
			list = append(list, filepath.Join(home, "Android", "Sdk", "platform-tools", exeName))
			list = append(list, filepath.Join(home, ".android", "platform-tools", exeName))
		}
		list = append(list, filepath.Join("/usr", "bin", exeName))
		list = append(list, filepath.Join("/usr", "local", "bin", exeName))
	}

	return list
}
