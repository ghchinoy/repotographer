package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghchinoy/repotographer/internal/cluster"
	"github.com/ghchinoy/repotographer/internal/github"
	"github.com/ghchinoy/repotographer/internal/model"
	"github.com/ghchinoy/repotographer/internal/render"
	"github.com/ghchinoy/repotographer/internal/taxonomy"
	"github.com/spf13/cobra"
)

var Version = "0.1.0"

// RootCmd is the main entry point for the repotographer CLI.
var RootCmd = &cobra.Command{
	Use:   "repotographer",
	Short: "Map and visualize public GitHub repositories into concept graphs with AI-suggested taxonomy",
	Long: `repotographer connects to GitHub CLI (gh) to fetch repositories for a user or organization,
builds a structured concept graph with technology hubs and connectivity states,
proposes an intelligent domain taxonomy using Gemini (or Vertex AI),
and renders interactive Cytoscape HTML visualizers, Graphviz DOT, and PNG graphs.`,
}

func init() {
	RootCmd.AddCommand(mapCmd)
	RootCmd.AddCommand(suggestCmd)
	RootCmd.AddCommand(renderCmd)
	RootCmd.AddCommand(versionCmd)

	// Flags for map
	mapCmd.Flags().String("type", "auto", "Account type: 'user', 'org', or 'auto'")
	mapCmd.Flags().StringP("out", "o", "./out", "Output directory for generated files")
	mapCmd.Flags().String("format", "html,dot,png,json", "Comma-separated list of formats to generate: html,dot,png,json")
	mapCmd.Flags().Bool("llm", true, "Use Gemini / Vertex AI to refine domain taxonomy")
	mapCmd.Flags().Bool("vertex", false, "Use Google Cloud Vertex AI instead of Gemini API key")
	mapCmd.Flags().String("model", "gemini-3.7-flash", "Gemini model to use for taxonomy suggestions")
	mapCmd.Flags().IntP("limit", "n", 500, "Maximum number of repositories to fetch")
	mapCmd.Flags().Bool("include-forks", false, "Include forked repositories in mapping")
	mapCmd.Flags().Bool("include-archived", false, "Include archived repositories in mapping")
	mapCmd.Flags().Bool("min-signal", true, "Filter out completely empty repositories with no desc/topics/lang/stars")
	mapCmd.Flags().String("taxonomy", "", "Optional path to a custom taxonomy.json to apply instead of generating one")

	// Flags for suggest
	suggestCmd.Flags().String("type", "auto", "Account type: 'user', 'org', or 'auto'")
	suggestCmd.Flags().StringP("out", "o", "taxonomy.json", "Output path for the generated taxonomy JSON")
	suggestCmd.Flags().Bool("vertex", false, "Use Google Cloud Vertex AI instead of Gemini API key")
	suggestCmd.Flags().String("model", "gemini-3.7-flash", "Gemini model to use for taxonomy suggestions")
	suggestCmd.Flags().IntP("limit", "n", 500, "Maximum number of repositories to fetch")

	// Flags for render
	renderCmd.Flags().StringP("out", "o", "./out", "Output directory for rendered assets")
	renderCmd.Flags().String("format", "html,dot,png", "Comma-separated list of formats: html,dot,png")
	renderCmd.Flags().String("taxonomy", "", "Optional path to a custom taxonomy.json to apply before rendering")
}

