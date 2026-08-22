package cluster

import (
	"testing"

	"github.com/ghchinoy/repotographer/internal/model"
)

func TestComputeConnectivity(t *testing.T) {
	// Connected case
	if got := ComputeConnectivity(true, "", nil, nil); got != "connected" {
		t.Errorf("expected 'connected', got '%s'", got)
	}

	// Standalone with description
	if got := ComputeConnectivity(false, "A real project description", nil, nil); got != "standalone" {
		t.Errorf("expected 'standalone', got '%s'", got)
	}

	// Standalone with topics
	if got := ComputeConnectivity(false, "", []string{"go", "cli"}, nil); got != "standalone" {
		t.Errorf("expected 'standalone', got '%s'", got)
	}

	// Standalone with meaningful stack
	if got := ComputeConnectivity(false, "", nil, []string{"mcp", "ai"}); got != "standalone" {
		t.Errorf("expected 'standalone', got '%s'", got)
	}

	// Bare case
	if got := ComputeConnectivity(false, "", nil, []string{"go"}); got != "bare" {
		t.Errorf("expected 'bare', got '%s'", got)
	}
}

func TestBuildDeterministicGraph(t *testing.T) {
	repos := []model.Repo{
		{
			Name:        "agent-mcp-server",
			Description: "Model Context Protocol server for Go agents",
			PrimaryLanguage: &model.Language{Name: "Go"},
			RepositoryTopics: []model.Topic{{Name: "mcp"}, {Name: "agent"}},
		},
		{
			Name:        "voice-tts-engine",
			Description: "Fast speech synthesis in Rust",
			PrimaryLanguage: &model.Language{Name: "Rust"},
			RepositoryTopics: []model.Topic{{Name: "audio"}, {Name: "tts"}},
		},
	}

	graph, tax := BuildDeterministicGraph("test-user", "user", repos, true)

	if graph.Stats.IncludedRepos != 2 {
		t.Fatalf("expected 2 included repos, got %d", graph.Stats.IncludedRepos)
	}
	if len(graph.Domains) == 0 {
		t.Fatalf("expected at least 1 domain, got %d", len(graph.Domains))
	}
	if len(graph.Tech) == 0 {
		t.Fatalf("expected at least 1 tech hub, got %d", len(graph.Tech))
	}
	if len(tax.Assignments) != 2 {
		t.Fatalf("expected 2 taxonomy assignments, got %d", len(tax.Assignments))
	}
}
