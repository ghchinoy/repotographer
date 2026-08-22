package cluster

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ghchinoy/repotographer/internal/model"
)

// Standard tech signatures for detection
type TechSignature struct {
	ID       string
	Label    string
	Kind     string // "language", "protocol", "framework", "platform"
	Keywords []string
	Stack    []string
	MinCount int
}

var DefaultSignatures = []TechSignature{
	{ID: "tech_mcp", Label: "Model Context Protocol (MCP)", Kind: "protocol", Keywords: []string{"mcp", "model context protocol", "model-context-protocol"}, Stack: []string{"ai", "mcp"}},
	{ID: "tech_a2a", Label: "A2A Protocol", Kind: "protocol", Keywords: []string{"a2a", "agent-to-agent", "agent to agent"}, Stack: []string{"ai", "a2a"}},
	{ID: "tech_wasm", Label: "WebAssembly (WASM)", Kind: "platform", Keywords: []string{"wasm", "webassembly", "wasi"}, Stack: []string{"systems", "wasm"}},
	{ID: "tech_ai", Label: "AI & LLM Runtimes", Kind: "framework", Keywords: []string{"llm", "gemini", "gemma", "openai", "agent", "genai", "prompt", "mlx", "rag", "transformer"}, Stack: []string{"ai"}},
	{ID: "tech_audio", Label: "Voice & Speech Audio", Kind: "domain", Keywords: []string{"audio", "speech", "tts", "voice", "vox", "whisper", "kokoro", "stt", "sound"}, Stack: []string{"audio"}},
	{ID: "tech_cli", Label: "CLIs & TUIs", Kind: "framework", Keywords: []string{"cli", "tui", "bubbletea", "bubble tea", "cobra", "terminal", "command-line"}, Stack: []string{"cli"}},
	{ID: "tech_go", Label: "Go", Kind: "language", Keywords: []string{"go", "golang"}, Stack: []string{"go"}},
	{ID: "tech_rust", Label: "Rust", Kind: "language", Keywords: []string{"rust", "cargo"}, Stack: []string{"rust"}},
	{ID: "tech_python", Label: "Python", Kind: "language", Keywords: []string{"python", "pytorch", "pip", "fastapi"}, Stack: []string{"python"}},
	{ID: "tech_typescript", Label: "TypeScript / JavaScript", Kind: "language", Keywords: []string{"typescript", "javascript", "nodejs", "react", "lit", "vue", "nextjs", "web component"}, Stack: []string{"frontend", "javascript"}},
}

// ComputeConnectivity calculates Definition B connectivity state.
func ComputeConnectivity(hasEdges bool, desc string, topics []string, stack []string) string {
	if hasEdges {
		return "connected"
	}
	hasDesc := len(strings.TrimSpace(desc)) > 5
	hasTopics := len(topics) > 0

	// Meaningful stack tags (excluding generic language tags)
	meaningfulStackCount := 0
	for _, s := range stack {
		if s != "go" && s != "rust" && s != "javascript" && s != "typescript" && s != "python" {
			meaningfulStackCount++
		}
	}

	if hasDesc || hasTopics || meaningfulStackCount > 0 {
		return "standalone"
	}
	return "bare"
}

// DetectionResult holds stack tags and matched tech hubs for a repository.
type DetectionResult struct {
	Stack     []string
	TechEdges []model.Edge
}

// DetectStackAndEdges identifies technology stack tags and connects repos to tech hubs.
func DetectStackAndEdges(repo model.Repo, nodeID string, activeTechMap map[string]model.Tech) DetectionResult {
	name := strings.ToLower(repo.Name)
	desc := strings.ToLower(repo.Description)
	var topicNames []string
	for _, t := range repo.RepositoryTopics {
		topicNames = append(topicNames, strings.ToLower(t.Name))
	}
	lang := ""
	if repo.PrimaryLanguage != nil {
		lang = strings.ToLower(repo.PrimaryLanguage.Name)
	}

	fullText := fmt.Sprintf("%s %s %s %s", name, desc, strings.Join(topicNames, " "), lang)

	stackSet := make(map[string]bool)
	if lang != "" {
		stackSet[lang] = true
	}

	var edges []model.Edge

	for _, sig := range DefaultSignatures {
		matched := false
		// Check primary language match
		if sig.Kind == "language" && lang != "" {
			for _, kw := range sig.Keywords {
				if lang == kw {
					matched = true
					break
				}
			}
		}

		// Check topics and text
		if !matched {
			for _, kw := range sig.Keywords {
				// Exact topic match
				for _, t := range topicNames {
					if t == kw {
						matched = true
						break
					}
				}
				if matched {
					break
				}

				// Word-boundary check in full text
				if strings.Contains(fullText, kw) {
					matched = true
					break
				}
			}
		}

		if matched {
			for _, s := range sig.Stack {
				stackSet[s] = true
			}
			if _, exists := activeTechMap[sig.ID]; exists {
				edges = append(edges, model.Edge{
					ID:     fmt.Sprintf("%s->%s", nodeID, sig.ID),
					Source: nodeID,
					Target: sig.ID,
					Label:  "implements",
				})
			}
		}
	}

	var stack []string
	for s := range stackSet {
		stack = append(stack, s)
	}
	sort.Strings(stack)

	return DetectionResult{
		Stack:     stack,
		TechEdges: edges,
	}
}

