package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ImFlog/human-release/pkg/models"
)

// Client handles LLM API calls
type Client struct {
	config *models.LLMConfig
	http   *http.Client
}

// NewClient creates a new LLM client
func NewClient(config *models.LLMConfig) *Client {
	return &Client{
		config: config,
		http:   &http.Client{},
	}
}

// SummarizeCommit generates a summary for a commit using the LLM
func (c *Client) SummarizeCommit(commit *models.Commit) (string, error) {
	switch c.config.Provider {
	case "mistral":
		return c.summarizeWithMistral(commit)
	case "openai":
		return c.summarizeWithOpenAI(commit)
	default:
		return "", fmt.Errorf("unsupported provider: %s", c.config.Provider)
	}
}

// summarizeWithMistral calls Mistral API for commit summarization
func (c *Client) summarizeWithMistral(commit *models.Commit) (string, error) {
	prompt := fmt.Sprintf(`Summarize the following git commit in a concise manner (1-2 sentences).
Focus on what was changed and why.

Commit Message: %s
Files Changed: %v
Diff Stats:
%s

Summary:`, commit.Message, commit.Files, commit.Diff)

	baseURL := c.config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.mistral.ai"
	}

	reqBody := map[string]interface{}{
		"model": c.config.Model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens": 200,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", baseURL+"/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.ApiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return result.Choices[0].Message.Content, nil
}

// summarizeWithOpenAI calls OpenAI API for commit summarization
func (c *Client) summarizeWithOpenAI(commit *models.Commit) (string, error) {
	// Similar implementation for OpenAI
	prompt := fmt.Sprintf(`Summarize the following git commit in a concise manner (1-2 sentences).
Focus on what was changed and why.

Commit Message: %s
Files Changed: %v
Diff Stats:
%s

Summary:`, commit.Message, commit.Files, commit.Diff)

	baseURL := c.config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}

	reqBody := map[string]interface{}{
		"model": c.config.Model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens": 200,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", baseURL+"/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.ApiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return result.Choices[0].Message.Content, nil
}
