package android

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/TheJenilDGohel/MobRig/internal/platform"
)

// ADBClient wraps the adb executable to provide low-level Android operations.
type ADBClient struct {
	adbPath string
}

// NewADBClient creates an ADBClient with the verified path to adb.
func NewADBClient(adbPath string) *ADBClient {
	return &ADBClient{adbPath: adbPath}
}

// Path returns the adb executable path.
func (c *ADBClient) Path() string {
	return c.adbPath
}

// run executes an adb command with context and returns combined stdout.
func (c *ADBClient) run(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, c.adbPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		errStr := strings.TrimSpace(stderr.String())
		if errStr == "" {
			errStr = strings.TrimSpace(stdout.String())
		}
		if errStr == "" {
			errStr = err.Error()
		}
		return nil, fmt.Errorf("adb %s: %s", strings.Join(args, " "), errStr)
	}
	return stdout.Bytes(), nil
}

// Devices discovers all connected devices by parsing `adb devices -l`.
func (c *ADBClient) Devices(ctx context.Context) ([]platform.DeviceInfo, error) {
	out, err := c.run(ctx, "devices", "-l")
	if err != nil {
		return nil, err
	}
	return ParseDevicesList(string(out)), nil
}

// ParseDevicesList parses raw `adb devices -l` output into a slice of DeviceInfo.
func ParseDevicesList(output string) []platform.DeviceInfo {
	var devices []platform.DeviceInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "List of devices") || strings.HasPrefix(line, "*") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		id := fields[0]
		rawStatus := fields[1]

		var status platform.DeviceStatus
		switch rawStatus {
		case "device":
			status = platform.StatusConnected
		case "offline":
			status = platform.StatusOffline
		case "unauthorized":
			status = platform.StatusUnauthorized
		default:
			status = platform.StatusDisconnected
		}

		// Determine connection type
		var connType platform.ConnectionType
		if strings.HasPrefix(id, "emulator-") {
			connType = platform.ConnectionEmulator
		} else if strings.Contains(id, ":") {
			connType = platform.ConnectionWiFi
		} else {
			connType = platform.ConnectionUSB
		}

		// Parse key:value metadata (product:..., model:..., transport_id:...)
		model := ""
		transportID := ""
		for _, f := range fields[2:] {
			parts := strings.SplitN(f, ":", 2)
			if len(parts) == 2 {
				switch parts[0] {
				case "model":
					model = strings.ReplaceAll(parts[1], "_", " ")
				case "transport_id":
					transportID = parts[1]
				}
			}
		}

		if model == "" {
			model = id
		}

		devices = append(devices, platform.DeviceInfo{
			ID:             id,
			Platform:       platform.PlatformAndroid,
			Model:          model,
			Manufacturer:   "Android",
			Status:         status,
			ConnectionType: connType,
			TransportID:    transportID,
		})
	}

	return devices
}

// GetDeviceInfo queries detailed properties from the device.
func (c *ADBClient) GetDeviceInfo(ctx context.Context, id string) (*platform.DeviceInfo, error) {
	out, err := c.run(ctx, "-s", id, "shell", "getprop")
	if err != nil {
		return nil, fmt.Errorf("get device properties: %w", err)
	}

	props := parseGetProp(string(out))

	info := &platform.DeviceInfo{
		ID:             id,
		Platform:       platform.PlatformAndroid,
		Model:          props["ro.product.model"],
		Manufacturer:   props["ro.product.manufacturer"],
		OSVersion:      props["ro.build.version.release"],
		Status:         platform.StatusConnected,
		ConnectionType: platform.ConnectionUSB,
	}

	if strings.HasPrefix(id, "emulator-") {
		info.ConnectionType = platform.ConnectionEmulator
	} else if strings.Contains(id, ":") {
		info.ConnectionType = platform.ConnectionWiFi
	}

	if apiStr, ok := props["ro.build.version.sdk"]; ok {
		if api, err := strconv.Atoi(apiStr); err == nil {
			info.APILevel = api
		}
	}

	// Query screen dimensions via `wm size`
	if wmSize, err := c.run(ctx, "-s", id, "shell", "wm", "size"); err == nil {
		w, h := parseWMSize(string(wmSize))
		info.ScreenWidth = w
		info.ScreenHeight = h
	}

	// Query screen density via `wm density`
	if wmDensity, err := c.run(ctx, "-s", id, "shell", "wm", "density"); err == nil {
		info.ScreenDensity = parseWMDensity(string(wmDensity))
	}

	return info, nil
}

