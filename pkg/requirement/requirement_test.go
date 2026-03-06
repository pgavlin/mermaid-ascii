package requirement

import (
	"strings"
	"testing"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

func TestIsRequirementDiagram(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid", "requirementDiagram", true},
		{"with comment", "%% comment\nrequirementDiagram", true},
		{"not requirement", "graph TD", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRequirementDiagram(tt.input); got != tt.want {
				t.Errorf("IsRequirementDiagram() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseRequirement(t *testing.T) {
	input := `requirementDiagram
requirement "Test Req" {
  id: 1
  text: "Must do something"
  risk: high
  verifymethod: test
}
element "Test Element" {
  type: "simulation"
  docref: "doc1"
}
"Test Req" - satisfies -> "Test Element"
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(d.Requirements) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(d.Requirements))
	}
	req := d.Requirements[0]
	if req.Name != "Test Req" {
		t.Errorf("expected name 'Test Req', got %q", req.Name)
	}
	if req.ID != "1" {
		t.Errorf("expected id '1', got %q", req.ID)
	}
	if req.Text != "Must do something" {
		t.Errorf("expected text 'Must do something', got %q", req.Text)
	}
	if req.Risk != "high" {
		t.Errorf("expected risk 'high', got %q", req.Risk)
	}
	if req.VerifyMethod != "test" {
		t.Errorf("expected verifymethod 'test', got %q", req.VerifyMethod)
	}

	if len(d.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(d.Elements))
	}
	elem := d.Elements[0]
	if elem.Name != "Test Element" {
		t.Errorf("expected element name 'Test Element', got %q", elem.Name)
	}
	if elem.Type != "simulation" {
		t.Errorf("expected type 'simulation', got %q", elem.Type)
	}

	if len(d.Relationships) != 1 {
		t.Fatalf("expected 1 relationship, got %d", len(d.Relationships))
	}
	rel := d.Relationships[0]
	if rel.Type != "satisfies" {
		t.Errorf("expected relationship type 'satisfies', got %q", rel.Type)
	}
}

func TestRenderRequirement(t *testing.T) {
	input := `requirementDiagram
requirement "Auth" {
  id: 1
  text: "Must authenticate"
  risk: high
  verifymethod: test
}
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
	if !strings.Contains(result, "Auth") {
		t.Error("expected output to contain 'Auth'")
	}
	if !strings.Contains(result, "<<requirement>>") {
		t.Error("expected output to contain '<<requirement>>'")
	}
}

func TestRenderASCII(t *testing.T) {
	input := `requirementDiagram
requirement "Test" {
  id: 1
  text: "desc"
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
	input := `requirementDiagram
requirement "Test" {
  id: 1
  text: "desc"
}
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
	if !strings.Contains(result, "┌") {
		t.Error("expected Unicode box-drawing characters when config is nil")
	}
}

func TestRenderReqElement(t *testing.T) {
	input := `requirementDiagram
element "Test Element" {
  type: "simulation"
  docref: "doc-ref-123"
}
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
	if !strings.Contains(result, "<<element>>") {
		t.Error("expected output to contain '<<element>>'")
	}
	if !strings.Contains(result, "Test Element") {
		t.Error("expected output to contain 'Test Element'")
	}
	if !strings.Contains(result, "Type: simulation") {
		t.Error("expected output to contain 'Type: simulation'")
	}
	if !strings.Contains(result, "DocRef: doc-ref-123") {
		t.Error("expected output to contain 'DocRef: doc-ref-123'")
	}
}

func TestRenderReqElementNoType(t *testing.T) {
	input := `requirementDiagram
element "Bare Element" {
}
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
	if !strings.Contains(result, "<<element>>") {
		t.Error("expected output to contain '<<element>>'")
	}
	if !strings.Contains(result, "Bare Element") {
		t.Error("expected output to contain 'Bare Element'")
	}
	// Should NOT contain Type: or DocRef: lines
	if strings.Contains(result, "Type:") {
		t.Error("expected output NOT to contain 'Type:' for element without type")
	}
	if strings.Contains(result, "DocRef:") {
		t.Error("expected output NOT to contain 'DocRef:' for element without docref")
	}
}

func TestRenderDesignConstraint(t *testing.T) {
	input := `requirementDiagram
designConstraint "Weight Limit" {
  id: DC001
  text: "Must weigh less than 5kg"
  risk: medium
  verifymethod: inspection
}
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(d.Requirements) != 1 {
		t.Fatalf("expected 1 requirement, got %d", len(d.Requirements))
	}
	if d.Requirements[0].Type != "designConstraint" {
		t.Errorf("expected type 'designConstraint', got %q", d.Requirements[0].Type)
	}

	config := diagram.NewTestConfig(false, "cli")
	result, err := Render(d, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	if !strings.Contains(result, "<<designConstraint>>") {
		t.Error("expected output to contain '<<designConstraint>>'")
	}
	if !strings.Contains(result, "Weight Limit") {
		t.Error("expected output to contain 'Weight Limit'")
	}
	if !strings.Contains(result, "Risk: medium") {
		t.Error("expected output to contain 'Risk: medium'")
	}
	if !strings.Contains(result, "Verify: inspection") {
		t.Error("expected output to contain 'Verify: inspection'")
	}
}

func TestRenderFunctionalRequirement(t *testing.T) {
	input := `requirementDiagram
functionalRequirement "Login" {
  id: FR001
  text: "System shall support login"
}
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
	if !strings.Contains(result, "<<functionalRequirement>>") {
		t.Error("expected output to contain '<<functionalRequirement>>'")
	}
	if !strings.Contains(result, "Login") {
		t.Error("expected output to contain 'Login'")
	}
}

func TestRenderWithRelationships(t *testing.T) {
	input := `requirementDiagram
requirement "Req A" {
  id: 1
  text: "Requirement A"
}
element "Elem B" {
  type: "component"
}
"Elem B" - satisfies -> "Req A"
"Req A" - traces -> "Elem B"
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(d.Relationships) != 2 {
		t.Fatalf("expected 2 relationships, got %d", len(d.Relationships))
	}

	config := diagram.NewTestConfig(false, "cli")
	result, err := Render(d, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	if !strings.Contains(result, "[satisfies]") {
		t.Error("expected output to contain '[satisfies]'")
	}
	if !strings.Contains(result, "[traces]") {
		t.Error("expected output to contain '[traces]'")
	}
	if !strings.Contains(result, "Elem B") {
		t.Error("expected output to contain 'Elem B'")
	}
	if !strings.Contains(result, "Req A") {
		t.Error("expected output to contain 'Req A'")
	}
}

func TestRenderRelationshipsASCII(t *testing.T) {
	input := `requirementDiagram
requirement "R1" {
  id: 1
  text: "desc"
}
element "E1" {
  type: "module"
}
"E1" - verifies -> "R1"
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
	if !strings.Contains(result, "-->") {
		t.Error("expected ASCII arrow '-->' in relationship output")
	}
	if !strings.Contains(result, "[verifies]") {
		t.Error("expected output to contain '[verifies]'")
	}
}

func TestParseRiskAndVerifyMethod(t *testing.T) {
	input := `requirementDiagram
requirement "Safety Req" {
  id: SR001
  text: "Must be safe"
  risk: high
  verifymethod: analysis
}
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	req := d.Requirements[0]
	if req.Risk != "high" {
		t.Errorf("expected risk 'high', got %q", req.Risk)
	}
	if req.VerifyMethod != "analysis" {
		t.Errorf("expected verifymethod 'analysis', got %q", req.VerifyMethod)
	}
}

func TestParseMultipleRequirementTypes(t *testing.T) {
	input := `requirementDiagram
requirement "Generic" {
  id: R1
  text: "Generic req"
}
interfaceRequirement "Interface" {
  id: IR1
  text: "Interface req"
}
performanceRequirement "Performance" {
  id: PR1
  text: "Perf req"
}
physicalRequirement "Physical" {
  id: PHR1
  text: "Physical req"
}
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(d.Requirements) != 4 {
		t.Fatalf("expected 4 requirements, got %d", len(d.Requirements))
	}
	expectedTypes := []string{"requirement", "interfaceRequirement", "performanceRequirement", "physicalRequirement"}
	for i, et := range expectedTypes {
		if d.Requirements[i].Type != et {
			t.Errorf("requirement %d: expected type %q, got %q", i, et, d.Requirements[i].Type)
		}
	}
}

func TestRenderMultipleElements(t *testing.T) {
	input := `requirementDiagram
element "Frontend" {
  type: "application"
  docref: "frontend-spec"
}
element "Backend" {
  type: "service"
  docref: "backend-spec"
}
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
	if !strings.Contains(result, "Frontend") {
		t.Error("expected output to contain 'Frontend'")
	}
	if !strings.Contains(result, "Backend") {
		t.Error("expected output to contain 'Backend'")
	}
	// Count occurrences of <<element>>
	count := strings.Count(result, "<<element>>")
	if count != 2 {
		t.Errorf("expected 2 occurrences of '<<element>>', got %d", count)
	}
}

func TestParseNotRequirementDiagram(t *testing.T) {
	_, err := Parse("graph TD\nA --> B")
	if err == nil {
		t.Error("expected error for non-requirement diagram input")
	}
}

func TestParseAllRelationshipTypes(t *testing.T) {
	input := `requirementDiagram
requirement "A" {
  id: 1
}
requirement "B" {
  id: 2
}
"A" - traces -> "B"
"A" - copies -> "B"
"A" - derives -> "B"
"A" - satisfies -> "B"
"A" - verifies -> "B"
"A" - refines -> "B"
"A" - contains -> "B"
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(d.Relationships) != 7 {
		t.Fatalf("expected 7 relationships, got %d", len(d.Relationships))
	}
	expectedTypes := []string{"traces", "copies", "derives", "satisfies", "verifies", "refines", "contains"}
	for i, et := range expectedTypes {
		if d.Relationships[i].Type != et {
			t.Errorf("relationship %d: expected type %q, got %q", i, et, d.Relationships[i].Type)
		}
	}
}
