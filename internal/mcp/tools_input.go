package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/TheJenilDGohel/MobRig/internal/platform"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// --- tap ---

type TapInput struct {
	DeviceID string `json:"device_id" jsonschema:"The serial or identifier of the device"`
	Index    *int   `json:"index,omitempty" jsonschema:"Element index from get_ui_tree to tap (preferred)"`
	X        int    `json:"x,omitempty" jsonschema:"X coordinate on screen (if not using index)"`
	Y        int    `json:"y,omitempty" jsonschema:"Y coordinate on screen (if not using index)"`
}

type ActionOutput struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Duration string `json:"duration,omitempty"`
}

func handleTap(registry *platform.AdapterRegistry) mcp.ToolHandlerFor[TapInput, ActionOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in TapInput) (*mcp.CallToolResult, ActionOutput, error) {
		adapter := registry.FindAdapter(ctx, in.DeviceID)
		if adapter == nil {
			return ErrorResult("Device %q not found", in.DeviceID), ActionOutput{Success: false}, nil
		}

		x, y, desc, err := resolveCoordinates(ctx, adapter, in.DeviceID, in.Index, in.X, in.Y)
		if err != nil {
			return ErrorResult("Could not resolve tap location: %v", err), ActionOutput{Success: false}, nil
		}

		res, err := adapter.Tap(ctx, in.DeviceID, x, y)
		if err != nil {
			return ErrorResult("Tap failed: %v", err), ActionOutput{Success: false}, nil
		}

		msg := fmt.Sprintf("Tapped %s at (%d, %d)", desc, x, y)
		return TextResult(msg), ActionOutput{
			Success:  res.Success,
			Message:  msg,
			Duration: res.Duration.String(),
		}, nil
	}
}

// --- double_tap ---

type DoubleTapInput struct {
	DeviceID string `json:"device_id" jsonschema:"The serial or identifier of the device"`
	Index    *int   `json:"index,omitempty" jsonschema:"Element index from get_ui_tree to double-tap"`
	X        int    `json:"x,omitempty" jsonschema:"X coordinate"`
	Y        int    `json:"y,omitempty" jsonschema:"Y coordinate"`
}

func handleDoubleTap(registry *platform.AdapterRegistry) mcp.ToolHandlerFor[DoubleTapInput, ActionOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in DoubleTapInput) (*mcp.CallToolResult, ActionOutput, error) {
		adapter := registry.FindAdapter(ctx, in.DeviceID)
		if adapter == nil {
			return ErrorResult("Device %q not found", in.DeviceID), ActionOutput{Success: false}, nil
		}

		x, y, desc, err := resolveCoordinates(ctx, adapter, in.DeviceID, in.Index, in.X, in.Y)
		if err != nil {
			return ErrorResult("Could not resolve tap location: %v", err), ActionOutput{Success: false}, nil
		}

		res, err := adapter.DoubleTap(ctx, in.DeviceID, x, y)
		if err != nil {
			return ErrorResult("Double-tap failed: %v", err), ActionOutput{Success: false}, nil
		}

		msg := fmt.Sprintf("Double-tapped %s at (%d, %d)", desc, x, y)
		return TextResult(msg), ActionOutput{
			Success:  res.Success,
			Message:  msg,
			Duration: res.Duration.String(),
		}, nil
	}
}

// --- long_press ---

type LongPressInput struct {
	DeviceID   string `json:"device_id" jsonschema:"The serial or identifier of the device"`
	Index      *int   `json:"index,omitempty" jsonschema:"Element index from get_ui_tree to long press"`
	X          int    `json:"x,omitempty" jsonschema:"X coordinate"`
	Y          int    `json:"y,omitempty" jsonschema:"Y coordinate"`
	DurationMs int    `json:"duration_ms,omitempty" jsonschema:"Press duration in milliseconds (default 1000)"`
}

func handleLongPress(registry *platform.AdapterRegistry) mcp.ToolHandlerFor[LongPressInput, ActionOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in LongPressInput) (*mcp.CallToolResult, ActionOutput, error) {
		adapter := registry.FindAdapter(ctx, in.DeviceID)
		if adapter == nil {
			return ErrorResult("Device %q not found", in.DeviceID), ActionOutput{Success: false}, nil
		}

		x, y, desc, err := resolveCoordinates(ctx, adapter, in.DeviceID, in.Index, in.X, in.Y)
		if err != nil {
			return ErrorResult("Could not resolve long-press location: %v", err), ActionOutput{Success: false}, nil
		}

		dur := in.DurationMs
		if dur <= 0 {
			dur = 1000
		}

		res, err := adapter.LongPress(ctx, in.DeviceID, x, y, dur)
		if err != nil {
			return ErrorResult("Long-press failed: %v", err), ActionOutput{Success: false}, nil
		}

		msg := fmt.Sprintf("Long-pressed %s at (%d, %d) for %dms", desc, x, y, dur)
		return TextResult(msg), ActionOutput{
			Success:  res.Success,
			Message:  msg,
			Duration: res.Duration.String(),
		}, nil
	}
}