// BuildDeterministicGraph constructs an initial concept graph and taxonomy from raw repos.
func BuildDeterministicGraph(account string, accountType string, rawRepos []model.Repo, minSignal bool) (*model.Graph, *model.Taxonomy) {
	// 1. Identify which tech hubs are actually relevant in this dataset
	techFrequency := make(map[string]int)
	for _, repo := range rawRepos {
		res := DetectStackAndEdges(repo, "temp", nil)
		for _, s := range res.Stack {
			techFrequency[s]++
		}
	}

	activeTechMap := make(map[string]model.Tech)
	var activeTechList []model.Tech
	for _, sig := range DefaultSignatures {
		// Include tech hub if at least 1 repo matches
		count := 0
		for _, s := range sig.Stack {
			count += techFrequency[s]
		}
		if count > 0 {
			tech := model.Tech{
				ID:    sig.ID,
				Label: sig.Label,
				Kind:  sig.Kind,
				Stack: sig.Stack,
			}
			activeTechMap[sig.ID] = tech
			activeTechList = append(activeTechList, tech)
		}
	}

	// 2. Pre-cluster repos into default heuristic domains
	defaultDomains := []model.Domain{
		{ID: "domain_agents", Label: "AI Agents & Protocols", Blurb: "Autonomous agent architectures, MCP tooling, and protocol implementations"},
		{ID: "domain_systems", Label: "Systems & Runtimes", Blurb: "High-performance runtimes, WebAssembly, and low-level system software"},
		{ID: "domain_audio", Label: "Audio & Multimodal", Blurb: "Speech synthesis, audio processing, and multimodal live streaming"},
		{ID: "domain_cli", Label: "Developer Tools & CLIs", Blurb: "Terminal user interfaces, command-line utilities, and cloud infrastructure"},
		{ID: "domain_web", Label: "Web & Frontend", Blurb: "Web applications, UI components, and client-side interfaces"},
		{ID: "domain_general", Label: "Core & Utility Libraries", Blurb: "General purpose software, experimental tooling, and foundational libraries"},
	}

	var nodes []model.Node
	var edges []model.Edge
	assignments := make(map[string]string)

	// Add tech hub nodes
	for _, tech := range activeTechList {
		nodes = append(nodes, model.Node{
			ID:     tech.ID,
			Label:  tech.Label,
			Kind:   tech.Kind,
			Type:   "tech",
			Stack:  tech.Stack,
			Source: "system",
		})
	}

	connectedCount := 0
	standaloneCount := 0
	bareCount := 0
	includedRepos := 0

	for _, repo := range rawRepos {
		if repo.IsFork || repo.IsPrivate || repo.IsArchived {
			continue
		}

		hasTopics := len(repo.RepositoryTopics) > 0
		hasDesc := len(strings.TrimSpace(repo.Description)) > 5
		hasLang := repo.PrimaryLanguage != nil && repo.PrimaryLanguage.Name != ""
		hasStars := repo.StargazerCount > 0

		if minSignal && !hasTopics && !hasDesc && !hasLang && !hasStars {
			continue
		}

		cleanName := strings.ReplaceAll(repo.Name, ".", "_")
		cleanName = strings.ReplaceAll(cleanName, "-", "_")
		nodeID := fmt.Sprintf("repo_%s", cleanName)

		detection := DetectStackAndEdges(repo, nodeID, activeTechMap)

		// Determine domain assignment
		domainID := assignDeterministicDomain(repo, detection.Stack)
		assignments[nodeID] = domainID

		hasEdges := len(detection.TechEdges) > 0
		var topicStrings []string
		for _, t := range repo.RepositoryTopics {
			topicStrings = append(topicStrings, t.Name)
		}

		connectivity := ComputeConnectivity(hasEdges, repo.Description, topicStrings, detection.Stack)
		switch connectivity {
		case "connected":
			connectedCount++
		case "standalone":
			standaloneCount++
		case "bare":
			bareCount++
		}

		pushedStr := ""
		if repo.PushedAt != nil {
			pushedStr = repo.PushedAt.Format(time.RFC3339)
		}
		createdStr := ""
		if repo.CreatedAt != nil {
			createdStr = repo.CreatedAt.Format(time.RFC3339)
		}

		langName := ""
		if repo.PrimaryLanguage != nil {
			langName = repo.PrimaryLanguage.Name
		}

		node := model.Node{
			ID:           nodeID,
			Label:        repo.Name,
			Parent:       domainID,
			Type:         "repo",
			URL:          repo.URL,
			Homepage:     repo.HomepageURL,
			Desc:         repo.Description,
			Stars:        repo.StargazerCount,
			Language:     langName,
			Topics:       topicStrings,
			Stack:        detection.Stack,
			IsFork:       repo.IsFork,
			IsArchived:   repo.IsArchived,
			Status:       "active",
			Source:       "auto",
			PushedAt:     pushedStr,
			CreatedAt:    createdStr,
			Connectivity: connectivity,
		}

		nodes = append(nodes, node)
		edges = append(edges, detection.TechEdges...)
		includedRepos++
	}

	// Filter out unused domains
	domainRepoCount := make(map[string]int)
	for _, domainID := range assignments {
		domainRepoCount[domainID]++
	}
	var usedDomains []model.Domain
	for _, d := range defaultDomains {
		if domainRepoCount[d.ID] > 0 {
			usedDomains = append(usedDomains, d)
		}
	}

	stats := model.Stats{
		TotalGithubRepos: len(rawRepos),
		IncludedRepos:    includedRepos,
		DomainCount:      len(usedDomains),
		TechCount:        len(activeTechList),
		EdgeCount:        len(edges),
		ConnectedCount:   connectedCount,
		StandaloneCount:  standaloneCount,
		BareCount:        bareCount,
	}

	nowStr := time.Now().UTC().Format(time.RFC3339)

	graph := &model.Graph{
		Generated:   nowStr,
		Topic:       "repos",
		Account:     account,
		AccountType: accountType,
		Stats:       stats,
		Domains:     usedDomains,
		Tech:        activeTechList,
		Nodes:       nodes,
		Edges:       edges,
	}

	taxonomy := &model.Taxonomy{
		Generated:   nowStr,
		Account:     account,
		Domains:     usedDomains,
		Assignments: assignments,
	}

	return graph, taxonomy
}

