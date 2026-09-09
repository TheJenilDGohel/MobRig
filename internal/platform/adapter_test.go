package platform

import (
	"context"
	"testing"
)

type mockAdapter struct {
	platform string
	devices  []DeviceInfo
}

func (m *mockAdapter) Platform() string {
	return m.platform
}

func (m *mockAdapter) ListDevices(_ context.Context) ([]DeviceInfo, error) {
	return m.devices, nil
}

func (m *mockAdapter) GetScreenshot(_ context.Context, _ string) (*Screenshot, error) {
	return &Screenshot{Width: 1080, Height: 1920}, nil
}

func (m *mockAdapter) GetUITree(_ context.Context, _ string, _ UITreeOptions) (*UITree, error) {
	return &UITree{}, nil
}

func (m *mockAdapter) Tap(_ context.Context, _ string, _, _ int) (*ActionResult, error) {
	return &ActionResult{Success: true, Message: "Tapped"}, nil
}

func (m *mockAdapter) DoubleTap(_ context.Context, _ string, _, _ int) (*ActionResult, error) {
	return &ActionResult{Success: true, Message: "Double tapped"}, nil
}

func (m *mockAdapter) LongPress(_ context.Context, _ string, _, _, _ int) (*ActionResult, error) {
	return &ActionResult{Success: true, Message: "Long pressed"}, nil
}

func (m *mockAdapter) Swipe(_ context.Context, _ string, _, _, _, _, _ int) (*ActionResult, error) {
	return &ActionResult{Success: true, Message: "Swiped"}, nil
}

func (m *mockAdapter) TypeText(_ context.Context, _ string, _ string, _ bool) (*ActionResult, error) {
	return &ActionResult{Success: true, Message: "Typed"}, nil
}

func (m *mockAdapter) PressButton(_ context.Context, _ string, _ int) (*ActionResult, error) {
	return &ActionResult{Success: true, Message: "Pressed button"}, nil
}

func (m *mockAdapter) GoHome(_ context.Context, _ string) (*ActionResult, error) {
	return &ActionResult{Success: true, Message: "Went home"}, nil
}

func (m *mockAdapter) LaunchApp(_ context.Context, _ string, _ string) (*ActionResult, error) {
	return &ActionResult{Success: true, Message: "Launched app"}, nil
}

func (m *mockAdapter) TerminateApp(_ context.Context, _ string, _ string) (*ActionResult, error) {
	return &ActionResult{Success: true, Message: "Terminated app"}, nil
}

func (m *mockAdapter) InstallApp(_ context.Context, _ string, _ string) (*ActionResult, error) {
	return &ActionResult{Success: true, Message: "Installed app"}, nil
}

func (m *mockAdapter) UninstallApp(_ context.Context, _ string, _ string) (*ActionResult, error) {
	return &ActionResult{Success: true, Message: "Uninstalled app"}, nil
}

func (m *mockAdapter) OpenURL(_ context.Context, _ string, _ string) (*ActionResult, error) {
	return &ActionResult{Success: true, Message: "Opened URL"}, nil
}

func (m *mockAdapter) ClearAppData(_ context.Context, _ string, _ string) (*ActionResult, error) {
	return &ActionResult{Success: true, Message: "Cleared app data"}, nil
}

func (m *mockAdapter) GetDeviceInfo(_ context.Context, deviceID string) (*DeviceInfo, error) {
	for _, d := range m.devices {
		if d.ID == deviceID {
			return &d, nil
		}
	}
	return nil, nil
}

func TestBoundsCalculations(t *testing.T) {
	b := Bounds{
		Left:   100,
		Top:    200,
		Right:  300,
		Bottom: 600,
	}

	if b.Width() != 200 {
		t.Errorf("expected width 200, got %d", b.Width())
	}
	if b.Height() != 400 {
		t.Errorf("expected height 400, got %d", b.Height())
	}
	if b.CenterX() != 200 {
		t.Errorf("expected CenterX 200, got %d", b.CenterX())
	}
	if b.CenterY() != 400 {
		t.Errorf("expected CenterY 400, got %d", b.CenterY())
	}
}

func TestAdapterRegistry(t *testing.T) {
	registry := NewAdapterRegistry()

	androidMock := &mockAdapter{
		platform: "android",
		devices: []DeviceInfo{
			{ID: "dev-android-1", Platform: PlatformAndroid, Status: StatusConnected},
			{ID: "dev-android-2", Platform: PlatformAndroid, Status: StatusOffline},
		},
	}

	iosMock := &mockAdapter{
		platform: "ios",
		devices: []DeviceInfo{
			{ID: "dev-ios-1", Platform: PlatformIOS, Status: StatusConnected},
		},
	}

	registry.Register(androidMock)
	registry.Register(iosMock)

	ctx := context.Background()
	all, err := registry.ListAllDevices(ctx)
	if err != nil {
		t.Fatalf("ListAllDevices failed: %v", err)
	}

	if len(all) != 3 {
		t.Fatalf("expected 3 devices, got %d", len(all))
	}

	foundAndroid := registry.FindAdapter(ctx, "dev-android-1")
	if foundAndroid == nil || foundAndroid.Platform() != "android" {
		t.Errorf("expected android adapter for dev-android-1")
	}

	foundIOS := registry.FindAdapter(ctx, "dev-ios-1")
	if foundIOS == nil || foundIOS.Platform() != "ios" {
		t.Errorf("expected ios adapter for dev-ios-1")
	}

	notFound := registry.FindAdapter(ctx, "nonexistent-id")
	if notFound != nil {
		t.Errorf("expected nil for nonexistent-id, got %v", notFound)
	}
}
