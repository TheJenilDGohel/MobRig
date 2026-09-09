package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/TheJenilDGohel/MobRig/internal/config"
	"github.com/TheJenilDGohel/MobRig/internal/platform"
)

// mockServerAdapter implements platform.PlatformAdapter for HTTP testing.
type mockServerAdapter struct {
	platform.PlatformAdapter
	tappedX      int
	tappedY      int
	longPressedX int
	longPressedY int
	swipedX1     int
	swipedY1     int
	swipedX2     int
	swipedY2     int
	typedText    string
	lastKeyCode  int
	wentHome     bool
	launchedApp  string
	termApp      string
	installedApp string
	uninstalled  string
	clearedApp   string
	openedURL    string
}

func (m *mockServerAdapter) Platform() string {
	return "android"
}

func (m *mockServerAdapter) ListDevices(ctx context.Context) ([]platform.DeviceInfo, error) {
	return []platform.DeviceInfo{
		{
			ID:             "mock-pixel-8",
			Platform:       platform.PlatformAndroid,
			Model:          "Pixel 8",
			Status:         platform.StatusConnected,
			ConnectionType: platform.ConnectionUSB,
		},
	}, nil
}

func (m *mockServerAdapter) GetDeviceInfo(ctx context.Context, deviceID string) (*platform.DeviceInfo, error) {
	if deviceID != "mock-pixel-8" {
		return nil, fmt.Errorf("device not found")
	}
	return &platform.DeviceInfo{
		ID:           "mock-pixel-8",
		Platform:     platform.PlatformAndroid,
		Model:        "Pixel 8",
		Manufacturer: "Google",
		OSVersion:    "14",
		APILevel:     34,
		ScreenWidth:  1080,
		ScreenHeight: 2400,
	}, nil
}

func (m *mockServerAdapter) GetScreenshot(ctx context.Context, deviceID string) (*platform.Screenshot, error) {
	pngData := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
	return &platform.Screenshot{
		Data:      pngData,
		Base64:    "iVBORw0KGgo=",
		Width:     1080,
		Height:    2400,
		Format:    "png",
		Timestamp: time.Now(),
		SizeBytes: len(pngData),
	}, nil
}

func (m *mockServerAdapter) GetUITree(ctx context.Context, deviceID string, opts platform.UITreeOptions) (*platform.UITree, error) {
	return &platform.UITree{
		Elements: []platform.UIElement{
			{
				Index:      0,
				Type:       "Button",
				Text:       "Login",
				ResourceID: "com.example.app:id/btn_login",
				Bounds:     platform.Bounds{Left: 100, Top: 200, Right: 300, Bottom: 400}, // Center: (200, 300)
				Clickable:  true,
			},
			{
				Index:      1,
				Type:       "TextView",
				Text:       "Welcome to MobRig",
				ResourceID: "com.example.app:id/tv_title",
				Bounds:     platform.Bounds{Left: 50, Top: 50, Right: 500, Bottom: 100},
			},
		},
		TokenCount: 150,
		Format:     "compact",
		AppPackage: "com.example.app",
	}, nil
}

func (m *mockServerAdapter) Tap(ctx context.Context, deviceID string, x, y int) (*platform.ActionResult, error) {
	m.tappedX = x
	m.tappedY = y
	return &platform.ActionResult{Success: true, Message: "Tapped", Duration: 15 * time.Millisecond}, nil
}

func (m *mockServerAdapter) DoubleTap(ctx context.Context, deviceID string, x, y int) (*platform.ActionResult, error) {
	return &platform.ActionResult{Success: true, Message: "Double-tapped", Duration: 30 * time.Millisecond}, nil
}

func (m *mockServerAdapter) LongPress(ctx context.Context, deviceID string, x, y, durationMs int) (*platform.ActionResult, error) {
	m.longPressedX = x
	m.longPressedY = y
	return &platform.ActionResult{Success: true, Message: "Long-pressed", Duration: time.Duration(durationMs) * time.Millisecond}, nil
}

func (m *mockServerAdapter) Swipe(ctx context.Context, deviceID string, x1, y1, x2, y2, durationMs int) (*platform.ActionResult, error) {
	m.swipedX1 = x1
	m.swipedY1 = y1
	m.swipedX2 = x2
	m.swipedY2 = y2
	return &platform.ActionResult{Success: true, Message: "Swiped", Duration: time.Duration(durationMs) * time.Millisecond}, nil
}

