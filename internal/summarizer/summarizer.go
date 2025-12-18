package summarizer

import (
	"strings"
	"sync"

	"github.com/ImFlog/human-release/internal/llm"
	"github.com/ImFlog/human-release/pkg/models"
)

// Summarizer handles commit summarization using LLM
type Summarizer struct {
	llmClient *llm.Client
}

// NewSummarizer creates a new commit summarizer
func NewSummarizer(llmClient *llm.Client) *Summarizer {
	return &Summarizer{
		llmClient: llmClient,
	}
}

// SummarizeCommit generates a summary for a single commit
func (s *Summarizer) SummarizeCommit(commit *models.Commit) (*models.CommitSummary, error) {
	summary, err := s.llmClient.SummarizeCommit(commit)
	if err != nil {
		return nil, err
	}

	return &models.CommitSummary{
		CommitHash: commit.Hash,
		Title:      extractTitle(commit.Message),
		Summary:    summary,
	}, nil
}

// SummarizeCommits generates summaries for multiple commits (with concurrency)
func (s *Summarizer) SummarizeCommits(commits []models.Commit) ([]models.CommitSummary, error) {
	results := make([]models.CommitSummary, len(commits))
	errors := make([]error, len(commits))
	var wg sync.WaitGroup

	// Limit concurrency to avoid rate limiting
	maxConcurrency := 5
	semaphore := make(chan struct{}, maxConcurrency)

	for i := range commits {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			commit := commits[idx]
			summary, err := s.SummarizeCommit(&commit)
			if err != nil {
				errors[idx] = err
				return
			}
			results[idx] = *summary
		}(i)
	}

	wg.Wait()

	// Check for errors
	for _, err := range errors {
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}

// extractTitle extracts the first line of the commit message as title
func extractTitle(message string) string {
	lines := strings.Split(message, "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return ""
}
