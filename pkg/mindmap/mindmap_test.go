package mindmap

import (
	"strings"
	"testing"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

func TestIsMindmapDiagram(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"mindmap\n  root", true},
		{"%% comment\nmindmap", true},
		{"graph LR\n  A-->B", false},
		{"", false},
	}
	for _, tt := range tests {
		got := IsMindmapDiagram(tt.input)
		if got != tt.want {
			t.Errorf("IsMindmapDiagram(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParse(t *testing.T) {
	input := `mindmap
  root
    Child 1
      Grandchild 1
      Grandchild 2
    Child 2
    Child 3`

	md, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if md.Root == nil {
		t.Fatal("Root is nil")
	}

	if md.Root.Text != "root" {
		t.Errorf("Root text = %q, want %q", md.Root.Text, "root")
	}

	if len(md.Root.Children) != 3 {
		t.Fatalf("Root children count = %d, want 3", len(md.Root.Children))
	}

	if md.Root.Children[0].Text != "Child 1" {
		t.Errorf("Child 0 text = %q, want %q", md.Root.Children[0].Text, "Child 1")
	}

	if len(md.Root.Children[0].Children) != 2 {
		t.Errorf("Child 0 children count = %d, want 2", len(md.Root.Children[0].Children))
	}
}

func TestParseWithShapes(t *testing.T) {
	input := `mindmap
  root
    [Square]
    (Rounded)
    {{Hexagon}}`

	md, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if md.Root.Children[0].Shape != ShapeSquare {
		t.Errorf("Child 0 shape = %d, want ShapeSquare", md.Root.Children[0].Shape)
	}
	if md.Root.Children[1].Shape != ShapeRounded {
		t.Errorf("Child 1 shape = %d, want ShapeRounded", md.Root.Children[1].Shape)
	}
	if md.Root.Children[2].Shape != ShapeHexagon {
		t.Errorf("Child 2 shape = %d, want ShapeHexagon", md.Root.Children[2].Shape)
	}
}

func TestRender(t *testing.T) {
	input := `mindmap
  Project
    Backend
      API
      Database
    Frontend
      UI
      Tests`

	md, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	output, err := Render(md, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(output, "Project") {
		t.Error("Output should contain root")
	}
	if !strings.Contains(output, "Backend") {
		t.Error("Output should contain child")
	}
	if !strings.Contains(output, "API") {
		t.Error("Output should contain grandchild")
	}
}

func TestRenderASCII(t *testing.T) {
	input := `mindmap
  Root
    Child`

	md, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	config.UseAscii = true
	output, err := Render(md, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if strings.Contains(output, "├") || strings.Contains(output, "│") {
		t.Error("ASCII output should not contain Unicode characters")
	}
}

func TestRenderNodeShapes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
		ascii    bool
	}{
		{
			name: "square shape unicode",
			input: `mindmap
  root
    [Square]`,
			contains: []string{"┌", "┐", "└", "┘", "Square"},
			ascii:    false,
		},
		{
			name: "square shape ascii",
			input: `mindmap
  root
    [Square]`,
			contains: []string{"[Square]"},
			ascii:    true,
		},
		{
			name: "rounded shape unicode",
			input: `mindmap
  root
    (Rounded)`,
			contains: []string{"╭", "╮", "╰", "╯", "Rounded"},
			ascii:    false,
		},
		{
			name: "rounded shape ascii",
			input: `mindmap
  root
    (Rounded)`,
			contains: []string{"(Rounded)"},
			ascii:    true,
		},
		{
			name: "hexagon shape unicode",
			input: `mindmap
  root
    {{Hexagon}}`,
			contains: []string{"⬡", "Hexagon"},
			ascii:    false,
		},
		{
			name: "hexagon shape ascii",
			input: `mindmap
  root
    {{Hexagon}}`,
			contains: []string{"{{Hexagon}}"},
			ascii:    true,
		},
		{
			name: "bang shape",
			input: `mindmap
  root
    ))Bang((`,
			contains: []string{"💥", "Bang"},
			ascii:    false,
		},
		{
			name: "cloud shape",
			input: `mindmap
  root
    )Cloud(`,
			contains: []string{"☁", "Cloud"},
			ascii:    false,
		},
		{
			name: "default shape",
			input: `mindmap
  root
    PlainText`,
			contains: []string{"PlainText"},
			ascii:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}
			config := diagram.DefaultConfig()
			config.UseAscii = tt.ascii
			output, err := Render(md, config)
			if err != nil {
				t.Fatalf("Render failed: %v", err)
			}
			for _, s := range tt.contains {
				if !strings.Contains(output, s) {
					t.Errorf("output should contain %q, got:\n%s", s, output)
				}
			}
		})
	}
}

func TestRenderNilDiagram(t *testing.T) {
	config := diagram.DefaultConfig()

	_, err := Render(nil, config)
	if err == nil {
		t.Error("expected error for nil diagram")
	}

	md := &MindmapDiagram{Root: nil}
	_, err = Render(md, config)
	if err == nil {
		t.Error("expected error for nil root")
	}
}

func TestRenderNilConfig(t *testing.T) {
	input := `mindmap
  Root
    Child`
	md, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	output, err := Render(md, nil)
	if err != nil {
		t.Fatalf("Render with nil config failed: %v", err)
	}
	if !strings.Contains(output, "Root") {
		t.Error("output should contain Root")
	}
}

