package setup

import (
	"context"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/TheJenilDGohel/MobRig/internal/platform/android"
)

// EnvironmentReport contains diagnostics about the host development environment.
type EnvironmentReport struct {
	ADBFound        bool   `json:"adb_found"`
	ADBPath         string `json:"adb_path,omitempty"`
	ADBVersion      string `json:"adb_version,omitempty"`
	AndroidCLIFound bool   `json:"android_cli_found"`
	AndroidCLIPath  string `json:"android_cli_path,omitempty"`
	CLIInstallCmd   string `json:"cli_install_cmd"`
	Ready           bool   `json:"ready"`
}

// CheckEnvironment inspects ADB and Google's android CLI availability.
func CheckEnvironment(customADB, customCLI string) *EnvironmentReport {
	report := &EnvironmentReport{
		CLIInstallCmd: GetCLIInstallCommand(),
	}

	// 1. Detect ADB
	if adbPath, err := android.DetectADB(customADB); err == nil {
		report.ADBFound = true
		report.ADBPath = adbPath
		report.ADBVersion = getADBVersion(adbPath)
	}

	// 2. Detect Google Android CLI
	if cliPath, ok := android.DetectAndroidCLI(customCLI); ok {
		report.AndroidCLIFound = true
		report.AndroidCLIPath = cliPath
	}

	// Ready if at least ADB is found (we can use ADB fallback mode for full automation)
	report.Ready = report.ADBFound

	return report
}

// GetCLIInstallCommand returns the official Google CLI installation script command for the current OS.
func GetCLIInstallCommand() string {
	if runtime.GOOS == "windows" {
		return "irm https://dl.google.com/android/cli/install.ps1 | iex"
	}
	return "curl -fsSL https://dl.google.com/android/cli/install.sh | bash"
}

func getADBVersion(adbPath string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, adbPath, "version").Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(out), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return ""
}
