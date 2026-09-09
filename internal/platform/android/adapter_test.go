package android

import (
	"testing"
)

func TestAndroidAdapter_Basic(t *testing.T) {
	adb := NewADBClient("adb")
	adapter := NewAdapter(adb, nil, nil)

	if adapter.Platform() != "android" {
		t.Errorf("expected platform 'android', got %q", adapter.Platform())
	}

	if adapter.HasCLI() {
		t.Error("expected HasCLI() to be false when cli is nil")
	}

	cli := NewAndroidCLI("android")
	adapterWithCLI := NewAdapter(adb, cli, nil)
	if !adapterWithCLI.HasCLI() {
		t.Error("expected HasCLI() to be true when cli is provided")
	}
}
