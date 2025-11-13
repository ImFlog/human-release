package generator

import (
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/ImFlog/human-release/pkg/models"
	"github.com/ImFlog/human-release/internal/summarizer"
)

// Generator handles release notes generation
type Generator struct {
	templatePath string
}

// NewGenerator creates a new release notes generator
func NewGenerator(templatePath string) *Generator {
	return &Generator{
		templatePath: templatePath,
	}
}

// Generate creates release notes from commit summaries
func (g *Generator) Generate(fromRef, toRef string, summaries []models.CommitSummary) (*models.ReleaseNotes, error) {
	organized := summarizer.OrganizeSummaries(summaries)

	releaseNotes := &models.ReleaseNotes{
		FromRef:   fromRef,
		ToRef:     toRef,
		Date:      time.Now(),
		Summaries: summaries,
	}

	// Try to load custom template if provided
	var content string
	var err error
	if g.templatePath != "" && g.templatePath != "default" {
		content, err = g.renderFromFile(fromRef, toRef, organized)
		if err != nil {
			// Fall back to default template
			content = g.renderDefault(fromRef, toRef, organized)
		}
	} else {
		content = g.renderDefault(fromRef, toRef, organized)
	}

	releaseNotes.Content = content
	return releaseNotes, nil
}

// renderDefault renders using the default template
func (g *Generator) renderDefault(fromRef, toRef string, organized map[string][]models.CommitSummary) string {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("# Release Notes: %s...%s\n\n", fromRef, toRef))
	output.WriteString(fmt.Sprintf("*Generated on %s*\n\n", time.Now().Format(time.RFC1123)))

	// Define category order
	categoryOrder := []string{
		"feature",
		"bugfix",
		"performance",
		"refactor",
		"test",
		"docs",
		"ci/cd",
		"chore",
		"other",
	}

	for _, category := range categoryOrder {
		summaries, exists := organized[category]
		if !exists || len(summaries) == 0 {
			continue
		}

		output.WriteString(fmt.Sprintf("## %s\n\n", capitalizeCategory(category)))

		for _, summary := range summaries {
			output.WriteString(fmt.Sprintf("- **%s** - %s\n", summary.Title, summary.Description))
			output.WriteString(fmt.Sprintf("  - Commit: `%s`\n", summary.CommitHash[:8]))
		}
		output.WriteString("\n")
	}

	return output.String()
}

// renderFromFile renders using a custom template file
func (g *Generator) renderFromFile(fromRef, toRef string, organized map[string][]models.CommitSummary) (string, error) {
	// Read template file
	tmplContent, err := os.ReadFile(g.templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template file: %w", err)
	}

	// Parse and execute template
	tmpl, err := template.New("release_notes").Parse(string(tmplContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	data := map[string]interface{}{
		"FromRef":    fromRef,
		"ToRef":      toRef,
		"Date":       time.Now(),
		"Organized":  organized,
	}

	var output strings.Builder
	if err := tmpl.Execute(&output, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return output.String(), nil
}

// capitalizeCategory returns a nice display name for a category
func capitalizeCategory(category string) string {
	categoryNames := map[string]string{
		"feature":     "Features",
		"bugfix":      "Bug Fixes",
		"performance": "Performance Improvements",
		"refactor":    "Refactoring",
		"test":        "Tests",
		"docs":        "Documentation",
		"ci/cd":       "CI/CD",
		"chore":       "Chores",
		"other":       "Other",
	}

	if name, exists := categoryNames[category]; exists {
		return name
	}
	return strings.ToTitle(category)
}

// WriteToFile writes the release notes to a file
func (g *Generator) WriteToFile(releaseNotes *models.ReleaseNotes, filePath string) error {
	return os.WriteFile(filePath, []byte(releaseNotes.Content), 0644)
}