// Screenshot captures the screen via `adb exec-out screencap -p`.
func (c *ADBClient) Screenshot(ctx context.Context, id string) (*platform.Screenshot, error) {
	start := time.Now()
	data, err := c.run(ctx, "-s", id, "exec-out", "screencap", "-p")
	if err != nil {
		return nil, fmt.Errorf("screencap: %w", err)
	}

	if len(data) < 24 {
		return nil, fmt.Errorf("screencap returned invalid data: too short (%d bytes)", len(data))
	}

	// Check PNG magic bytes: 0x89 'P' 'N' 'G' 0x0D 0x0A 0x1A 0x0A
	pngMagic := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
	if !bytes.HasPrefix(data, pngMagic) {
		return nil, fmt.Errorf("screencap returned non-PNG data")
	}

	// Read width and height from PNG IHDR chunk (bytes 16..24)
	width := int(binary.BigEndian.Uint32(data[16:20]))
	height := int(binary.BigEndian.Uint32(data[20:24]))

	b64 := base64.StdEncoding.EncodeToString(data)

	return &platform.Screenshot{
		Data:      data,
		Base64:    b64,
		Width:     width,
		Height:    height,
		Format:    "png",
		Timestamp: start,
		SizeBytes: len(data),
	}, nil
}

// Tap performs a single tap at (x, y).
func (c *ADBClient) Tap(ctx context.Context, id string, x, y int) (*platform.ActionResult, error) {
	start := time.Now()
	_, err := c.run(ctx, "-s", id, "shell", "input", "tap", strconv.Itoa(x), strconv.Itoa(y))
	duration := time.Since(start)
	if err != nil {
		return &platform.ActionResult{
			Success:  false,
			Duration: duration,
			Error:    err.Error(),
			Message:  fmt.Sprintf("Failed to tap at (%d, %d)", x, y),
		}, err
	}
	return &platform.ActionResult{
		Success:  true,
		Duration: duration,
		Message:  fmt.Sprintf("Tapped at (%d, %d)", x, y),
	}, nil
}

// DoubleTap performs a double tap at (x, y) with a 100ms interval.
func (c *ADBClient) DoubleTap(ctx context.Context, id string, x, y int) (*platform.ActionResult, error) {
	start := time.Now()
	xStr, yStr := strconv.Itoa(x), strconv.Itoa(y)
	cmdStr := fmt.Sprintf("input tap %s %s && sleep 0.1 && input tap %s %s", xStr, yStr, xStr, yStr)
	_, err := c.run(ctx, "-s", id, "shell", cmdStr)
	duration := time.Since(start)
	if err != nil {
		return &platform.ActionResult{
			Success:  false,
			Duration: duration,
			Error:    err.Error(),
			Message:  fmt.Sprintf("Failed to double-tap at (%d, %d)", x, y),
		}, err
	}
	return &platform.ActionResult{
		Success:  true,
		Duration: duration,
		Message:  fmt.Sprintf("Double-tapped at (%d, %d)", x, y),
	}, nil
}

// LongPress performs a long press by holding a tap as a zero-distance swipe.
func (c *ADBClient) LongPress(ctx context.Context, id string, x, y int, durationMs int) (*platform.ActionResult, error) {
	if durationMs <= 0 {
		durationMs = 1000 // default 1 second
	}
	start := time.Now()
	xStr, yStr, dStr := strconv.Itoa(x), strconv.Itoa(y), strconv.Itoa(durationMs)
	_, err := c.run(ctx, "-s", id, "shell", "input", "swipe", xStr, yStr, xStr, yStr, dStr)
	duration := time.Since(start)
	if err != nil {
		return &platform.ActionResult{
			Success:  false,
			Duration: duration,
			Error:    err.Error(),
			Message:  fmt.Sprintf("Failed to long-press at (%d, %d) for %dms", x, y, durationMs),
		}, err
	}
	return &platform.ActionResult{
		Success:  true,
		Duration: duration,
		Message:  fmt.Sprintf("Long-pressed at (%d, %d) for %dms", x, y, durationMs),
	}, nil
}

// Swipe performs a swipe gesture from (x1, y1) to (x2, y2).
func (c *ADBClient) Swipe(ctx context.Context, id string, x1, y1, x2, y2 int, durationMs int) (*platform.ActionResult, error) {
	if durationMs <= 0 {
		durationMs = 300 // default 300ms
	}
	start := time.Now()
	args := []string{
		"-s", id, "shell", "input", "swipe",
		strconv.Itoa(x1), strconv.Itoa(y1),
		strconv.Itoa(x2), strconv.Itoa(y2),
		strconv.Itoa(durationMs),
	}
	_, err := c.run(ctx, args...)
	duration := time.Since(start)
	if err != nil {
		return &platform.ActionResult{
			Success:  false,
			Duration: duration,
			Error:    err.Error(),
			Message:  fmt.Sprintf("Failed to swipe from (%d, %d) to (%d, %d)", x1, y1, x2, y2),
		}, err
	}
	return &platform.ActionResult{
		Success:  true,
		Duration: duration,
		Message:  fmt.Sprintf("Swiped from (%d, %d) to (%d, %d)", x1, y1, x2, y2),
	}, nil
}

