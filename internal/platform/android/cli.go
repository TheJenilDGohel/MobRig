package android

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/TheJenilDGohel/MobRig/internal/platform"
)

// AndroidCLI wraps Google's official `android` command line tool.
type AndroidCLI struct {
	cliPath string
}

// NewAndroidCLI creates an AndroidCLI wrapper with verified binary path.
func NewAndroidCLI(cliPath string) *AndroidCLI {
	return &AndroidCLI{cliPath: cliPath}
}

// Path returns the path to the android executable.
func (c *AndroidCLI) Path() string {
	return c.cliPath
}

// GetLayout extracts the screen UI hierarchy using `android layout`.
func (c *AndroidCLI) GetLayout(ctx context.Context, deviceID string, opts platform.UITreeOptions) (*platform.UITree, error) {
	args := []string{"layout", "--device", deviceID, "--format", "json"}
	if opts.DiffFromPrevious {
		args = append(args, "--diff")
	}

	cmd := exec.CommandContext(ctx, c.cliPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errStr := strings.TrimSpace(stderr.String())
		if errStr == "" {
			errStr = err.Error()
		}
		return nil, fmt.Errorf("android layout: %s", errStr)
	}

	raw := stdout.Bytes()
	return ParseAndroidCLILayoutJSON(raw, opts)
}

// ParseAndroidCLILayoutJSON parses json layout output from `android layout`.
func ParseAndroidCLILayoutJSON(raw []byte, opts platform.UITreeOptions) (*platform.UITree, error) {
	// Flexible structure to parse both flat list and hierarchical object output
	var parsed struct {
		AppPackage string               `json:"app_package,omitempty"`
		ScreenName string               `json:"screen_name,omitempty"`
		Elements   []platform.UIElement `json:"elements,omitempty"`
		Nodes      []platform.UIElement `json:"nodes,omitempty"`
	}

	// Try unmarshaling into wrapper struct
	if err := json.Unmarshal(raw, &parsed); err == nil && (len(parsed.Elements) > 0 || len(parsed.Nodes) > 0) {
		elems := parsed.Elements
		if len(elems) == 0 {
			elems = parsed.Nodes
		}
		filtered := filterElements(elems, opts)
		return &platform.UITree{
			Elements:   filtered,
			AppPackage: parsed.AppPackage,
			ScreenName: parsed.ScreenName,
			Format:     "compact",
			TokenCount: estimateTokenCount(filtered),
			Timestamp:  time.Now(),
		}, nil
	}

	// Try unmarshaling directly into an array of elements
	var elemList []platform.UIElement
	if err := json.Unmarshal(raw, &elemList); err == nil {
		filtered := filterElements(elemList, opts)
		return &platform.UITree{
			Elements:   filtered,
			Format:     "compact",
			TokenCount: estimateTokenCount(filtered),
			Timestamp:  time.Now(),
		}, nil
	}

	return nil, fmt.Errorf("unrecognized json format from android layout")
}