func TestDeeplyNestedMindmap(t *testing.T) {
	input := `mindmap
  Root
    Level1
      Level2
        Level3
          Level4`

	md, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if md.Root.Text != "Root" {
		t.Errorf("Root text = %q, want %q", md.Root.Text, "Root")
	}

	// Traverse down the tree
	node := md.Root
	expected := []string{"Root", "Level1", "Level2", "Level3", "Level4"}
	for i, exp := range expected {
		if node.Text != exp {
			t.Errorf("depth %d: text = %q, want %q", i, node.Text, exp)
		}
		if i < len(expected)-1 {
			if len(node.Children) == 0 {
				t.Fatalf("depth %d: expected children", i)
			}
			node = node.Children[0]
		}
	}

	// Also render to make sure deep nesting renders without error
	config := diagram.DefaultConfig()
	output, err := Render(md, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("output should contain %q", exp)
		}
	}
}

func TestParseWithComments(t *testing.T) {
	input := `%% This is a comment
mindmap
  %% Another comment
  Root
    %% Comment between nodes
    Child1
    Child2`

	md, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if md.Root.Text != "Root" {
		t.Errorf("Root text = %q, want %q", md.Root.Text, "Root")
	}
	if len(md.Root.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(md.Root.Children))
	}
	if md.Root.Children[0].Text != "Child1" {
		t.Errorf("Child 0 text = %q, want %q", md.Root.Children[0].Text, "Child1")
	}
	if md.Root.Children[1].Text != "Child2" {
		t.Errorf("Child 1 text = %q, want %q", md.Root.Children[1].Text, "Child2")
	}
}

func TestParseEmptyInput(t *testing.T) {
	_, err := Parse("")
	if err == nil {
		t.Error("expected error for empty input")
	}
}

func TestParseMindmapOnly(t *testing.T) {
	_, err := Parse("mindmap")
	if err == nil {
		t.Error("expected error for mindmap with no nodes")
	}
}

func TestParseNotMindmap(t *testing.T) {
	_, err := Parse("graph LR\n  A-->B")
	if err == nil {
		t.Error("expected error for non-mindmap input")
	}
}

func TestParseNodeTextShapes(t *testing.T) {
	tests := []struct {
		input string
		text  string
		shape NodeShape
	}{
		{"[Square]", "Square", ShapeSquare},
		{"(Rounded)", "Rounded", ShapeRounded},
		{"{{Hexagon}}", "Hexagon", ShapeHexagon},
		{"))Bang((", "Bang", ShapeBang},
		{")Cloud(", "Cloud", ShapeCloud},
		{"PlainText", "PlainText", ShapeDefault},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			node := parseNodeText(tt.input)
			if node.Text != tt.text {
				t.Errorf("parseNodeText(%q).Text = %q, want %q", tt.input, node.Text, tt.text)
			}
			if node.Shape != tt.shape {
				t.Errorf("parseNodeText(%q).Shape = %d, want %d", tt.input, node.Shape, tt.shape)
			}
		})
	}
}

func TestNodeSpecialCharacters(t *testing.T) {
	input := `mindmap
  Root & Base
    Child: Item #1
      Grand/child (special)`

	md, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if md.Root.Text != "Root & Base" {
		t.Errorf("Root text = %q, want %q", md.Root.Text, "Root & Base")
	}
	if md.Root.Children[0].Text != "Child: Item #1" {
		t.Errorf("Child text = %q, want %q", md.Root.Children[0].Text, "Child: Item #1")
	}

	config := diagram.DefaultConfig()
	output, err := Render(md, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if !strings.Contains(output, "Root & Base") {
		t.Error("output should contain special chars in root")
	}
}

func TestParseBlankLinesIgnored(t *testing.T) {
	input := `mindmap

  Root

    Child1

    Child2
`

	md, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if md.Root.Text != "Root" {
		t.Errorf("Root text = %q, want %q", md.Root.Text, "Root")
	}
	if len(md.Root.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(md.Root.Children))
	}
}

func TestRenderAllShapesInTree(t *testing.T) {
	input := `mindmap
  root
    [Square]
    (Rounded)
    {{Hexagon}}
    ))Bang((
    )Cloud(`

	md, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Unicode rendering
	config := diagram.DefaultConfig()
	output, err := Render(md, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if !strings.Contains(output, "Square") {
		t.Error("output should contain Square")
	}
	if !strings.Contains(output, "Rounded") {
		t.Error("output should contain Rounded")
	}
	if !strings.Contains(output, "Hexagon") {
		t.Error("output should contain Hexagon")
	}
	if !strings.Contains(output, "Bang") {
		t.Error("output should contain Bang")
	}
	if !strings.Contains(output, "Cloud") {
		t.Error("output should contain Cloud")
	}

	// ASCII rendering
	config.UseAscii = true
	outputAscii, err := Render(md, config)
	if err != nil {
		t.Fatalf("Render ASCII failed: %v", err)
	}
	if !strings.Contains(outputAscii, "[Square]") {
		t.Error("ASCII output should contain [Square]")
	}
	if !strings.Contains(outputAscii, "(Rounded)") {
		t.Error("ASCII output should contain (Rounded)")
	}
	if !strings.Contains(outputAscii, "{{Hexagon}}") {
		t.Error("ASCII output should contain {{Hexagon}}")
	}
}

func TestRenderMultipleChildrenBranching(t *testing.T) {
	input := `mindmap
  Root
    A
    B
    C`

	md, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	output, err := Render(md, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// First children should use branch char, last should use last-branch char
	if !strings.Contains(output, "├") {
		t.Error("output should contain branch char ├ for non-last children")
	}
	if !strings.Contains(output, "└") {
		t.Error("output should contain last-branch char └ for last child")
	}
}