// TypeText types text into the currently focused input.
// Spaces are encoded as %s per Android input text convention.
func (c *ADBClient) TypeText(ctx context.Context, id string, text string, submit bool) (*platform.ActionResult, error) {
	start := time.Now()
	escaped := EscapeInputText(text)

	var cmdStr string
	if submit {
		cmdStr = fmt.Sprintf("input text '%s' && input keyevent 66", escaped)
	} else {
		cmdStr = fmt.Sprintf("input text '%s'", escaped)
	}

	_, err := c.run(ctx, "-s", id, "shell", cmdStr)
	duration := time.Since(start)
	if err != nil {
		return &platform.ActionResult{
			Success:  false,
			Duration: duration,
			Error:    err.Error(),
			Message:  fmt.Sprintf("Failed to type text %q", text),
		}, err
	}
	msg := fmt.Sprintf("Typed text %q", text)
	if submit {
		msg += " and submitted"
	}
	return &platform.ActionResult{
		Success:  true,
		Duration: duration,
		Message:  msg,
	}, nil
}

// EscapeInputText escapes a string for Android's `input text` command.
func EscapeInputText(text string) string {
	// Android input text treats %s as space
	var sb strings.Builder
	for _, ch := range text {
		switch ch {
		case ' ':
			sb.WriteString("%s")
		case '\'':
			sb.WriteString(`\'`)
		case '"':
			sb.WriteString(`\"`)
		case '&':
			sb.WriteString(`\&`)
		case '<':
			sb.WriteString(`\<`)
		case '>':
			sb.WriteString(`\>`)
		case ';':
			sb.WriteString(`\;`)
		case '|':
			sb.WriteString(`\|`)
		case '(':
			sb.WriteString(`\(`)
		case ')':
			sb.WriteString(`\)`)
		case '$':
			sb.WriteString(`\$`)
		case '\\':
			sb.WriteString(`\\`)
		default:
			sb.WriteRune(ch)
		}
	}
	return sb.String()
}

// PressButton presses a hardware or system key by key code.
func (c *ADBClient) PressButton(ctx context.Context, id string, keyCode int) (*platform.ActionResult, error) {
	start := time.Now()
	_, err := c.run(ctx, "-s", id, "shell", "input", "keyevent", strconv.Itoa(keyCode))
	duration := time.Since(start)
	if err != nil {
		return &platform.ActionResult{
			Success:  false,
			Duration: duration,
			Error:    err.Error(),
			Message:  fmt.Sprintf("Failed to press key %d", keyCode),
		}, err
	}
	return &platform.ActionResult{
		Success:  true,
		Duration: duration,
		Message:  fmt.Sprintf("Pressed key %d", keyCode),
	}, nil
}

// GoHome presses the Android home key (key code 3).
func (c *ADBClient) GoHome(ctx context.Context, id string) (*platform.ActionResult, error) {
	return c.PressButton(ctx, id, 3)
}

// LaunchApp launches an app by package name.
func (c *ADBClient) LaunchApp(ctx context.Context, id string, packageName string) (*platform.ActionResult, error) {
	start := time.Now()
	_, err := c.run(ctx, "-s", id, "shell", "monkey", "-p", packageName, "-c", "android.intent.category.LAUNCHER", "1")
	duration := time.Since(start)
	if err != nil {
		return &platform.ActionResult{
			Success:  false,
			Duration: duration,
			Error:    err.Error(),
			Message:  fmt.Sprintf("Failed to launch app %s", packageName),
		}, err
	}
	return &platform.ActionResult{
		Success:  true,
		Duration: duration,
		Message:  fmt.Sprintf("Launched app %s", packageName),
	}, nil
}

// TerminateApp force-stops an app by package name.
func (c *ADBClient) TerminateApp(ctx context.Context, id string, packageName string) (*platform.ActionResult, error) {
	start := time.Now()
	_, err := c.run(ctx, "-s", id, "shell", "am", "force-stop", packageName)
	duration := time.Since(start)
	if err != nil {
		return &platform.ActionResult{
			Success:  false,
			Duration: duration,
			Error:    err.Error(),
			Message:  fmt.Sprintf("Failed to terminate app %s", packageName),
		}, err
	}
	return &platform.ActionResult{
		Success:  true,
		Duration: duration,
		Message:  fmt.Sprintf("Terminated app %s", packageName),
	}, nil
}

// InstallApp installs an APK from a local path.
func (c *ADBClient) InstallApp(ctx context.Context, id string, appPath string) (*platform.ActionResult, error) {
	start := time.Now()
	_, err := c.run(ctx, "-s", id, "install", "-r", appPath)
	duration := time.Since(start)
	if err != nil {
		return &platform.ActionResult{
			Success:  false,
			Duration: duration,
			Error:    err.Error(),
			Message:  fmt.Sprintf("Failed to install app from %s", appPath),
		}, err
	}
	return &platform.ActionResult{
		Success:  true,
		Duration: duration,
		Message:  fmt.Sprintf("Installed app from %s", appPath),
	}, nil
}