func (m *mockServerAdapter) TypeText(ctx context.Context, deviceID, text string, submit bool) (*platform.ActionResult, error) {
	m.typedText = text
	return &platform.ActionResult{Success: true, Message: "Typed", Duration: 20 * time.Millisecond}, nil
}

func (m *mockServerAdapter) PressButton(ctx context.Context, deviceID string, keyCode int) (*platform.ActionResult, error) {
	m.lastKeyCode = keyCode
	return &platform.ActionResult{Success: true, Message: "Key pressed", Duration: 10 * time.Millisecond}, nil
}

func (m *mockServerAdapter) GoHome(ctx context.Context, deviceID string) (*platform.ActionResult, error) {
	m.wentHome = true
	return &platform.ActionResult{Success: true, Message: "Home pressed", Duration: 10 * time.Millisecond}, nil
}

func (m *mockServerAdapter) LaunchApp(ctx context.Context, deviceID, pkg string) (*platform.ActionResult, error) {
	m.launchedApp = pkg
	return &platform.ActionResult{Success: true, Message: "App launched", Duration: 50 * time.Millisecond}, nil
}

func (m *mockServerAdapter) TerminateApp(ctx context.Context, deviceID, pkg string) (*platform.ActionResult, error) {
	m.termApp = pkg
	return &platform.ActionResult{Success: true, Message: "App terminated", Duration: 20 * time.Millisecond}, nil
}

func (m *mockServerAdapter) InstallApp(ctx context.Context, deviceID, path string) (*platform.ActionResult, error) {
	m.installedApp = path
	return &platform.ActionResult{Success: true, Message: "App installed", Duration: 500 * time.Millisecond}, nil
}

func (m *mockServerAdapter) UninstallApp(ctx context.Context, deviceID, pkg string) (*platform.ActionResult, error) {
	m.uninstalled = pkg
	return &platform.ActionResult{Success: true, Message: "App uninstalled", Duration: 200 * time.Millisecond}, nil
}

func (m *mockServerAdapter) ClearAppData(ctx context.Context, deviceID, pkg string) (*platform.ActionResult, error) {
	m.clearedApp = pkg
	return &platform.ActionResult{Success: true, Message: "Data cleared", Duration: 80 * time.Millisecond}, nil
}

func (m *mockServerAdapter) OpenURL(ctx context.Context, deviceID, url string) (*platform.ActionResult, error) {
	m.openedURL = url
	return &platform.ActionResult{Success: true, Message: "URL opened", Duration: 40 * time.Millisecond}, nil
}

func setupTestServer() (*Server, *mockServerAdapter) {
	cfg := &config.Config{
		HTTPPort: 8686,
		HTTPHost: "127.0.0.1",
		Version:  "3.0.0-test",
	}
	registry := platform.NewAdapterRegistry()
	mock := &mockServerAdapter{}
	registry.Register(mock)

	srv := New(cfg, registry, nil)
	return srv, mock
}

func TestHandleHealth(t *testing.T) {
	srv, _ := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode JSON response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", resp.Status)
	}
	if resp.Version != "3.0.0-test" {
		t.Errorf("expected version '3.0.0-test', got %q", resp.Version)
	}
}

func TestHandleListDevices(t *testing.T) {
	srv, _ := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/devices", nil)
	w := httptest.NewRecorder()

	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp DevicesListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if resp.Count != 1 {
		t.Fatalf("expected 1 device, got %d", resp.Count)
	}
	if resp.Devices[0].ID != "mock-pixel-8" {
		t.Errorf("expected ID 'mock-pixel-8', got %q", resp.Devices[0].ID)
	}
}

func TestHandleGetDevice(t *testing.T) {
	srv, _ := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/devices/mock-pixel-8", nil)
	w := httptest.NewRecorder()

	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp DeviceDetailResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if resp.Device == nil || resp.Device.Model != "Pixel 8" {
		t.Errorf("expected model 'Pixel 8', got %+v", resp.Device)
	}
}

