package render

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/ghchinoy/repotographer/internal/model"
)

//go:embed templates/graph.html.tmpl
var htmlTemplateContent string

type templateData struct {
	Account     string
	AccountType string
	Stats       model.Stats
	GraphJSON   template.JS
}

// RenderHTML produces a self-contained interactive Cytoscape.js HTML visualizer.
func RenderHTML(graph *model.Graph, outputPath string) error {
	graphBytes, err := json.MarshalIndent(graph, "    ", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize graph data for template: %w", err)
	}

	tmpl, err := template.New("graph").Parse(htmlTemplateContent)
	if err != nil {
		return fmt.Errorf("failed to parse HTML graph template: %w", err)
	}

	data := templateData{
		Account:     graph.Account,
		AccountType: graph.AccountType,
		Stats:       graph.Stats,
		GraphJSON:   template.JS(string(graphBytes)),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute HTML graph template: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write HTML visualizer file: %w", err)
	}

	return nil
}
