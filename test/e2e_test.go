package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/TheJenilDGohel/MobRig/internal/config"
	"github.com/TheJenilDGohel/MobRig/internal/database"
	mobrigmcp "github.com/TheJenilDGohel/MobRig/internal/mcp"
	"github.com/TheJenilDGohel/MobRig/internal/platform"
	"github.com/TheJenilDGohel/MobRig/internal/server"
	"github.com/TheJenilDGohel/MobRig/internal/setup"
)

type e2eMockAdapter struct {
	platform.PlatformAdapter
	tappedX int
	tappedY int
}

func (m *e2eMockAdapter) Platform() string {
	return "android"
}

func (m *e2eMockAdapter) ListDevices(ctx context.Context) ([]platform.DeviceInfo, error) {
	return []platform.DeviceInfo{
		{
			ID:             "e2e-pixel-8",
			Platform:       platform.PlatformAndroid,
			Model:          "Pixel 8 Pro",
			Manufacturer:   "Google",
			OSVersion:      "14",
			APILevel:       34,
			ScreenWidth:    1080,
			ScreenHeight:   2400,
			Status:         platform.StatusConnected,
			ConnectionType: platform.ConnectionUSB,
		},
	}, nil
}

func (m *e2eMockAdapter) GetDeviceInfo(ctx context.Context, deviceID string) (*platform.DeviceInfo, error) {
	if deviceID != "e2e-pixel-8" {
		return nil, fmt.Errorf("device not found")
	}
	return &platform.DeviceInfo{
		ID:           "e2e-pixel-8",
		Platform:     platform.PlatformAndroid,
		Model:        "Pixel 8 Pro",
		Manufacturer: "Google",
		OSVersion:    "14",
		APILevel:     34,
		ScreenWidth:  1080,
		ScreenHeight: 2400,
	}, nil
}

func (m *e2eMockAdapter) GetScreenshot(ctx context.Context, deviceID string) (*platform.Screenshot, error) {
	png := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
	return &platform.Screenshot{
		Data:      png,
		Base64:    "iVBORw0KGgo=",
		Width:     1080,
		Height:    2400,
		Format:    "png",
		SizeBytes: len(png),
	}, nil
}

func (m *e2eMockAdapter) GetUITree(ctx context.Context, deviceID string, opts platform.UITreeOptions) (*platform.UITree, error) {
	return &platform.UITree{
		Elements: []platform.UIElement{
			{
				Index:      0,
				Type:       "Button",
				Text:       "Submit",
				ResourceID: "com.example:id/submit",
				Bounds:     platform.Bounds{Left: 100, Top: 500, Right: 300, Bottom: 600}, // Center: (200, 550)
			},
		},
		TokenCount: 80,
		AppPackage: "com.example.app",
	}, nil
}

func (m *e2eMockAdapter) Tap(ctx context.Context, deviceID string, x, y int) (*platform.ActionResult, error) {
	m.tappedX = x
	m.tappedY = y
	return &platform.ActionResult{Success: true, Message: "Tapped", Duration: 10 * time.Millisecond}, nil
}

func (m *e2eMockAdapter) DoubleTap(ctx context.Context, deviceID string, x, y int) (*platform.ActionResult, error) {
	return &platform.ActionResult{Success: true, Message: "Double tapped", Duration: 10 * time.Millisecond}, nil
}

func (m *e2eMockAdapter) LongPress(ctx context.Context, deviceID string, x, y, durMs int) (*platform.ActionResult, error) {
	return &platform.ActionResult{Success: true, Message: "Long pressed", Duration: 10 * time.Millisecond}, nil
}

func (m *e2eMockAdapter) Swipe(ctx context.Context, deviceID string, x1, y1, x2, y2, durMs int) (*platform.ActionResult, error) {
	return &platform.ActionResult{Success: true, Message: "Swiped", Duration: 10 * time.Millisecond}, nil
}

func (m *e2eMockAdapter) TypeText(ctx context.Context, deviceID, text string, submit bool) (*platform.ActionResult, error) {
	return &platform.ActionResult{Success: true, Message: "Typed", Duration: 10 * time.Millisecond}, nil
}

func (m *e2eMockAdapter) PressButton(ctx context.Context, deviceID string, key int) (*platform.ActionResult, error) {
	return &platform.ActionResult{Success: true, Message: "Key pressed", Duration: 10 * time.Millisecond}, nil
}

func (m *e2eMockAdapter) GoHome(ctx context.Context, deviceID string) (*platform.ActionResult, error) {
	return &platform.ActionResult{Success: true, Message: "Home pressed", Duration: 10 * time.Millisecond}, nil
}

func (m *e2eMockAdapter) LaunchApp(ctx context.Context, deviceID, pkg string) (*platform.ActionResult, error) {
	return &platform.ActionResult{Success: true, Message: "Launched", Duration: 10 * time.Millisecond}, nil
}

func (m *e2eMockAdapter) TerminateApp(ctx context.Context, deviceID, pkg string) (*platform.ActionResult, error) {
	return &platform.ActionResult{Success: true, Message: "Terminated", Duration: 10 * time.Millisecond}, nil
}

func (m *e2eMockAdapter) InstallApp(ctx context.Context, deviceID, path string) (*platform.ActionResult, error) {
	return &platform.ActionResult{Success: true, Message: "Installed", Duration: 10 * time.Millisecond}, nil
}

func (m *e2eMockAdapter) UninstallApp(ctx context.Context, deviceID, pkg string) (*platform.ActionResult, error) {
	return &platform.ActionResult{Success: true, Message: "Uninstalled", Duration: 10 * time.Millisecond}, nil
}

func (m *e2eMockAdapter) ClearAppData(ctx context.Context, deviceID, pkg string) (*platform.ActionResult, error) {
	return &platform.ActionResult{Success: true, Message: "Cleared", Duration: 10 * time.Millisecond}, nil
}

