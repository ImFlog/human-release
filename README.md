# human-release

A powerful CLI tool that generates beautifully formatted release notes from git commit history using AI (LLM-powered).

## Overview

**human-release** automates the tedious process of writing release notes by:
1. **Parsing git history** between any two references (tags, commits, or branches)
2. **Summarizing commits** intelligently using an LLM (Mistral, OpenAI, etc.)
3. **Organizing summaries** by category (Features, Bug Fixes, Performance, etc.)
4. **Generating release notes** in professional Markdown format

No more manually reviewing every commit to write release notes. Just point it at your repository, and get beautifully formatted release notes in seconds.

## Features

- 🚀 **Multi-LLM Support** - Works with Mistral, OpenAI, and extensible for other providers
- 📝 **Smart Categorization** - Automatically categorizes commits (feature, bugfix, docs, refactor, etc.)
- ⚡ **Concurrent Processing** - Summarizes multiple commits in parallel for speed
- 🎯 **Flexible References** - Supports commit hashes, tags, and branch names
- 🔧 **Customizable** - Custom Markdown templates for release notes format
- 📋 **Unified Configuration** - TOML config files + environment variables via Viper
- 🔐 **Secure** - API keys via environment variables (not in config files)
- 📦 **Zero External Dependencies** - Pure Go implementation using go-git library

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/ImFlog/human-release.git
cd human-release

# Build the binary
make build

# Optional: Install globally
make install
```

### Requirements

- Go 1.20 or higher

## Quick Start

### 1. Initialize Configuration

```bash
./human-release config init
```

This creates a `config.toml` file in the current directory with default settings.

### 2. Set Your LLM API Key

```bash
export HUMAN_RELEASE_LLM_API_KEY=your-api-key-here
```

Or add it to `config.toml` (not recommended for production).

### 3. Generate Release Notes

```bash
./human-release --from v1.0.0 --to v1.1.0
```

Output appears on stdout. To save to a file:

```bash
./human-release --from v1.0.0 --to v1.1.0 --output RELEASE_NOTES.md
```

## Usage

### Basic Usage

```bash
# Generate release notes between two tags
./human-release --from v1.0.0 --to v1.1.0

# Using commit hashes
./human-release --from abc123 --to def456

# Using branches
./human-release --from main --to feature-branch

# Save to file
./human-release --from v1.0.0 --to v1.1.0 --output release_notes.md

# Use custom repository path
./human-release --repo /path/to/repo --from v1.0.0 --to v1.1.0

# Use custom template
./human-release --from v1.0.0 --to v1.1.0 --template custom_template.md
```

### Command-Line Flags

```
-f, --from string      Starting git reference (commit hash or tag) [required]
-t, --to string        Ending git reference (commit hash or tag) [required]
-o, --output string    Output file path (default: stdout)
-c, --config string    Path to configuration file (TOML)
-r, --repo string      Path to git repository (default: current directory)
    --template string  Custom markdown template path
-h, --help             Show help message
```

## Configuration

### Configuration File (TOML)

Create a `config.toml` file:

```toml
# LLM Provider Configuration
[llm]
provider = "mistral"          # or "openai"
model = "mistral-large-latest"
# base_url = ""               # Optional: for custom endpoints
# api_key = ""                # Recommended to use env var instead

# Template Configuration
[template]
path = "templates/release_notes.md"  # Custom template path (optional)

# Git Configuration (can be overridden by CLI flags)
[git]
# from_ref = "v1.0.0"
# to_ref = "v1.1.0"
```

### Environment Variables

Set environment variables with the `HUMAN_RELEASE_` prefix:

```bash
# Required: LLM API Key
export HUMAN_RELEASE_LLM_API_KEY=sk-xxx...

# Optional: Custom LLM endpoint
export HUMAN_RELEASE_LLM_BASE_URL=https://api.example.com

