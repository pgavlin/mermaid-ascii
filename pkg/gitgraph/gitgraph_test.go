package gitgraph

import (
	"strings"
	"testing"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

func TestIsGitGraph(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"gitGraph\n  commit", true},
		{"%% comment\ngitGraph", true},
		{"graph LR\n  A-->B", false},
		{"", false},
	}
	for _, tt := range tests {
		got := IsGitGraph(tt.input)
		if got != tt.want {
			t.Errorf("IsGitGraph(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParse(t *testing.T) {
	input := `gitGraph
    commit id: "c1" msg: "Initial commit"
    commit id: "c2" msg: "Add feature"
    branch develop
    checkout develop
    commit id: "c3" msg: "Dev work"
    checkout main
    merge develop`

	gg, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(gg.Commits) != 4 {
		t.Errorf("Commits count = %d, want 4", len(gg.Commits))
	}

	if len(gg.Branches) != 2 {
		t.Errorf("Branches count = %d, want 2", len(gg.Branches))
	}

	if gg.Commits[0].ID != "c1" {
		t.Errorf("Commit 0 ID = %q, want %q", gg.Commits[0].ID, "c1")
	}

	if gg.Commits[2].Branch != "develop" {
		t.Errorf("Commit 2 branch = %q, want %q", gg.Commits[2].Branch, "develop")
	}
}

func TestParseWithTags(t *testing.T) {
	input := `gitGraph
    commit id: "c1" tag: "v1.0"
    commit id: "c2" tag: "v2.0"`

	gg, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if gg.Commits[0].Tag != "v1.0" {
		t.Errorf("Commit 0 tag = %q, want %q", gg.Commits[0].Tag, "v1.0")
	}
}

func TestParseNoCommits(t *testing.T) {
	_, err := Parse("gitGraph\n  branch develop")
	if err == nil {
		t.Error("Expected error for no commits")
	}
}

func TestRender(t *testing.T) {
	input := `gitGraph
    commit id: "c1"
    commit id: "c2"
    branch develop
    checkout develop
    commit id: "c3"
    checkout main
    merge develop`

	gg, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	output, err := Render(gg, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(output, "c1") {
		t.Error("Output should contain commit ID")
	}
	if !strings.Contains(output, "main") {
		t.Error("Output should contain branch name")
	}
	if !strings.Contains(output, "●") {
		t.Error("Unicode output should contain commit dot")
	}
}

func TestRenderASCII(t *testing.T) {
	input := `gitGraph
    commit id: "c1"
    commit id: "c2"`

	gg, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	config.UseAscii = true
	output, err := Render(gg, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if strings.Contains(output, "●") {
		t.Error("ASCII output should not contain Unicode characters")
	}
	if !strings.Contains(output, "*") {
		t.Error("ASCII output should contain * commit character")
	}
}

func TestParseCherryPick(t *testing.T) {
	input := `gitGraph
    commit id: "c1"
    branch develop
    checkout develop
    commit id: "c2"
    checkout main
    cherry-pick id: "c2"`

	gg, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should have 3 commits: c1, c2, cherry-pick
	if len(gg.Commits) != 3 {
		t.Fatalf("Commits count = %d, want 3", len(gg.Commits))
	}

	cp := gg.Commits[2]
	if !strings.Contains(cp.Message, "cherry-pick c2") {
		t.Errorf("Cherry-pick commit message = %q, want it to contain 'cherry-pick c2'", cp.Message)
	}
	if cp.Branch != "main" {
		t.Errorf("Cherry-pick commit branch = %q, want 'main'", cp.Branch)
	}
	if len(cp.Parents) != 1 || cp.Parents[0] != "c2" {
		t.Errorf("Cherry-pick commit parents = %v, want ['c2']", cp.Parents)
	}
}

func TestParseTagWithCustomName(t *testing.T) {
	input := `gitGraph
    commit id: "c1" tag: "release-1.0"
    commit id: "c2" tag: "release-2.0-beta"`

	gg, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if gg.Commits[0].Tag != "release-1.0" {
		t.Errorf("Commit 0 tag = %q, want %q", gg.Commits[0].Tag, "release-1.0")
	}
	if gg.Commits[1].Tag != "release-2.0-beta" {
		t.Errorf("Commit 1 tag = %q, want %q", gg.Commits[1].Tag, "release-2.0-beta")
	}
}

func TestParseMergeWithCustomID(t *testing.T) {
	input := `gitGraph
    commit id: "c1"
    branch develop
    checkout develop
    commit id: "c2"
    checkout main
    merge develop id: "merge-1" tag: "v1.0"`

	gg, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	mergeCommit := gg.Commits[len(gg.Commits)-1]
	if mergeCommit.ID != "merge-1" {
		t.Errorf("Merge commit ID = %q, want 'merge-1'", mergeCommit.ID)
	}
	if mergeCommit.Tag != "v1.0" {
		t.Errorf("Merge commit tag = %q, want 'v1.0'", mergeCommit.Tag)
	}
	if len(mergeCommit.Parents) != 1 || mergeCommit.Parents[0] != "develop" {
		t.Errorf("Merge commit parents = %v, want ['develop']", mergeCommit.Parents)
	}
}

func TestRenderNilDiagram(t *testing.T) {
	_, err := Render(nil, nil)
	if err == nil {
		t.Error("Expected error for nil diagram")
	}
}

func TestRenderNilConfig(t *testing.T) {
	input := `gitGraph
    commit id: "c1"
    commit id: "c2"`

	gg, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	output, err := Render(gg, nil)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(output, "c1") {
		t.Error("Output should contain commit ID c1")
	}
}

func TestParseInvalidKeyword(t *testing.T) {
	_, err := Parse("flowchart\n  commit id: \"c1\"")
	if err == nil {
		t.Error("Expected error for invalid keyword")
	}
}

func TestParseCommitTypes(t *testing.T) {
	input := `gitGraph
    commit id: "c1" type: NORMAL
    commit id: "c2" type: REVERSE
    commit id: "c3" type: HIGHLIGHT`

	gg, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if gg.Commits[0].Type != Normal {
		t.Errorf("Commit c1 type = %d, want Normal (%d)", gg.Commits[0].Type, Normal)
	}
	if gg.Commits[1].Type != Reverse {
		t.Errorf("Commit c2 type = %d, want Reverse (%d)", gg.Commits[1].Type, Reverse)
	}
	if gg.Commits[2].Type != Highlight {
		t.Errorf("Commit c3 type = %d, want Highlight (%d)", gg.Commits[2].Type, Highlight)
	}
}

func TestRenderCommitTypes(t *testing.T) {
	input := `gitGraph
    commit id: "c1" type: REVERSE
    commit id: "c2" type: HIGHLIGHT`

	gg, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	output, err := Render(gg, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(output, "▲") {
		t.Error("Output should contain reverse commit char ▲")
	}
	if !strings.Contains(output, "◆") {
		t.Error("Output should contain highlight commit char ◆")
	}
}

func TestRenderWithTag(t *testing.T) {
	input := `gitGraph
    commit id: "c1" tag: "v1.0"`

	gg, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	output, err := Render(gg, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(output, "(tag: v1.0)") {
		t.Error("Output should contain tag annotation")
	}
}
