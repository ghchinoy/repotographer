package taxonomy

import (
	"os"
	"testing"
)

func TestResolveGCPProject(t *testing.T) {
	// Clean env
	os.Unsetenv("GOOGLE_CLOUD_PROJECT")
	os.Unsetenv("PROJECT_ID")
	os.Unsetenv("GCP_PROJECT")

	// 1. Explicit option wins
	if got := ResolveGCPProject("my-opt-project"); got != "my-opt-project" {
		t.Errorf("expected 'my-opt-project', got '%s'", got)
	}

	// 2. Fallback to GOOGLE_CLOUD_PROJECT
	os.Setenv("GOOGLE_CLOUD_PROJECT", "proj-google-cloud")
	if got := ResolveGCPProject(""); got != "proj-google-cloud" {
		t.Errorf("expected 'proj-google-cloud', got '%s'", got)
	}

	// 3. Fallback to PROJECT_ID when GOOGLE_CLOUD_PROJECT unset
	os.Unsetenv("GOOGLE_CLOUD_PROJECT")
	os.Setenv("PROJECT_ID", "proj-project-id")
	if got := ResolveGCPProject(""); got != "proj-project-id" {
		t.Errorf("expected 'proj-project-id', got '%s'", got)
	}

	// 4. Fallback to GCP_PROJECT when others unset
	os.Unsetenv("PROJECT_ID")
	os.Setenv("GCP_PROJECT", "proj-gcp-project")
	if got := ResolveGCPProject(""); got != "proj-gcp-project" {
		t.Errorf("expected 'proj-gcp-project', got '%s'", got)
	}

	// Clean up
	os.Unsetenv("GCP_PROJECT")
}

func TestResolveLLMConfigPrecedence(t *testing.T) {
	os.Unsetenv("GEMINI_API_KEY")
	os.Unsetenv("GOOGLE_CLOUD_PROJECT")
	os.Unsetenv("PROJECT_ID")
	os.Unsetenv("GCP_PROJECT")
	os.Unsetenv("GOOGLE_GENAI_USE_VERTEXAI")

	// Case 1: Disabled
	cfg, err := ResolveLLMConfig(LLMOptions{Enabled: false})
	if err != nil || cfg.Enabled || cfg.Backend != BackendNone {
		t.Errorf("expected disabled config, got %+v, err: %v", cfg, err)
	}

	// Case 2: No credentials -> BackendNone, Enabled: false
	cfg, err = ResolveLLMConfig(LLMOptions{Enabled: true})
	if err != nil || cfg.Enabled || cfg.Backend != BackendNone {
		t.Errorf("expected no backend without credentials, got %+v, err: %v", cfg, err)
	}

	// Case 3: Explicit API key -> BackendGemini
	cfg, err = ResolveLLMConfig(LLMOptions{Enabled: true, APIKey: "test-api-key"})
	if err != nil || !cfg.Enabled || cfg.Backend != BackendGemini || cfg.APIKey != "test-api-key" {
		t.Errorf("expected Gemini backend with API key, got %+v, err: %v", cfg, err)
	}

	// Case 4: Ambient PROJECT_ID with no API key -> Auto Vertex
	os.Setenv("PROJECT_ID", "ambient-project-123")
	cfg, err = ResolveLLMConfig(LLMOptions{Enabled: true})
	if err != nil || !cfg.Enabled || cfg.Backend != BackendVertex || !cfg.AutoVertex || cfg.Project != "ambient-project-123" {
		t.Errorf("expected Auto Vertex backend with ambient project, got %+v, err: %v", cfg, err)
	}

	// Case 5: Ambient GEMINI_API_KEY takes precedence over ambient project unless explicit vertex
	os.Setenv("GEMINI_API_KEY", "ambient-gemini-key")
	cfg, err = ResolveLLMConfig(LLMOptions{Enabled: true})
	if err != nil || !cfg.Enabled || cfg.Backend != BackendGemini || cfg.APIKey != "ambient-gemini-key" {
		t.Errorf("expected Gemini backend to win over ambient project, got %+v, err: %v", cfg, err)
	}

	// Case 6: Explicit --vertex flag forces Vertex even if GEMINI_API_KEY present
	cfg, err = ResolveLLMConfig(LLMOptions{Enabled: true, UseVertex: true})
	if err != nil || !cfg.Enabled || cfg.Backend != BackendVertex || cfg.AutoVertex || cfg.Project != "ambient-project-123" {
		t.Errorf("expected explicit Vertex to win over API key, got %+v, err: %v", cfg, err)
	}

	// Clean up
	os.Unsetenv("GEMINI_API_KEY")
	os.Unsetenv("PROJECT_ID")
}
