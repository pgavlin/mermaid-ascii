package blockdiagram

import (
	"strings"
	"testing"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

func TestIsBlockDiagram(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid", "block-beta", true},
		{"with comment", "%% comment\nblock-beta", true},
		{"not block", "graph TD", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsBlockDiagram(tt.input); got != tt.want {
				t.Errorf("IsBlockDiagram() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseColumns(t *testing.T) {
	input := `block-beta
columns 3
a["Block A"]
b["Block B"]
c["Block C"]
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if d.Columns != 3 {
		t.Errorf("expected 3 columns, got %d", d.Columns)
	}
	if len(d.Blocks) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(d.Blocks))
	}
	if d.Blocks[0].Label != "Block A" {
		t.Errorf("expected label 'Block A', got %q", d.Blocks[0].Label)
	}
}

func TestParseNestedBlock(t *testing.T) {
	input := `block-beta
columns 2
block:group
  a["Inner A"]
  b["Inner B"]
end
c["Outer"]
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(d.Blocks) != 2 {
		t.Fatalf("expected 2 top-level blocks, got %d", len(d.Blocks))
	}
	if len(d.Blocks[0].Children) != 2 {
		t.Fatalf("expected 2 children in first block, got %d", len(d.Blocks[0].Children))
	}
}

func TestRenderGrid(t *testing.T) {
	input := `block-beta
columns 2
a["A"]
b["B"]
c["C"]
d["D"]
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	config := diagram.NewTestConfig(false, "cli")
	result, err := Render(d, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	if !strings.Contains(result, "A") {
		t.Error("expected output to contain 'A'")
	}
	if !strings.Contains(result, "B") {
		t.Error("expected output to contain 'B'")
	}
	// Should have multiple rows
	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) < 6 {
		t.Errorf("expected at least 6 lines for 2x2 grid, got %d", len(lines))
	}
}

func TestRenderASCII(t *testing.T) {
	input := `block-beta
a["Test"]
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	config := diagram.NewTestConfig(true, "cli")
	result, err := Render(d, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	if !strings.Contains(result, "+") {
		t.Error("expected ASCII characters")
	}
}

func TestParseEmpty(t *testing.T) {
	_, err := Parse("")
	if err == nil {
		t.Error("expected error for empty input")
	}
}

func TestParseSimpleNames(t *testing.T) {
	input := `block-beta
columns 2
foo
bar
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(d.Blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(d.Blocks))
	}
	if d.Blocks[0].ID != "foo" {
		t.Errorf("expected ID 'foo', got %q", d.Blocks[0].ID)
	}
	if d.Blocks[0].Label != "foo" {
		t.Errorf("expected label 'foo', got %q", d.Blocks[0].Label)
	}
}

func TestParseColumns3(t *testing.T) {
	input := `block-beta
columns 3
a["Alpha"]
b["Beta"]
c["Gamma"]
d["Delta"]
e["Epsilon"]
f["Zeta"]
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if d.Columns != 3 {
		t.Errorf("expected 3 columns, got %d", d.Columns)
	}
	if len(d.Blocks) != 6 {
		t.Fatalf("expected 6 blocks, got %d", len(d.Blocks))
	}
	// Verify all labels
	expected := []string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon", "Zeta"}
	for i, exp := range expected {
		if d.Blocks[i].Label != exp {
			t.Errorf("block %d: expected label %q, got %q", i, exp, d.Blocks[i].Label)
		}
	}
}

func TestParseSpanSyntax(t *testing.T) {
	input := `block-beta
columns 3
a["Small"]:1
b["Wide"]:2
c["Full"]:3
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(d.Blocks) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(d.Blocks))
	}
	if d.Blocks[0].Span != 1 {
		t.Errorf("block 0: expected span 1, got %d", d.Blocks[0].Span)
	}
	if d.Blocks[1].Span != 2 {
		t.Errorf("block 1: expected span 2, got %d", d.Blocks[1].Span)
	}
	if d.Blocks[2].Span != 3 {
		t.Errorf("block 2: expected span 3, got %d", d.Blocks[2].Span)
	}
}

func TestRenderSpanBlocks(t *testing.T) {
	input := `block-beta
columns 3
a["A"]:1
b["B"]:2
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	config := diagram.NewTestConfig(false, "cli")
	result, err := Render(d, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	if !strings.Contains(result, "A") {
		t.Error("expected output to contain 'A'")
	}
	if !strings.Contains(result, "B") {
		t.Error("expected output to contain 'B'")
	}
	// The span-2 block should render wider
	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) < 3 {
		t.Errorf("expected at least 3 lines, got %d", len(lines))
	}
}

func TestParseNestedBlockWithChildren(t *testing.T) {
	input := `block-beta
columns 2
block:container
  columns 2
  x["X"]
  y["Y"]
end
z["Z"]
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(d.Blocks) != 2 {
		t.Fatalf("expected 2 top-level blocks, got %d", len(d.Blocks))
	}
	container := d.Blocks[0]
	if container.ID != "container" {
		t.Errorf("expected container ID 'container', got %q", container.ID)
	}
	if container.Columns != 2 {
		t.Errorf("expected container columns 2, got %d", container.Columns)
	}
	if len(container.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(container.Children))
	}
	if container.Children[0].Label != "X" {
		t.Errorf("expected first child label 'X', got %q", container.Children[0].Label)
	}
	if container.Children[1].Label != "Y" {
		t.Errorf("expected second child label 'Y', got %q", container.Children[1].Label)
	}
}

func TestRenderNestedBlockWithChildren(t *testing.T) {
	input := `block-beta
columns 2
block:parent
  columns 2
  a["Child1"]
  b["Child2"]
end
c["Sibling"]
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	config := diagram.NewTestConfig(false, "cli")
	result, err := Render(d, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	if !strings.Contains(result, "parent") {
		t.Error("expected output to contain 'parent'")
	}
	if !strings.Contains(result, "Child1") {
		t.Error("expected output to contain 'Child1'")
	}
	if !strings.Contains(result, "Child2") {
		t.Error("expected output to contain 'Child2'")
	}
	if !strings.Contains(result, "Sibling") {
		t.Error("expected output to contain 'Sibling'")
	}
}

func TestRenderNilDiagram(t *testing.T) {
	config := diagram.NewTestConfig(false, "cli")
	_, err := Render(nil, config)
	if err == nil {
		t.Error("expected error for nil diagram")
	}
	if !strings.Contains(err.Error(), "nil diagram") {
		t.Errorf("expected 'nil diagram' error, got %q", err.Error())
	}
}

func TestRenderNilConfig(t *testing.T) {
	input := `block-beta
a["Test"]
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	// Render with nil config should use defaults (Unicode)
	result, err := Render(d, nil)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	if !strings.Contains(result, "Test") {
		t.Error("expected output to contain 'Test'")
	}
	// Should use Unicode characters by default
	if !strings.Contains(result, "┌") {
		t.Error("expected Unicode box-drawing characters when config is nil")
	}
}

func TestParseBlockWithoutLabel(t *testing.T) {
	input := `block-beta
block
  inner["Inside"]
end
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(d.Blocks) != 1 {
		t.Fatalf("expected 1 top-level block, got %d", len(d.Blocks))
	}
	b := d.Blocks[0]
	// Block without label should have auto-generated ID and empty label
	if b.Label != "" {
		t.Errorf("expected empty label for unnamed block, got %q", b.Label)
	}
	if len(b.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(b.Children))
	}
	if b.Children[0].Label != "Inside" {
		t.Errorf("expected child label 'Inside', got %q", b.Children[0].Label)
	}
}

func TestRenderBlockWithoutLabel(t *testing.T) {
	input := `block-beta
block
  a["Content"]
end
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	config := diagram.NewTestConfig(false, "cli")
	result, err := Render(d, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	if !strings.Contains(result, "Content") {
		t.Error("expected output to contain child 'Content'")
	}
}

func TestParseRoundBracketLabel(t *testing.T) {
	input := `block-beta
a("Round Label")
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(d.Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(d.Blocks))
	}
	if d.Blocks[0].Label != "Round Label" {
		t.Errorf("expected label 'Round Label', got %q", d.Blocks[0].Label)
	}
}

func TestParseNoContentError(t *testing.T) {
	input := "%% just a comment"
	_, err := Parse(input)
	if err == nil {
		t.Error("expected error for comment-only input")
	}
}

func TestParseNotBlockBeta(t *testing.T) {
	input := "graph TD\nA --> B"
	_, err := Parse(input)
	if err == nil {
		t.Error("expected error for non-block-beta input")
	}
}

func TestRenderColumns3Grid(t *testing.T) {
	input := `block-beta
columns 3
a["A"]
b["B"]
c["C"]
d["D"]
e["E"]
f["F"]
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	config := diagram.NewTestConfig(false, "cli")
	result, err := Render(d, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(result), "\n")
	// 3 columns, 6 blocks = 2 rows, each row is 3 lines (top, mid, bot)
	if len(lines) < 6 {
		t.Errorf("expected at least 6 lines for 2 rows of 3, got %d", len(lines))
	}
	for _, label := range []string{"A", "B", "C", "D", "E", "F"} {
		if !strings.Contains(result, label) {
			t.Errorf("expected output to contain %q", label)
		}
	}
}
