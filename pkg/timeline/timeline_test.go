package timeline

import (
	"strings"
	"testing"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

func TestIsTimelineDiagram(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"timeline\n  2024 : Event", true},
		{"%% comment\ntimeline", true},
		{"graph LR\n  A-->B", false},
		{"", false},
	}
	for _, tt := range tests {
		got := IsTimelineDiagram(tt.input)
		if got != tt.want {
			t.Errorf("IsTimelineDiagram(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParse(t *testing.T) {
	input := `timeline
    title History
    section Ancient
    3000 BC : Pyramids built
    500 BC : Democracy in Athens
    section Modern
    1969 : Moon landing`

	td, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if td.Title != "History" {
		t.Errorf("Title = %q, want %q", td.Title, "History")
	}

	if len(td.Sections) != 2 {
		t.Errorf("Sections count = %d, want 2", len(td.Sections))
	}

	if len(td.Events) != 3 {
		t.Errorf("Events count = %d, want 3", len(td.Events))
	}

	if td.Events[0].Period != "3000 BC" {
		t.Errorf("Event 0 period = %q, want %q", td.Events[0].Period, "3000 BC")
	}
}

func TestParseEmptyInput(t *testing.T) {
	_, err := Parse("")
	if err == nil {
		t.Error("Expected error for empty input")
	}
}

func TestRender(t *testing.T) {
	input := `timeline
    title History
    section Ancient
    3000 BC : Pyramids
    section Modern
    1969 : Moon landing`

	td, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	output, err := Render(td, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(output, "History") {
		t.Error("Output should contain title")
	}
	if !strings.Contains(output, "3000 BC") {
		t.Error("Output should contain period")
	}
	if !strings.Contains(output, "Pyramids") {
		t.Error("Output should contain event")
	}
}

func TestRenderASCII(t *testing.T) {
	input := `timeline
    2024 : Event`

	td, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	config.UseAscii = true
	output, err := Render(td, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if strings.Contains(output, "│") || strings.Contains(output, "─") {
		t.Error("ASCII output should not contain Unicode characters")
	}
}
