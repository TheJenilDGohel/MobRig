package mcp

import (
	"context"
	"fmt"

	"github.com/TheJenilDGohel/MobRig/internal/platform"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// --- launch_app ---

type LaunchAppInput struct {
	DeviceID    string `json:"device_id" jsonschema:"The serial or identifier of the device"`
	PackageName string `json:"package_name" jsonschema:"The package name to launch (e.g. com.example.app)"`
}

func handleLaunchApp(registry *platform.AdapterRegistry) mcp.ToolHandlerFor[LaunchAppInput, ActionOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in LaunchAppInput) (*mcp.CallToolResult, ActionOutput, error) {
		adapter := registry.FindAdapter(ctx, in.DeviceID)
		if adapter == nil {
			return ErrorResult("Device %q not found", in.DeviceID), ActionOutput{Success: false}, nil
		}

		res, err := adapter.LaunchApp(ctx, in.DeviceID, in.PackageName)
		if err != nil {
			return ErrorResult("Launch app failed: %v", err), ActionOutput{Success: false}, nil
		}

		msg := fmt.Sprintf("Launched application %s", in.PackageName)
		return TextResult(msg), ActionOutput{
			Success:  res.Success,
			Message:  msg,
			Duration: res.Duration.String(),
		}, nil
	}
}

// --- terminate_app ---

type TerminateAppInput struct {
	DeviceID    string `json:"device_id" jsonschema:"The serial or identifier of the device"`
	PackageName string `json:"package_name" jsonschema:"The package name to terminate (e.g. com.example.app)"`
}

func handleTerminateApp(registry *platform.AdapterRegistry) mcp.ToolHandlerFor[TerminateAppInput, ActionOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in TerminateAppInput) (*mcp.CallToolResult, ActionOutput, error) {
		adapter := registry.FindAdapter(ctx, in.DeviceID)
		if adapter == nil {
			return ErrorResult("Device %q not found", in.DeviceID), ActionOutput{Success: false}, nil
		}

		res, err := adapter.TerminateApp(ctx, in.DeviceID, in.PackageName)
		if err != nil {
			return ErrorResult("Terminate app failed: %v", err), ActionOutput{Success: false}, nil
		}

		msg := fmt.Sprintf("Terminated application %s", in.PackageName)
		return TextResult(msg), ActionOutput{
			Success:  res.Success,
			Message:  msg,
			Duration: res.Duration.String(),
		}, nil
	}
}

// --- install_app ---

type InstallAppInput struct {
	DeviceID string `json:"device_id" jsonschema:"The serial or identifier of the device"`
	AppPath  string `json:"app_path" jsonschema:"Local filesystem path to APK file"`
}

func handleInstallApp(registry *platform.AdapterRegistry) mcp.ToolHandlerFor[InstallAppInput, ActionOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in InstallAppInput) (*mcp.CallToolResult, ActionOutput, error) {
		adapter := registry.FindAdapter(ctx, in.DeviceID)
		if adapter == nil {
			return ErrorResult("Device %q not found", in.DeviceID), ActionOutput{Success: false}, nil
		}

		res, err := adapter.InstallApp(ctx, in.DeviceID, in.AppPath)
		if err != nil {
			return ErrorResult("Install app failed: %v", err), ActionOutput{Success: false}, nil
		}

		msg := fmt.Sprintf("Installed application from %s", in.AppPath)
		return TextResult(msg), ActionOutput{
			Success:  res.Success,
			Message:  msg,
			Duration: res.Duration.String(),
		}, nil
	}
}

// --- uninstall_app ---

type UninstallAppInput struct {
	DeviceID    string `json:"device_id" jsonschema:"The serial or identifier of the device"`
	PackageName string `json:"package_name" jsonschema:"Package name to uninstall"`
}

func handleUninstallApp(registry *platform.AdapterRegistry) mcp.ToolHandlerFor[UninstallAppInput, ActionOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in UninstallAppInput) (*mcp.CallToolResult, ActionOutput, error) {
		adapter := registry.FindAdapter(ctx, in.DeviceID)
		if adapter == nil {
			return ErrorResult("Device %q not found", in.DeviceID), ActionOutput{Success: false}, nil
		}

		res, err := adapter.UninstallApp(ctx, in.DeviceID, in.PackageName)
		if err != nil {
			return ErrorResult("Uninstall app failed: %v", err), ActionOutput{Success: false}, nil
		}

		msg := fmt.Sprintf("Uninstalled application %s", in.PackageName)
		return TextResult(msg), ActionOutput{
			Success:  res.Success,
			Message:  msg,
			Duration: res.Duration.String(),
		}, nil
	}
}

// --- open_url ---

type OpenURLInput struct {
	DeviceID string `json:"device_id" jsonschema:"The serial or identifier of the device"`
	URL      string `json:"url" jsonschema:"URL or deep link to open on the device"`
}

func handleOpenURL(registry *platform.AdapterRegistry) mcp.ToolHandlerFor[OpenURLInput, ActionOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in OpenURLInput) (*mcp.CallToolResult, ActionOutput, error) {
		adapter := registry.FindAdapter(ctx, in.DeviceID)
		if adapter == nil {
			return ErrorResult("Device %q not found", in.DeviceID), ActionOutput{Success: false}, nil
		}

		res, err := adapter.OpenURL(ctx, in.DeviceID, in.URL)
		if err != nil {
			return ErrorResult("Open URL failed: %v", err), ActionOutput{Success: false}, nil
		}

		msg := fmt.Sprintf("Opened URL %s", in.URL)
		return TextResult(msg), ActionOutput{
			Success:  res.Success,
			Message:  msg,
			Duration: res.Duration.String(),
		}, nil
	}
}

// --- clear_app_data ---

type ClearAppDataInput struct {
	DeviceID    string `json:"device_id" jsonschema:"The serial or identifier of the device"`
	PackageName string `json:"package_name" jsonschema:"Package name to clear all data/cache for"`
}

func handleClearAppData(registry *platform.AdapterRegistry) mcp.ToolHandlerFor[ClearAppDataInput, ActionOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in ClearAppDataInput) (*mcp.CallToolResult, ActionOutput, error) {
		adapter := registry.FindAdapter(ctx, in.DeviceID)
		if adapter == nil {
			return ErrorResult("Device %q not found", in.DeviceID), ActionOutput{Success: false}, nil
		}

		res, err := adapter.ClearAppData(ctx, in.DeviceID, in.PackageName)
		if err != nil {
			return ErrorResult("Clear app data failed: %v", err), ActionOutput{Success: false}, nil
		}

		msg := fmt.Sprintf("Cleared app data for %s", in.PackageName)
		return TextResult(msg), ActionOutput{
			Success:  res.Success,
			Message:  msg,
			Duration: res.Duration.String(),
		}, nil
	}
}
