package mcp

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestMCPConfigGenerators(t *testing.T) {
	envMap := map[string]string{
		"GEMINI_API_KEY": "test-key-123",
	}

	// 1. Test Claude format
	cfg := StdMCPServersConfig{
		MCPServers: map[string]StdMCPServerEntry{
			"repotographer": {
				Command: "/usr/local/bin/repotographer",
				Args:    []string{"mcp"},
				Env:     envMap,
			},
		},
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal Claude config: %v", err)
	}
	if !strings.Contains(string(data), "mcpServers") || !strings.Contains(string(data), "test-key-123") {
		t.Errorf("unexpected Claude config output: %s", string(data))
	}

	// 2. Test OpenCode format
	ocCfg := OpenCodeMCPConfig{
		MCP: map[string]OpenCodeMCPEntry{
			"repotographer": {
				Type:        "local",
				Command:     []string{"/usr/local/bin/repotographer", "mcp"},
				Environment: envMap,
				Enabled:     true,
			},
		},
	}
	data, err = json.Marshal(ocCfg)
	if err != nil {
		t.Fatalf("failed to marshal OpenCode config: %v", err)
	}
	if !strings.Contains(string(data), "\"type\":\"local\"") || !strings.Contains(string(data), "\"enabled\":true") {
		t.Errorf("unexpected OpenCode config output: %s", string(data))
	}
}

func TestMCPConfigCLIExecution(t *testing.T) {
	var buf bytes.Buffer
	mcpConfigCmd.SetOut(&buf)
	mcpConfigCmd.SetErr(&buf)
	mcpConfigCmd.SetArgs([]string{"--client", "claude"})

	err := mcpConfigCmd.Execute()
	if err != nil {
		t.Fatalf("mcp config execution failed: %v", err)
	}
}
