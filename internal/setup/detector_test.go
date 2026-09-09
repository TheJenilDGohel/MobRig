package setup

import (
	"runtime"
	"strings"
	"testing"
)

func TestCheckEnvironment(t *testing.T) {
	report := CheckEnvironment("", "")

	if report.CLIInstallCmd == "" {
		t.Error("expected non-empty CLI install command")
	}

	if runtime.GOOS == "windows" {
		if !strings.Contains(report.CLIInstallCmd, "install.ps1") {
			t.Errorf("expected PowerShell script for Windows, got: %s", report.CLIInstallCmd)
		}
	} else {
		if !strings.Contains(report.CLIInstallCmd, "install.sh") {
			t.Errorf("expected bash script for non-Windows, got: %s", report.CLIInstallCmd)
		}
	}
}

func TestGetCLIInstallCommand(t *testing.T) {
	cmd := GetCLIInstallCommand()
	if cmd == "" {
		t.Fatal("expected install command string")
	}
}
