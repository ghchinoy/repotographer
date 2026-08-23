---
title: Connectivity & Detection Model
description: How repotographer classifies repository connectivity and detects technology hubs.
---

`repotographer` applies a deterministic classification model to evaluate repository connectivity and identify shared technology hubs across a portfolio.

## Tri-State Connectivity (Definition B)

Every repository node is assigned to one of three connectivity tiers:

| Connectivity Tier | Visual Style | Definition & Criteria |
|---|---|---|
| **`connected`** | Solid blue border, full opacity | Has at least one graph edge connecting to a detected technology hub. |
| **`standalone`** | Dashed gray border, slight opacity | Has no outgoing graph edges, but contains descriptive metadata (description longer than 5 characters, topics, or stack tags). |
| **`bare`** | Dotted border, dimmed opacity | Contains minimal metadata with no edges, topics, or non-trivial descriptions. |

---

## Standalone Sub-Bucket Isolation

In large portfolios, dozens of standalone repositories can overwhelm a visual graph. `repotographer` isolates unlinked nodes in interactive sub-buckets within the HTML visualizer:

1. When a domain is collapsed, child nodes remain hidden.
2. Expanding a domain renders only its `connected` repositories.
3. Standalone repositories remain inside an interactive `⊕ Standalone · N` pill node.
4. Clicking the sub-bucket pill unpacks and tiles the standalone repositories inside that domain container on demand.

---

## Technology Hub Signatures

`repotographer` scans repository names, descriptions, languages, and topics for standard technology patterns:

| Technology Hub | Node ID | Kind | Trigger Keywords & Stack |
|---|---|---|---|
| **Model Context Protocol** | `tech_mcp` | Protocol | `mcp`, `model context protocol` |
| **A2A Protocol** | `tech_a2a` | Protocol | `a2a`, `agent-to-agent` |
| **WebAssembly** | `tech_wasm` | Platform | `wasm`, `webassembly`, `wasi` |
| **AI & LLM Runtimes** | `tech_ai` | Framework | `llm`, `gemini`, `gemma`, `openai`, `agent`, `genai`, `prompt`, `mlx`, `rag` |
| **Voice & Speech Audio** | `tech_audio` | Domain | `audio`, `speech`, `tts`, `voice`, `vox`, `whisper`, `kokoro`, `stt` |
| **CLIs & TUIs** | `tech_cli` | Framework | `cli`, `tui`, `bubbletea`, `cobra`, `terminal` |
| **Go** | `tech_go` | Language | Primary language `Go`, or `golang` in topics/text |
| **Rust** | `tech_rust` | Language | Primary language `Rust`, or `cargo` in topics/text |
| **Python** | `tech_python` | Language | Primary language `Python`, `pytorch`, `fastapi` |
| **TypeScript / JavaScript** | `tech_typescript` | Language | `typescript`, `javascript`, `nodejs`, `react`, `lit`, `vue`, `web component` |

A technology hub node is added to the graph whenever at least one repository matches its signature.
