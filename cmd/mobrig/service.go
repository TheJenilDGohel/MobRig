package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/TheJenilDGohel/MobRig/internal/config"
	"github.com/TheJenilDGohel/MobRig/internal/database"
	"github.com/TheJenilDGohel/MobRig/internal/platform"
	"github.com/TheJenilDGohel/MobRig/internal/setup"
)

// MobRigService is the Go backend service exposed to the Svelte frontend via Wails bindings.
// Methods on this struct are callable from JavaScript.
type MobRigService struct {
	config   *config.Config
	db       *database.Database
	repos    *database.Repositories
	registry *platform.AdapterRegistry
	started  time.Time
}

// NewMobRigService creates the service with all dependencies injected.
func NewMobRigService(cfg *config.Config, db *database.Database, repos *database.Repositories, registry *platform.AdapterRegistry) *MobRigService {
	return &MobRigService{
		config:   cfg,
		db:       db,
		repos:    repos,
		registry: registry,
		started:  time.Now(),
	}
}

// --- Wails-bound methods (callable from Svelte UI) ---

// HealthInfo is returned by GetHealth.
type HealthInfo struct {
	Status   string `json:"status"`
	Version  string `json:"version"`
	Uptime   string `json:"uptime"`
	HTTPPort int    `json:"http_port"`
	Debug    bool   `json:"debug"`
}

// GetHealth returns the application health status.
func (s *MobRigService) GetHealth() HealthInfo {
	return HealthInfo{
		Status:   "ok",
		Version:  s.config.Version,
		Uptime:   time.Since(s.started).Round(time.Second).String(),
		HTTPPort: s.config.HTTPPort,
		Debug:    s.config.Debug,
	}
}

// ConfigInfo is a safe subset of config exposed to the UI.
type ConfigInfo struct {
	DataDir  string `json:"data_dir"`
	HTTPPort int    `json:"http_port"`
	HTTPHost string `json:"http_host"`
	DBPath   string `json:"db_path"`
	Debug    bool   `json:"debug"`
	Version  string `json:"version"`
}

// GetConfig returns configuration info safe for UI display.
func (s *MobRigService) GetConfig() ConfigInfo {
	return ConfigInfo{
		DataDir:  s.config.DataDir,
		HTTPPort: s.config.HTTPPort,
		HTTPHost: s.config.HTTPHost,
		DBPath:   s.config.DBPath,
		Debug:    s.config.Debug,
		Version:  s.config.Version,
	}
}

// DeviceListItem is a simplified device representation for the Hub UI.
type DeviceListItem struct {
	ID             string `json:"id"`
	Name           string `json:"name"`     // model or friendly name
	OS             string `json:"os"`       // e.g., "Android 14"
	Status         string `json:"status"`   // connected, offline, busy
	Platform       string `json:"platform"` // android, ios
	ConnectionType string `json:"connection_type"`
}

// GetDevices returns the current list of connected devices.
func (s *MobRigService) GetDevices() []DeviceListItem {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	devices, err := s.registry.ListAllDevices(ctx)
	if err != nil {
		return []DeviceListItem{}
	}

	items := make([]DeviceListItem, 0, len(devices))
	for _, d := range devices {
		osName := "Android"
		if d.OSVersion != "" {
			osName += " " + d.OSVersion
		}
		name := d.Model
		if name == "" {
			name = d.ID
		}

		items = append(items, DeviceListItem{
			ID:             d.ID,
			Name:           name,
			OS:             osName,
			Status:         string(d.Status),
			Platform:       string(d.Platform),
			ConnectionType: string(d.ConnectionType),
		})
	}
	return items
}

// GetDeviceInfo returns detailed specs for a specific device.
func (s *MobRigService) GetDeviceInfo(deviceID string) (*platform.DeviceInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	adapter := s.registry.FindAdapter(ctx, deviceID)
	if adapter == nil {
		return nil, fmt.Errorf("device %q not found", deviceID)
	}
	return adapter.GetDeviceInfo(ctx, deviceID)
}

// GetScreenshot captures a live screenshot for the given device.
func (s *MobRigService) GetScreenshot(deviceID string) (*platform.Screenshot, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	adapter := s.registry.FindAdapter(ctx, deviceID)
	if adapter == nil {
		return nil, fmt.Errorf("device %q not found", deviceID)
	}
	return adapter.GetScreenshot(ctx, deviceID)
}

// QuickAction executes a common action (home, back, open_url) from the Desktop Hub.
func (s *MobRigService) QuickAction(deviceID string, action string, param string) (*platform.ActionResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	adapter := s.registry.FindAdapter(ctx, deviceID)
	if adapter == nil {
		return nil, fmt.Errorf("device %q not found", deviceID)
	}

	switch action {
	case "home":
		return adapter.GoHome(ctx, deviceID)
	case "back":
		return adapter.PressButton(ctx, deviceID, 4) // KEYCODE_BACK
	case "open_url":
		return adapter.OpenURL(ctx, deviceID, param)
	default:
		return nil, fmt.Errorf("unknown action %q", action)
	}
}

// GetSetupStatus returns the environment setup report.
func (s *MobRigService) GetSetupStatus() *setup.EnvironmentReport {
	return setup.CheckEnvironment(s.config.ADBPath, s.config.AndroidCLIPath)
}

// AutoConfigureMCP generates MCP configs for Claude Desktop, Cursor, and Claude Code.
func (s *MobRigService) AutoConfigureMCP(workspaceDir string) []string {
	execPath, err := os.Executable()
	if err != nil {
		execPath = "mobrig"
	}
	if workspaceDir == "" {
		workspaceDir = "."
	}
	configured, _ := setup.GenerateAllConfigs(workspaceDir, execPath)
	return configured
}

// Greet is a test method to verify Wails bindings work.
func (s *MobRigService) Greet(name string) string {
	return "Hello " + name + " from MobRig v" + s.config.Version + "!"
}
