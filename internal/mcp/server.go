package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ghchinoy/repotographer/internal/cli"
	"github.com/ghchinoy/repotographer/internal/cluster"
	"github.com/ghchinoy/repotographer/internal/github"
	"github.com/ghchinoy/repotographer/internal/model"
	"github.com/ghchinoy/repotographer/internal/render"
	"github.com/ghchinoy/repotographer/internal/taxonomy"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

func init() {
	cli.RootCmd.AddCommand(mcpCmd)
}

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start repotographer as an MCP (Model Context Protocol) stdio server",
	RunE: func(cmd *cobra.Command, args []string) error {
		return RunMCPServer(cmd.Context())
	},
}

// MapAccountParams defines the input for the map_github_account MCP tool.
type MapAccountParams struct {
	Owner       string `json:"owner" jsonschema:"GitHub username or organization name"`
	AccountType string `json:"account_type,omitempty" jsonschema:"Optional account type: 'user' or 'org'"`
	Limit       int    `json:"limit,omitempty" jsonschema:"Maximum number of repositories to analyze (default 500)"`
	UseLLM      bool   `json:"use_llm,omitempty" jsonschema:"Whether to refine domain taxonomy with Gemini (default true)"`
	Model       string `json:"model,omitempty" jsonschema:"Gemini model ID to use (default gemini-3.7-flash)"`
	UseVertex   bool   `json:"use_vertex,omitempty" jsonschema:"Whether to use Google Cloud Vertex AI"`
}

// MapAccountResult defines the output of the map_github_account MCP tool.
type MapAccountResult struct {
	Graph    *model.Graph    `json:"graph"`
	Taxonomy *model.Taxonomy `json:"taxonomy"`
	LLMUsed  bool            `json:"llm_used"`
	LLMNote  string          `json:"llm_note,omitempty"`
}

// SuggestTaxonomyParams defines the input for the suggest_taxonomy MCP tool.
type SuggestTaxonomyParams struct {
	Owner       string `json:"owner" jsonschema:"GitHub username or organization name"`
	AccountType string `json:"account_type,omitempty" jsonschema:"Optional account type: 'user' or 'org'"`
	Limit       int    `json:"limit,omitempty" jsonschema:"Maximum number of repositories to analyze"`
	Model       string `json:"model,omitempty" jsonschema:"Gemini model ID to use"`
	UseVertex   bool   `json:"use_vertex,omitempty" jsonschema:"Whether to use Google Cloud Vertex AI"`
}

// SuggestTaxonomyResult defines the output of the suggest_taxonomy MCP tool.
type SuggestTaxonomyResult struct {
	Taxonomy *model.Taxonomy `json:"taxonomy"`
	LLMUsed  bool            `json:"llm_used"`
	LLMNote  string          `json:"llm_note,omitempty"`
}

// RenderGraphParams defines the input for the render_graph MCP tool.
type RenderGraphParams struct {
	Graph   model.Graph `json:"graph" jsonschema:"The concept graph data structure to render"`
	OutDir  string      `json:"out_dir,omitempty" jsonschema:"Directory to save rendered files (defaults to temporary directory)"`
	Formats []string    `json:"formats,omitempty" jsonschema:"List of formats to produce: html, dot, png, json"`
}

// RenderGraphResult defines the output of the render_graph MCP tool.
type RenderGraphResult struct {
	HTMLPath string `json:"html_path,omitempty"`
	DOTPath  string `json:"dot_path,omitempty"`
	PNGPath  string `json:"png_path,omitempty"`
	JSONPath string `json:"json_path,omitempty"`
}

