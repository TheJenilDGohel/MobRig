package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

// Version is the current MobRig engine version. Injected via -ldflags at build time.
var Version = "0.1.0-dev"

const (
	DefaultHTTPPort = 8686
	DefaultHTTPHost = "127.0.0.1"
	AppDirName      = ".mobrig"
	DBFileName      = "mobrig.db"
	ConfigFileName  = "config.json"
)

// Config holds all application configuration.
type Config struct {
	// Paths
	DataDir string `json:"data_dir"` // ~/.mobrig/ — DB, logs, recordings
	DBPath  string `json:"db_path"`  // ~/.mobrig/mobrig.db

	// Server
	HTTPPort int    `json:"http_port"` // 8686 default
	HTTPHost string `json:"http_host"` // 127.0.0.1 default (localhost-only per R027)

	// Android
	ADBPath        string `json:"adb_path"`         // auto-detected or explicit
	AndroidCLIPath string `json:"android_cli_path"` // auto-detected or explicit

	// App
	Debug   bool   `json:"debug"`
	Version string `json:"-"` // set from build-time ldflags, not config file
}

// Load builds a Config by layering: defaults → config file → env vars → CLI overrides.
// CLI overrides are applied by the caller after Load returns.
func Load() (*Config, error) {
	cfg := defaults()

	// Check if MOBRIG_DATA_DIR is set to determine custom config location
	if v := os.Getenv("MOBRIG_DATA_DIR"); v != "" {
		cfg.DataDir = v
		cfg.DBPath = filepath.Join(v, DBFileName)
	}

	// Layer 2: config file (if it exists)
	configPath := filepath.Join(cfg.DataDir, ConfigFileName)
	if data, err := os.ReadFile(configPath); err == nil {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
		}
		// Re-derive DBPath if DataDir was overridden in config file
		if cfg.DBPath == "" {
			cfg.DBPath = filepath.Join(cfg.DataDir, DBFileName)
		}
	}

	// Layer 3: environment variables
	applyEnv(cfg)

	// Always set version from build-time constant
	cfg.Version = Version

	return cfg, nil
}

// EnsureDataDir creates the data directory if it doesn't exist.
func (c *Config) EnsureDataDir() error {
	return os.MkdirAll(c.DataDir, 0o755)
}

// defaults returns a Config populated with sensible default values.
func defaults() *Config {
	dataDir := defaultDataDir()
	return &Config{
		DataDir:  dataDir,
		DBPath:   filepath.Join(dataDir, DBFileName),
		HTTPPort: DefaultHTTPPort,
		HTTPHost: DefaultHTTPHost,
		Debug:    false,
		Version:  Version,
	}
}

// defaultDataDir returns the platform-appropriate data directory.
//   - Windows: %LOCALAPPDATA%\.mobrig
//   - macOS:   ~/Library/Application Support/.mobrig
//   - Linux:   ~/.mobrig (XDG_DATA_HOME fallback)
func defaultDataDir() string {
	switch runtime.GOOS {
	case "windows":
		if local := os.Getenv("LOCALAPPDATA"); local != "" {
			return filepath.Join(local, AppDirName)
		}
	case "darwin":
		home, _ := os.UserHomeDir()
		if home != "" {
			return filepath.Join(home, "Library", "Application Support", AppDirName)
		}
	}

	// Linux / fallback: use XDG_DATA_HOME or ~/.mobrig
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, AppDirName)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, AppDirName)
}

// applyEnv overrides config values from MOBRIG_* environment variables.
func applyEnv(cfg *Config) {
	if v := os.Getenv("MOBRIG_DATA_DIR"); v != "" {
		cfg.DataDir = v
		cfg.DBPath = filepath.Join(v, DBFileName)
	}
	if v := os.Getenv("MOBRIG_DB_PATH"); v != "" {
		cfg.DBPath = v
	}
	if v := os.Getenv("MOBRIG_HTTP_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.HTTPPort = port
		}
	}
	if v := os.Getenv("MOBRIG_HTTP_HOST"); v != "" {
		cfg.HTTPHost = v
	}
	if v := os.Getenv("MOBRIG_ADB_PATH"); v != "" {
		cfg.ADBPath = v
	}
	if v := os.Getenv("MOBRIG_ANDROID_CLI_PATH"); v != "" {
		cfg.AndroidCLIPath = v
	}
	if v := os.Getenv("MOBRIG_DEBUG"); v == "1" || v == "true" {
		cfg.Debug = true
	}
}
