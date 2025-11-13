package models

import "time"

// Commit represents a git commit with its changes
type Commit struct {
	Hash      string
	Author    string
	Date      time.Time
	Message   string
	Files     []string
	Diff      string
	Summary   string
}

// CommitSummary is the LLM-generated summary of a commit
type CommitSummary struct {
	CommitHash  string
	Title       string
	Description string
	Category    string // e.g., "feature", "bugfix", "docs", "refactor"
}

// ReleaseNotes represents the final release notes
type ReleaseNotes struct {
	Version    string
	Date       time.Time
	FromRef    string
	ToRef      string
	Summaries  []CommitSummary
	Content    string // Rendered markdown
}

// Config represents the application configuration
type Config struct {
	LLM LLMConfig
	Git GitConfig
	Template TemplateConfig
}

// LLMConfig holds LLM provider settings
type LLMConfig struct {
	Provider string // "openai", "anthropic", "other"
	Model    string
	ApiKey   string
	BaseURL  string // optional, for custom endpoints
}

// GitConfig holds git-related settings
type GitConfig struct {
	FromRef string
	ToRef   string
}

// TemplateConfig holds template-related settings
type TemplateConfig struct {
	Path string
}
