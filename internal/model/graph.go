package model

import "time"

// Repo represents a raw repository fetched from GitHub CLI.
type Repo struct {
	Name             string     `json:"name"`
	NameWithOwner    string     `json:"nameWithOwner"`
	Description      string     `json:"description"`
	URL              string     `json:"url"`
	HomepageURL      string     `json:"homepageUrl"`
	PrimaryLanguage  *Language  `json:"primaryLanguage,omitempty"`
	RepositoryTopics []Topic    `json:"repositoryTopics,omitempty"`
	StargazerCount   int        `json:"stargazerCount"`
	PushedAt         *time.Time `json:"pushedAt,omitempty"`
	CreatedAt        *time.Time `json:"createdAt,omitempty"`
	IsFork           bool       `json:"isFork"`
	IsArchived       bool       `json:"isArchived"`
	IsPrivate        bool       `json:"isPrivate"`
}

// Language represents GitHub primaryLanguage.
type Language struct {
	Name string `json:"name"`
}

// Topic represents GitHub repositoryTopics item.
type Topic struct {
	Name string `json:"name"`
}

// Node represents a node in the concept graph (repo or tech hub).
type Node struct {
	ID           string   `json:"id"`
	Label        string   `json:"label"`
	Parent       string   `json:"parent,omitempty"`
	Type         string   `json:"type"` // "repo" or "tech"
	Kind         string   `json:"kind,omitempty"`
	URL          string   `json:"url,omitempty"`
	Homepage     string   `json:"homepage,omitempty"`
	Desc         string   `json:"desc,omitempty"`
	Stars        int      `json:"stars,omitempty"`
	Language     string   `json:"language,omitempty"`
	Topics       []string `json:"topics,omitempty"`
	Stack        []string `json:"stack,omitempty"`
	IsFork       bool     `json:"isFork,omitempty"`
	IsArchived   bool     `json:"isArchived,omitempty"`
	Status       string   `json:"status,omitempty"`
	Source       string   `json:"source,omitempty"`
	PushedAt     string   `json:"pushedAt,omitempty"`
	CreatedAt    string   `json:"createdAt,omitempty"`
	Connectivity string   `json:"connectivity,omitempty"` // "connected", "standalone", "bare"
	Locked       bool     `json:"locked,omitempty"`
}

// Edge represents a directed relation between two nodes.
type Edge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
	Label  string `json:"label,omitempty"`
}

// Domain represents a higher-level thematic pillar.
type Domain struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Blurb string `json:"blurb"`
}

// Tech represents a technology hub node.
type Tech struct {
	ID    string   `json:"id"`
	Label string   `json:"label"`
	Kind  string   `json:"kind"`
	Stack []string `json:"stack"`
}

// Stats provides summary metrics of the concept graph.
type Stats struct {
	TotalGithubRepos int `json:"totalGithubRepos"`
	IncludedRepos    int `json:"includedRepos"`
	DomainCount      int `json:"domainCount"`
	TechCount        int `json:"techCount"`
	EdgeCount        int `json:"edgeCount"`
	ConnectedCount   int `json:"connectedCount"`
	StandaloneCount  int `json:"standaloneCount"`
	BareCount        int `json:"bareCount"`
}

// Graph is the full serialization format compatible with ghc.wtf concept-graph components.
type Graph struct {
	Generated   string   `json:"generated"`
	Topic       string   `json:"topic"`
	Account     string   `json:"account"`
	AccountType string   `json:"accountType"` // "user" or "org"
	Stats       Stats    `json:"stats"`
	Domains     []Domain `json:"domains"`
	Tech        []Tech   `json:"tech"`
	Nodes       []Node   `json:"nodes"`
	Edges       []Edge   `json:"edges"`
}

// Taxonomy represents the domain structure and repository assignments for curation.
type Taxonomy struct {
	Generated   string            `json:"generated"`
	Account     string            `json:"account"`
	Domains     []Domain          `json:"domains"`
	Assignments map[string]string `json:"assignments"` // repo_id -> domain_id
}
