package platform

import "context"

// PlatformAdapter is the interface that all device platform adapters must implement.
// S03 (Android Adapter) implements this for Android via dual-path (android CLI + ADB fallback).
// Future iOS adapter (M003) will also implement this interface.
//
// All methods accept a context for cancellation and timeouts.
// All methods return structured types + NLP-friendly messages in ActionResult.
type PlatformAdapter interface {
	// Platform returns the platform name (e.g., "android", "ios").
	Platform() string

	// ListDevices discovers and returns all connected devices for this platform.
	ListDevices(ctx context.Context) ([]DeviceInfo, error)

	// GetScreenshot captures the current screen of the specified device.
	GetScreenshot(ctx context.Context, deviceID string) (*Screenshot, error)

	// GetUITree reads the current UI hierarchy of the specified device.
	GetUITree(ctx context.Context, deviceID string, opts UITreeOptions) (*UITree, error)

	// Tap performs a tap at the given coordinates on the device.
	Tap(ctx context.Context, deviceID string, x, y int) (*ActionResult, error)

	// DoubleTap performs a double-tap at the given coordinates.
	DoubleTap(ctx context.Context, deviceID string, x, y int) (*ActionResult, error)

	// LongPress performs a long press at the given coordinates.
	LongPress(ctx context.Context, deviceID string, x, y int, durationMs int) (*ActionResult, error)

	// Swipe performs a swipe gesture from (x1,y1) to (x2,y2).
	Swipe(ctx context.Context, deviceID string, x1, y1, x2, y2, durationMs int) (*ActionResult, error)

	// TypeText types text into the currently focused field on the device.
	// If submit is true, presses Enter/Return after typing.
	TypeText(ctx context.Context, deviceID string, text string, submit bool) (*ActionResult, error)

	// PressButton presses a hardware/system button by key code.
	PressButton(ctx context.Context, deviceID string, keyCode int) (*ActionResult, error)

	// GoHome presses the home button.
	GoHome(ctx context.Context, deviceID string) (*ActionResult, error)

	// LaunchApp starts the specified app on the device.
	LaunchApp(ctx context.Context, deviceID string, packageName string) (*ActionResult, error)

	// TerminateApp force-stops the specified app.
	TerminateApp(ctx context.Context, deviceID string, packageName string) (*ActionResult, error)

	// InstallApp installs an app from the specified local path.
	InstallApp(ctx context.Context, deviceID string, appPath string) (*ActionResult, error)

	// UninstallApp removes the specified app from the device.
	UninstallApp(ctx context.Context, deviceID string, packageName string) (*ActionResult, error)

	// OpenURL opens a URL on the device (via default browser or deep link handler).
	OpenURL(ctx context.Context, deviceID string, url string) (*ActionResult, error)

	// ClearAppData clears all data for the specified app.
	ClearAppData(ctx context.Context, deviceID string, packageName string) (*ActionResult, error)

	// GetDeviceInfo returns detailed information about a specific device.
	GetDeviceInfo(ctx context.Context, deviceID string) (*DeviceInfo, error)
}

// AdapterRegistry manages multiple platform adapters (e.g., Android + iOS).
type AdapterRegistry struct {
	adapters []PlatformAdapter
}

// NewAdapterRegistry creates an empty registry.
func NewAdapterRegistry() *AdapterRegistry {
	return &AdapterRegistry{}
}

// Register adds a platform adapter to the registry.
func (r *AdapterRegistry) Register(adapter PlatformAdapter) {
	r.adapters = append(r.adapters, adapter)
}

// ListAllDevices queries all registered adapters and aggregates their devices.
func (r *AdapterRegistry) ListAllDevices(ctx context.Context) ([]DeviceInfo, error) {
	var all []DeviceInfo
	for _, adapter := range r.adapters {
		devices, err := adapter.ListDevices(ctx)
		if err != nil {
			// Log error but continue with other adapters
			continue
		}
		all = append(all, devices...)
	}
	return all, nil
}

// FindAdapter returns the adapter that owns the given device ID, or nil.
func (r *AdapterRegistry) FindAdapter(ctx context.Context, deviceID string) PlatformAdapter {
	for _, adapter := range r.adapters {
		devices, err := adapter.ListDevices(ctx)
		if err != nil {
			continue
		}
		for _, d := range devices {
			if d.ID == deviceID {
				return adapter
			}
		}
	}
	return nil
}