var mapCmd = &cobra.Command{
	Use:   "map <owner>",
	Short: "Fetch repos, construct concept graph, refine taxonomy with AI, and render visual outputs",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		owner := args[0]
		outDir, _ := cmd.Flags().GetString("out")
		formatsStr, _ := cmd.Flags().GetString("format")
		useLLM, _ := cmd.Flags().GetBool("llm")
		useVertex, _ := cmd.Flags().GetBool("vertex")
		modelName, _ := cmd.Flags().GetString("model")
		limit, _ := cmd.Flags().GetInt("limit")
		includeForks, _ := cmd.Flags().GetBool("include-forks")
		includeArchived, _ := cmd.Flags().GetBool("include-archived")
		minSignal, _ := cmd.Flags().GetBool("min-signal")
		customTaxPath, _ := cmd.Flags().GetString("taxonomy")
		accType, _ := cmd.Flags().GetString("type")

		fmt.Printf("🔍 Fetching public repositories for '%s' via `gh`...\n", owner)
		repos, err := github.FetchRepos(github.FetchOptions{
			Owner:           owner,
			AccountType:     accType,
			Limit:           limit,
			IncludeForks:    includeForks,
			IncludeArchived: includeArchived,
		})
		if err != nil {
			return err
		}
		fmt.Printf("✓ Retrieved %d repositories from GitHub.\n", len(repos))

		fmt.Println("🧩 Building concept graph and detecting technology hubs...")
		graph, tax := cluster.BuildDeterministicGraph(owner, accType, repos, minSignal)
		fmt.Printf("✓ Concept graph initialized: %d repos included across %d tech hubs.\n",
			graph.Stats.IncludedRepos, graph.Stats.TechCount)

		// Apply custom or AI taxonomy
		if customTaxPath != "" {
			fmt.Printf("📂 Loading custom taxonomy from '%s'...\n", customTaxPath)
			customTax, err := taxonomy.LoadTaxonomy(customTaxPath)
			if err != nil {
				return err
			}
			taxonomy.ApplyTaxonomy(graph, customTax)
			tax = customTax
			fmt.Printf("✓ Applied custom taxonomy with %d domains.\n", len(customTax.Domains))
		} else if useLLM {
			fmt.Printf("✨ Proposing domain taxonomy with Gemini (%s)...\n", modelName)
			llmTax, usedLLM, err := taxonomy.SuggestTaxonomy(context.Background(), graph, taxonomy.LLMOptions{
				Enabled:   true,
				Model:     modelName,
				UseVertex: useVertex,
			})
			if err != nil {
				fmt.Printf("⚠️  Gemini taxonomy generation notice: %v\n", err)
				fmt.Println("   Continuing with deterministic heuristic taxonomy.")
			} else if usedLLM {
				tax = llmTax
				taxonomy.ApplyTaxonomy(graph, tax)
				fmt.Printf("✓ Gemini proposed %d thematic domains.\n", len(tax.Domains))
			}
		}

		if err := os.MkdirAll(outDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		formats := strings.Split(formatsStr, ",")
		formatSet := make(map[string]bool)
		for _, f := range formats {
			formatSet[strings.TrimSpace(strings.ToLower(f))] = true
		}

		// Save graph JSON
		graphPath := filepath.Join(outDir, "graph.json")
		if formatSet["json"] {
			graphData, _ := json.MarshalIndent(graph, "", "  ")
			if err := os.WriteFile(graphPath, graphData, 0644); err != nil {
				return fmt.Errorf("failed to write graph.json: %w", err)
			}
			fmt.Printf("📄 Graph JSON: %s\n", graphPath)
		}

		// Save taxonomy JSON (curation surface)
		taxPath := filepath.Join(outDir, "taxonomy.json")
		if err := taxonomy.SaveTaxonomy(taxPath, tax); err != nil {
			return fmt.Errorf("failed to write taxonomy.json: %w", err)
		}
		fmt.Printf("📝 Taxonomy Curation File: %s\n", taxPath)

		// Render HTML
		if formatSet["html"] {
			htmlPath := filepath.Join(outDir, "graph.html")
			if err := render.RenderHTML(graph, htmlPath); err != nil {
				return err
			}
			fmt.Printf("🌐 Interactive HTML Visualizer: %s\n", htmlPath)
		}

		// Render DOT
		dotPath := filepath.Join(outDir, "graph.dot")
		if formatSet["dot"] || formatSet["png"] {
			if err := render.RenderDOT(graph, dotPath); err != nil {
				return err
			}
			if formatSet["dot"] {
				fmt.Printf("📊 Graphviz DOT: %s\n", dotPath)
			}
		}

		// Render PNG
		if formatSet["png"] {
			pngPath := filepath.Join(outDir, "graph.png")
			rendered, err := render.RenderPNG(dotPath, pngPath)
			if err != nil {
				fmt.Printf("⚠️  PNG rendering failed: %v\n", err)
			} else if rendered {
				fmt.Printf("🖼️  Static Graph Image: %s\n", pngPath)
			} else {
				fmt.Println("ℹ️  Graphviz `dot` was not found on PATH. Skipped PNG generation (install Graphviz to enable).")
			}
		}

		fmt.Printf("\n🎉 Success! Concept map created in %s\n", outDir)
		return nil
	},
}

