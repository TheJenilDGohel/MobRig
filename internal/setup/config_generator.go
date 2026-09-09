package setup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// MCPConfigEntry represents the standard JSON entry for an MCP server.
type MCPConfigEntry struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

// GetClaudeDesktopConfigPath returns the platform-specific path to Claude Desktop's MCP config.
func GetClaudeDesktopConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, "Claude", "claude_desktop_config.json"), nil
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json"), nil
	default: // linux & others
		return filepath.Join(home, ".config", "Claude", "claude_desktop_config.json"), nil
	}
}

// GetCursorWorkspaceConfigPath returns the path to .cursor/mcp.json in the specified workspace directory.
func GetCursorWorkspaceConfigPath(workspaceDir string) string {
	if workspaceDir == "" {
		workspaceDir = "."
	}
	return filepath.Join(workspaceDir, ".cursor", "mcp.json")
}

// GetClaudeCodeWorkspaceConfigPath returns the path to .mcp.json in the specified workspace directory.
func GetClaudeCodeWorkspaceConfigPath(workspaceDir string) string {
	if workspaceDir == "" {
		workspaceDir = "."
	}
	return filepath.Join(workspaceDir, ".mcp.json")
}

// MergeMCPConfig safely merges the "mobrig" server configuration into existing JSON content.
// It preserves all other keys and servers present in the existing configuration.
func MergeMCPConfig(existing []byte, binaryPath string) ([]byte, error) {
	var root map[string]any

	trimmed := strings.TrimSpace(string(existing))
	if trimmed == "" {
		root = make(map[string]any)
	} else {
		if err := json.Unmarshal(existing, &root); err != nil {
			return nil, fmt.Errorf("failed to parse existing JSON: %w", err)
		}
	}

	serversRaw, exists := root["mcpServers"]
	var servers map[string]any
	if exists {
		var ok bool
		servers, ok = serversRaw.(map[string]any)
		if !ok {
			servers = make(map[string]any)
		}
	} else {
		servers = make(map[string]any)
	}

	servers["mobrig"] = map[string]any{
		"command": binaryPath,
		"args":    []any{"mcp"},
	}

	root["mcpServers"] = servers

	return json.MarshalIndent(root, "", "  ")
}

// WriteMCPConfig writes or merges the "mobrig" MCP server configuration into the given file path.
func WriteMCPConfig(filePath string, binaryPath string) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	var existingContent []byte
	if data, err := os.ReadFile(filePath); err == nil {
		existingContent = data
	}

	merged, err := MergeMCPConfig(existingContent, binaryPath)
	if err != nil {
		return fmt.Errorf("failed to merge MCP config: %w", err)
	}

	if err := os.WriteFile(filePath, merged, 0644); err != nil {
		return fmt.Errorf("failed to write config to %s: %w", filePath, err)
	}

	return nil
}

// GenerateAllConfigs auto-configures MCP files across Claude Desktop, Cursor, and Claude Code.
// Returns list of successfully configured file paths.
func GenerateAllConfigs(workspaceDir string, binaryPath string) ([]string, []error) {
	var configured []string
	var errs []error

	// 1. Claude Desktop (if directory exists or can be written)
	if cdPath, err := GetClaudeDesktopConfigPath(); err == nil {
		if err := WriteMCPConfig(cdPath, binaryPath); err == nil {
			configured = append(configured, cdPath)
		} else {
			errs = append(errs, err)
		}
	}

	// 2. Cursor workspace (.cursor/mcp.json)
	cursorPath := GetCursorWorkspaceConfigPath(workspaceDir)
	if err := WriteMCPConfig(cursorPath, binaryPath); err == nil {
		configured = append(configured, cursorPath)
	} else {
		errs = append(errs, err)
	}

	// 3. Claude Code workspace (.mcp.json)
	ccPath := GetClaudeCodeWorkspaceConfigPath(workspaceDir)
	if err := WriteMCPConfig(ccPath, binaryPath); err == nil {
		configured = append(configured, ccPath)
	} else {
		errs = append(errs, err)
	}

	return configured, errs
}