func TestHandleGetDevice_NotFound(t *testing.T) {
	srv, _ := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/devices/nonexistent-device", nil)
	w := httptest.NewRecorder()

	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}

	var prob ProblemDetail
	if err := json.Unmarshal(w.Body.Bytes(), &prob); err != nil {
		t.Fatalf("failed to parse problem detail: %v", err)
	}

	if prob.Status != 404 {
		t.Errorf("expected problem status 404, got %d", prob.Status)
	}
}

func TestHandleGetScreenshot(t *testing.T) {
	srv, _ := setupTestServer()

	// JSON response
	req := httptest.NewRequest(http.MethodGet, "/api/v1/devices/mock-pixel-8/screenshot", nil)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	var resp ScreenshotResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse screenshot JSON: %v", err)
	}
	if resp.Width != 1080 || resp.Height != 2400 {
		t.Errorf("unexpected dimensions: %dx%d", resp.Width, resp.Height)
	}

	// Raw binary response
	rawReq := httptest.NewRequest(http.MethodGet, "/api/v1/devices/mock-pixel-8/screenshot?raw=true", nil)
	rawW := httptest.NewRecorder()
	srv.mux.ServeHTTP(rawW, rawReq)

	if rawW.Code != http.StatusOK {
		t.Fatalf("expected status 200 for raw screenshot, got %d", rawW.Code)
	}
	if rawW.Header().Get("Content-Type") != "image/png" {
		t.Errorf("expected image/png content type, got %q", rawW.Header().Get("Content-Type"))
	}
}

func TestHandleGetUITree(t *testing.T) {
	srv, _ := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/devices/mock-pixel-8/ui-tree?visible_only=true", nil)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp UITreeResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode UI tree JSON: %v", err)
	}

	if resp.Tree == nil || len(resp.Tree.Elements) != 2 {
		t.Fatalf("expected 2 elements in tree, got %+v", resp.Tree)
	}
}

func TestHandleFindElements(t *testing.T) {
	srv, _ := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/devices/mock-pixel-8/elements?query=Login", nil)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp ElementsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse elements JSON: %v", err)
	}

	if resp.Count != 1 || resp.Elements[0].Text != "Login" {
		t.Errorf("expected 1 element with text 'Login', got %+v", resp)
	}
}