func (m *e2eMockAdapter) OpenURL(ctx context.Context, deviceID, url string) (*platform.ActionResult, error) {
	return &platform.ActionResult{Success: true, Message: "Opened", Duration: 10 * time.Millisecond}, nil
}

func TestEndToEndSystem(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "mobrig_e2e.db")

	// 1. Initialize Database & Migrations
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	repos := database.NewRepositories(db)

	// 2. Initialize Platform Adapter Registry
	registry := platform.NewAdapterRegistry()
	mock := &e2eMockAdapter{}
	registry.Register(mock)

	// 3. Initialize HTTP Server
	cfg := &config.Config{
		DataDir:  tmpDir,
		DBPath:   dbPath,
		HTTPPort: 0, // OS assigns random available port
		HTTPHost: "127.0.0.1",
		Version:  "3.0.0-e2e",
		Debug:    true,
	}

	srv := server.New(cfg, registry, repos)
	if _, err := srv.Listen(); err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	time.Sleep(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	defer func() {
		_ = srv.Shutdown(ctx)
	}()

	baseURL := fmt.Sprintf("http://%s", srv.Addr())

	// Test 4: Health Check
	res, err := http.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("failed to query health: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 health, got %d", res.StatusCode)
	}
	var health server.HealthResponse
	_ = json.NewDecoder(res.Body).Decode(&health)
	res.Body.Close()
	if health.Version != "3.0.0-e2e" {
		t.Errorf("expected version 3.0.0-e2e, got %s", health.Version)
	}

	// Test 5: List Devices
	res, err = http.Get(baseURL + "/api/v1/devices")
	if err != nil {
		t.Fatalf("failed to query devices: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 devices, got %d", res.StatusCode)
	}
	var devList server.DevicesListResponse
	_ = json.NewDecoder(res.Body).Decode(&devList)
	res.Body.Close()
	if devList.Count != 1 || devList.Devices[0].ID != "e2e-pixel-8" {
		t.Fatalf("unexpected devices: %+v", devList)
	}

	// Test 6: Device Detail
	res, err = http.Get(baseURL + "/api/v1/devices/e2e-pixel-8")
	if err != nil {
		t.Fatalf("failed to query device detail: %v", err)
	}
	var devDetail server.DeviceDetailResponse
	_ = json.NewDecoder(res.Body).Decode(&devDetail)
	res.Body.Close()
	if devDetail.Device == nil || devDetail.Device.Model != "Pixel 8 Pro" {
		t.Errorf("unexpected device detail: %+v", devDetail.Device)
	}

	// Test 7: Live Screenshot (JSON)
	res, err = http.Get(baseURL + "/api/v1/devices/e2e-pixel-8/screenshot")
	if err != nil {
		t.Fatalf("failed to get screenshot: %v", err)
	}
	var shot server.ScreenshotResponse
	_ = json.NewDecoder(res.Body).Decode(&shot)
	res.Body.Close()
	if shot.Width != 1080 || shot.Height != 2400 {
		t.Errorf("unexpected screenshot dimensions: %dx%d", shot.Width, shot.Height)
	}

	// Test 8: UI Tree
	res, err = http.Get(baseURL + "/api/v1/devices/e2e-pixel-8/ui-tree")
	if err != nil {
		t.Fatalf("failed to get UI tree: %v", err)
	}
	var uiTree server.UITreeResponse
	_ = json.NewDecoder(res.Body).Decode(&uiTree)
	res.Body.Close()
	if uiTree.Tree == nil || len(uiTree.Tree.Elements) != 1 {
		t.Fatalf("unexpected UI tree: %+v", uiTree)
	}

	// Test 9: Tap by Element Index (Element 0 Center is 200, 550)
	tapBody := bytes.NewBufferString(`{"index": 0}`)
	res, err = http.Post(baseURL+"/api/v1/devices/e2e-pixel-8/tap", "application/json", tapBody)
	if err != nil {
		t.Fatalf("failed to tap: %v", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 tap, got %d", res.StatusCode)
	}
	res.Body.Close()
	if mock.tappedX != 200 || mock.tappedY != 550 {
		t.Errorf("expected tap at (200, 550), got (%d, %d)", mock.tappedX, mock.tappedY)
	}

	// Test 10: RFC 7807 Error on Nonexistent Device
	res, err = http.Get(baseURL + "/api/v1/devices/ghost-device")
	if err != nil {
		t.Fatalf("failed request: %v", err)
	}
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", res.StatusCode)
	}
	var prob server.ProblemDetail
	_ = json.NewDecoder(res.Body).Decode(&prob)
	res.Body.Close()
	if prob.Status != 404 || prob.Title != "Not Found" {
		t.Errorf("unexpected problem detail: %+v", prob)
	}

	// Test 11: MCP Server Creation
	mcpServer, err := mobrigmcp.NewServer(registry, repos, "3.0.0-e2e")
	if err != nil {
		t.Fatalf("failed to create MCP server: %v", err)
	}
	if mcpServer.UnderlyingServer() == nil {
		t.Fatal("expected non-nil MCP underlying server")
	}

	// Test 12: Auto-Setup Diagnostics & Config Generation
	report := setup.CheckEnvironment("", "")
	if report.CLIInstallCmd == "" {
		t.Error("expected non-empty CLI install command in setup report")
	}

	configured, errs := setup.GenerateAllConfigs(tmpDir, "C:\\test\\mobrig.exe")
	if len(configured) == 0 {
		t.Error("expected at least one MCP config file generated")
	}
	if len(errs) > 0 {
		// Log notices if any
		for _, e := range errs {
			t.Logf("config generation notice: %v", e)
		}
	}
}
