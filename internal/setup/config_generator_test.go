package setup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMergeMCPConfig_Empty(t *testing.T) {
	merged, err := MergeMCPConfig(nil, "/path/to/mobrig")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var root map[string]any
	if err := json.Unmarshal(merged, &root); err != nil {
		t.Fatalf("failed to parse merged JSON: %v", err)
	}

	servers, ok := root["mcpServers"].(map[string]any)
	if !ok {
		t.Fatal("expected mcpServers object")
	}

	mobrig, ok := servers["mobrig"].(map[string]any)
	if !ok {
		t.Fatal("expected mobrig server entry")
	}

	if mobrig["command"] != "/path/to/mobrig" {
		t.Errorf("expected command '/path/to/mobrig', got %v", mobrig["command"])
	}
}

func TestMergeMCPConfig_PreserveExisting(t *testing.T) {
	existing := []byte(`{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem"]
    },
    "mobrig": {
      "command": "/old/path/to/mobrig",
      "args": ["old"]
    }
  }
}`)

	merged, err := MergeMCPConfig(existing, "/new/path/to/mobrig")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var root map[string]any
	if err := json.Unmarshal(merged, &root); err != nil {
		t.Fatalf("failed to parse merged JSON: %v", err)
	}

	servers := root["mcpServers"].(map[string]any)

	// Verify filesystem server preserved
	if _, ok := servers["filesystem"]; !ok {
		t.Error("expected filesystem server to be preserved")
	}

	// Verify mobrig server updated
	mobrig := servers["mobrig"].(map[string]any)
	if mobrig["command"] != "/new/path/to/mobrig" {
		t.Errorf("expected updated command, got %v", mobrig["command"])
	}
}

func TestWriteMCPConfig_And_GenerateAll(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, ".cursor", "mcp.json")

	err := WriteMCPConfig(testFile, "C:\\mobrig.exe")
	if err != nil {
		t.Fatalf("failed to write MCP config: %v", err)
	}

	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}

	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("invalid JSON written: %v", err)
	}

	servers := root["mcpServers"].(map[string]any)
	if _, ok := servers["mobrig"]; !ok {
		t.Error("missing mobrig in generated config")
	}

	// Test GenerateAllConfigs
	configured, _ := GenerateAllConfigs(tmpDir, "C:\\mobrig.exe")
	if len(configured) == 0 {
		t.Error("expected at least one config file generated")
	}
}
