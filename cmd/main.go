package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ImFlog/human-release/internal/config"
	"github.com/ImFlog/human-release/internal/generator"
	"github.com/ImFlog/human-release/internal/git"
	"github.com/ImFlog/human-release/internal/llm"
	"github.com/ImFlog/human-release/internal/summarizer"
	"github.com/spf13/cobra"
)

var (
	configPath   string
	fromRef      string
	toRef        string
	outputFile   string
	templatePath string
	repoPath     string
)

var rootCmd = &cobra.Command{
	Use:   "human-release",
	Short: "Generate release notes from git history using AI",
	Long: `human-release is a CLI tool that parses git history between two references,
summarizes commits using an LLM, and generates beautifully formatted release notes.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := generateReleaseNotes(); err != nil {
			log.Fatalf("Error: %v", err)
		}
	},
}

func init() {
	rootCmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to configuration file (TOML)")
	rootCmd.Flags().StringVarP(&fromRef, "from", "f", "", "Starting git reference (commit hash or tag)")
	rootCmd.Flags().StringVarP(&toRef, "to", "t", "", "Ending git reference (commit hash or tag)")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (default: stdout)")
	rootCmd.Flags().StringVarP(&templatePath, "template", "", "", "Custom markdown template path")
	rootCmd.Flags().StringVarP(&repoPath, "repo", "r", ".", "Path to git repository")

	rootCmd.MarkFlagRequired("from")
	rootCmd.MarkFlagRequired("to")
}

func generateReleaseNotes() error {
	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Override config with CLI flags if provided
	if fromRef != "" {
		cfg.Git.FromRef = fromRef
	}
	if toRef != "" {
		cfg.Git.ToRef = toRef
	}
	if templatePath != "" {
		cfg.Template.Path = templatePath
	}

	// Validate required fields
	if cfg.Git.FromRef == "" || cfg.Git.ToRef == "" {
		return fmt.Errorf("both --from and --to references are required")
	}

	if cfg.LLM.ApiKey == "" {
		return fmt.Errorf("LLM API key is not configured. Set it via HUMAN_RELEASE_LLM_API_KEY environment variable")
	}

	// Initialize git parser
	gitParser, err := git.NewParser(repoPath)
	if err != nil {
		return fmt.Errorf("failed to initialize git parser: %w", err)
	}

	// Get commits
	fmt.Printf("Fetching commits from %s to %s...\n", cfg.Git.FromRef, cfg.Git.ToRef)
	commits, err := gitParser.GetCommits(cfg.Git.FromRef, cfg.Git.ToRef)
	if err != nil {
		return fmt.Errorf("failed to get commits: %w", err)
	}

	if len(commits) == 0 {
		return fmt.Errorf("no commits found between %s and %s", cfg.Git.FromRef, cfg.Git.ToRef)
	}

	fmt.Printf("Found %d commits. Generating summaries...\n", len(commits))

	// Initialize LLM client
	llmClient := llm.NewClient(&cfg.LLM)

	// Initialize summarizer
	sum := summarizer.NewSummarizer(llmClient)

	// Summarize commits
	summaries, err := sum.SummarizeCommits(commits)
	if err != nil {
		return fmt.Errorf("failed to summarize commits: %w", err)
	}

	fmt.Printf("Generated %d summaries. Creating release notes...\n", len(summaries))

	// Generate release notes
	templatePath := cfg.Template.Path
	if templatePath == "" {
		templatePath = "default"
	}
	gen := generator.NewGenerator(templatePath)
	releaseNotes, err := gen.Generate(cfg.Git.FromRef, cfg.Git.ToRef, summaries)
	if err != nil {
		return fmt.Errorf("failed to generate release notes: %w", err)
	}

	// Output results
	if outputFile != "" {
		if err := gen.WriteToFile(releaseNotes, outputFile); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Release notes written to %s\n", outputFile)
	} else {
		fmt.Println(releaseNotes.Content)
	}

	return nil
}

func main() {
	// Create config command
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management commands",
	}

	// Create config init subcommand
	configInitCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return initConfig()
		},
	}

	configCmd.AddCommand(configInitCmd)
	rootCmd.AddCommand(configCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// initConfig creates a default configuration file
func initConfig() error {
	configContent := `# human-release configuration
# This file can be overridden by environment variables with HUMAN_RELEASE_ prefix

[llm]
provider = "mistral"
model = "mistral-large-latest"
# base_url = "" # Optional: for custom endpoints
# api_key = "" # Set via HUMAN_RELEASE_LLM_API_KEY environment variable

[template]
path = "templates/release_notes.md"
`

	// Create config file in current directory
	if err := os.WriteFile("config.toml", []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}

	fmt.Println("Created config.toml")
	fmt.Println("Set your API key: export HUMAN_RELEASE_LLM_API_KEY=<your-api-key>")

	return nil
}
