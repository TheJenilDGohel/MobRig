package mcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/TheJenilDGohel/MobRig/internal/platform"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// --- list_devices ---

type ListDevicesInput struct{}

type DeviceItemOutput struct {
	ID             string `json:"id"`
	Platform       string `json:"platform"`
	Model          string `json:"model"`
	Status         string `json:"status"`
	ConnectionType string `json:"connection_type"`
}

type ListDevicesOutput struct {
	Devices []DeviceItemOutput `json:"devices"`
	Count   int                `json:"count"`
}

func handleListDevices(registry *platform.AdapterRegistry) mcp.ToolHandlerFor[ListDevicesInput, ListDevicesOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in ListDevicesInput) (*mcp.CallToolResult, ListDevicesOutput, error) {
		devices, err := registry.ListAllDevices(ctx)
		if err != nil {
			return ErrorResult("Failed to list devices: %v", err), ListDevicesOutput{}, nil
		}

		out := ListDevicesOutput{
			Devices: make([]DeviceItemOutput, 0, len(devices)),
			Count:   len(devices),
		}

		var sb strings.Builder
		if len(devices) == 0 {
			sb.WriteString("No mobile devices connected. Please connect an Android device via USB/WiFi or start an emulator.")
		} else {
			sb.WriteString(fmt.Sprintf("Found %d connected device(s):\n", len(devices)))
			for _, d := range devices {
				out.Devices = append(out.Devices, DeviceItemOutput{
					ID:             d.ID,
					Platform:       string(d.Platform),
					Model:          d.Model,
					Status:         string(d.Status),
					ConnectionType: string(d.ConnectionType),
				})
				sb.WriteString(fmt.Sprintf("- %s (%s, %s, %s)\n", d.ID, d.Model, d.ConnectionType, d.Status))
			}
		}

		return TextResult(sb.String()), out, nil
	}
}

// --- get_device_info ---

type GetDeviceInfoInput struct {
	DeviceID string `json:"device_id" jsonschema:"The serial or identifier of the device"`
}

type GetDeviceInfoOutput struct {
	DeviceInfo *platform.DeviceInfo `json:"device_info,omitempty"`
}

func handleGetDeviceInfo(registry *platform.AdapterRegistry) mcp.ToolHandlerFor[GetDeviceInfoInput, GetDeviceInfoOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in GetDeviceInfoInput) (*mcp.CallToolResult, GetDeviceInfoOutput, error) {
		adapter := registry.FindAdapter(ctx, in.DeviceID)
		if adapter == nil {
			return ErrorResult("Device %q not found or adapter not registered", in.DeviceID), GetDeviceInfoOutput{}, nil
		}

		info, err := adapter.GetDeviceInfo(ctx, in.DeviceID)
		if err != nil {
			return ErrorResult("Failed to get device info: %v", err), GetDeviceInfoOutput{}, nil
		}

		msg := fmt.Sprintf("Device %s:\n- Model: %s (%s)\n- OS: Android %s (API %d)\n- Screen: %dx%d (density %.1f)\n- Connection: %s",
			info.ID, info.Model, info.Manufacturer, info.OSVersion, info.APILevel,
			info.ScreenWidth, info.ScreenHeight, info.ScreenDensity, info.ConnectionType)

		return TextResult(msg), GetDeviceInfoOutput{DeviceInfo: info}, nil
	}
}

// --- get_screenshot ---

type GetScreenshotInput struct {
	DeviceID string `json:"device_id" jsonschema:"The serial or identifier of the device"`
}

type GetScreenshotOutput struct {
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	SizeBytes int    `json:"size_bytes"`
	Format    string `json:"format"`
}

