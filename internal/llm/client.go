package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/mistral"

	"github.com/ImFlog/human-release/pkg/models"
)

// Client handles LLM API calls
type Client struct {
	llm llms.Model
}

// NewClient creates a new LLM client
func NewClient(config *models.LLMConfig) (*Client, error) {
	switch config.Provider {
	case models.ProviderMistral:
		// Initialize Mistral client if needed
		llm, err := mistral.New(mistral.WithAPIKey(config.ApiKey))
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Mistral client: %w", err)
		}
		return &Client{llm: llm}, nil
		// TODO: Implement other providers
	default:
		return nil, fmt.Errorf("unsupported provider: %s", config.Provider)
	}
}

// SummarizeCommit generates a summary for a commit using the LLM
func (c *Client) SummarizeCommit(commit *models.Commit) (*models.AICommitSummary, error) {
	prompt := fmt.Sprintf(`You are an expert software engineer skilled in summarizing code changes.
Use the following commit information to generate a concise summary and categorize the change.

Commit Message: %s
Files Changed: %v
Diff Stats:
%s

Response with the following JSON format:
{
  "breaking_changes": ["..."],
  "features": ["..."],
  "bug_fixes": ["..."],
  "build": ["..."],
  "refactor": ["..."],
  "docs": ["..."],
  "chores": ["..."],
  "other_changes": ["..."]
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