// RunMCPServer registers all tools and starts the Model Context Protocol stdio server.
func RunMCPServer(ctx context.Context) error {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "repotographer",
		Version: cli.Version,
	}, nil)

	// Tool 1: map_github_account
	mcp.AddTool(server, &mcp.Tool{
		Name:        "map_github_account",
		Description: "Fetches public GitHub repositories for a user/org, maps technology connections, and derives domain taxonomy.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args MapAccountParams) (*mcp.CallToolResult, any, error) {
		limit := args.Limit
		if limit <= 0 {
			limit = 500
		}
		modelName := args.Model
		if modelName == "" {
			modelName = "gemini-3.7-flash"
		}

		repos, err := github.FetchRepos(github.FetchOptions{
			Owner:       args.Owner,
			AccountType: args.AccountType,
			Limit:       limit,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("github fetch error: %w", err)
		}

		graph, tax := cluster.BuildDeterministicGraph(args.Owner, args.AccountType, repos, true)

		llmUsed := false
		llmNote := ""
		if args.UseLLM {
			llmTax, usedLLM, err := taxonomy.SuggestTaxonomy(ctx, graph, taxonomy.LLMOptions{
				Enabled:   true,
				Model:     modelName,
				UseVertex: args.UseVertex,
			})
			if err != nil {
				llmNote = fmt.Sprintf("LLM taxonomy refinement unavailable: %v (used deterministic taxonomy)", err)
			} else if usedLLM && llmTax != nil {
				tax = llmTax
				taxonomy.ApplyTaxonomy(graph, tax)
				llmUsed = true
			} else {
				llmNote = "Deterministic heuristic taxonomy used"
			}
		} else {
			llmNote = "LLM refinement disabled by parameter"
		}

		res := MapAccountResult{
			Graph:    graph,
			Taxonomy: tax,
			LLMUsed:  llmUsed,
			LLMNote:  llmNote,
		}

		resJSON, _ := json.Marshal(res)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(resJSON)},
			},
		}, res, nil
	})

	// Tool 2: suggest_taxonomy
	mcp.AddTool(server, &mcp.Tool{
		Name:        "suggest_taxonomy",
		Description: "Proposes an intelligent domain pillar taxonomy for a GitHub account's repositories using Gemini (with automatic heuristic fallback).",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args SuggestTaxonomyParams) (*mcp.CallToolResult, any, error) {
		limit := args.Limit
		if limit <= 0 {
			limit = 500
		}
		modelName := args.Model
		if modelName == "" {
			modelName = "gemini-3.7-flash"
		}

		repos, err := github.FetchRepos(github.FetchOptions{
			Owner:       args.Owner,
			AccountType: args.AccountType,
			Limit:       limit,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("github fetch error: %w", err)
		}

		graph, detTax := cluster.BuildDeterministicGraph(args.Owner, args.AccountType, repos, true)
		tax := detTax

		llmTax, usedLLM, err := taxonomy.SuggestTaxonomy(ctx, graph, taxonomy.LLMOptions{
			Enabled:   true,
			Model:     modelName,
			UseVertex: args.UseVertex,
		})
		llmUsed := false
		llmNote := ""
		if err != nil {
			llmNote = fmt.Sprintf("LLM taxonomy refinement unavailable: %v (used deterministic taxonomy)", err)
		} else if usedLLM && llmTax != nil {
			tax = llmTax
			llmUsed = true
		} else {
			llmNote = "Deterministic heuristic taxonomy used"
		}

		res := SuggestTaxonomyResult{
			Taxonomy: tax,
			LLMUsed:  llmUsed,
			LLMNote:  llmNote,
		}

		resJSON, _ := json.Marshal(res)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(resJSON)},
			},
		}, res, nil
	})

	// Tool 3: render_graph
	mcp.AddTool(server, &mcp.Tool{
		Name:        "render_graph",
		Description: "Renders a concept graph data structure to interactive HTML, Graphviz DOT, and PNG.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args RenderGraphParams) (*mcp.CallToolResult, any, error) {
		outDir := args.OutDir
		if outDir == "" {
			tmpDir, err := os.MkdirTemp("", "repotographer-*")
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create temporary output directory: %w", err)
			}
			outDir = tmpDir
		}

		if err := os.MkdirAll(outDir, 0755); err != nil {
			return nil, nil, fmt.Errorf("failed to create output directory: %w", err)
		}

		formats := args.Formats
		if len(formats) == 0 {
			formats = []string{"html", "dot", "png", "json"}
		}
		formatSet := make(map[string]bool)
		for _, f := range formats {
			formatSet[f] = true
		}

		result := RenderGraphResult{}

		if formatSet["json"] {
			p := filepath.Join(outDir, "graph.json")
			data, _ := json.MarshalIndent(args.Graph, "", "  ")
			_ = os.WriteFile(p, data, 0644)
			result.JSONPath = p
		}

		if formatSet["html"] {
			p := filepath.Join(outDir, "graph.html")
			if err := render.RenderHTML(&args.Graph, p); err == nil {
				result.HTMLPath = p
			}
		}

		dotPath := filepath.Join(outDir, "graph.dot")
		if formatSet["dot"] || formatSet["png"] {
			if err := render.RenderDOT(&args.Graph, dotPath); err == nil {
				if formatSet["dot"] {
					result.DOTPath = dotPath
				}
			}
		}

		if formatSet["png"] {
			pngPath := filepath.Join(outDir, "graph.png")
			if rendered, err := render.RenderPNG(dotPath, pngPath); err == nil && rendered {
				result.PNGPath = pngPath
			}
		}

		resJSON, _ := json.Marshal(result)
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(resJSON)},
			},
		}, result, nil
	})

	return server.Run(ctx, &mcp.StdioTransport{})
}
