package git

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/ImFlog/human-release/pkg/models"
)

// Parser handles git repository operations
type Parser struct {
	repo *git.Repository
}

// NewParser creates a new git parser for the given repository path
func NewParser(repoPath string) (*Parser, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	return &Parser{repo: repo}, nil
}

// GetCommits retrieves commits between two references (commit hashes or tags)
func (p *Parser) GetCommits(fromRef, toRef string) ([]models.Commit, error) {
	fromHash, err := p.resolveRef(fromRef)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve 'from' reference: %w", err)
	}

	toHash, err := p.resolveRef(toRef)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve 'to' reference: %w", err)
	}

	// Get commits reachable from toHash but not from fromHash
	iter, err := p.repo.Log(&git.LogOptions{
		From: toHash,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get log: %w", err)
	}

	var commits []models.Commit
	seen := make(map[plumbing.Hash]bool)

	err = iter.ForEach(func(c *object.Commit) error {
		if c.Hash == fromHash {
			return nil // Stop when we reach the from commit
		}

		if seen[c.Hash] {
			return nil
		}
		seen[c.Hash] = true

		// Get the diff for this commit
		var patch bytes.Buffer
		stats, err := c.Stats()
		if err == nil {
			for _, stat := range stats {
				patch.WriteString(stat.String() + "\n")
			}
		}

		files := []string{}
		for _, stat := range stats {
			files = append(files, stat.Name)
		}

		commit := models.Commit{
			Hash:    c.Hash.String(),
			Author:  c.Author.Name,
			Date:    c.Author.When,
			Message: c.Message,
			Files:   files,
			Diff:    patch.String(),
		}

		commits = append(commits, commit)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate commits: %w", err)
	}

	return commits, nil
}

// resolveRef converts a reference (tag, branch, or commit hash) to a commit hash
func (p *Parser) resolveRef(ref string) (plumbing.Hash, error) {
	// Try to parse as a hash first
	if hash := plumbing.NewHash(ref); !hash.IsZero() {
		return hash, nil
	}

	// Try to resolve as a reference (tag or branch)
	h, err := p.repo.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return plumbing.ZeroHash, fmt.Errorf("failed to resolve reference %s: %w", ref, err)
	}

	return *h, nil
}

// GetCommitChanges returns the detailed changes for a specific commit
func (p *Parser) GetCommitChanges(commitHash string) (string, error) {
	hash := plumbing.NewHash(commitHash)
	commit, err := p.repo.CommitObject(hash)
	if err != nil {
		return "", fmt.Errorf("failed to get commit: %w", err)
	}

	// Get parent commit for diff
	var parent *object.Commit
	if commit.NumParents() > 0 {
		parent, err = commit.Parent(0)
		if err != nil {
			return "", fmt.Errorf("failed to get parent commit: %w", err)
		}
	}

	// Get tree diff
	var currentTree *object.Tree
	var parentTree *object.Tree

	currentTree, err = commit.Tree()
	if err != nil {
		return "", fmt.Errorf("failed to get commit tree: %w", err)
	}

	if parent != nil {
		parentTree, err = parent.Tree()
		if err != nil {
			return "", fmt.Errorf("failed to get parent tree: %w", err)
		}
	}

	// Get patch
	var patch bytes.Buffer
	if parentTree != nil {
		changes, err := object.DiffTree(parentTree, currentTree)
		if err != nil {
			return "", fmt.Errorf("failed to diff trees: %w", err)
		}

		for _, change := range changes {
			from := change.From.Name
			to := change.To.Name

			// Determine change type
			if from == "" {
				patch.WriteString(fmt.Sprintf("A\t%s\n", to))
			} else if to == "" {
				patch.WriteString(fmt.Sprintf("D\t%s\n", from))
			} else if from == to {
				patch.WriteString(fmt.Sprintf("M\t%s\n", to))
			} else {
				patch.WriteString(fmt.Sprintf("R\t%s -> %s\n", from, to))
			}
		}
	} else {
		// First commit - show all files as added
		iter := currentTree.Files()
		iter.ForEach(func(f *object.File) error {
			patch.WriteString(fmt.Sprintf("A\t%s\n", f.Name))
			return nil
		})
	}

	return patch.String(), nil
}

// GetTags returns all tags in the repository
func (p *Parser) GetTags() ([]string, error) {
	tags, err := p.repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	var tagNames []string
	err = tags.ForEach(func(ref *plumbing.Reference) error {
		tagNames = append(tagNames, strings.TrimPrefix(ref.Name().String(), "refs/tags/"))
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate tags: %w", err)
	}

	return tagNames, nil
}
