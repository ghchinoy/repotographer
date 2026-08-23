# Agent Instructions

Quick-reference guidelines for developing, testing, documenting, and releasing `repotographer`.

## 1. Project Overview

`repotographer` is a Go CLI and Model Context Protocol (MCP) server that maps public GitHub repositories into interactive concept graphs and AI-suggested domain taxonomies.

- **Stack:** Go 1.25+, Cobra CLI, MCP Go SDK, Google GenAI SDK.
- **Outputs:** `graph.{html,png,dot,json}` and `taxonomy.json`.
- **Key Dependencies:** `gh` CLI (authenticated) is required; Graphviz `dot` is optional (PNG generation skips gracefully if missing).
- **LLM Support:** Gemini or Vertex AI. If credentials are missing, it falls back to deterministic keyword heuristics.

---

## 2. Build and Quality Gates

Run the standard check suite before committing any code changes:

```bash
make test && make vet && make build && ./bin/repotographer version
```

- `make test` runs all package tests.
- `make vet` runs Go vet analysis.
- `make build` compiles `./bin/repotographer`.

---

## 3. Documentation Site

The documentation site is built with Astro Starlight and Catppuccin theme in `docs-site/`.

- **Base path:** `/repotographer` (hosted on GitHub Pages at `https://ghchinoy.github.io/repotographer/`).
- **File format:** Use `.mdx` for all documentation pages in `docs-site/src/content/docs/`.
- **Local build verification:**
  ```bash
  cd docs-site && npm run build
  ```
- **Deployment:** Automatically built and published by `.github/workflows/deploy-docs.yml` on pushes to `main`.

---

## 4. Editorial and Prose Standards

All written documentation, skill guides, and README updates must meet strict editorial standards:

- **Quality floor:** Docstats Axis B score of 10.0 with 0 AI-tell flags and 0 em dashes.
- **Verification tool:** Run docstats from `~/projects/docstats`:
  ```bash
  uv run python -c '
  import anyio
  from fastapi_app import analyze_fastapi
  from models import TextSourceModel

  async def check():
      with open("/path/to/file.mdx", "r") as f:
          res = await analyze_fastapi(TextSourceModel(text=f.read()))
      print("AI Score:", res.ai_patterns.ai_tell_score, "Flags:", res.ai_patterns.flags)

  anyio.run(check)
  '
  ```

---

## 5. Release Process (GoReleaser)

Releases are automated via GoReleaser and GitHub Actions.

### Version Synchronization
When bumping versions, update all four locations together:
1. `internal/cli/run.go` (`var Version = "X.Y.Z"`)
2. `plugins/repotographer/plugin.json` (`"version": "X.Y.Z"`)
3. `plugins/repotographer/skills/repotographer/SKILL.md` (`version: "X.Y.Z"`)
4. `docs-site/src/content/docs/getting-started/installation.mdx` (version check output)

### Release Steps
1. Ensure all tests and doc builds pass cleanly.
2. Commit changes using Conventional Commits (`feat:`, `fix:`, `docs:`, `ci:`).
3. Validate GoReleaser locally (optional):
   ```bash
   goreleaser check
   goreleaser release --snapshot --clean --skip=publish
   ```
4. Tag and push the release:
   ```bash
   git tag -a vX.Y.Z -m "Release vX.Y.Z"
   git push origin vX.Y.Z
   ```
5. `.github/workflows/release.yml` builds multi-platform binaries (Darwin arm64/x86_64, Linux arm64/x86_64, Windows x86_64), generates checksums, and publishes the GitHub Release.

---

## 6. Issue Tracking

- Use GitHub Issues (`gh issue`) for bug reports and feature requests.
- Reference issue numbers in commit messages (for example: `fix(render): resolve DOMContentLoaded race condition (fixes #5)`).
- Close resolved issues with a reference to the fixing commit hash.
