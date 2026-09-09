package android

import (
	"context"
	"fmt"
	"log"

	"github.com/TheJenilDGohel/MobRig/internal/database"
	"github.com/TheJenilDGohel/MobRig/internal/platform"
)

// AndroidAdapter implements platform.PlatformAdapter for Android devices.
// It provides a dual-path execution strategy:
// 1. Primary: Google's official `android` CLI when installed
// 2. Fallback: Direct `adb` CLI commands + high-speed compact XML parsing
type AndroidAdapter struct {
	adb   *ADBClient
	cli   *AndroidCLI
	repos *database.Repositories
}

// NewAdapter creates an AndroidAdapter with configured ADB client and optional CLI.
func NewAdapter(adb *ADBClient, cli *AndroidCLI, repos *database.Repositories) *AndroidAdapter {
	return &AndroidAdapter{
		adb:   adb,
		cli:   cli,
		repos: repos,
	}
}

// Platform returns "android".
func (a *AndroidAdapter) Platform() string {
	return string(platform.PlatformAndroid)
}

// HasCLI returns true if the Google android CLI is available and active.
func (a *AndroidAdapter) HasCLI() bool {
	return a.cli != nil
}

// ListDevices discovers and returns all connected Android devices.
func (a *AndroidAdapter) ListDevices(ctx context.Context) ([]platform.DeviceInfo, error) {
	devices, err := a.adb.Devices(ctx)
	if err != nil {
		return nil, err
	}

	// Persist connected devices into SQLite repository if available
	if a.repos != nil {
		for _, dev := range devices {
			go func(d platform.DeviceInfo) {
				_ = a.repos.Devices.Upsert(context.Background(), database.Device{
					ID:             d.ID,
					Platform:       string(d.Platform),
					Model:          d.Model,
					Manufacturer:   d.Manufacturer,
					ConnectionType: string(d.ConnectionType),
				})
			}(dev)
		}
	}

	return devices, nil
}

// GetDeviceInfo queries detailed hardware and OS specs for a device.
func (a *AndroidAdapter) GetDeviceInfo(ctx context.Context, deviceID string) (*platform.DeviceInfo, error) {
	info, err := a.adb.GetDeviceInfo(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	// Update SQLite device entry with full specs
	if a.repos != nil {
		_ = a.repos.Devices.Upsert(ctx, database.Device{
			ID:             info.ID,
			Platform:       string(info.Platform),
			Model:          info.Model,
			Manufacturer:   info.Manufacturer,
			OSVersion:      info.OSVersion,
			APILevel:       info.APILevel,
			ScreenWidth:    info.ScreenWidth,
			ScreenHeight:   info.ScreenHeight,
			ScreenDensity:  info.ScreenDensity,
			ConnectionType: string(info.ConnectionType),
		})
	}

	return info, nil
}

// GetScreenshot captures the current screen as a PNG screenshot.
func (a *AndroidAdapter) GetScreenshot(ctx context.Context, deviceID string) (*platform.Screenshot, error) {
	return a.adb.Screenshot(ctx, deviceID)
}

// GetUITree extracts the compact UI hierarchy using the dual-path strategy.
func (a *AndroidAdapter) GetUITree(ctx context.Context, deviceID string, opts platform.UITreeOptions) (*platform.UITree, error) {
	// Path 1: If Google's android CLI is available, try it first
	if a.cli != nil {
		tree, err := a.cli.GetLayout(ctx, deviceID, opts)
		if err == nil {
			return tree, nil
		}
		// Log warning and fallback to Path 2
		log.Printf("[android-adapter] android CLI layout failed (%v), falling back to ADB uiautomator", err)
	}

	// Path 2: Fallback via adb uiautomator dump + compact XML parser
	xmlContent, err := a.adb.DumpUIHierarchy(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("adb uiautomator dump: %w", err)
	}

	return ParseUITreeXML(xmlContent, opts)
}

// Tap performs a touch event at (x, y).
func (a *AndroidAdapter) Tap(ctx context.Context, deviceID string, x, y int) (*platform.ActionResult, error) {
	return a.adb.Tap(ctx, deviceID, x, y)
}

// DoubleTap performs two taps in quick succession at (x, y).
func (a *AndroidAdapter) DoubleTap(ctx context.Context, deviceID string, x, y int) (*platform.ActionResult, error) {
	return a.adb.DoubleTap(ctx, deviceID, x, y)
}

// LongPress performs a long press at (x, y).
func (a *AndroidAdapter) LongPress(ctx context.Context, deviceID string, x, y int, durationMs int) (*platform.ActionResult, error) {
	return a.adb.LongPress(ctx, deviceID, x, y, durationMs)
}

// Swipe performs a swipe gesture from (x1, y1) to (x2, y2).
func (a *AndroidAdapter) Swipe(ctx context.Context, deviceID string, x1, y1, x2, y2 int, durationMs int) (*platform.ActionResult, error) {
	return a.adb.Swipe(ctx, deviceID, x1, y1, x2, y2, durationMs)
}

// TypeText sends text input to the currently focused element.
func (a *AndroidAdapter) TypeText(ctx context.Context, deviceID string, text string, submit bool) (*platform.ActionResult, error) {
	return a.adb.TypeText(ctx, deviceID, text, submit)
}

// PressButton simulates a hardware/system key event.
func (a *AndroidAdapter) PressButton(ctx context.Context, deviceID string, keyCode int) (*platform.ActionResult, error) {
	return a.adb.PressButton(ctx, deviceID, keyCode)
}

// GoHome navigates to the device home screen.
func (a *AndroidAdapter) GoHome(ctx context.Context, deviceID string) (*platform.ActionResult, error) {
	return a.adb.GoHome(ctx, deviceID)
}

// LaunchApp starts an application.
func (a *AndroidAdapter) LaunchApp(ctx context.Context, deviceID string, packageName string) (*platform.ActionResult, error) {
	return a.adb.LaunchApp(ctx, deviceID, packageName)
}

// TerminateApp force-stops an application.
func (a *AndroidAdapter) TerminateApp(ctx context.Context, deviceID string, packageName string) (*platform.ActionResult, error) {
	return a.adb.TerminateApp(ctx, deviceID, packageName)
}

// InstallApp installs an APK package from a local filesystem path.
func (a *AndroidAdapter) InstallApp(ctx context.Context, deviceID string, appPath string) (*platform.ActionResult, error) {
	return a.adb.InstallApp(ctx, deviceID, appPath)
}

// UninstallApp removes an application from the device.
func (a *AndroidAdapter) UninstallApp(ctx context.Context, deviceID string, packageName string) (*platform.ActionResult, error) {
	return a.adb.UninstallApp(ctx, deviceID, packageName)
}

// OpenURL opens a URL or deep link on the device.
func (a *AndroidAdapter) OpenURL(ctx context.Context, deviceID string, url string) (*platform.ActionResult, error) {
	return a.adb.OpenURL(ctx, deviceID, url)
}

// ClearAppData resets all state and caches for the given app.
func (a *AndroidAdapter) ClearAppData(ctx context.Context, deviceID string, packageName string) (*platform.ActionResult, error) {
	return a.adb.ClearAppData(ctx, deviceID, packageName)
}