func assignDeterministicDomain(repo model.Repo, stack []string) string {
	name := strings.ToLower(repo.Name)
	desc := strings.ToLower(repo.Description)
	var topics []string
	for _, t := range repo.RepositoryTopics {
		topics = append(topics, strings.ToLower(t.Name))
	}
	fullText := fmt.Sprintf("%s %s %s", name, desc, strings.Join(topics, " "))

	stackSet := make(map[string]bool)
	for _, s := range stack {
		stackSet[s] = true
	}

	if stackSet["mcp"] || stackSet["a2a"] || strings.Contains(fullText, "agent") || strings.Contains(fullText, "delegation") || strings.Contains(fullText, "skills") {
		return "domain_agents"
	}
	if stackSet["audio"] || strings.Contains(fullText, "speech") || strings.Contains(fullText, "voice") || strings.Contains(fullText, "tts") {
		return "domain_audio"
	}
	if stackSet["wasm"] || stackSet["rust"] || strings.Contains(fullText, "runtime") || strings.Contains(fullText, "compiler") || strings.Contains(fullText, "codec") {
		return "domain_systems"
	}
	if stackSet["cli"] || strings.Contains(fullText, "tui") || strings.Contains(fullText, "scanner") || strings.Contains(fullText, "tool") {
		return "domain_cli"
	}
	if stackSet["frontend"] || strings.Contains(fullText, "web") || strings.Contains(fullText, "ui") || strings.Contains(fullText, "component") {
		return "domain_web"
	}

	return "domain_general"
}