var suggestCmd = &cobra.Command{
	Use:   "suggest <owner>",
	Short: "Suggest an AI domain taxonomy for a GitHub account and save taxonomy.json",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		owner := args[0]
		outPath, _ := cmd.Flags().GetString("out")
		useVertex, _ := cmd.Flags().GetBool("vertex")
		modelName, _ := cmd.Flags().GetString("model")
		limit, _ := cmd.Flags().GetInt("limit")
		accType, _ := cmd.Flags().GetString("type")

		fmt.Printf("🔍 Fetching public repositories for '%s' via `gh`...\n", owner)
		repos, err := github.FetchRepos(github.FetchOptions{
			Owner:        owner,
			AccountType:  accType,
			Limit:        limit,
			IncludeForks: false,
		})
		if err != nil {
			return err
		}

		graph, detTax := cluster.BuildDeterministicGraph(owner, accType, repos, true)
		tax := detTax

		fmt.Printf("✨ Asking Gemini (%s) to propose domain taxonomy...\n", modelName)
		llmTax, usedLLM, err := taxonomy.SuggestTaxonomy(context.Background(), graph, taxonomy.LLMOptions{
			Enabled:   true,
			Model:     modelName,
			UseVertex: useVertex,
		})
		if err != nil {
			fmt.Printf("⚠️  LLM taxonomy generation notice: %v\n", err)
			fmt.Println("   Generated deterministic heuristic taxonomy.")
		} else if usedLLM && llmTax != nil {
			tax = llmTax
		} else {
			fmt.Println("ℹ️  Using deterministic heuristic taxonomy.")
		}

		if err := taxonomy.SaveTaxonomy(outPath, tax); err != nil {
			return err
		}

		fmt.Printf("✓ Taxonomy with %d domains saved to %s\n", len(tax.Domains), outPath)
		return nil
	},
}

var renderCmd = &cobra.Command{
	Use:   "render <graph.json>",
	Short: "Render visual assets (html, dot, png) from an existing graph.json",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		graphPath := args[0]
		outDir, _ := cmd.Flags().GetString("out")
		formatsStr, _ := cmd.Flags().GetString("format")
		taxPath, _ := cmd.Flags().GetString("taxonomy")

		data, err := os.ReadFile(graphPath)
		if err != nil {
			return fmt.Errorf("failed to read graph file: %w", err)
		}
		var graph model.Graph
		if err := json.Unmarshal(data, &graph); err != nil {
			return fmt.Errorf("failed to parse graph JSON: %w", err)
		}

		if taxPath != "" {
			tax, err := taxonomy.LoadTaxonomy(taxPath)
			if err != nil {
				return err
			}
			taxonomy.ApplyTaxonomy(&graph, tax)
			fmt.Printf("✓ Applied taxonomy from %s\n", taxPath)
		}

		if err := os.MkdirAll(outDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		formats := strings.Split(formatsStr, ",")
		formatSet := make(map[string]bool)
		for _, f := range formats {
			formatSet[strings.TrimSpace(strings.ToLower(f))] = true
		}

		if formatSet["html"] {
			htmlPath := filepath.Join(outDir, "graph.html")
			if err := render.RenderHTML(&graph, htmlPath); err != nil {
				return err
			}
			fmt.Printf("🌐 HTML: %s\n", htmlPath)
		}

		dotPath := filepath.Join(outDir, "graph.dot")
		if formatSet["dot"] || formatSet["png"] {
			if err := render.RenderDOT(&graph, dotPath); err != nil {
				return err
			}
			if formatSet["dot"] {
				fmt.Printf("📊 DOT: %s\n", dotPath)
			}
		}

		if formatSet["png"] {
			pngPath := filepath.Join(outDir, "graph.png")
			rendered, err := render.RenderPNG(dotPath, pngPath)
			if err != nil {
				fmt.Printf("⚠️  PNG rendering failed: %v\n", err)
			} else if rendered {
				fmt.Printf("🖼️  PNG: %s\n", pngPath)
			} else {
				fmt.Println("ℹ️  Graphviz `dot` not found. Skipped PNG.")
			}
		}

		return nil
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print repotographer version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("repotographer v%s\n", Version)
	},
}
