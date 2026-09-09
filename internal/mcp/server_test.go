package mcp

import (
	"context"
	"testing"
	"time"

	"github.com/TheJenilDGohel/MobRig/internal/platform"
)

// mockAdapter implements platform.PlatformAdapter for testing.
type mockAdapter struct {
	platform.PlatformAdapter
	tappedX int
	tappedY int
}

func (m *mockAdapter) Platform() string {
	return "android"
}

func (m *mockAdapter) ListDevices(ctx context.Context) ([]platform.DeviceInfo, error) {
	return []platform.DeviceInfo{
		{
			ID:             "mock-device-1",
			Platform:       platform.PlatformAndroid,
			Model:          "Pixel 8",
			Status:         platform.StatusConnected,
			ConnectionType: platform.ConnectionUSB,
		},
	}, nil
}

func (m *mockAdapter) GetDeviceInfo(ctx context.Context, deviceID string) (*platform.DeviceInfo, error) {
	return &platform.DeviceInfo{
		ID:           "mock-device-1",
		Platform:     platform.PlatformAndroid,
		Model:        "Pixel 8",
		ScreenWidth:  1080,
		ScreenHeight: 2400,
	}, nil
}

func (m *mockAdapter) GetUITree(ctx context.Context, deviceID string, opts platform.UITreeOptions) (*platform.UITree, error) {
	return &platform.UITree{
		Elements: []platform.UIElement{
			{
				Index:      0,
				Type:       "Button",
				Text:       "Login",
				ResourceID: "com.example.app:id/login",
				Bounds:     platform.Bounds{Left: 100, Top: 200, Right: 300, Bottom: 300}, // Center: (200, 250)
				Clickable:  true,
			},
		},
	}, nil
}

func (m *mockAdapter) Tap(ctx context.Context, deviceID string, x, y int) (*platform.ActionResult, error) {
	m.tappedX = x
	m.tappedY = y
	return &platform.ActionResult{
		Success:  true,
		Message:  "tapped",
		Duration: 10 * time.Millisecond,
	}, nil
}

func TestNewServer(t *testing.T) {
	registry := platform.NewAdapterRegistry()
	server, err := NewServer(registry, nil, "3.0.0-test")
	if err != nil {
		t.Fatalf("failed to create MCP server: %v", err)
	}

	if server.UnderlyingServer() == nil {
		t.Fatal("expected non-nil underlying server")
	}
}

func TestHandleListDevices(t *testing.T) {
	registry := platform.NewAdapterRegistry()
	mock := &mockAdapter{}
	registry.Register(mock)

	h := handleListDevices(registry)
	res, out, err := h(context.Background(), nil, ListDevicesInput{})
	if err != nil {
		t.Fatalf("unexpected handler error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success result, got error: %+v", res)
	}
	if out.Count != 1 {
		t.Fatalf("expected 1 device, got %d", out.Count)
	}
	if out.Devices[0].ID != "mock-device-1" {
		t.Errorf("expected device ID 'mock-device-1', got %q", out.Devices[0].ID)
	}
}

func TestHandleTap_WithIndex(t *testing.T) {
	registry := platform.NewAdapterRegistry()
	mock := &mockAdapter{}
	registry.Register(mock)

	h := handleTap(registry)
	idx := 0
	res, out, err := h(context.Background(), nil, TapInput{
		DeviceID: "mock-device-1",
		Index:    &idx,
	})
	if err != nil {
		t.Fatalf("unexpected handler error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected success result, got error: %+v", res)
	}
	if !out.Success {
		t.Error("expected out.Success to be true")
	}

	// Resolved center of [100,200][300,300] should be (200, 250)
	if mock.tappedX != 200 || mock.tappedY != 250 {
		t.Errorf("expected tap at (200, 250), got (%d, %d)", mock.tappedX, mock.tappedY)
	}
}

func TestHandleFindElement(t *testing.T) {
	registry := platform.NewAdapterRegistry()
	mock := &mockAdapter{}
	registry.Register(mock)

	h := handleFindElement(registry)

	// By text
	res, out, err := h(context.Background(), nil, FindElementInput{
		DeviceID: "mock-device-1",
		Text:     "Login",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.IsError || !out.Found {
		t.Error("expected element to be found")
	}
	if out.Element.ResourceID != "com.example.app:id/login" {
		t.Errorf("expected resource id 'com.example.app:id/login', got %q", out.Element.ResourceID)
	}

	// Non-existent
	res, out, err = h(context.Background(), nil, FindElementInput{
		DeviceID: "mock-device-1",
		Text:     "NonExistent",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.IsError || out.Found {
		t.Error("expected not found error")
	}
}

func TestResponseHelpers(t *testing.T) {
	tr := TextResult("hello")
	if tr.IsError || len(tr.Content) != 1 {
		t.Errorf("unexpected text result: %+v", tr)
	}

	er := ErrorResult("bad error: %d", 404)
	if !er.IsError || len(er.Content) != 1 {
		t.Errorf("unexpected error result: %+v", er)
	}

	sr := ScreenshotResult([]byte("fake-data"), 100, 200, "dev-1")
	if sr.IsError || len(sr.Content) != 2 {
		t.Errorf("unexpected screenshot result: %+v", sr)
	}
}
