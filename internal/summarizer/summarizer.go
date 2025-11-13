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

	category := categorizeCommit(commit.Message, summary)

	return &models.CommitSummary{
		CommitHash:  commit.Hash,
		Title:       extractTitle(commit.Message),
		Description: summary,
		Category:    category,
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

// categorizeCommit determines the category of a commit based on its message and summary
func categorizeCommit(message, summary string) string {
	lowerMsg := strings.ToLower(message)
	lowerSum := strings.ToLower(summary)

	if strings.Contains(lowerMsg, "feat:") || strings.Contains(lowerMsg, "feature") {
		return "feature"
	}
	if strings.Contains(lowerMsg, "fix:") || strings.Contains(lowerMsg, "bug") {
		return "bugfix"
	}
	if strings.Contains(lowerMsg, "docs:") || strings.Contains(lowerMsg, "documentation") {
		return "docs"
	}
	if strings.Contains(lowerMsg, "refactor:") || strings.Contains(lowerMsg, "refactoring") {
		return "refactor"
	}
	if strings.Contains(lowerMsg, "perf:") || strings.Contains(lowerMsg, "performance") {
		return "performance"
	}
	if strings.Contains(lowerMsg, "test:") || strings.Contains(lowerMsg, "testing") {
		return "test"
	}
	if strings.Contains(lowerMsg, "ci:") || strings.Contains(lowerMsg, "cd:") {
		return "ci/cd"
	}
	if strings.Contains(lowerMsg, "chore:") {
		return "chore"
	}

	// Fallback categorization based on summary
	if strings.Contains(lowerSum, "added") || strings.Contains(lowerSum, "new") {
		return "feature"
	}
	if strings.Contains(lowerSum, "fixed") || strings.Contains(lowerSum, "bug") {
		return "bugfix"
	}

	return "other"
}

// extractTitle extracts the first line of the commit message as title
func extractTitle(message string) string {
	lines := strings.Split(message, "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return ""
}

// OrganizeSummaries groups commit summaries by category
func OrganizeSummaries(summaries []models.CommitSummary) map[string][]models.CommitSummary {
	organized := make(map[string][]models.CommitSummary)

	for _, summary := range summaries {
		organized[summary.Category] = append(organized[summary.Category], summary)
	}

	return organized
}