// --- swipe ---

type SwipeInput struct {
	DeviceID   string `json:"device_id" jsonschema:"The serial or identifier of the device"`
	Direction  string `json:"direction,omitempty" jsonschema:"Swipe direction: up, down, left, right (e.g. 'up' scrolls down)"`
	X1         int    `json:"x1,omitempty" jsonschema:"Starting X coordinate"`
	Y1         int    `json:"y1,omitempty" jsonschema:"Starting Y coordinate"`
	X2         int    `json:"x2,omitempty" jsonschema:"Ending X coordinate"`
	Y2         int    `json:"y2,omitempty" jsonschema:"Ending Y coordinate"`
	DurationMs int    `json:"duration_ms,omitempty" jsonschema:"Swipe duration in milliseconds (default 300)"`
}

func handleSwipe(registry *platform.AdapterRegistry) mcp.ToolHandlerFor[SwipeInput, ActionOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in SwipeInput) (*mcp.CallToolResult, ActionOutput, error) {
		adapter := registry.FindAdapter(ctx, in.DeviceID)
		if adapter == nil {
			return ErrorResult("Device %q not found", in.DeviceID), ActionOutput{Success: false}, nil
		}

		x1, y1, x2, y2 := in.X1, in.Y1, in.X2, in.Y2
		dur := in.DurationMs
		if dur <= 0 {
			dur = 300
		}

		// If direction is provided instead of coordinates, calculate based on screen size
		if in.Direction != "" {
			info, err := adapter.GetDeviceInfo(ctx, in.DeviceID)
			width, height := 1080, 2400
			if err == nil && info.ScreenWidth > 0 && info.ScreenHeight > 0 {
				width = info.ScreenWidth
				height = info.ScreenHeight
			}

			midX := width / 2
			midY := height / 2

			switch strings.ToLower(in.Direction) {
			case "up": // Scrolls content down
				x1, y1 = midX, int(float64(height)*0.75)
				x2, y2 = midX, int(float64(height)*0.25)
			case "down": // Scrolls content up
				x1, y1 = midX, int(float64(height)*0.25)
				x2, y2 = midX, int(float64(height)*0.75)
			case "left": // Swipes to the next page on right
				x1, y1 = int(float64(width)*0.8), midY
				x2, y2 = int(float64(width)*0.2), midY
			case "right": // Swipes to the previous page on left
				x1, y1 = int(float64(width)*0.2), midY
				x2, y2 = int(float64(width)*0.8), midY
			default:
				return ErrorResult("Unknown direction %q: use 'up', 'down', 'left', or 'right'", in.Direction), ActionOutput{Success: false}, nil
			}
		}

		res, err := adapter.Swipe(ctx, in.DeviceID, x1, y1, x2, y2, dur)
		if err != nil {
			return ErrorResult("Swipe failed: %v", err), ActionOutput{Success: false}, nil
		}

		msg := fmt.Sprintf("Swiped from (%d, %d) to (%d, %d)", x1, y1, x2, y2)
		if in.Direction != "" {
			msg = fmt.Sprintf("Swiped %s (%d, %d)->(%d, %d)", in.Direction, x1, y1, x2, y2)
		}

		return TextResult(msg), ActionOutput{
			Success:  res.Success,
			Message:  msg,
			Duration: res.Duration.String(),
		}, nil
	}
}

// --- type_text ---

type TypeTextInput struct {
	DeviceID string `json:"device_id" jsonschema:"The serial or identifier of the device"`
	Text     string `json:"text" jsonschema:"The text string to type"`
	Submit   bool   `json:"submit,omitempty" jsonschema:"If true presses Enter/Submit after typing"`
	Index    *int   `json:"index,omitempty" jsonschema:"Optional input field index to tap before typing"`
}

