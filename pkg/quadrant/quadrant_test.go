package quadrant

import (
	"strings"
	"testing"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

func TestIsQuadrantChart(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"quadrantChart\n  title Test", true},
		{"%% comment\nquadrantChart", true},
		{"graph LR\n  A-->B", false},
		{"", false},
	}
	for _, tt := range tests {
		got := IsQuadrantChart(tt.input)
		if got != tt.want {
			t.Errorf("IsQuadrantChart(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParse(t *testing.T) {
	input := `quadrantChart
    title Reach and engagement
    x-axis Low Reach --> High Reach
    y-axis Low Engagement --> High Engagement
    quadrant-1 We should expand
    quadrant-2 Need to promote
    quadrant-3 Re-evaluate
    quadrant-4 May be improved
    Campaign A: [0.3, 0.6]
    Campaign B: [0.45, 0.23]
    Campaign C: [0.57, 0.69]`

	qc, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if qc.Title != "Reach and engagement" {
		t.Errorf("Title = %q, want %q", qc.Title, "Reach and engagement")
	}

	if qc.XAxisLeft != "Low Reach" {
		t.Errorf("XAxisLeft = %q, want %q", qc.XAxisLeft, "Low Reach")
	}

	if qc.XAxisRight != "High Reach" {
		t.Errorf("XAxisRight = %q, want %q", qc.XAxisRight, "High Reach")
	}

	if qc.Quadrant1 != "We should expand" {
		t.Errorf("Quadrant1 = %q, want %q", qc.Quadrant1, "We should expand")
	}

	if len(qc.Points) != 3 {
		t.Fatalf("Points count = %d, want 3", len(qc.Points))
	}

	if qc.Points[0].Label != "Campaign A" {
		t.Errorf("Point 0 label = %q, want %q", qc.Points[0].Label, "Campaign A")
	}
}

func TestRender(t *testing.T) {
	input := `quadrantChart
    title Test Chart
    x-axis Low --> High
    y-axis Small --> Large
    quadrant-1 Great
    quadrant-2 Good
    quadrant-3 Bad
    quadrant-4 OK
    Point A: [0.8, 0.9]
    Point B: [0.2, 0.3]`

	qc, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	output, err := Render(qc, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(output, "Test Chart") {
		t.Error("Output should contain title")
	}
	if !strings.Contains(output, "Great") {
		t.Error("Output should contain quadrant label")
	}
	if !strings.Contains(output, "Point A") {
		t.Error("Output should contain point label")
	}
}

func TestRenderASCII(t *testing.T) {
	input := `quadrantChart
    x-axis Low --> High
    y-axis Small --> Large
    A: [0.5, 0.5]`

	qc, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	config.UseAscii = true
	output, err := Render(qc, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if strings.Contains(output, "│") || strings.Contains(output, "─") || strings.Contains(output, "●") {
		t.Error("ASCII output should not contain Unicode characters")
	}
}
