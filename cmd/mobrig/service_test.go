package main

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/TheJenilDGohel/MobRig/internal/config"
	"github.com/TheJenilDGohel/MobRig/internal/database"
	"github.com/TheJenilDGohel/MobRig/internal/platform"
)

type mockServiceAdapter struct {
	platform.PlatformAdapter
	wentHome  bool
	openedURL string
}

func (m *mockServiceAdapter) Platform() string {
	return "android"
}

func (m *mockServiceAdapter) ListDevices(ctx context.Context) ([]platform.DeviceInfo, error) {
	return []platform.DeviceInfo{
		{
			ID:             "test-dev-1",
			Platform:       platform.PlatformAndroid,
			Model:          "Pixel 8",
			Status:         platform.StatusConnected,
			ConnectionType: platform.ConnectionUSB,
		},
	}, nil
}

func (m *mockServiceAdapter) GetDeviceInfo(ctx context.Context, deviceID string) (*platform.DeviceInfo, error) {
	return &platform.DeviceInfo{
		ID:           "test-dev-1",
		Platform:     platform.PlatformAndroid,
		Model:        "Pixel 8",
		Manufacturer: "Google",
		OSVersion:    "14",
		APILevel:     34,
		ScreenWidth:  1080,
		ScreenHeight: 2400,
	}, nil
}

func (m *mockServiceAdapter) GetScreenshot(ctx context.Context, deviceID string) (*platform.Screenshot, error) {
	return &platform.Screenshot{
		Width:     1080,
		Height:    2400,
		Format:    "png",
		Base64:    "mockbase64",
		SizeBytes: 100,
	}, nil
}

func (m *mockServiceAdapter) GoHome(ctx context.Context, deviceID string) (*platform.ActionResult, error) {
	m.wentHome = true
	return &platform.ActionResult{Success: true, Message: "Home", Duration: 10 * time.Millisecond}, nil
}

func (m *mockServiceAdapter) PressButton(ctx context.Context, deviceID string, key int) (*platform.ActionResult, error) {
	return &platform.ActionResult{Success: true, Message: "Key", Duration: 10 * time.Millisecond}, nil
}

func (m *mockServiceAdapter) OpenURL(ctx context.Context, deviceID string, url string) (*platform.ActionResult, error) {
	m.openedURL = url
	return &platform.ActionResult{Success: true, Message: "URL opened", Duration: 10 * time.Millisecond}, nil
}

func TestMobRigService(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "mobrig.db")

	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	cfg := &config.Config{
		DataDir:  tmpDir,
		DBPath:   dbPath,
		HTTPPort: 8686,
		HTTPHost: "127.0.0.1",
		Version:  "3.0.0-test",
		Debug:    true,
	}

	mock := &mockServiceAdapter{}
	registry := platform.NewAdapterRegistry()
	registry.Register(mock)

	repos := database.NewRepositories(db)
	svc := NewMobRigService(cfg, db, repos, registry)

	// Test GetHealth
	health := svc.GetHealth()
	if health.Status != "ok" {
		t.Errorf("expected health status 'ok', got %q", health.Status)
	}
	if health.Version != "3.0.0-test" {
		t.Errorf("expected version '3.0.0-test', got %q", health.Version)
	}
	if health.HTTPPort != 8686 {
		t.Errorf("expected port 8686, got %d", health.HTTPPort)
	}
	if !health.Debug {
		t.Error("expected debug true")
	}

	// Test GetConfig
	cfgInfo := svc.GetConfig()
	if cfgInfo.DataDir != tmpDir {
		t.Errorf("expected data dir %q, got %q", tmpDir, cfgInfo.DataDir)
	}
	if cfgInfo.HTTPPort != 8686 {
		t.Errorf("expected port 8686, got %d", cfgInfo.HTTPPort)
	}

	// Test GetDevices
	devices := svc.GetDevices()
	if len(devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(devices))
	}
	if devices[0].ID != "test-dev-1" {
		t.Errorf("expected test-dev-1, got %q", devices[0].ID)
	}

	// Test GetDeviceInfo
	info, err := svc.GetDeviceInfo("test-dev-1")
	if err != nil {
		t.Fatalf("failed to get device info: %v", err)
	}
	if info.Model != "Pixel 8" {
		t.Errorf("expected Pixel 8, got %q", info.Model)
	}

	// Test GetScreenshot
	shot, err := svc.GetScreenshot("test-dev-1")
	if err != nil {
		t.Fatalf("failed to get screenshot: %v", err)
	}
	if shot.Width != 1080 {
		t.Errorf("expected width 1080, got %d", shot.Width)
	}

	// Test QuickAction home
	res, err := svc.QuickAction("test-dev-1", "home", "")
	if err != nil || !res.Success || !mock.wentHome {
		t.Errorf("QuickAction home failed: %v", err)
	}

	// Test QuickAction open_url
	res, err = svc.QuickAction("test-dev-1", "open_url", "https://mobrig.dev")
	if err != nil || !res.Success || mock.openedURL != "https://mobrig.dev" {
		t.Errorf("QuickAction open_url failed: %v", err)
	}

	// Test GetSetupStatus
	status := svc.GetSetupStatus()
	if status == nil {
		t.Fatal("expected non-nil setup status")
	}

	// Test AutoConfigureMCP
	configured := svc.AutoConfigureMCP(tmpDir)
	if configured == nil {
		t.Error("expected non-nil configured slice")
	}

	// Test Greet
	msg := svc.Greet("Tester")
	if msg != "Hello Tester from MobRig v3.0.0-test!" {
		t.Errorf("unexpected greeting: %q", msg)
	}
}
