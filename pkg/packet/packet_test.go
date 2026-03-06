package packet

import (
	"strings"
	"testing"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

func TestIsPacketDiagram(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"packet-beta\n  0-3: \"Version\"", true},
		{"%% comment\npacket-beta", true},
		{"graph LR\n  A-->B", false},
		{"", false},
	}
	for _, tt := range tests {
		got := IsPacketDiagram(tt.input)
		if got != tt.want {
			t.Errorf("IsPacketDiagram(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParse(t *testing.T) {
	input := `packet-beta
    0-3: "Version"
    4-7: "IHL"
    8-15: "Type of Service"
    16-31: "Total Length"`

	pd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(pd.Fields) != 4 {
		t.Fatalf("Fields count = %d, want 4", len(pd.Fields))
	}

	if pd.Fields[0].Label != "Version" {
		t.Errorf("Field 0 label = %q, want %q", pd.Fields[0].Label, "Version")
	}

	if pd.Fields[0].StartBit != 0 || pd.Fields[0].EndBit != 3 {
		t.Errorf("Field 0 bits = %d-%d, want 0-3", pd.Fields[0].StartBit, pd.Fields[0].EndBit)
	}
}

func TestRender(t *testing.T) {
	input := `packet-beta
    0-3: "Version"
    4-7: "IHL"
    8-15: "Type of Service"`

	pd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	output, err := Render(pd, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(output, "Version") {
		t.Error("Output should contain field label")
	}
	if !strings.Contains(output, "IHL") {
		t.Error("Output should contain field label")
	}
}

func TestRenderASCII(t *testing.T) {
	input := `packet-beta
    0-7: "Header"
    8-15: "Data"`

	pd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	config.UseAscii = true
	output, err := Render(pd, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if strings.Contains(output, "│") || strings.Contains(output, "─") {
		t.Error("ASCII output should not contain Unicode characters")
	}
}

func TestParseBitRangeSyntax(t *testing.T) {
	input := `packet-beta
    0-7: "Header"
    8-15: "Payload"
    16: "Flag"`

	pd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(pd.Fields) != 3 {
		t.Fatalf("Fields count = %d, want 3", len(pd.Fields))
	}

	// Check range field
	if pd.Fields[0].StartBit != 0 || pd.Fields[0].EndBit != 7 {
		t.Errorf("Field 0 bits = %d-%d, want 0-7", pd.Fields[0].StartBit, pd.Fields[0].EndBit)
	}

	// Check single bit field (no range)
	if pd.Fields[2].StartBit != 16 || pd.Fields[2].EndBit != 16 {
		t.Errorf("Field 2 bits = %d-%d, want 16-16", pd.Fields[2].StartBit, pd.Fields[2].EndBit)
	}
	if pd.Fields[2].Label != "Flag" {
		t.Errorf("Field 2 label = %q, want 'Flag'", pd.Fields[2].Label)
	}
}

func TestRenderNilDiagram(t *testing.T) {
	_, err := Render(nil, nil)
	if err == nil {
		t.Error("Expected error for nil diagram")
	}
}

func TestRenderNilConfig(t *testing.T) {
	input := `packet-beta
    0-7: "Header"
    8-15: "Data"`

	pd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	output, err := Render(pd, nil)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(output, "Header") {
		t.Error("Output should contain field label")
	}
}

func TestParseInvalidInput(t *testing.T) {
	// Wrong keyword
	_, err := Parse("graph\n  0-7: \"Header\"")
	if err == nil {
		t.Error("Expected error for invalid keyword")
	}

	// Empty input
	_, err = Parse("")
	if err == nil {
		t.Error("Expected error for empty input")
	}

	// No fields
	_, err = Parse("packet-beta\n  some invalid line")
	if err == nil {
		t.Error("Expected error for no fields")
	}
}

func TestRenderEmptyFields(t *testing.T) {
	pd := &PacketDiagram{Fields: []*Field{}}
	_, err := Render(pd, nil)
	if err == nil {
		t.Error("Expected error for empty fields")
	}
}

func TestRenderMultipleRows(t *testing.T) {
	input := `packet-beta
    0-15: "First Half"
    16-31: "Second Half"
    32-47: "Third (row 2)"`

	pd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	output, err := Render(pd, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(output, "First Half") {
		t.Error("Output should contain 'First Half'")
	}
	if !strings.Contains(output, "Third") {
		t.Error("Output should contain field from second row")
	}
}
