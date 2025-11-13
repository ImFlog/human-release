# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**human-release** is a CLI tool that generates release notes from git commit history using AI. The tool:
1. Parses git history between two references (commits or tags)
2. Summarizes each commit using an LLM
3. Organizes summaries by category (features, bugfixes, docs, etc.)
4. Generates formatted release notes in Markdown

## Architecture

The project is organized into clear, modular packages with separation of concerns:

```
internal/
├── git/       - Git repository parsing and commit extraction
├── llm/       - LLM client abstraction for multi-provider support
├── summarizer - Commit summarization with LLM and categorization
├── generator  - Release notes generation from commit summaries
└── config     - Configuration management via Viper (TOML + env vars)

cmd/                - CLI entry point using Cobra framework
pkg/models/         - Data structures (Commit, CommitSummary, ReleaseNotes, Config)
templates/          - Default Markdown template for release notes
```

### Key Components

**Git Parser** (`internal/git/parser.go`)
- Uses `go-git` library to parse git repositories
- Extracts commits between two references with file changes
- Supports commit hashes, branch names, and tags as references

**LLM Client** (`internal/llm/client.go`)
- Multi-provider abstraction (currently supports Mistral, OpenAI)
- HTTP-based API calls with proper error handling
- Generates concise commit summaries via LLM prompts

**Summarizer** (`internal/summarizer/summarizer.go`)
- Processes commits concurrently (max 5 at a time to avoid rate limiting)
- Categorizes commits based on message patterns and summary content
- Categories: feature, bugfix, performance, refactor, test, docs, ci/cd, chore, other

**Generator** (`internal/generator/release_notes.go`)
- Groups commit summaries by category
- Renders release notes from default or custom Markdown templates
- Outputs to file or stdout

**Configuration** (`internal/config/config.go`)
- Uses **Viper** for unified configuration management
- TOML files for non-sensitive settings
- Environment variables for secrets (API keys)
- CLI flags override all other configuration sources
- Environment variable prefix: `HUMAN_RELEASE_`

## Configuration

### TOML Configuration File

Set via `--config` or `HUMAN_RELEASE_CONFIG`:
```toml
[llm]
provider = "mistral"  # or "openai"
model = "mistral-large-latest"
base_url = ""  # Optional, for custom endpoints

[template]
path = "templates/release_notes.md"  # Custom template path
```

### Environment Variables

- `HUMAN_RELEASE_LLM_API_KEY` - API key for the LLM provider (required)
- `HUMAN_RELEASE_LLM_BASE_URL` - Custom LLM endpoint (optional)
- `HUMAN_RELEASE_CONFIG` - Path to config file (optional)

## Common Development Commands

### Build and Installation

```bash
# Build the binary
make build

# Install globally
make install

# Clean build artifacts
make clean
```

### Running the Tool

```bash
# Generate release notes between two references
./human-release --from v1.0.0 --to v1.1.0

# Output to file
./human-release --from v1.0.0 --to v1.1.0 --output release_notes.md

# Use custom config and template
./human-release --config config.toml --template custom_template.md \
  --from v1.0.0 --to v1.1.0

# Operate on a different repository
./human-release --repo /path/to/repo --from main --to feature-branch
```

### Initialize Configuration

```bash
./human-release config init
# Creates config.toml with defaults in current directory
```

### Testing and Linting

```bash
# Run all tests
make test

# Run a single test
make test-single TEST=github.com/ImFlog/human-release/internal/git

# Format code
make fmt

# Run linters (requires golangci-lint)
make lint
```

## Development Workflow

### Adding a New LLM Provider

1. Add provider case in `internal/llm/client.go` → `SummarizeCommit()` method
2. Implement provider-specific API call (following Mistral/OpenAI pattern)
3. Update config examples in `config.example.toml`
4. Update documentation

### Extending Commit Categorization

Edit `internal/summarizer/categorizeCommit()` to add keywords or logic for new categories. Categories are displayed in `capitalizeCategory()` mapping.

### Custom Release Notes Templates

Create a Markdown template with Go template syntax. Available data:
- `{{ .FromRef }}` - Starting reference
- `{{ .ToRef }}` - Ending reference
- `{{ .Date }}` - Generation date
- `{{ .Organized }}` - Map of category → []CommitSummary
  - Each summary has: CommitHash, Title, Description, Category

Example:
```markdown
# Release {{ .ToRef }}
{{ range $category, $summaries := .Organized }}
## {{ $category }}
{{ range $summaries }}
- {{ .Title }}: {{ .Description }}
{{ end }}
{{ end }}
```

## Key Design Decisions

1. **Viper for Configuration** - Seamless TOML + env var support with CLI flag overrides
2. **Modular Architecture** - Each package has single responsibility for testability
3. **Concurrent Summarization** - Limits to 5 concurrent LLM calls to avoid rate limiting
4. **go-git Library** - Pure Go git parsing without external dependencies
5. **Interface-based LLM Client** - Easy to add new providers without modifying existing code

## Dependencies

- `github.com/go-git/go-git/v5` - Git repository parsing
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management

## Troubleshooting

**"API key is not configured"**
- Set `HUMAN_RELEASE_LLM_API_KEY` environment variable
- Or add `api_key` to TOML config (not recommended for production)

**"No commits found between refs"**
- Verify both references exist in the repository
- Check they're in the correct order (from should be ancestor of to)

**Rate limiting from LLM provider**
- Default concurrency is 5. Reduce in `internal/summarizer/SummarizeCommits()` semaphore value
- Add delays between requests if needed

**Template parsing errors**
- Check template syntax is valid Go template syntax
- Test with `text/template` package locally
