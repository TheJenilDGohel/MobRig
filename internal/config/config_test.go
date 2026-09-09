package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaults(t *testing.T) {
	cfg := defaults()

	if cfg.HTTPPort != DefaultHTTPPort {
		t.Errorf("expected HTTPPort %d, got %d", DefaultHTTPPort, cfg.HTTPPort)
	}
	if cfg.HTTPHost != DefaultHTTPHost {
		t.Errorf("expected HTTPHost %q, got %q", DefaultHTTPHost, cfg.HTTPHost)
	}
	if cfg.DataDir == "" {
		t.Error("expected non-empty DataDir")
	}
	if cfg.DBPath != filepath.Join(cfg.DataDir, DBFileName) {
		t.Errorf("expected DBPath %q, got %q", filepath.Join(cfg.DataDir, DBFileName), cfg.DBPath)
	}
	if cfg.Debug {
		t.Error("expected Debug to be false by default")
	}
}

func TestApplyEnv(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("MOBRIG_DATA_DIR", tmpDir)
	t.Setenv("MOBRIG_HTTP_PORT", "9999")
	t.Setenv("MOBRIG_HTTP_HOST", "0.0.0.0")
	t.Setenv("MOBRIG_DEBUG", "1")
	t.Setenv("MOBRIG_ADB_PATH", "/custom/adb")
	t.Setenv("MOBRIG_ANDROID_CLI_PATH", "/custom/android")

	cfg := defaults()
	applyEnv(cfg)

	if cfg.DataDir != tmpDir {
		t.Errorf("expected DataDir %q, got %q", tmpDir, cfg.DataDir)
	}
	if cfg.HTTPPort != 9999 {
		t.Errorf("expected HTTPPort 9999, got %d", cfg.HTTPPort)
	}
	if cfg.HTTPHost != "0.0.0.0" {
		t.Errorf("expected HTTPHost '0.0.0.0', got %q", cfg.HTTPHost)
	}
	if !cfg.Debug {
		t.Error("expected Debug to be true")
	}
	if cfg.ADBPath != "/custom/adb" {
		t.Errorf("expected ADBPath '/custom/adb', got %q", cfg.ADBPath)
	}
	if cfg.AndroidCLIPath != "/custom/android" {
		t.Errorf("expected AndroidCLIPath '/custom/android', got %q", cfg.AndroidCLIPath)
	}
}

func TestLoadWithConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("MOBRIG_DATA_DIR", tmpDir)

	// Write mock config.json
	fileCfg := map[string]any{
		"http_port": 8888,
		"debug":     true,
	}
	cfgBytes, err := json.Marshal(fileCfg)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, ConfigFileName), cfgBytes, 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.HTTPPort != 8888 {
		t.Errorf("expected HTTPPort 8888 from config file, got %d", cfg.HTTPPort)
	}
	if !cfg.Debug {
		t.Error("expected Debug true from config file")
	}
}

func TestEnsureDataDir(t *testing.T) {
	tmpDir := filepath.Join(t.TempDir(), "sub", "mobrig_test")
	cfg := &Config{DataDir: tmpDir}

	if err := cfg.EnsureDataDir(); err != nil {
		t.Fatalf("EnsureDataDir failed: %v", err)
	}

	info, err := os.Stat(tmpDir)
	if err != nil {
		t.Fatalf("data directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("expected %s to be directory", tmpDir)
	}
}
