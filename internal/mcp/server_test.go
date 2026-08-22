package mcp

import (
	"context"
	"testing"

	"github.com/ghchinoy/repotographer/internal/cli"
	"github.com/ghchinoy/repotographer/internal/model"
	mcp_sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMCPServerToolsRegistration(t *testing.T) {
	ctx := context.Background()

	server := mcp_sdk.NewServer(&mcp_sdk.Implementation{
		Name:    "repotographer-test",
		Version: cli.Version,
	}, nil)

	// Register tools
	mcp_sdk.AddTool(server, &mcp_sdk.Tool{
		Name:        "map_github_account",
		Description: "Maps github account",
	}, func(ctx context.Context, req *mcp_sdk.CallToolRequest, args MapAccountParams) (*mcp_sdk.CallToolResult, any, error) {
		return &mcp_sdk.CallToolResult{}, nil, nil
	})

	mcp_sdk.AddTool(server, &mcp_sdk.Tool{
		Name:        "suggest_taxonomy",
		Description: "Suggests taxonomy",
	}, func(ctx context.Context, req *mcp_sdk.CallToolRequest, args SuggestTaxonomyParams) (*mcp_sdk.CallToolResult, any, error) {
		return &mcp_sdk.CallToolResult{}, nil, nil
	})

	mcp_sdk.AddTool(server, &mcp_sdk.Tool{
		Name:        "render_graph",
		Description: "Renders graph",
	}, func(ctx context.Context, req *mcp_sdk.CallToolRequest, args RenderGraphParams) (*mcp_sdk.CallToolResult, any, error) {
		return &mcp_sdk.CallToolResult{}, nil, nil
	})

	t1, t2 := mcp_sdk.NewInMemoryTransports()
	ss, err := server.Connect(ctx, t1, nil)
	if err != nil {
		t.Fatalf("failed to connect server: %v", err)
	}
	defer ss.Close()

	client := mcp_sdk.NewClient(&mcp_sdk.Implementation{Name: "test-client", Version: "1.0.0"}, nil)
	cs, err := client.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatalf("failed to connect client: %v", err)
	}
	defer cs.Close()

	toolsResult, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	expectedTools := map[string]bool{
		"map_github_account": false,
		"suggest_taxonomy":   false,
		"render_graph":       false,
	}

	for _, tool := range toolsResult.Tools {
		if _, ok := expectedTools[tool.Name]; ok {
			expectedTools[tool.Name] = true
		}
	}

	for name, found := range expectedTools {
		if !found {
			t.Errorf("expected tool '%s' was not registered in MCP server", name)
		}
	}
}

func TestRenderGraphToolExecution(t *testing.T) {
	ctx := context.Background()

	server := mcp_sdk.NewServer(&mcp_sdk.Implementation{
		Name:    "repotographer-test",
		Version: cli.Version,
	}, nil)

	mcp_sdk.AddTool(server, &mcp_sdk.Tool{
		Name:        "render_graph",
		Description: "Renders graph",
	}, func(ctx context.Context, req *mcp_sdk.CallToolRequest, args RenderGraphParams) (*mcp_sdk.CallToolResult, any, error) {
		tmpDir := t.TempDir()
		args.OutDir = tmpDir
		res := RenderGraphResult{
			HTMLPath: tmpDir + "/graph.html",
			JSONPath: tmpDir + "/graph.json",
		}
		return &mcp_sdk.CallToolResult{
			Content: []mcp_sdk.Content{&mcp_sdk.TextContent{Text: "rendered"}},
		}, res, nil
	})

	t1, t2 := mcp_sdk.NewInMemoryTransports()
	ss, err := server.Connect(ctx, t1, nil)
	if err != nil {
		t.Fatalf("failed to connect server: %v", err)
	}
	defer ss.Close()

	client := mcp_sdk.NewClient(&mcp_sdk.Implementation{Name: "test-client", Version: "1.0.0"}, nil)
	cs, err := client.Connect(ctx, t2, nil)
	if err != nil {
		t.Fatalf("failed to connect client: %v", err)
	}
	defer cs.Close()

	testGraph := model.Graph{
		Account: "test-user",
		Stats:   model.Stats{IncludedRepos: 1},
	}

	callRes, err := cs.CallTool(ctx, &mcp_sdk.CallToolParams{
		Name: "render_graph",
		Arguments: map[string]any{
			"graph":   testGraph,
			"formats": []string{"html", "json"},
		},
	})
	if err != nil {
		t.Fatalf("CallTool render_graph failed: %v", err)
	}

	if len(callRes.Content) == 0 {
		t.Fatalf("expected non-empty content response from render_graph")
	}
}
