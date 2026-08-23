package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghchinoy/repotographer/internal/taxonomy"
	"github.com/spf13/cobra"
)

func init() {
	mcpCmd.AddCommand(mcpConfigCmd)
}

var mcpConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Print MCP client configuration snippets for Claude Desktop, Cursor, OpenCode, and Antigravity",
	Long: `Prints formatted JSON configuration blocks for integrating repotographer with popular MCP clients.
Resolves the installed repotographer binary path and configures recommended environment variables.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, _ := cmd.Flags().GetString("client")
		useEnv, _ := cmd.Flags().GetBool("env")

		binPath, err := os.Executable()
		if err != nil {
			binPath = "repotographer"
		} else {
			if resolved, err := filepath.EvalSymlinks(binPath); err == nil {
				binPath = resolved
			}
		}

		envMap := make(map[string]string)
		if useEnv {
			apiKey := taxonomy.ResolveAPIKey("")
			if apiKey != "" {
				envMap["GEMINI_API_KEY"] = apiKey
			} else {
				envMap["GEMINI_API_KEY"] = "your-gemini-api-key"
			}

			project := taxonomy.ResolveGCPProject("")
			if project != "" {
				envMap["GOOGLE_CLOUD_PROJECT"] = project
				envMap["GOOGLE_GENAI_USE_VERTEXAI"] = "true"
			}
		} else {
			envMap["GEMINI_API_KEY"] = "your-gemini-api-key"
		}

		client = strings.ToLower(strings.TrimSpace(client))

		switch client {
		case "claude", "claude-desktop":
			printClaudeConfig(binPath, envMap)
		case "cursor":
			printCursorConfig(binPath, envMap)
		case "opencode":
			printOpenCodeConfig(binPath, envMap)
		case "antigravity":
			printAntigravityConfig(binPath, envMap)
		case "all":
			printAllConfigs(binPath, envMap)
		default:
			return fmt.Errorf("unknown client '%s'. Supported clients: all, claude, cursor, opencode, antigravity", client)
		}

		return nil
	},
}

func init() {
	mcpConfigCmd.Flags().StringP("client", "c", "all", "Target client: all, claude, cursor, opencode, antigravity")
	mcpConfigCmd.Flags().Bool("env", false, "Populate config with detected environment variables instead of placeholders")
}

// Format 1: Standard mcpServers wrapper (Claude Desktop, Cursor)
type StdMCPServersConfig struct {
	MCPServers map[string]StdMCPServerEntry `json:"mcpServers"`
}

type StdMCPServerEntry struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

// Format 2: OpenCode / Antigravity local tool declaration
type OpenCodeMCPConfig struct {
	MCP map[string]OpenCodeMCPEntry `json:"mcp"`
}

type OpenCodeMCPEntry struct {
	Type        string            `json:"type"`
	Command     []string          `json:"command"`
	Environment map[string]string `json:"environment,omitempty"`
	Enabled     bool              `json:"enabled"`
}

func printClaudeConfig(binPath string, envMap map[string]string) {
	cfg := StdMCPServersConfig{
		MCPServers: map[string]StdMCPServerEntry{
			"repotographer": {
				Command: binPath,
				Args:    []string{"mcp"},
				Env:     envMap,
			},
		},
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	fmt.Println(string(data))
}

func printCursorConfig(binPath string, envMap map[string]string) {
	cfg := StdMCPServersConfig{
		MCPServers: map[string]StdMCPServerEntry{
			"repotographer": {
				Command: binPath,
				Args:    []string{"mcp"},
				Env:     envMap,
			},
		},
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	fmt.Println(string(data))
}

func printOpenCodeConfig(binPath string, envMap map[string]string) {
	cfg := OpenCodeMCPConfig{
		MCP: map[string]OpenCodeMCPEntry{
			"repotographer": {
				Type:        "local",
				Command:     []string{binPath, "mcp"},
				Environment: envMap,
				Enabled:     true,
			},
		},
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	fmt.Println(string(data))
}

func printAntigravityConfig(binPath string, envMap map[string]string) {
	cfg := StdMCPServersConfig{
		MCPServers: map[string]StdMCPServerEntry{
			"repotographer": {
				Command: binPath,
				Args:    []string{"mcp"},
				Env:     envMap,
			},
		},
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	fmt.Println(string(data))
}

func printAllConfigs(binPath string, envMap map[string]string) {
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("  Claude Desktop Config (~/.config/Claude/claude_desktop_config.json)")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	printClaudeConfig(binPath, envMap)
	fmt.Println()

	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("  Cursor Config (~/.cursor/mcp.json)")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	printCursorConfig(binPath, envMap)
	fmt.Println()

	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("  OpenCode Config (opencode.json)")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	printOpenCodeConfig(binPath, envMap)
	fmt.Println()

	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("  Antigravity Config (.antigravity/mcp_config.json)")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	printAntigravityConfig(binPath, envMap)
}
