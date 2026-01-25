package models

import "time"

// Commit represents a git commit with its changes
type Commit struct {
	Hash    string
	Author  string
	Date    time.Time
	Message string
	Files   []string
	Diff    string
	Summary string
}

type AICommitSummary struct {
	BreakingChanges []string `json:"breaking_changes"`
	Features        []string `json:"feat"`
	BugFixes        []string `json:"fix"`
	Refactor        []string `json:"refactor"`
	Others          []string `json:"other_changes"`
}

// CommitSummary is the LLM-generated summary of a commit
type CommitSummary struct {
	CommitHash string
	Title      string
	Summary    *AICommitSummary
}

// ReleaseNotes represents the final release notes
type ReleaseNotes struct {
	Version   string
	Date      time.Time
	FromRef   string
	ToRef     string
	Summaries []CommitSummary
	Content   string // Rendered markdown
}

// Config represents the application configuration
type Config struct {
	LLM      LLMConfig
	Git      GitConfig
	Template TemplateConfig
}

type LLMProvider string

const (
	ProviderOpenAI    LLMProvider = "openai"
	ProviderAnthropic LLMProvider = "anthropic"
	ProviderMistral   LLMProvider = "mistral"
	ProviderGemini    LLMProvider = "gemini"
	ProviderOllama    LLMProvider = "ollama"
)

// LLMConfig holds LLM provider settings
type LLMConfig struct {
	Provider LLMProvider
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
