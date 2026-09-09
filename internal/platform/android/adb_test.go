package android

import (
	"testing"

	"github.com/TheJenilDGohel/MobRig/internal/platform"
)

func TestParseDevicesList(t *testing.T) {
	raw := `List of devices attached
emulator-5554          device product:sdk_gphone64_x86_64 model:Pixel_8 device:emu64x transport_id:1
192.168.1.100:5555     offline transport_id:2
00008110-001A2B3C      unauthorized transport_id:3
dev-usb-123            device model:Galaxy_S24 transport_id:4
`
	devices := ParseDevicesList(raw)
	if len(devices) != 4 {
		t.Fatalf("expected 4 devices, got %d", len(devices))
	}

	// 1. Emulator
	d0 := devices[0]
	if d0.ID != "emulator-5554" || d0.Status != platform.StatusConnected || d0.ConnectionType != platform.ConnectionEmulator {
		t.Errorf("d0 mismatch: %+v", d0)
	}
	if d0.Model != "Pixel 8" {
		t.Errorf("expected model 'Pixel 8', got %q", d0.Model)
	}

	// 2. WiFi Offline
	d1 := devices[1]
	if d1.ID != "192.168.1.100:5555" || d1.Status != platform.StatusOffline || d1.ConnectionType != platform.ConnectionWiFi {
		t.Errorf("d1 mismatch: %+v", d1)
	}

	// 3. Unauthorized
	d2 := devices[2]
	if d2.ID != "00008110-001A2B3C" || d2.Status != platform.StatusUnauthorized {
		t.Errorf("d2 mismatch: %+v", d2)
	}

	// 4. USB Device
	d3 := devices[3]
	if d3.ID != "dev-usb-123" || d3.Status != platform.StatusConnected || d3.ConnectionType != platform.ConnectionUSB {
		t.Errorf("d3 mismatch: %+v", d3)
	}
	if d3.Model != "Galaxy S24" {
		t.Errorf("expected model 'Galaxy S24', got %q", d3.Model)
	}
}

func TestEscapeInputText(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello world", "hello%sworld"},
		{"user@example.com", "user@example.com"},
		{"hello 'world'", `hello%s\'world\'`},
		{"a&b<c>d", `a\&b\<c\>d`},
		{"$100 (usd)", `\$100%s\(usd\)`},
	}

	for _, tt := range tests {
		got := EscapeInputText(tt.input)
		if got != tt.expected {
			t.Errorf("EscapeInputText(%q) = %q, expected %q", tt.input, got, tt.expected)
		}
	}
}

func TestParseWMSizeAndDensity(t *testing.T) {
	sizeOutput := "Physical size: 1080x2400\nOverride size: 1080x2340\n"
	w, h := parseWMSize(sizeOutput)
	if w != 1080 || h != 2400 {
		t.Errorf("expected 1080x2400, got %dx%d", w, h)
	}

	densityOutput := "Physical density: 420\n"
	density := parseWMDensity(densityOutput)
	if density != 420 {
		t.Errorf("expected density 420, got %f", density)
	}
}

func TestParseGetProp(t *testing.T) {
	raw := `[ro.product.manufacturer]: [Google]
[ro.product.model]: [Pixel 8]
[ro.build.version.release]: [14]
[ro.build.version.sdk]: [34]
`
	props := parseGetProp(raw)
	if props["ro.product.manufacturer"] != "Google" {
		t.Errorf("expected Google, got %q", props["ro.product.manufacturer"])
	}
	if props["ro.product.model"] != "Pixel 8" {
		t.Errorf("expected Pixel 8, got %q", props["ro.product.model"])
	}
	if props["ro.build.version.sdk"] != "34" {
		t.Errorf("expected 34, got %q", props["ro.build.version.sdk"])
	}
}