# Optional: Config file path
export HUMAN_RELEASE_CONFIG=./config.toml
```

### Configuration Priority

CLI flags > Environment Variables > Config File > Defaults

## Supported LLM Providers

### Mistral (Default)

```toml
[llm]
provider = "mistral"
model = "mistral-large-latest"
# Available models: mistral-large-latest, mistral-medium-latest
```

Get your API key: https://console.mistral.ai/

### OpenAI

```toml
[llm]
provider = "openai"
model = "gpt-4"
# Available models: gpt-4, gpt-4-turbo, gpt-3.5-turbo
```

Get your API key: https://platform.openai.com/api-keys

### Adding Custom Providers

To add a new LLM provider, extend `internal/llm/client.go` and implement the `summarizeWith[Provider]` function. See existing implementations for the pattern.

## Custom Templates

Create a Markdown template with Go template syntax. Place it in your template file and reference it:

```bash
./human-release --from v1.0.0 --to v1.1.0 --template ./my_template.md
```

### Template Variables

Available variables in your template:

- `{{ .FromRef }}` - Starting reference
- `{{ .ToRef }}` - Ending reference
- `{{ .Date }}` - Generation timestamp (time.Time)
- `{{ .Organized }}` - Map of categories to commit summaries

Each summary in `.Organized` has:
- `CommitHash` - Full commit hash
- `Title` - Commit message first line
- `Description` - AI-generated summary
- `Category` - Auto-detected category

### Example Custom Template

```markdown
# Release {{ .ToRef }}

Generated on {{ .Date.Format "2006-01-02" }}

{{ range $category, $summaries := .Organized }}
## {{ $category | upper }}

{{ range $summaries }}
- **{{ .Title }}** ({{ .CommitHash | truncateHash }})

  {{ .Description }}

{{ end }}
{{ end }}

