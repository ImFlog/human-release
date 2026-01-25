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
For each commit define a description of the change, It should not exceed 100 characters.
Then categorize the change into one of the following categories:
* breaking_changes = a change that changes an breaks existing usage (a breaking API contract for instance)
* feat = a new feature
* fix = a bug fix, a minor change
* refactor = a code change that does not affect the behavior of the code
* other_changes = a change that does not fit into any of the above categories

Use the following commit information to generate a summary and categorize the changes.

Commit Message: %s
Files Changed: %v
Diff Stats:
%s

Respond with the following JSON format:
{
  "breaking_changes": ["...", "..."],
  "feat": ["...", "..."],
  "fix": ["...", "..."],
  "refactor": ["...", "..."],
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
