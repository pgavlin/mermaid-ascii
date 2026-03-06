package sankey

import (
	"strings"
	"testing"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

func TestIsSankeyDiagram(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid", "sankey-beta", true},
		{"with comment", "%% comment\nsankey-beta", true},
		{"not sankey", "graph TD", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSankeyDiagram(tt.input); got != tt.want {
				t.Errorf("IsSankeyDiagram() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseFlows(t *testing.T) {
	input := `sankey-beta
Source A,Target X,100
Source B,Target Y,200
Source A,Target Z,50
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(d.Flows) != 3 {
		t.Fatalf("expected 3 flows, got %d", len(d.Flows))
	}
	if d.Flows[0].Source != "Source A" {
		t.Errorf("expected source 'Source A', got %q", d.Flows[0].Source)
	}
	if d.Flows[0].Target != "Target X" {
		t.Errorf("expected target 'Target X', got %q", d.Flows[0].Target)
	}
	if d.Flows[0].Value != 100 {
		t.Errorf("expected value 100, got %f", d.Flows[0].Value)
	}
}

func TestParseQuotedFields(t *testing.T) {
	input := `sankey-beta
"Source, A",Target X,100
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(d.Flows) != 1 {
		t.Fatalf("expected 1 flow, got %d", len(d.Flows))
	}
	if d.Flows[0].Source != "Source, A" {
		t.Errorf("expected source 'Source, A', got %q", d.Flows[0].Source)
	}
}

func TestRender(t *testing.T) {
	input := `sankey-beta
Electricity,Heat,50
Electricity,Light,30
Gas,Heat,20
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
	if !strings.Contains(result, "Electricity") {
		t.Error("expected output to contain 'Electricity'")
	}
	if !strings.Contains(result, "Heat") {
		t.Error("expected output to contain 'Heat'")
	}
	if !strings.Contains(result, "█") {
		t.Error("expected output to contain bar characters")
	}
}

func TestRenderASCII(t *testing.T) {
	input := `sankey-beta
A,B,100
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
	if !strings.Contains(result, "#") {
		t.Error("expected ASCII bar characters")
	}
	if strings.Contains(result, "█") {
		t.Error("expected no Unicode bar characters in ASCII mode")
	}
}

func TestRenderProportionalBars(t *testing.T) {
	input := `sankey-beta
A,B,100
C,D,50
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
	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	// First line should have more # than second
	count1 := strings.Count(lines[0], "#")
	count2 := strings.Count(lines[1], "#")
	if count1 <= count2 {
		t.Errorf("expected first flow to have more bar chars than second, got %d vs %d", count1, count2)
	}
}

func TestParseEmpty(t *testing.T) {
	_, err := Parse("")
	if err == nil {
		t.Error("expected error for empty input")
	}
}

func TestRenderNilDiagram(t *testing.T) {
	config := diagram.NewTestConfig(true, "cli")
	_, err := Render(nil, config)
	if err == nil {
		t.Error("expected error for nil diagram")
	}
}
