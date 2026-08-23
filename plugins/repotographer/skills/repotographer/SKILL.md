---
name: repotographer
description: Map public GitHub repositories into structured concept graphs with AI-suggested domain taxonomies, tri-state connectivity analysis, and interactive Cytoscape or Graphviz visual outputs. Use when exploring a user or organization's GitHub portfolio, creating architectural landscape maps, or generating visual repository taxonomies.
license: Apache-2.0
metadata:
  version: "0.2.0"
---

# repotographer

This skill guides coding assistants and developers through mapping, categorizing, and visualizing public GitHub repository ecosystems.

## Available MCP Tools

The `repotographer` MCP server provides three core tools:

1. `map_github_account`: Fetches public repositories, detects technology connections, derives a domain taxonomy, and returns the complete graph and taxonomy objects.
2. `suggest_taxonomy`: Proposes high-level thematic domain pillars and assigns repositories without generating full graph layout geometry.
3. `render_graph`: Renders a concept graph into interactive HTML (`graph.html`), Graphviz DOT (`graph.dot`), and PNG (`graph.png`).

---

## The Four-Phase Workflow

### 1. Portfolio Ingestion & Initial Mapping

When a user asks to analyze or visualize a GitHub user or organization:
- Call `map_github_account` with the target `owner` (and optional `account_type: "user" | "org"`).
- By default, Gemini 3.7 Flash proposes a 3 to 6 pillar domain taxonomy.
- If LLM credentials are absent, the tool automatically falls back to deterministic stack heuristics.

```json
// Example MCP Tool Call
{
  "name": "map_github_account",
  "arguments": {
    "owner": "ghchinoy",
    "use_llm": true
  }
}
```

### 2. Connectivity Model Analysis

Examine the returned nodes against the tri-state connectivity model:
- **`connected`**: Repositories with outgoing edges to detected technology hubs (MCP, WASM, Go, Rust, Audio, Python, TypeScript).
- **`standalone`**: Independent projects containing metadata (descriptions or topics) but no shared hub links.
- **`bare`**: Minimal repositories lacking descriptions, topics, and non-trivial stack tags.

### 3. Human-in-the-Loop Taxonomy Curation

When the user wants to refine domain groupings or tailor their narrative:
1. Save or inspect `taxonomy.json`.
2. Allow the user to adjust domain labels, descriptions, or project assignments.
3. Keep domain IDs unique and prefixed with `domain_` (for example, `domain_agent_systems`).

### 4. Diagram & Visualizer Rendering

Render the final graph to disk:
- Pass the graph structure and output directory to `render_graph`.
- The tool generates self-contained interactive Cytoscape visualizers (`graph.html`) and static Graphviz diagrams (`graph.png`, `graph.dot`).

```json
// Example MCP Tool Call
{
  "name": "render_graph",
  "arguments": {
    "graph": "<GraphObject>",
    "out_dir": "./out",
    "formats": ["html", "dot", "png", "json"]
  }
}
```

---

## CLI Equivalent Commands

When operating in terminal sessions without an active MCP connection:

```bash
# Complete one-step mapping
repotographer map <owner> --out ./out

# Suggest taxonomy only
repotographer suggest <owner> --out taxonomy.json

# Render from curated taxonomy
repotographer render ./out/graph.json --taxonomy taxonomy.json --out ./dist

# Print client configurations
repotographer mcp config
```
