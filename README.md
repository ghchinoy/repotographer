# repotographer 🗺️

**GitHub Repository Cartographer & Taxonomy Explorer (CLI + MCP Server)**

`repotographer` maps the public GitHub repositories of any user or organization into structured, interactive concept graphs and AI-suggested domain taxonomies.

Available as both a **standalone CLI tool** and an **MCP (Model Context Protocol) stdio server**.

---

## ✨ Features

- **GitHub Ecosystem Analysis:** Fetches public repositories using GitHub CLI (`gh`), automatically filtering active, non-forked codebases.
- **Technology Hub Detection:** Detects shared technology stacks (MCP, A2A, WASM, AI runtimes, Speech/Audio, Go, Rust, Python, TypeScript) and links them to repos.
- **Tri-State Connectivity:** Classifies nodes into `connected` (linked to shared tech), `standalone` (independent with metadata), and `bare` (minimal metadata).
- **AI-Assisted Domain Taxonomy:** Leverages Google Gen AI SDK (`gemini-3.7-flash` or Vertex AI) to cluster repositories into cohesive thematic pillars.
- **Curator-in-the-Loop Workflow:** Emits `taxonomy.json` as an editable file for custom human refinement and re-rendering.
- **Multi-Format Visual Output:**
  - 🌐 **Self-Contained Interactive HTML:** Cytoscape.js + fcose web visualizer with collapsible domain containers, standalone sub-buckets, filter chips, and an inspector drawer.
  - 📊 **Graphviz DOT:** Clustered digraphs formatted by domain and connectivity.
  - 🖼️ **Static PNG Image:** High-resolution diagram generated automatically when Graphviz is available.
  - 📄 **Structured JSON:** Standardized graph dataset.
- **MCP Server:** Native support for LLM agents to map repositories and render diagrams.

---

## 📋 Prerequisites

1. **GitHub CLI (`gh`) [Required]**
   `repotographer` uses `gh` to fetch public repositories.
   ```bash
   brew install gh
   gh auth login
   ```
   > *Note: `gh` must be authenticated with GitHub before running `repotographer`.*

2. **Graphviz (`dot`) [Optional - for PNG rendering]**
   If Graphviz is installed on your PATH, `repotographer` will automatically generate `.png` images. If absent, it gracefully outputs `.dot` and interactive `.html` without errors.
   ```bash
   brew install graphviz
   ```

3. **Gemini API Key or Google Cloud Vertex AI [Optional - for AI taxonomy]**
   - **Gemini Developer API (Default):**
     ```bash
     export GEMINI_API_KEY="your-api-key"
     ```
   - **Google Cloud Vertex AI:**
     ```bash
     export GOOGLE_GENAI_USE_VERTEXAI="true"
     export GOOGLE_CLOUD_PROJECT="your-project-id"
     export GOOGLE_CLOUD_LOCATION="global"
     ```

---

## 🚀 Installation

### From Source (Go 1.23+)
```bash
git clone https://github.com/ghchinoy/repotographer.git
cd repotographer
go install .
```

---

## 💻 CLI Usage

### 1. Complete Map & Visualize (One-Step)
Fetches repos, clusters tech hubs, asks Gemini for domain taxonomy, and renders all formats:

```bash
repotographer map ghchinoy --out ./out
```

**Output in `./out`:**
- `graph.html` — Interactive visualizer (double-click to open in any browser)
- `graph.png` — High-resolution static image
- `graph.dot` — Graphviz source file
- `graph.json` — Machine-readable concept graph
- `taxonomy.json` — Human-editable domain taxonomy

#### Offline / No-LLM Mode:
```bash
repotographer map ghchinoy --llm=false --out ./out
```

#### Use Vertex AI Backend:
```bash
repotographer map google --type org --vertex --out ./out-google
```

---

### 2. Suggest & Curate Taxonomy
If you want to review or hand-edit the domain taxonomy before generating final diagrams:

**Step 1: Generate AI taxonomy suggestion:**
```bash
repotographer suggest ghchinoy --out taxonomy.json
```

**Step 2: Inspect and hand-tune `taxonomy.json`:**
```json
{
  "domains": [
    {
      "id": "domain_agents",
      "label": "AI Agents & MCP Protocols",
      "blurb": "Multi-agent coordination and protocol servers"
    }
  ],
  "assignments": {
    "repo_my_project": "domain_agents"
  }
}
```

**Step 3: Re-render with your curated taxonomy:**
```bash
repotographer render ./out/graph.json --taxonomy taxonomy.json --out ./out
```

---

## 🤖 MCP (Model Context Protocol) Server

`repotographer` includes a built-in MCP server over `stdio` implementing the official MCP specification.

### Configure with `opencode`, Claude Desktop, or Cursor

Add to your MCP configuration (e.g. `claude_desktop_config.json` or `opencode.json`):

```json
{
  "mcpServers": {
    "repotographer": {
      "command": "repotographer",
      "args": ["mcp"],
      "env": {
        "GEMINI_API_KEY": "your-gemini-api-key"
      }
    }
  }
}
```

### Exposed MCP Tools

| Tool | Parameters | Description |
|---|---|---|
| `map_github_account` | `owner` (string, required)<br>`account_type` ("user" \| "org")<br>`limit` (int)<br>`use_llm` (bool)<br>`model` (string)<br>`use_vertex` (bool) | Fetches GitHub repositories, maps tech connections, derives taxonomy, and returns graph + taxonomy data structures. |
| `suggest_taxonomy` | `owner` (string, required)<br>`account_type` ("user" \| "org")<br>`model` (string)<br>`use_vertex` (bool) | Analyzes repositories and returns a proposed domain pillar taxonomy with project assignments. |
| `render_graph` | `graph` (object, required)<br>`out_dir` (string)<br>`formats` (array: "html", "dot", "png", "json") | Renders a concept graph to disk and returns the generated file paths. |

---

## 🏛️ Architecture & Connectivity Model

`repotographer` classifies repository nodes into **three connectivity tiers (Definition B)**:
1. **`connected`** (Solid border, primary glow): Repositories that link to at least one detected technology hub.
2. **`standalone`** (Dashed border, muted): Standalone repositories with descriptive signal (topics, description, or stack tags). In the interactive HTML visualizer, these are isolated into interactive `⊕ Standalone · N` sub-buckets to prevent initial clutter.
3. **`bare`** (Dotted border, subtle): Minimal repositories with no descriptions or topics.

---

## 📄 License

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.
