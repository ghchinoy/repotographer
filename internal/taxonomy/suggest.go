package taxonomy

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/ghchinoy/repotographer/internal/model"
)

// SuggestTaxonomy produces a taxonomy for the given graph using Gemini if enabled, or deterministic heuristics as fallback.
func SuggestTaxonomy(ctx context.Context, graph *model.Graph, opts LLMOptions) (*model.Taxonomy, bool, error) {
	if !opts.Enabled {
		// Return current graph domains & assignments
		assignments := make(map[string]string)
		for _, n := range graph.Nodes {
			if n.Type == "repo" && n.Parent != "" {
				assignments[n.ID] = n.Parent
			}
		}
		tax := &model.Taxonomy{
			Generated:   time.Now().UTC().Format(time.RFC3339),
			Account:     graph.Account,
			Domains:     graph.Domains,
			Assignments: assignments,
		}
		return tax, false, nil
	}

	llmTax, err := ProposeWithGemini(ctx, graph.Account, graph.Nodes, opts)
	if err != nil {
		// Return warning but don't hard fail if offline fallback is available
		return nil, false, err
	}

	llmTax.Generated = time.Now().UTC().Format(time.RFC3339)
	return llmTax, true, nil
}

// ApplyTaxonomy updates a graph's domains and repository parents from a taxonomy.
func ApplyTaxonomy(graph *model.Graph, taxonomy *model.Taxonomy) {
	if taxonomy == nil || len(taxonomy.Domains) == 0 {
		return
	}

	graph.Domains = taxonomy.Domains
	graph.Stats.DomainCount = len(taxonomy.Domains)

	validDomainIDs := make(map[string]bool)
	for _, d := range taxonomy.Domains {
		validDomainIDs[d.ID] = true
	}

	fallbackDomain := taxonomy.Domains[0].ID

	for i := range graph.Nodes {
		if graph.Nodes[i].Type == "repo" {
			if assigned, ok := taxonomy.Assignments[graph.Nodes[i].ID]; ok && validDomainIDs[assigned] {
				graph.Nodes[i].Parent = assigned
			} else if !validDomainIDs[graph.Nodes[i].Parent] {
				graph.Nodes[i].Parent = fallbackDomain
			}
		}
	}
}

// LoadTaxonomy reads a taxonomy.json file from disk.
func LoadTaxonomy(filePath string) (*model.Taxonomy, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read taxonomy file: %w", err)
	}
	var tax model.Taxonomy
	if err := json.Unmarshal(data, &tax); err != nil {
		return nil, fmt.Errorf("failed to parse taxonomy JSON: %w", err)
	}
	return &tax, nil
}

// SaveTaxonomy writes a taxonomy to disk.
func SaveTaxonomy(filePath string, tax *model.Taxonomy) error {
	data, err := json.MarshalIndent(tax, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize taxonomy JSON: %w", err)
	}
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write taxonomy file: %w", err)
	}
	return nil
}
