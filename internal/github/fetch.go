package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/ghchinoy/repotographer/internal/model"
)

// FetchOptions configures GitHub repository fetching.
type FetchOptions struct {
	Owner        string
	AccountType  string // "user", "org", or "" (auto)
	Limit        int
	IncludeForks bool
	IncludeArchived bool
}

// CheckPrerequisites verifies that `gh` is installed and authenticated.
func CheckPrerequisites() error {
	ghPath, err := exec.LookPath("gh")
	if err != nil {
		return fmt.Errorf("GitHub CLI (`gh`) is required but not found in PATH. Please install it from https://cli.github.com/")
	}
	_ = ghPath

	cmd := exec.Command("gh", "auth", "status")
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		// Even if auth status returns non-zero, check if stderr mentions logged in
		stderr := errBuf.String()
		if !strings.Contains(stderr, "Logged in to") && !strings.Contains(stderr, "Active account: true") {
			return fmt.Errorf("GitHub CLI is not authenticated. Please run `gh auth login` before running repotographer: %s", stderr)
		}
	}

	return nil
}

// FetchRepos retrieves public repositories for a given user or organization using GitHub CLI.
func FetchRepos(opts FetchOptions) ([]model.Repo, error) {
	if err := CheckPrerequisites(); err != nil {
		return nil, err
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = 500
	}

	args := []string{
		"repo", "list", opts.Owner,
		"--limit", fmt.Sprintf("%d", limit),
		"--visibility", "public",
		"--json", "name,nameWithOwner,description,url,homepageUrl,primaryLanguage,repositoryTopics,stargazerCount,pushedAt,createdAt,isFork,isArchived,isPrivate",
	}

	if !opts.IncludeForks {
		args = append(args, "--source")
	}
	if !opts.IncludeArchived {
		args = append(args, "--no-archived")
	}

	cmd := exec.Command("gh", args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return nil, fmt.Errorf("failed to fetch repositories for %s via `gh`: %s", opts.Owner, errMsg)
	}

	var repos []model.Repo
	if err := json.Unmarshal(stdout.Bytes(), &repos); err != nil {
		return nil, fmt.Errorf("failed to parse GitHub CLI response: %w", err)
	}

	return repos, nil
}