func TestHandleTap_CoordinatesAndIndex(t *testing.T) {
	srv, mock := setupTestServer()

	// 1. By Coordinates
	body := bytes.NewBufferString(`{"x": 150, "y": 250}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/devices/mock-pixel-8/tap", body)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if mock.tappedX != 150 || mock.tappedY != 250 {
		t.Errorf("expected tapped at (150, 250), got (%d, %d)", mock.tappedX, mock.tappedY)
	}

	// 2. By Index (element 0 bounds [100, 200][300, 400] -> center is 200, 300)
	bodyIndex := bytes.NewBufferString(`{"index": 0}`)
	reqIndex := httptest.NewRequest(http.MethodPost, "/api/v1/devices/mock-pixel-8/tap", bodyIndex)
	wIndex := httptest.NewRecorder()
	srv.mux.ServeHTTP(wIndex, reqIndex)

	if wIndex.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", wIndex.Code, wIndex.Body.String())
	}
	if mock.tappedX != 200 || mock.tappedY != 300 {
		t.Errorf("expected tapped at (200, 300), got (%d, %d)", mock.tappedX, mock.tappedY)
	}
}

func TestHandleSwipe_Direction(t *testing.T) {
	srv, mock := setupTestServer()

	// Swipe "up" on 1080x2400 screen with distance 0.5:
	// delta = 2400 * 0.5 / 2 = 600
	// midX = 540, midY = 1200
	// start: (540, 1800), end: (540, 600)
	body := bytes.NewBufferString(`{"direction": "up", "distance": 0.5}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/devices/mock-pixel-8/swipe", body)
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if mock.swipedX1 != 540 || mock.swipedY1 != 1800 || mock.swipedX2 != 540 || mock.swipedY2 != 600 {
		t.Errorf("unexpected swipe coordinates: (%d, %d) -> (%d, %d)",
			mock.swipedX1, mock.swipedY1, mock.swipedX2, mock.swipedY2)
	}
}

func TestHandleAppLifecycle(t *testing.T) {
	srv, mock := setupTestServer()

	// Launch
	req := httptest.NewRequest(http.MethodPost, "/api/v1/devices/mock-pixel-8/apps/launch", bytes.NewBufferString(`{"package_name": "com.test.app"}`))
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK || mock.launchedApp != "com.test.app" {
		t.Errorf("launch app failed: code %d, app %s", w.Code, mock.launchedApp)
	}

	// Terminate
	req = httptest.NewRequest(http.MethodPost, "/api/v1/devices/mock-pixel-8/apps/terminate", bytes.NewBufferString(`{"package_name": "com.test.app"}`))
	w = httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK || mock.termApp != "com.test.app" {
		t.Errorf("terminate app failed: code %d, app %s", w.Code, mock.termApp)
	}

	// Install
	req = httptest.NewRequest(http.MethodPost, "/api/v1/devices/mock-pixel-8/apps/install", bytes.NewBufferString(`{"app_path": "C:\\test\\app.apk"}`))
	w = httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK || mock.installedApp != "C:\\test\\app.apk" {
		t.Errorf("install app failed: code %d, app %s", w.Code, mock.installedApp)
	}

	// Uninstall
	req = httptest.NewRequest(http.MethodPost, "/api/v1/devices/mock-pixel-8/apps/uninstall", bytes.NewBufferString(`{"package_name": "com.test.app"}`))
	w = httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK || mock.uninstalled != "com.test.app" {
		t.Errorf("uninstall app failed: code %d, app %s", w.Code, mock.uninstalled)
	}

	// Clear data
	req = httptest.NewRequest(http.MethodPost, "/api/v1/devices/mock-pixel-8/apps/clear-data", bytes.NewBufferString(`{"package_name": "com.test.app"}`))
	w = httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK || mock.clearedApp != "com.test.app" {
		t.Errorf("clear data failed: code %d, app %s", w.Code, mock.clearedApp)
	}

	// Open URL
	req = httptest.NewRequest(http.MethodPost, "/api/v1/devices/mock-pixel-8/open-url", bytes.NewBufferString(`{"url": "https://example.com"}`))
	w = httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK || mock.openedURL != "https://example.com" {
		t.Errorf("open url failed: code %d, url %s", w.Code, mock.openedURL)
	}
}

func TestHandleInputs(t *testing.T) {
	srv, mock := setupTestServer()

	// Double tap
	req := httptest.NewRequest(http.MethodPost, "/api/v1/devices/mock-pixel-8/double-tap", bytes.NewBufferString(`{"x": 100, "y": 200}`))
	w := httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("double-tap failed: code %d", w.Code)
	}

	// Long press
	req = httptest.NewRequest(http.MethodPost, "/api/v1/devices/mock-pixel-8/long-press", bytes.NewBufferString(`{"x": 100, "y": 200, "duration_ms": 1500}`))
	w = httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK || mock.longPressedX != 100 || mock.longPressedY != 200 {
		t.Errorf("long-press failed: code %d", w.Code)
	}

	// Type text
	req = httptest.NewRequest(http.MethodPost, "/api/v1/devices/mock-pixel-8/type", bytes.NewBufferString(`{"text": "hello world", "submit": true}`))
	w = httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK || mock.typedText != "hello world" {
		t.Errorf("type text failed: code %d", w.Code)
	}

	// Press key
	req = httptest.NewRequest(http.MethodPost, "/api/v1/devices/mock-pixel-8/key", bytes.NewBufferString(`{"key_code": 4}`))
	w = httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK || mock.lastKeyCode != 4 {
		t.Errorf("press key failed: code %d", w.Code)
	}

	// Home
	req = httptest.NewRequest(http.MethodPost, "/api/v1/devices/mock-pixel-8/home", nil)
	w = httptest.NewRecorder()
	srv.mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK || !mock.wentHome {
		t.Errorf("go home failed: code %d", w.Code)
	}
}

func TestCORS(t *testing.T) {
	srv, _ := setupTestServer()

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/devices", nil)
	w := httptest.NewRecorder()

	srv.srv.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content for OPTIONS, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected CORS header *, got %q", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestServerStartAndShutdown(t *testing.T) {
	srv, _ := setupTestServer()
	srv.srv.Addr = "127.0.0.1:0"

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	time.Sleep(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("failed to shutdown server: %v", err)
	}

	err := <-errCh
	if err != nil {
		t.Errorf("srv.Start() returned error after shutdown: %v", err)
	}
}
