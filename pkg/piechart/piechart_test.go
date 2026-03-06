package piechart

import (
	"strings"
	"testing"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

func TestIsPieChart(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"pie\n  \"A\" : 50", true},
		{"pie title My Chart\n  \"A\" : 50", true},
		{"%% comment\npie", true},
		{"graph LR\n  A-->B", false},
		{"", false},
	}
	for _, tt := range tests {
		got := IsPieChart(tt.input)
		if got != tt.want {
			t.Errorf("IsPieChart(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParse(t *testing.T) {
	input := `pie title Pets
    "Dogs" : 45
    "Cats" : 30
    "Birds" : 15
    "Fish" : 10`

	pc, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if pc.Title != "Pets" {
		t.Errorf("Title = %q, want %q", pc.Title, "Pets")
	}

	if len(pc.Slices) != 4 {
		t.Fatalf("Slices count = %d, want 4", len(pc.Slices))
	}

	if pc.Slices[0].Label != "Dogs" {
		t.Errorf("Slice 0 label = %q, want %q", pc.Slices[0].Label, "Dogs")
	}

	if pc.Slices[0].Value != 45 {
		t.Errorf("Slice 0 value = %f, want 45", pc.Slices[0].Value)
	}

	// Check percentages
	if pc.Slices[0].Percentage != 45 {
		t.Errorf("Slice 0 percentage = %f, want 45", pc.Slices[0].Percentage)
	}
}

func TestParseWithTitleDirective(t *testing.T) {
	input := `pie
    title Pets
    "Dogs" : 60
    "Cats" : 40`

	pc, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if pc.Title != "Pets" {
		t.Errorf("Title = %q, want %q", pc.Title, "Pets")
	}
}

func TestParseEmptyInput(t *testing.T) {
	_, err := Parse("")
	if err == nil {
		t.Error("Expected error for empty input")
	}
}

func TestParseNoData(t *testing.T) {
	_, err := Parse("pie\n  title Empty")
	if err == nil {
		t.Error("Expected error for no data")
	}
}

func TestRender(t *testing.T) {
	input := `pie title Pets
    "Dogs" : 45
    "Cats" : 30
    "Birds" : 15
    "Fish" : 10`

	pc, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	output, err := Render(pc, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(output, "Pets") {
		t.Error("Output should contain title")
	}
	if !strings.Contains(output, "Dogs") {
		t.Error("Output should contain slice label")
	}
	if !strings.Contains(output, "%") {
		t.Error("Output should contain percentage")
	}
	if !strings.Contains(output, "█") {
		t.Error("Unicode output should contain bar fill character")
	}
}

func TestRenderASCII(t *testing.T) {
	input := `pie
    "A" : 50
    "B" : 50`

	pc, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	config.UseAscii = true
	output, err := Render(pc, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if strings.Contains(output, "█") {
		t.Error("ASCII output should not contain Unicode characters")
	}
	if !strings.Contains(output, "#") {
		t.Error("ASCII output should contain # bar fill character")
	}
}

func TestParseWithShowData(t *testing.T) {
	input := `pie showData
    title Chart
    "Dogs" : 50
    "Cats" : 50`

	// "pie showData" starts with "pie " so it should be accepted
	// but "showData" is not "title X" so the title won't be set from first line
	pc, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if pc.Title != "Chart" {
		t.Errorf("Title = %q, want 'Chart'", pc.Title)
	}
	if len(pc.Slices) != 2 {
		t.Fatalf("Slices count = %d, want 2", len(pc.Slices))
	}
}

func TestParseShowDataDirective(t *testing.T) {
	input := `pie
    showData
    "A" : 30
    "B" : 70`

	pc, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(pc.Slices) != 2 {
		t.Fatalf("Slices count = %d, want 2", len(pc.Slices))
	}
}

func TestRenderNilDiagram(t *testing.T) {
	_, err := Render(nil, nil)
	if err == nil {
		t.Error("Expected error for nil diagram")
	}
}

func TestRenderNilConfig(t *testing.T) {
	input := `pie
    "A" : 60
    "B" : 40`

	pc, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	output, err := Render(pc, nil)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(output, "A") {
		t.Error("Output should contain slice label")
	}
}

func TestParseNoDataEntries(t *testing.T) {
	_, err := Parse("pie\n  showData")
	if err == nil {
		t.Error("Expected error for no data entries")
	}
}

func TestRenderEmptySlices(t *testing.T) {
	pc := &PieChart{Slices: []*Slice{}}
	_, err := Render(pc, nil)
	if err == nil {
		t.Error("Expected error for empty slices")
	}
}

func TestParseInvalidKeyword(t *testing.T) {
	_, err := Parse("graph\n  \"A\" : 50")
	if err == nil {
		t.Error("Expected error for invalid keyword")
	}
}