func handleGetScreenshot(registry *platform.AdapterRegistry) mcp.ToolHandlerFor[GetScreenshotInput, GetScreenshotOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in GetScreenshotInput) (*mcp.CallToolResult, GetScreenshotOutput, error) {
		adapter := registry.FindAdapter(ctx, in.DeviceID)
		if adapter == nil {
			return ErrorResult("Device %q not found", in.DeviceID), GetScreenshotOutput{}, nil
		}

		shot, err := adapter.GetScreenshot(ctx, in.DeviceID)
		if err != nil {
			return ErrorResult("Failed to capture screenshot: %v", err), GetScreenshotOutput{}, nil
		}

		out := GetScreenshotOutput{
			Width:     shot.Width,
			Height:    shot.Height,
			SizeBytes: shot.SizeBytes,
			Format:    shot.Format,
		}

		return ScreenshotResult([]byte(shot.Base64), shot.Width, shot.Height, in.DeviceID), out, nil
	}
}

// --- get_ui_tree ---

type GetUITreeInput struct {
	DeviceID         string `json:"device_id" jsonschema:"The serial or identifier of the device"`
	VisibleOnly      bool   `json:"visible_only,omitempty" jsonschema:"If true only returns visible and informative UI elements (recommended)"`
	TextFilter       string `json:"text_filter,omitempty" jsonschema:"Filter elements by text or content description substring"`
	DiffFromPrevious bool   `json:"diff_from_previous,omitempty" jsonschema:"Only return elements that changed since the last inspection"`
	MaxDepth         int    `json:"max_depth,omitempty" jsonschema:"Limit hierarchy depth"`
}

type GetUITreeOutput struct {
	UITree *platform.UITree `json:"ui_tree,omitempty"`
}

func handleGetUITree(registry *platform.AdapterRegistry) mcp.ToolHandlerFor[GetUITreeInput, GetUITreeOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in GetUITreeInput) (*mcp.CallToolResult, GetUITreeOutput, error) {
		adapter := registry.FindAdapter(ctx, in.DeviceID)
		if adapter == nil {
			return ErrorResult("Device %q not found", in.DeviceID), GetUITreeOutput{}, nil
		}

		tree, err := adapter.GetUITree(ctx, in.DeviceID, platform.UITreeOptions{
			VisibleOnly:      in.VisibleOnly,
			TextFilter:       in.TextFilter,
			DiffFromPrevious: in.DiffFromPrevious,
			MaxDepth:         in.MaxDepth,
		})
		if err != nil {
			return ErrorResult("Failed to read UI tree: %v", err), GetUITreeOutput{}, nil
		}

		// Produce compact text summary
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("UI Tree for %s (app: %s, elements: %d, ~%d tokens):\n",
			in.DeviceID, tree.AppPackage, len(tree.Elements), tree.TokenCount))

		for _, e := range tree.Elements {
			desc := e.Type
			if e.Text != "" {
				desc += fmt.Sprintf(" text=%q", e.Text)
			}
			if e.ContentDesc != "" {
				desc += fmt.Sprintf(" desc=%q", e.ContentDesc)
			}
			if e.ResourceID != "" {
				desc += fmt.Sprintf(" id=%q", e.ResourceID)
			}
			props := ""
			if e.Clickable {
				props += " [clickable]"
			}
			if e.Scrollable {
				props += " [scrollable]"
			}
			if e.Focused {
				props += " [focused]"
			}

			sb.WriteString(fmt.Sprintf("[%d] %s at (%d,%d)-(%d,%d)%s\n",
				e.Index, desc, e.Bounds.Left, e.Bounds.Top, e.Bounds.Right, e.Bounds.Bottom, props))
		}

		return TextResult(sb.String()), GetUITreeOutput{UITree: tree}, nil
	}
}

// --- find_element ---

type FindElementInput struct {
	DeviceID string `json:"device_id" jsonschema:"The serial or identifier of the device"`
	Text     string `json:"text,omitempty" jsonschema:"Text or content description substring to find"`
	Selector string `json:"selector,omitempty" jsonschema:"Resource ID or class name to find"`
}

type FindElementOutput struct {
	Found   bool                `json:"found"`
	Element *platform.UIElement `json:"element,omitempty"`
}