// UninstallApp uninstalls an app package.
func (c *ADBClient) UninstallApp(ctx context.Context, id string, packageName string) (*platform.ActionResult, error) {
	start := time.Now()
	_, err := c.run(ctx, "-s", id, "uninstall", packageName)
	duration := time.Since(start)
	if err != nil {
		return &platform.ActionResult{
			Success:  false,
			Duration: duration,
			Error:    err.Error(),
			Message:  fmt.Sprintf("Failed to uninstall app %s", packageName),
		}, err
	}
	return &platform.ActionResult{
		Success:  true,
		Duration: duration,
		Message:  fmt.Sprintf("Uninstalled app %s", packageName),
	}, nil
}

// ClearAppData clears all stored data for an app.
func (c *ADBClient) ClearAppData(ctx context.Context, id string, packageName string) (*platform.ActionResult, error) {
	start := time.Now()
	_, err := c.run(ctx, "-s", id, "shell", "pm", "clear", packageName)
	duration := time.Since(start)
	if err != nil {
		return &platform.ActionResult{
			Success:  false,
			Duration: duration,
			Error:    err.Error(),
			Message:  fmt.Sprintf("Failed to clear data for %s", packageName),
		}, err
	}
	return &platform.ActionResult{
		Success:  true,
		Duration: duration,
		Message:  fmt.Sprintf("Cleared app data for %s", packageName),
	}, nil
}

// OpenURL opens a URL on the device via the default handler.
func (c *ADBClient) OpenURL(ctx context.Context, id string, rawURL string) (*platform.ActionResult, error) {
	start := time.Now()
	_, err := c.run(ctx, "-s", id, "shell", "am", "start", "-a", "android.intent.action.VIEW", "-d", rawURL)
	duration := time.Since(start)
	if err != nil {
		return &platform.ActionResult{
			Success:  false,
			Duration: duration,
			Error:    err.Error(),
			Message:  fmt.Sprintf("Failed to open URL %s", rawURL),
		}, err
	}
	return &platform.ActionResult{
		Success:  true,
		Duration: duration,
		Message:  fmt.Sprintf("Opened URL %s", rawURL),
	}, nil
}

// DumpUIHierarchy dumps the uiautomator XML string from the device.
// Attempts direct /dev/tty dump first for speed, falling back to temp file if necessary.
func (c *ADBClient) DumpUIHierarchy(ctx context.Context, id string) (string, error) {
	// 1. Try dumping directly to /dev/tty
	out, err := c.run(ctx, "-s", id, "exec-out", "uiautomator", "dump", "/dev/tty")
	xmlStr := string(out)
	if err == nil && strings.Contains(xmlStr, "<hierarchy") {
		return xmlStr, nil
	}

	// 2. Fallback: dump to /data/local/tmp/dump.xml and cat it
	tmpFile := "/data/local/tmp/mobrig_uidump.xml"
	if _, err := c.run(ctx, "-s", id, "shell", "uiautomator", "dump", tmpFile); err != nil {
		return "", fmt.Errorf("dump hierarchy to tmp file: %w", err)
	}

	catOut, err := c.run(ctx, "-s", id, "exec-out", "cat", tmpFile)
	if err != nil {
		return "", fmt.Errorf("read dump xml: %w", err)
	}

	return string(catOut), nil
}

// --- Helpers ---

func parseGetProp(output string) map[string]string {
	props := make(map[string]string)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[") && strings.Contains(line, "]: [") {
			parts := strings.SplitN(line, "]: [", 2)
			if len(parts) == 2 {
				key := strings.TrimPrefix(parts[0], "[")
				val := strings.TrimSuffix(parts[1], "]")
				props[key] = val
			}
		}
	}
	return props
}

func parseWMSize(output string) (int, int) {
	// e.g. "Physical size: 1080x2400"
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "size:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				dim := strings.TrimSpace(parts[len(parts)-1])
				wh := strings.Split(dim, "x")
				if len(wh) == 2 {
					w, _ := strconv.Atoi(strings.TrimSpace(wh[0]))
					h, _ := strconv.Atoi(strings.TrimSpace(wh[1]))
					return w, h
				}
			}
		}
	}
	return 0, 0
}

func parseWMDensity(output string) float64 {
	// e.g. "Physical density: 420"
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "density:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				denStr := strings.TrimSpace(parts[len(parts)-1])
				if d, err := strconv.ParseFloat(denStr, 64); err == nil {
					return d
				}
			}
		}
	}
	return 0
}