---
_Release notes auto-generated by [human-release](https://github.com/ImFlog/human-release)_
```

## Commit Categories

Commits are automatically categorized based on their message and content:

| Category        | Detection                            |
|-----------------|--------------------------------------|
| **feature**     | Keywords: "feat:", "feature"         |
| **bugfix**      | Keywords: "fix:", "bug"              |
| **performance** | Keywords: "perf:", "performance"     |
| **refactor**    | Keywords: "refactor:", "refactoring" |
| **test**        | Keywords: "test:", "testing"         |
| **docs**        | Keywords: "docs:", "documentation"   |
| **ci/cd**       | Keywords: "ci:", "cd:"               |
| **chore**       | Keywords: "chore:"                   |
| **other**       | Uncategorized commits                |

## Development

### Project Structure

```
human-release/
├── cmd/human-release/          # CLI entry point
├── internal/
│   ├── git/                    # Git repository operations
│   ├── llm/                    # LLM client abstraction
│   ├── summarizer/             # Commit summarization
│   ├── generator/              # Release notes generation
│   └── config/                 # Configuration management
├── pkg/models/                 # Data structures
├── templates/                  # Default templates
└── CLAUDE.md                   # Architecture & developer guide
```

### Build Commands

```bash
# Build binary
make build

# Run tests
make test

# Run a single test
make test-single TEST=./internal/git

# Format code
make fmt

# Run linters (requires golangci-lint)
make lint

# Clean build artifacts
make clean

# Install binary
make install
```

### Testing

```bash
# Run all tests
make test

# Run specific test
go test -v ./internal/git/...

# Run with coverage
go test -cover ./...
```

## Performance Considerations

- **Concurrent Summarization** - Default limit is 5 concurrent LLM calls to avoid rate limiting
- **Commit Count** - Performance scales linearly with number of commits to summarize
- **LLM Response Time** - Most of the time is spent waiting for LLM API responses

Typical performance:
- 10 commits: ~10-15 seconds
- 50 commits: ~50-75 seconds
- 100 commits: ~100-150 seconds

(Times vary based on LLM provider and API latency)

## Troubleshooting

### Error: "API key is not configured"

Set the LLM API key:
```bash
export HUMAN_RELEASE_LLM_API_KEY=your-api-key
```

Or add it to `config.toml`:
```toml
[llm]
api_key = "your-api-key"
```

### Error: "No commits found between refs"

Verify:
- Both references exist in the repository: `git log --oneline | grep <ref>`
- The `from` reference is an ancestor of `to`: `git merge-base --is-ancestor <from> <to>`
- You're in the correct repository: `git remote -v`

### Error: "API rate limit exceeded"

The LLM provider rate limited your requests. Solutions:
- Wait a few minutes and retry
- Reduce concurrency in `internal/summarizer/summarizer.go` (change semaphore size)
- Split large releases into smaller batches

### Error: "Template parsing error"

Verify your template:
- Use valid Go template syntax (check `text/template` documentation)
- Test with: `go run -template ./test.md`
- Check template file exists and is readable

## Examples

### Generate Release Notes from Latest Tag

```bash
# Get latest tag
LATEST=$(git describe --tags --abbrev=0)
PREVIOUS=$(git describe --tags --abbrev=0 $LATEST^)

# Generate release notes
./human-release --from $PREVIOUS --to $LATEST --output RELEASE_NOTES.md
```

### CI/CD Integration

**GitHub Actions Example:**

```yaml
name: Generate Release Notes

on:
  push:
    tags:
      - 'v*'

jobs:
  release-notes:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Build human-release
        run: go build -o human-release ./cmd/human-release

      - name: Generate release notes
        env:
          HUMAN_RELEASE_LLM_API_KEY: ${{ secrets.LLM_API_KEY }}
        run: |
          PREVIOUS_TAG=$(git describe --tags --abbrev=0 ${{ github.ref }}^)
          ./human-release --from $PREVIOUS_TAG --to ${{ github.ref_name }} \
            --output RELEASE_NOTES.md

      - name: Create release
        uses: softprops/action-gh-release@v1
        with:
          body_path: RELEASE_NOTES.md
```

## Architecture

For detailed architecture information, see [CLAUDE.md](./CLAUDE.md).

Key components:
- **Git Parser** - Pure Go git parsing with go-git
- **LLM Client** - Multi-provider abstraction over HTTP APIs
- **Summarizer** - Concurrent commit summarization and categorization
- **Generator** - Release notes generation with template support
- **Configuration** - Unified config management via Viper

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests and linters (`make test fmt lint`)
5. Commit with descriptive messages
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.

## Acknowledgments

- [go-git](https://github.com/go-git/go-git) - Pure Go implementation of git
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management
- [Mistral AI](https://mistral.ai/) & [OpenAI](https://openai.com/) - LLM providers

## Support

For issues, questions, or feature requests:
- Open an issue on [GitHub](https://github.com/ImFlog/human-release/issues)
- Check existing documentation in [CLAUDE.md](./CLAUDE.md)

## Roadmap
- `client.go` changes
  - [ ] Use langchain for LLM calls ? Or we could use an interface and implement for each provider manually (if langchain doesn't support multiple providers easily).
  - [ ] Put prompts in separate files for easier editing.
  - [ ] Tune prompt for better summaries (bullet points, sorted by fix / features / tech debt, etc.)
- `summarizer.go` changes
  - [ ] Improve commit categorization by letting the LLM decide category based on commit message + diff.
- `generator.go` changes
    - [ ] The format is not very good. Change it.
  - [ ] Support additional output formats (HTML, JSON, etc.)
- [ ] Add unit and integration tests for all components
- [ ] Add badges (build status, license, etc.)
- [ ] Add golangci-lint + configuration
- [ ] Additional LLM provider support (Anthropic Claude, Gemini, etc.)
- [ ] Docker image distribution
- [ ] Caching mechanism for commit summaries
- [ ] Interactive CLI for customizing output
- [ ] Performance optimizations for large repositories
- [ ] WebUI for release notes generation

---

**Made with ❤️ by ImFlog**
