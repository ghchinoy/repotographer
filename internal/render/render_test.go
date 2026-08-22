package render

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ghchinoy/repotographer/internal/model"
)

func TestRenderHTMLAndDOT(t *testing.T) {
	tmpDir := t.TempDir()

	graph := &model.Graph{
		Account:     "octocat",
		AccountType: "user",
		Stats:       model.Stats{IncludedRepos: 1, DomainCount: 1, TechCount: 1, ConnectedCount: 1},
		Domains: []model.Domain{
			{ID: "domain_tools", Label: "Developer Tooling", Blurb: "CLI tools and utilities"},
		},
		Tech: []model.Tech{
			{ID: "tech_go", Label: "Go", Kind: "language", Stack: []string{"go"}},
		},
		Nodes: []model.Node{
			{ID: "repo_tool", Label: "tool", Parent: "domain_tools", Type: "repo", Connectivity: "connected"},
			{ID: "tech_go", Label: "Go", Type: "tech"},
		},
		Edges: []model.Edge{
			{ID: "repo_tool->tech_go", Source: "repo_tool", Target: "tech_go", Label: "implements"},
		},
	}

	htmlPath := filepath.Join(tmpDir, "graph.html")
	if err := RenderHTML(graph, htmlPath); err != nil {
		t.Fatalf("RenderHTML failed: %v", err)
	}

	htmlContent, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("failed to read rendered HTML: %v", err)
	}
	if !strings.Contains(string(htmlContent), "octocat") {
		t.Errorf("expected HTML to contain account name 'octocat'")
	}
	if !strings.Contains(string(htmlContent), "cytoscape") {
		t.Errorf("expected HTML to include cytoscape script")
	}

	dotPath := filepath.Join(tmpDir, "graph.dot")
	if err := RenderDOT(graph, dotPath); err != nil {
		t.Fatalf("RenderDOT failed: %v", err)
	}

	dotContent, err := os.ReadFile(dotPath)
	if err != nil {
		t.Fatalf("failed to read rendered DOT: %v", err)
	}
	if !strings.Contains(string(dotContent), "digraph ConceptGraph") {
		t.Errorf("expected DOT to contain 'digraph ConceptGraph'")
	}
	if !strings.Contains(string(dotContent), "cluster_domain_tools") {
		t.Errorf("expected DOT to contain domain cluster 'cluster_domain_tools'")
	}
}
