package c4

import (
	"strings"
	"testing"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

func TestIsC4Diagram(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"C4Context", "C4Context", true},
		{"C4Container", "C4Container", true},
		{"C4Component", "C4Component", true},
		{"C4Dynamic", "C4Dynamic", true},
		{"not c4", "graph TD", false},
		{"empty", "", false},
		{"with comment", "%% comment\nC4Context", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsC4Diagram(tt.input); got != tt.want {
				t.Errorf("IsC4Diagram() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseBasic(t *testing.T) {
	input := `C4Context
Person(user, "User", "A user of the system")
System(system, "System", "The main system")
Rel(user, system, "Uses")
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(d.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(d.Elements))
	}
	if d.Elements[0].Label != "User" {
		t.Errorf("expected label 'User', got %q", d.Elements[0].Label)
	}
	if d.Elements[0].Kind != "person" {
		t.Errorf("expected kind 'person', got %q", d.Elements[0].Kind)
	}
	if d.Elements[1].Label != "System" {
		t.Errorf("expected label 'System', got %q", d.Elements[1].Label)
	}
	if len(d.Relationships) != 1 {
		t.Fatalf("expected 1 relationship, got %d", len(d.Relationships))
	}
	if d.Relationships[0].Label != "Uses" {
		t.Errorf("expected rel label 'Uses', got %q", d.Relationships[0].Label)
	}
}

func TestParseBoundary(t *testing.T) {
	input := `C4Context
System_Boundary(b1, "Boundary") {
  Container(c1, "Container 1", "Go", "A container")
}
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(d.Boundaries) != 1 {
		t.Fatalf("expected 1 boundary, got %d", len(d.Boundaries))
	}
	if d.Boundaries[0].Label != "Boundary" {
		t.Errorf("expected boundary label 'Boundary', got %q", d.Boundaries[0].Label)
	}
	if len(d.Boundaries[0].Elements) != 1 {
		t.Fatalf("expected 1 element in boundary, got %d", len(d.Boundaries[0].Elements))
	}
	if d.Boundaries[0].Elements[0].Label != "Container 1" {
		t.Errorf("expected element label 'Container 1', got %q", d.Boundaries[0].Elements[0].Label)
	}
}

func TestRender(t *testing.T) {
	input := `C4Context
Person(user, "User", "A user")
System(sys, "System", "Main system")
Rel(user, sys, "Uses")
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
	if !strings.Contains(result, "User") {
		t.Error("expected output to contain 'User'")
	}
	if !strings.Contains(result, "System") {
		t.Error("expected output to contain 'System'")
	}
	if !strings.Contains(result, "-->") {
		t.Error("expected output to contain '-->'")
	}
}

func TestRenderASCII(t *testing.T) {
	input := `C4Context
Person(user, "User", "A user")
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
		t.Error("expected ASCII box characters")
	}
	if strings.Contains(result, "┌") {
		t.Error("expected no Unicode box characters in ASCII mode")
	}
}

func TestParseEmpty(t *testing.T) {
	_, err := Parse("")
	if err == nil {
		t.Error("expected error for empty input")
	}
}

func TestParseInvalidKeyword(t *testing.T) {
	_, err := Parse("graph TD")
	if err == nil {
		t.Error("expected error for invalid keyword")
	}
}

func TestRenderBoundary(t *testing.T) {
	input := `C4Context
System_Boundary(b1, "My Boundary") {
  Container(c1, "Container 1", "Go", "First container")
  Container(c2, "Container 2", "Java", "Second container")
}
Person(user, "User", "A user")
Rel(user, c1, "Uses")
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(d.Boundaries) != 1 {
		t.Fatalf("expected 1 boundary, got %d", len(d.Boundaries))
	}
	if len(d.Boundaries[0].Elements) != 2 {
		t.Fatalf("expected 2 elements in boundary, got %d", len(d.Boundaries[0].Elements))
	}

	config := diagram.NewTestConfig(false, "cli")
	result, err := Render(d, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	if !strings.Contains(result, "My Boundary") {
		t.Error("expected output to contain boundary label 'My Boundary'")
	}
	if !strings.Contains(result, "Container 1") {
		t.Error("expected output to contain 'Container 1'")
	}
	if !strings.Contains(result, "Container 2") {
		t.Error("expected output to contain 'Container 2'")
	}
	if !strings.Contains(result, "┌") {
		t.Error("expected unicode box characters in boundary rendering")
	}
}

func TestRenderBoundaryASCII(t *testing.T) {
	input := `C4Context
System_Boundary(b1, "Boundary") {
  Container(c1, "Svc", "Go", "A service")
}
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
	if !strings.Contains(result, "Boundary") {
		t.Error("expected output to contain 'Boundary'")
	}
	if strings.Contains(result, "┌") {
		t.Error("expected no Unicode box characters in ASCII mode")
	}
	if !strings.Contains(result, "+") {
		t.Error("expected ASCII box characters")
	}
}

func TestParseContainerBoundary(t *testing.T) {
	input := `C4Container
Container_Boundary(cb1, "Container Boundary") {
  Component(comp1, "Component A", "Go", "Does things")
}
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(d.Boundaries) != 1 {
		t.Fatalf("expected 1 boundary, got %d", len(d.Boundaries))
	}
	if d.Boundaries[0].Label != "Container Boundary" {
		t.Errorf("expected boundary label 'Container Boundary', got %q", d.Boundaries[0].Label)
	}
	if len(d.Boundaries[0].Elements) != 1 {
		t.Fatalf("expected 1 element in boundary, got %d", len(d.Boundaries[0].Elements))
	}
	if d.Boundaries[0].Elements[0].Kind != "component" {
		t.Errorf("expected kind 'component', got %q", d.Boundaries[0].Elements[0].Kind)
	}
	if d.Boundaries[0].Elements[0].Technology != "Go" {
		t.Errorf("expected technology 'Go', got %q", d.Boundaries[0].Elements[0].Technology)
	}
}

func TestParseEnterpriseBoundary(t *testing.T) {
	input := `C4Context
Enterprise_Boundary(eb1, "Enterprise") {
  System(s1, "System A", "A system")
  System(s2, "System B", "Another system")
}
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(d.Boundaries) != 1 {
		t.Fatalf("expected 1 boundary, got %d", len(d.Boundaries))
	}
	if d.Boundaries[0].Label != "Enterprise" {
		t.Errorf("expected boundary label 'Enterprise', got %q", d.Boundaries[0].Label)
	}
	if len(d.Boundaries[0].Elements) != 2 {
		t.Fatalf("expected 2 elements in boundary, got %d", len(d.Boundaries[0].Elements))
	}

	config := diagram.NewTestConfig(false, "cli")
	result, err := Render(d, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	if !strings.Contains(result, "Enterprise") {
		t.Error("expected output to contain 'Enterprise'")
	}
	if !strings.Contains(result, "System A") {
		t.Error("expected output to contain 'System A'")
	}
	if !strings.Contains(result, "System B") {
		t.Error("expected output to contain 'System B'")
	}
}

func TestParseSystemExt(t *testing.T) {
	input := `C4Context
System_Ext(ext1, "External System", "An external system")
Person(user, "User", "A user")
Rel(user, ext1, "Calls")
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(d.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(d.Elements))
	}
	if d.Elements[0].Label != "External System" {
		t.Errorf("expected label 'External System', got %q", d.Elements[0].Label)
	}
	if d.Elements[0].Kind != "system" {
		t.Errorf("expected kind 'system', got %q", d.Elements[0].Kind)
	}
}

func TestParseContainerExt(t *testing.T) {
	input := `C4Container
Container_Ext(ext1, "External Container", "Python", "An external container")
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(d.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(d.Elements))
	}
	if d.Elements[0].Label != "External Container" {
		t.Errorf("expected label 'External Container', got %q", d.Elements[0].Label)
	}
	if d.Elements[0].Kind != "container" {
		t.Errorf("expected kind 'container', got %q", d.Elements[0].Kind)
	}
	if d.Elements[0].Technology != "Python" {
		t.Errorf("expected technology 'Python', got %q", d.Elements[0].Technology)
	}
}

func TestParseComponent(t *testing.T) {
	input := `C4Component
Component(comp1, "My Component", "Go", "Does stuff")
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(d.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(d.Elements))
	}
	if d.Elements[0].Kind != "component" {
		t.Errorf("expected kind 'component', got %q", d.Elements[0].Kind)
	}
	if d.Elements[0].Technology != "Go" {
		t.Errorf("expected technology 'Go', got %q", d.Elements[0].Technology)
	}
	if d.Elements[0].Description != "Does stuff" {
		t.Errorf("expected description 'Does stuff', got %q", d.Elements[0].Description)
	}
}

func TestRenderNilConfig(t *testing.T) {
	input := `C4Context
Person(user, "User", "A user")
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	result, err := Render(d, nil)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	if !strings.Contains(result, "User") {
		t.Error("expected output to contain 'User'")
	}
	// nil config should default to unicode
	if !strings.Contains(result, "┌") {
		t.Error("expected unicode box characters with nil config (default)")
	}
}

func TestRenderNilDiagram(t *testing.T) {
	_, err := Render(nil, nil)
	if err == nil {
		t.Error("expected error for nil diagram")
	}
}

func TestRenderEmptyDiagram(t *testing.T) {
	d := &C4Diagram{
		DiagramType: C4Context,
	}
	config := diagram.NewTestConfig(false, "cli")
	result, err := Render(d, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	// Should produce minimal output (just a newline)
	if result != "\n" {
		t.Errorf("expected empty diagram to produce just newline, got %q", result)
	}
}

func TestParseNoContent(t *testing.T) {
	input := `%% just a comment
%% another comment`
	_, err := Parse(input)
	if err == nil {
		t.Error("expected error for input with no content")
	}
}

func TestMultipleRelationships(t *testing.T) {
	input := `C4Context
Person(user, "User", "End user")
System(s1, "System A", "First system")
System(s2, "System B", "Second system")
Rel(user, s1, "Uses", "HTTPS")
Rel(s1, s2, "Calls", "gRPC")
Rel(user, s2, "Also uses")
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(d.Relationships) != 3 {
		t.Fatalf("expected 3 relationships, got %d", len(d.Relationships))
	}
	// Check technology field
	if d.Relationships[0].Technology != "HTTPS" {
		t.Errorf("expected technology 'HTTPS', got %q", d.Relationships[0].Technology)
	}
	if d.Relationships[1].Technology != "gRPC" {
		t.Errorf("expected technology 'gRPC', got %q", d.Relationships[1].Technology)
	}
	// Third relationship has no technology
	if d.Relationships[2].Technology != "" {
		t.Errorf("expected empty technology, got %q", d.Relationships[2].Technology)
	}

	config := diagram.NewTestConfig(false, "cli")
	result, err := Render(d, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	if !strings.Contains(result, "[HTTPS]") {
		t.Error("expected output to contain '[HTTPS]'")
	}
	if !strings.Contains(result, "[gRPC]") {
		t.Error("expected output to contain '[gRPC]'")
	}
	if !strings.Contains(result, "Also uses") {
		t.Error("expected output to contain 'Also uses'")
	}
}

func TestNestedBoundaries(t *testing.T) {
	input := `C4Context
Enterprise_Boundary(eb, "Enterprise") {
  System_Boundary(sb, "System Boundary") {
    Container(c1, "Service", "Go", "A service")
  }
}
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(d.Boundaries) != 1 {
		t.Fatalf("expected 1 top-level boundary, got %d", len(d.Boundaries))
	}
	if len(d.Boundaries[0].Boundaries) != 1 {
		t.Fatalf("expected 1 nested boundary, got %d", len(d.Boundaries[0].Boundaries))
	}
	if d.Boundaries[0].Boundaries[0].Label != "System Boundary" {
		t.Errorf("expected nested boundary label 'System Boundary', got %q", d.Boundaries[0].Boundaries[0].Label)
	}

	config := diagram.NewTestConfig(false, "cli")
	result, err := Render(d, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	if !strings.Contains(result, "Enterprise") {
		t.Error("expected output to contain 'Enterprise'")
	}
	if !strings.Contains(result, "System Boundary") {
		t.Error("expected output to contain 'System Boundary'")
	}
	if !strings.Contains(result, "Service") {
		t.Error("expected output to contain 'Service'")
	}
}

func TestParseDynamicDiagram(t *testing.T) {
	input := `C4Dynamic
Person(user, "User", "A user")
System(sys, "System", "A system")
Rel(user, sys, "Step 1")
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if d.DiagramType != C4Dynamic {
		t.Errorf("expected C4Dynamic diagram type, got %d", d.DiagramType)
	}
}

func TestParseComponentDiagram(t *testing.T) {
	input := `C4Component
Component(comp1, "Component A", "Go", "First")
Component(comp2, "Component B", "Java", "Second")
Rel(comp1, comp2, "Depends on")
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if d.DiagramType != C4Component {
		t.Errorf("expected C4Component diagram type, got %d", d.DiagramType)
	}
	if len(d.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(d.Elements))
	}
	if len(d.Relationships) != 1 {
		t.Fatalf("expected 1 relationship, got %d", len(d.Relationships))
	}
}