func handleTypeText(registry *platform.AdapterRegistry) mcp.ToolHandlerFor[TypeTextInput, ActionOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in TypeTextInput) (*mcp.CallToolResult, ActionOutput, error) {
		adapter := registry.FindAdapter(ctx, in.DeviceID)
		if adapter == nil {
			return ErrorResult("Device %q not found", in.DeviceID), ActionOutput{Success: false}, nil
		}

		// If an element index is specified, tap it first to focus it
		if in.Index != nil {
			x, y, _, err := resolveCoordinates(ctx, adapter, in.DeviceID, in.Index, 0, 0)
			if err == nil {
				_, _ = adapter.Tap(ctx, in.DeviceID, x, y)
			}
		}

		res, err := adapter.TypeText(ctx, in.DeviceID, in.Text, in.Submit)
		if err != nil {
			return ErrorResult("Type text failed: %v", err), ActionOutput{Success: false}, nil
		}

		msg := fmt.Sprintf("Typed text %q", in.Text)
		if in.Submit {
			msg += " and pressed Enter"
		}

		return TextResult(msg), ActionOutput{
			Success:  res.Success,
			Message:  msg,
			Duration: res.Duration.String(),
		}, nil
	}
}

// --- press_button ---

type PressButtonInput struct {
	DeviceID string `json:"device_id" jsonschema:"The serial or identifier of the device"`
	KeyCode  int    `json:"key_code" jsonschema:"Android keycode: 3=Home, 4=Back, 24=VolUp, 25=VolDown, 26=Power, 66=Enter"`
}

func handlePressButton(registry *platform.AdapterRegistry) mcp.ToolHandlerFor[PressButtonInput, ActionOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in PressButtonInput) (*mcp.CallToolResult, ActionOutput, error) {
		adapter := registry.FindAdapter(ctx, in.DeviceID)
		if adapter == nil {
			return ErrorResult("Device %q not found", in.DeviceID), ActionOutput{Success: false}, nil
		}

		res, err := adapter.PressButton(ctx, in.DeviceID, in.KeyCode)
		if err != nil {
			return ErrorResult("Press button failed: %v", err), ActionOutput{Success: false}, nil
		}

		keyName := formatKeyName(in.KeyCode)
		msg := fmt.Sprintf("Pressed key %d (%s)", in.KeyCode, keyName)

		return TextResult(msg), ActionOutput{
			Success:  res.Success,
			Message:  msg,
			Duration: res.Duration.String(),
		}, nil
	}
}

// --- go_home ---

type GoHomeInput struct {
	DeviceID string `json:"device_id" jsonschema:"The serial or identifier of the device"`
}

func handleGoHome(registry *platform.AdapterRegistry) mcp.ToolHandlerFor[GoHomeInput, ActionOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, in GoHomeInput) (*mcp.CallToolResult, ActionOutput, error) {
		adapter := registry.FindAdapter(ctx, in.DeviceID)
		if adapter == nil {
			return ErrorResult("Device %q not found", in.DeviceID), ActionOutput{Success: false}, nil
		}

		res, err := adapter.GoHome(ctx, in.DeviceID)
		if err != nil {
			return ErrorResult("Go home failed: %v", err), ActionOutput{Success: false}, nil
		}

		msg := "Pressed Home button"
		return TextResult(msg), ActionOutput{
			Success:  res.Success,
			Message:  msg,
			Duration: res.Duration.String(),
		}, nil
	}
}

// --- Helpers ---

func resolveCoordinates(ctx context.Context, adapter platform.PlatformAdapter, deviceID string, index *int, x, y int) (int, int, string, error) {
	if index != nil {
		tree, err := adapter.GetUITree(ctx, deviceID, platform.UITreeOptions{})
		if err != nil {
			return 0, 0, "", fmt.Errorf("read UI tree to resolve index %d: %w", *index, err)
		}
		if *index < 0 || *index >= len(tree.Elements) {
			return 0, 0, "", fmt.Errorf("element index %d out of bounds (found %d elements)", *index, len(tree.Elements))
		}
		elem := tree.Elements[*index]
		label := elem.Type
		if elem.Text != "" {
			label += fmt.Sprintf(" %q", elem.Text)
		}
		return elem.Bounds.CenterX(), elem.Bounds.CenterY(), label, nil
	}

	if x > 0 || y > 0 {
		return x, y, fmt.Sprintf("coordinates (%d, %d)", x, y), nil
	}

	return 0, 0, "", fmt.Errorf("either index or (x, y) coordinates must be provided")
}

func formatKeyName(code int) string {
	switch code {
	case 3:
		return "HOME"
	case 4:
		return "BACK"
	case 24:
		return "VOLUME_UP"
	case 25:
		return "VOLUME_DOWN"
	case 26:
		return "POWER"
	case 66:
		return "ENTER"
	case 82:
		return "MENU"
	case 187:
		return "APP_SWITCH"
	default:
		return "KEYCODE"
	}
}
