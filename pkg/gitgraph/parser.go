package gitgraph

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

const GitGraphKeyword = "gitGraph"

var (
	commitRegex   = regexp.MustCompile(`^\s*commit(?:\s+id:\s*"([^"]*)")?(?:\s+msg:\s*"([^"]*)")?(?:\s+tag:\s*"([^"]*)")?(?:\s+type:\s*(NORMAL|REVERSE|HIGHLIGHT))?\s*$`)
	branchRegex   = regexp.MustCompile(`^\s*branch\s+(\S+)\s*$`)
	checkoutRegex = regexp.MustCompile(`^\s*checkout\s+(\S+)\s*$`)
	mergeRegex    = regexp.MustCompile(`^\s*merge\s+(\S+)(?:\s+id:\s*"([^"]*)")?(?:\s+tag:\s*"([^"]*)")?\s*$`)
	cherryPickRegex = regexp.MustCompile(`^\s*cherry-pick\s+id:\s*"([^"]+)"\s*$`)
)

type GitGraph struct {
	Commits    []*Commit
	Branches   []*Branch
	CurrentBranch string
}

type Commit struct {
	ID      string
	Message string
	Tag     string
	Type    CommitType
	Branch  string
	Parents []string
	Lane    int
}

type Branch struct {
	Name      string
	Lane      int
	StartCommit string
}

type CommitType int

const (
	Normal CommitType = iota
	Reverse
	Highlight
)

func IsGitGraph(input string) bool {
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}
		return trimmed == GitGraphKeyword
	}
	return false
}

func Parse(input string) (*GitGraph, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty input")
	}

	rawLines := diagram.SplitLines(input)
	lines := diagram.RemoveComments(rawLines)
	if len(lines) == 0 {
		return nil, fmt.Errorf("no content found")
	}

	if strings.TrimSpace(lines[0]) != GitGraphKeyword {
		return nil, fmt.Errorf("expected %q keyword", GitGraphKeyword)
	}
	lines = lines[1:]

	gg := &GitGraph{
		Commits:       []*Commit{},
		Branches:      []*Branch{},
		CurrentBranch: "main",
	}

	// Create default main branch
	mainBranch := &Branch{Name: "main", Lane: 0}
	gg.Branches = append(gg.Branches, mainBranch)
	branchMap := map[string]*Branch{"main": mainBranch}
	nextLane := 1
	commitCounter := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if match := branchRegex.FindStringSubmatch(trimmed); match != nil {
			name := match[1]
			if _, exists := branchMap[name]; !exists {
				branch := &Branch{Name: name, Lane: nextLane}
				if len(gg.Commits) > 0 {
					branch.StartCommit = gg.Commits[len(gg.Commits)-1].ID
				}
				gg.Branches = append(gg.Branches, branch)
				branchMap[name] = branch
				nextLane++
			}
			continue
		}

		if match := checkoutRegex.FindStringSubmatch(trimmed); match != nil {
			name := match[1]
			if _, exists := branchMap[name]; exists {
				gg.CurrentBranch = name
			}
			continue
		}

		if match := commitRegex.FindStringSubmatch(trimmed); match != nil {
			id := match[1]
			if id == "" {
				id = fmt.Sprintf("c%d", commitCounter)
			}
			commitCounter++

			commitType := Normal
			if match[4] == "REVERSE" {
				commitType = Reverse
			} else if match[4] == "HIGHLIGHT" {
				commitType = Highlight
			}

			branch := branchMap[gg.CurrentBranch]
			commit := &Commit{
				ID:      id,
				Message: match[2],
				Tag:     match[3],
				Type:    commitType,
				Branch:  gg.CurrentBranch,
				Lane:    branch.Lane,
			}
			gg.Commits = append(gg.Commits, commit)
			continue
		}

		if match := mergeRegex.FindStringSubmatch(trimmed); match != nil {
			sourceBranch := match[1]
			id := match[2]
			if id == "" {
				id = fmt.Sprintf("m%d", commitCounter)
			}
			commitCounter++

			branch := branchMap[gg.CurrentBranch]
			commit := &Commit{
				ID:      id,
				Tag:     match[3],
				Branch:  gg.CurrentBranch,
				Lane:    branch.Lane,
				Parents: []string{sourceBranch},
			}
			gg.Commits = append(gg.Commits, commit)
			continue
		}

		if match := cherryPickRegex.FindStringSubmatch(trimmed); match != nil {
			commitCounter++
			branch := branchMap[gg.CurrentBranch]
			commit := &Commit{
				ID:      fmt.Sprintf("cp%d", commitCounter),
				Message: "cherry-pick " + match[1],
				Branch:  gg.CurrentBranch,
				Lane:    branch.Lane,
				Parents: []string{match[1]},
			}
			gg.Commits = append(gg.Commits, commit)
			continue
		}
	}

	if len(gg.Commits) == 0 {
		return nil, fmt.Errorf("no commits found")
	}

	return gg, nil
}
