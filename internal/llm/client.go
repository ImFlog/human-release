package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ImFlog/human-release/pkg/models"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

// Client handles LLM API calls
type Client struct {
	llm llms.Model
}

// NewClient creates a new LLM client
func NewClient(config *models.LLMConfig) (*Client, error) {
	var llm llms.Model
	var err error

	switch config.Provider {
	case models.ProviderOllama:
		opts := []ollama.Option{ollama.WithModel(config.Model)}
		if config.BaseURL != "" {
			opts = append(opts, ollama.WithServerURL(config.BaseURL))
		}
		llm, err = ollama.New(opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Ollama client: %w", err)
		}
	default:
		// TODO: Implement other providers
		return nil, fmt.Errorf("unsupported provider: %s", config.Provider)
	}
	return &Client{llm: llm}, nil
}

// SummarizeCommit generates a summary for a commit using the LLM
func (c *Client) SummarizeCommit(commit *models.Commit) (*models.AICommitSummary, error) {
	prompt := fmt.Sprintf(`You are an expert software engineer skilled in summarizing code changes.
Use the following commit information to generate a summary and categorize the changes.

Commit Message: %s
Files Changed: %v
Diff Stats:
%s

For each category, list the relevant changes found in the commit. Each change should be a one-liner description.
Respond with the following JSON format:
{
  "breaking_changes": ["...", "..."],
  "features": ["...", "..."],
  "bug_fixes": ["...", "..."],
  "build": ["...", "..."],
  "refactor": ["...", "..."],
  "docs": ["...", "..."],
  "chores": ["...", "..."],
  "tests": ["...", "..."],
  "other_changes": ["...", "..."]
}`, commit.Message, commit.Files, commit.Diff)

	response, err := c.llm.GenerateContent(context.Background(), []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	}, llms.WithJSONMode())

	if err != nil {
		return nil, fmt.Errorf("LLM generation error: %w", err)
	}

	var aiResult models.AICommitSummary
	if err := json.Unmarshal([]byte(response.Choices[0].Content), &aiResult); err != nil {
		return nil, fmt.Errorf("parsing AI response: %w", err)
	}
	return &aiResult, nil
}
