package taxonomy

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ghchinoy/repotographer/internal/model"
	"google.golang.org/genai"
)

// LLMOptions configures Gemini / Vertex AI taxonomy suggestions.
type LLMOptions struct {
	Enabled   bool
	Model     string
	UseVertex bool
	Project   string
	Location  string
	APIKey    string
}

// LLMTaxonomyResponse is the expected structured response from Gemini.
type LLMTaxonomyResponse struct {
	Domains []struct {
		ID    string `json:"id"`
		Label string `json:"label"`
		Blurb string `json:"blurb"`
	} `json:"domains"`
	Assignments []struct {
		RepoID   string `json:"repo_id"`
		DomainID string `json:"domain_id"`
	} `json:"assignments"`
}

// ProposeWithGemini sends repository metadata to Gemini to suggest refined domain pillars and assignments.
func ProposeWithGemini(ctx context.Context, account string, nodes []model.Node, opts LLMOptions) (*model.Taxonomy, error) {
	modelName := opts.Model
	if modelName == "" {
		modelName = "gemini-3.7-flash"
	}

	var client *genai.Client
	var err error

	useVertex := opts.UseVertex || os.Getenv("GOOGLE_GENAI_USE_VERTEXAI") == "true"
	if useVertex {
		project := opts.Project
		if project == "" {
			project = os.Getenv("GOOGLE_CLOUD_PROJECT")
		}
		location := opts.Location
		if location == "" {
			location = os.Getenv("GOOGLE_CLOUD_LOCATION")
		}
		if location == "" {
			location = "global"
		}
		if project == "" {
			return nil, fmt.Errorf("Vertex AI selected but GOOGLE_CLOUD_PROJECT is unset")
		}

		client, err = genai.NewClient(ctx, &genai.ClientConfig{
			Project:  project,
			Location: location,
			Backend:  genai.BackendVertexAI,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create Vertex AI GenAI client: %w", err)
		}
	} else {
		apiKey := opts.APIKey
		if apiKey == "" {
			apiKey = os.Getenv("GEMINI_API_KEY")
		}
		if apiKey == "" {
			return nil, fmt.Errorf("GEMINI_API_KEY environment variable is not set (and --vertex was not specified)")
		}

		client, err = genai.NewClient(ctx, &genai.ClientConfig{
			APIKey:  apiKey,
			Backend: genai.BackendGeminiAPI,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create Gemini API GenAI client: %w", err)
		}
	}

	// Prepare compact repository summaries
	type RepoSummary struct {
		ID       string   `json:"id"`
		Name     string   `json:"name"`
		Desc     string   `json:"desc"`
		Language string   `json:"language"`
		Topics   []string `json:"topics"`
		Stack    []string `json:"stack"`
	}

	var summaries []RepoSummary
	for _, n := range nodes {
		if n.Type != "repo" {
			continue
		}
		summaries = append(summaries, RepoSummary{
			ID:       n.ID,
			Name:     n.Label,
			Desc:     n.Desc,
			Language: n.Language,
			Topics:   n.Topics,
			Stack:    n.Stack,
		})
	}

	summariesJSON, err := json.Marshal(summaries)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal repo summaries: %w", err)
	}

	prompt := fmt.Sprintf(`You are an expert software taxonomy architect.
Analyze the following collection of %d public GitHub repositories owned by '%s' and propose an elegant, cohesive 3 to 6 pillar domain taxonomy that categorizes these projects meaningfully.

Repositories data:
%s

Instructions:
1. Propose between 3 and 6 top-level domains. Each domain must have:
   - "id": a unique identifier starting with "domain_" (e.g. "domain_agents", "domain_audio_systems", "domain_developer_tooling")
   - "label": a concise, compelling title (e.g. "AI Agents & Autonomous Systems", "Audio Processing & Speech Runtimes")
   - "blurb": a 1-sentence description of the core themes and technologies in this pillar
2. Map every repository ID to exactly one of your proposed domain IDs in the "assignments" object.
3. Ensure domain names reflect actual architectural groupings rather than generic catch-alls.
4. Output valid JSON matching this exact structure:
{
  "domains": [
    {"id": "domain_foo", "label": "Domain Label", "blurb": "Pillar description"}
  ],
  "assignments": [
    {"repo_id": "repo_foo", "domain_id": "domain_foo"}
  ]
}`, len(summaries), account, string(summariesJSON))

	responseSchema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"domains": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"id":    {Type: genai.TypeString, Description: "Unique domain ID like domain_agents"},
						"label": {Type: genai.TypeString, Description: "Short domain name"},
						"blurb": {Type: genai.TypeString, Description: "One sentence summary"},
					},
					Required: []string{"id", "label", "blurb"},
				},
			},
			"assignments": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"repo_id":   {Type: genai.TypeString, Description: "The repository node ID"},
						"domain_id": {Type: genai.TypeString, Description: "The assigned domain ID"},
					},
					Required: []string{"repo_id", "domain_id"},
				},
				Description: "List mapping each repository node ID to its assigned domain ID",
			},
		},
		Required: []string{"domains", "assignments"},
	}

	resp, err := client.Models.GenerateContent(ctx, modelName, genai.Text(prompt), &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema:   responseSchema,
		Temperature:      genai.Ptr(float32(0.2)),
	})
	if err != nil {
		return nil, fmt.Errorf("gemini taxonomy generation failed: %w", err)
	}

	respText := strings.TrimSpace(resp.Text())
	if respText == "" {
		return nil, fmt.Errorf("empty response received from Gemini")
	}

	var parsed LLMTaxonomyResponse
	if err := json.Unmarshal([]byte(respText), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse structured taxonomy from Gemini: %w (raw response: %s)", err, respText)
	}

	if len(parsed.Domains) == 0 {
		return nil, fmt.Errorf("gemini returned zero domains")
	}

	var domains []model.Domain
	validDomainIDs := make(map[string]bool)
	for _, d := range parsed.Domains {
		domains = append(domains, model.Domain{
			ID:    d.ID,
			Label: d.Label,
			Blurb: d.Blurb,
		})
		validDomainIDs[d.ID] = true
	}

	// Sanitize assignments
	cleanedAssignments := make(map[string]string)
	fallbackDomain := domains[0].ID
	for _, a := range parsed.Assignments {
		if validDomainIDs[a.DomainID] {
			cleanedAssignments[a.RepoID] = a.DomainID
		} else {
			cleanedAssignments[a.RepoID] = fallbackDomain
		}
	}

	taxonomy := &model.Taxonomy{
		Account:     account,
		Domains:     domains,
		Assignments: cleanedAssignments,
	}

	return taxonomy, nil
}