func handleFindElement(registry *platform.AdapterRegistry) mcp.ToolHandlerFor[FindElementInput, FindElementOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in FindElementInput) (*mcp.CallToolResult, FindElementOutput, error) {
		adapter := registry.FindAdapter(ctx, in.DeviceID)
		if adapter == nil {
			return ErrorResult("Device %q not found", in.DeviceID), FindElementOutput{}, nil
		}

		tree, err := adapter.GetUITree(ctx, in.DeviceID, platform.UITreeOptions{VisibleOnly: true})
		if err != nil {
			return ErrorResult("Failed to read UI tree: %v", err), FindElementOutput{}, nil
		}

		textLower := strings.ToLower(in.Text)
		selLower := strings.ToLower(in.Selector)

		for _, e := range tree.Elements {
			matchesText := textLower != "" && (strings.Contains(strings.ToLower(e.Text), textLower) || strings.Contains(strings.ToLower(e.ContentDesc), textLower))
			matchesSel := selLower != "" && (strings.Contains(strings.ToLower(e.ResourceID), selLower) || strings.EqualFold(e.Type, selLower))

			if (textLower != "" && matchesText) || (selLower != "" && matchesSel) {
				msg := fmt.Sprintf("Found element [%d] %s text=%q id=%q at center (%d, %d)",
					e.Index, e.Type, e.Text, e.ResourceID, e.Bounds.CenterX(), e.Bounds.CenterY())
				return TextResult(msg), FindElementOutput{Found: true, Element: &e}, nil
			}
		}

		return ErrorResult("Element matching text=%q selector=%q not found on screen", in.Text, in.Selector), FindElementOutput{Found: false}, nil
	}
}

// --- wait_for_element ---

type WaitForElementInput struct {
	DeviceID   string `json:"device_id" jsonschema:"The serial or identifier of the device"`
	Text       string `json:"text,omitempty" jsonschema:"Text or content description to wait for"`
	Selector   string `json:"selector,omitempty" jsonschema:"Resource ID to wait for"`
	TimeoutSec int    `json:"timeout_sec,omitempty" jsonschema:"Max seconds to wait (default 10)"`
}

type WaitForElementOutput struct {
	Found   bool                `json:"found"`
	Element *platform.UIElement `json:"element,omitempty"`
}

func handleWaitForElement(registry *platform.AdapterRegistry) mcp.ToolHandlerFor[WaitForElementInput, WaitForElementOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in WaitForElementInput) (*mcp.CallToolResult, WaitForElementOutput, error) {
		adapter := registry.FindAdapter(ctx, in.DeviceID)
		if adapter == nil {
			return ErrorResult("Device %q not found", in.DeviceID), WaitForElementOutput{}, nil
		}

		timeout := 10 * time.Second
		if in.TimeoutSec > 0 {
			timeout = time.Duration(in.TimeoutSec) * time.Second
		}

		deadline := time.Now().Add(timeout)
		textLower := strings.ToLower(in.Text)
		selLower := strings.ToLower(in.Selector)

		for time.Now().Before(deadline) {
			tree, err := adapter.GetUITree(ctx, in.DeviceID, platform.UITreeOptions{VisibleOnly: true})
			if err == nil {
				for _, e := range tree.Elements {
					matchesText := textLower != "" && (strings.Contains(strings.ToLower(e.Text), textLower) || strings.Contains(strings.ToLower(e.ContentDesc), textLower))
					matchesSel := selLower != "" && strings.Contains(strings.ToLower(e.ResourceID), selLower)

					if (textLower != "" && matchesText) || (selLower != "" && matchesSel) {
						msg := fmt.Sprintf("Found element [%d] %s text=%q at center (%d, %d)",
							e.Index, e.Type, e.Text, e.Bounds.CenterX(), e.Bounds.CenterY())
						return TextResult(msg), WaitForElementOutput{Found: true, Element: &e}, nil
					}
				}
			}
			time.Sleep(500 * time.Millisecond)
		}

		return ErrorResult("Timed out after %v waiting for element text=%q selector=%q", timeout, in.Text, in.Selector), WaitForElementOutput{Found: false}, nil
	}
}
