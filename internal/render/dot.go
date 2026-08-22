package render

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghchinoy/repotographer/internal/model"
)

// RenderDOT generates a Graphviz DOT representation of the concept graph.
func RenderDOT(graph *model.Graph, outputPath string) error {
	var sb strings.Builder

	sb.WriteString("digraph ConceptGraph {\n")
	sb.WriteString("  // Graph styling\n")
	sb.WriteString("  graph [rankdir=LR, bgcolor=\"#0f172a\", pad=\"0.5\", nodesep=\"0.6\", ranksep=\"0.8\", fontname=\"Helvetica\"];\n")
	sb.WriteString("  node [fontname=\"Helvetica\", fontsize=10, margin=\"0.15,0.08\"];\n")
	sb.WriteString("  edge [fontname=\"Helvetica\", fontsize=8, color=\"#475569\", fontcolor=\"#94a3b8\", arrowsize=0.6];\n\n")

	// 1. Tech hub nodes
	sb.WriteString("  // Core Technology Hubs\n")
	for _, tech := range graph.Tech {
		label := escapeDOTString(tech.Label)
		sb.WriteString(fmt.Sprintf("  \"%s\" [label=\"%s\", shape=diamond, style=filled, fillcolor=\"#f59e0b\", fontcolor=\"#0f172a\", color=\"#fbbf24\", penwidth=1.8];\n",
			tech.ID, label))
	}
	sb.WriteString("\n")

	// 2. Group repos into Domain subgraphs
	domainNodes := make(map[string][]model.Node)
	for _, n := range graph.Nodes {
		if n.Type == "repo" {
			domainNodes[n.Parent] = append(domainNodes[n.Parent], n)
		}
	}

	for _, domain := range graph.Domains {
		nodes := domainNodes[domain.ID]
		if len(nodes) == 0 {
			continue
		}

		clusterID := fmt.Sprintf("cluster_%s", domain.ID)
		domainLabel := escapeDOTString(domain.Label)

		sb.WriteString(fmt.Sprintf("  subgraph \"%s\" {\n", clusterID))
		sb.WriteString(fmt.Sprintf("    label=\"%s\";\n", domainLabel))
		sb.WriteString("    style=\"rounded,dashed\";\n")
		sb.WriteString("    color=\"#475569\";\n")
		sb.WriteString("    fontcolor=\"#38bdf8\";\n")
		sb.WriteString("    fontsize=12;\n")
		sb.WriteString("    penwidth=1.2;\n\n")

		for _, n := range nodes {
			label := escapeDOTString(n.Label)
			conn := n.Connectivity
			if conn == "" {
				conn = "connected"
			}

			var style, fillcolor, color, fontcolor string
			var penwidth float64

			switch conn {
			case "connected":
				style = "filled,rounded"
				fillcolor = "#1e293b"
				color = "#38bdf8"
				fontcolor = "#f8fafc"
				penwidth = 1.6
			case "standalone":
				style = "filled,rounded,dashed"
				fillcolor = "#0f172a"
				color = "#64748b"
				fontcolor = "#94a3b8"
				penwidth = 1.0
			case "bare":
				style = "filled,rounded,dotted"
				fillcolor = "#0f172a"
				color = "#475569"
				fontcolor = "#64748b"
				penwidth = 0.8
			}

			sb.WriteString(fmt.Sprintf("    \"%s\" [label=\"%s\", shape=box, style=\"%s\", fillcolor=\"%s\", color=\"%s\", fontcolor=\"%s\", penwidth=%.1f];\n",
				n.ID, label, style, fillcolor, color, fontcolor, penwidth))
		}

		sb.WriteString("  }\n\n")
	}

	// 3. Edges
	sb.WriteString("  // Graph Edges\n")
	for _, e := range graph.Edges {
		labelAttr := ""
		if e.Label != "" {
			labelAttr = fmt.Sprintf(", label=\"%s\"", escapeDOTString(e.Label))
		}
		sb.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\" [color=\"#38bdf8\"%s];\n", e.Source, e.Target, labelAttr))
	}

	sb.WriteString("}\n")

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory for DOT file: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("failed to write DOT file: %w", err)
	}

	return nil
}

func escapeDOTString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}
