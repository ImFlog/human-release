package generator

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/ImFlog/human-release/pkg/models"
)

//go:embed release_notes_default.tmpl
var defaultTemplate []byte

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
	releaseNotes := &models.ReleaseNotes{
		FromRef:   fromRef,
		ToRef:     toRef,
		Date:      time.Now(),
		Summaries: summaries,
	}

	// Try to load custom template if provided
	content, err := g.renderFromFile(fromRef, toRef, summaries)
	if err != nil {
		return nil, fmt.Errorf("failed to render release notes from template: %w", err)
	}

	releaseNotes.Content = content
	return releaseNotes, nil
}

// renderFromFile renders using a custom template file
func (g *Generator) renderFromFile(fromRef, toRef string, summaries []models.CommitSummary) (string, error) {
	var tmplContent []byte
	// Read template file
	if g.templatePath == "default" || g.templatePath == "" {
		tmplContent = defaultTemplate
	} else {
		var err error
		tmplContent, err = os.ReadFile(g.templatePath)
		if err != nil {
			return "", fmt.Errorf("failed to read template file: %w", err)
		}
	}

	// Parse and execute template
	tmpl, err := template.New("release_notes").Parse(string(tmplContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	type categorySummary struct {
		Title     string
		Hash      string
		Summaries string
	}

	// TODO: Flatten summaries by category for easier rendering
	categories := map[string][]categorySummary{}
	for _, sum := range summaries {
		if sum.Summary == nil {
			continue
		}
		for _, change := range sum.Summary.BreakingChanges {
			categories["Breaking change"] = append(categories["breaking_changes"], categorySummary{
				Title:     sum.Title,
				Hash:      sum.CommitHash,
				Summaries: change,
			})
		}
		for _, change := range sum.Summary.Features {
			categories["Features"] = append(categories["feature"], categorySummary{
				Title:     sum.Title,
				Hash:      sum.CommitHash,
				Summaries: change,
			})
		}
		for _, change := range sum.Summary.BugFixes {
			categories["Fixes"] = append(categories["bugfix"], categorySummary{
				Title:     sum.Title,
				Hash:      sum.CommitHash,
				Summaries: change,
			})
		}
		for _, change := range sum.Summary.Build {
			categories["Technical"] = append(categories["Technical"], categorySummary{
				Title:     sum.Title,
				Hash:      sum.CommitHash,
				Summaries: change,
			})
		}
		for _, change := range sum.Summary.Refactor {
			categories["Technical"] = append(categories["Technical"], categorySummary{
				Title:     sum.Title,
				Hash:      sum.CommitHash,
				Summaries: change,
			})
		}
		for _, change := range sum.Summary.Docs {
			categories["Misc"] = append(categories["Misc"], categorySummary{
				Title:     sum.Title,
				Hash:      sum.CommitHash,
				Summaries: change,
			})
		}
		for _, change := range sum.Summary.Chores {
			categories["Misc"] = append(categories["Misc"], categorySummary{
				Title:     sum.Title,
				Hash:      sum.CommitHash,
				Summaries: change,
			})
		}
		for _, change := range sum.Summary.Tests {
			categories["Technical"] = append(categories["Technical"], categorySummary{
				Title:     sum.Title,
				Hash:      sum.CommitHash,
				Summaries: change,
			})
		}
		for _, change := range sum.Summary.Others {
			categories["Misc"] = append(categories["Misc"], categorySummary{
				Title:     sum.Title,
				Hash:      sum.CommitHash,
				Summaries: change,
			})
		}
	}

	data := map[string]interface{}{
		"FromRef":              fromRef,
		"ToRef":                toRef,
		"Date":                 time.Now(),
		"CategorizedSummaries": categories,
	}

	var output strings.Builder
	if err := tmpl.Execute(&output, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return output.String(), nil
}

// WriteToFile writes the release notes to a file
func (g *Generator) WriteToFile(releaseNotes *models.ReleaseNotes, filePath string) error {
	return os.WriteFile(filePath, []byte(releaseNotes.Content), 0644)
}
