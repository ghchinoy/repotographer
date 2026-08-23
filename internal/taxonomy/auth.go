package taxonomy

import (
	"fmt"
	"os"
	"strings"
)

// BackendType represents the LLM provider backend.
type BackendType string

const (
	BackendNone   BackendType = "none"
	BackendGemini BackendType = "gemini"
	BackendVertex BackendType = "vertex"
)

// ResolvedLLMConfig contains the resolved authentication and backend parameters.
type ResolvedLLMConfig struct {
	Enabled    bool
	Backend    BackendType
	APIKey     string
	Project    string
	Location   string
	Model      string
	AutoVertex bool
}

// ResolveGCPProject returns the first non-empty GCP project ID from options or standard env vars.
// Checks: opts.Project -> GOOGLE_CLOUD_PROJECT -> PROJECT_ID -> GCP_PROJECT.
func ResolveGCPProject(projectOpt string) string {
	if strings.TrimSpace(projectOpt) != "" {
		return strings.TrimSpace(projectOpt)
	}
	for _, env := range []string{"GOOGLE_CLOUD_PROJECT", "PROJECT_ID", "GCP_PROJECT"} {
		if val := strings.TrimSpace(os.Getenv(env)); val != "" {
			return val
		}
	}
	return ""
}

// ResolveGCPLocation returns the first non-empty GCP location/region or "global".
// Checks: opts.Location -> GOOGLE_CLOUD_LOCATION -> GCP_REGION -> "global".
func ResolveGCPLocation(locationOpt string) string {
	if strings.TrimSpace(locationOpt) != "" {
		return strings.TrimSpace(locationOpt)
	}
	for _, env := range []string{"GOOGLE_CLOUD_LOCATION", "GCP_REGION"} {
		if val := strings.TrimSpace(os.Getenv(env)); val != "" {
			return val
		}
	}
	return "global"
}

// ResolveAPIKey returns the Gemini API key from options or GEMINI_API_KEY env var.
func ResolveAPIKey(apiKeyOpt string) string {
	if strings.TrimSpace(apiKeyOpt) != "" {
		return strings.TrimSpace(apiKeyOpt)
	}
	return strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
}

// ResolveLLMConfig determines the active LLM backend and credentials based on explicit options and ambient environment.
func ResolveLLMConfig(opts LLMOptions) (ResolvedLLMConfig, error) {
	if !opts.Enabled {
		return ResolvedLLMConfig{
			Enabled: false,
			Backend: BackendNone,
		}, nil
	}

	modelName := strings.TrimSpace(opts.Model)
	if modelName == "" {
		modelName = "gemini-3.7-flash"
	}

	project := ResolveGCPProject(opts.Project)
	location := ResolveGCPLocation(opts.Location)
	apiKey := ResolveAPIKey(opts.APIKey)

	explicitVertex := opts.UseVertex || strings.ToLower(os.Getenv("GOOGLE_GENAI_USE_VERTEXAI")) == "true"

	if explicitVertex {
		if project == "" {
			return ResolvedLLMConfig{}, fmt.Errorf("Vertex AI was selected, but no GCP project was found. Set GOOGLE_CLOUD_PROJECT, PROJECT_ID, or GCP_PROJECT")
		}
		return ResolvedLLMConfig{
			Enabled:    true,
			Backend:    BackendVertex,
			Project:    project,
			Location:   location,
			Model:      modelName,
			AutoVertex: false,
		}, nil
	}

	// If Gemini API key is available, use Gemini API backend
	if apiKey != "" {
		return ResolvedLLMConfig{
			Enabled:  true,
			Backend:  BackendGemini,
			APIKey:   apiKey,
			Model:    modelName,
			Project:  project,
			Location: location,
		}, nil
	}

	// If no Gemini API key, but a GCP project is present, auto-detect Vertex AI with ADC
	if project != "" {
		return ResolvedLLMConfig{
			Enabled:    true,
			Backend:    BackendVertex,
			Project:    project,
			Location:   location,
			Model:      modelName,
			AutoVertex: true,
		}, nil
	}

	// No credentials configured
	return ResolvedLLMConfig{
		Enabled: false,
		Backend: BackendNone,
		Model:   modelName,
	}, nil
}
