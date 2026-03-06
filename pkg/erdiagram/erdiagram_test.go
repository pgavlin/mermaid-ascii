package erdiagram

import (
	"strings"
	"testing"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

func TestIsERDiagram(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid", "erDiagram\n  CUSTOMER ||--o{ ORDER : places", true},
		{"with comment", "%% comment\nerDiagram\n  CUSTOMER ||--o{ ORDER : places", true},
		{"invalid", "classDiagram\n  class Animal", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsERDiagram(tt.input); got != tt.want {
				t.Errorf("IsERDiagram() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseRelationship(t *testing.T) {
	input := `erDiagram
CUSTOMER ||--o{ ORDER : places`
	erd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(erd.Entities) != 2 {
		t.Fatalf("expected 2 entities, got %d", len(erd.Entities))
	}

	if len(erd.Relationships) != 1 {
		t.Fatalf("expected 1 relationship, got %d", len(erd.Relationships))
	}

	rel := erd.Relationships[0]
	if rel.From != "CUSTOMER" {
		t.Errorf("expected from 'CUSTOMER', got %q", rel.From)
	}
	if rel.To != "ORDER" {
		t.Errorf("expected to 'ORDER', got %q", rel.To)
	}
	if rel.Label != "places" {
		t.Errorf("expected label 'places', got %q", rel.Label)
	}
	if rel.FromCardinality != ExactlyOne {
		t.Errorf("expected FromCardinality ExactlyOne, got %d", rel.FromCardinality)
	}
	if rel.ToCardinality != ZeroOrMany {
		t.Errorf("expected ToCardinality ZeroOrMany, got %d", rel.ToCardinality)
	}
}

func TestParseEntityAttributes(t *testing.T) {
	input := `erDiagram
CUSTOMER {
    string name
    int age
    string email PK
}`
	erd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(erd.Entities) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(erd.Entities))
	}

	entity := erd.Entities[0]
	if entity.Name != "CUSTOMER" {
		t.Errorf("expected entity name 'CUSTOMER', got %q", entity.Name)
	}
	if len(entity.Attributes) != 3 {
		t.Fatalf("expected 3 attributes, got %d", len(entity.Attributes))
	}

	// Check first attribute
	attr := entity.Attributes[0]
	if attr.Type != "string" || attr.Name != "name" {
		t.Errorf("attr 0: got type=%q name=%q", attr.Type, attr.Name)
	}

	// Check PK attribute
	attr = entity.Attributes[2]
	if attr.Constraint != PrimaryKey {
		t.Errorf("expected PrimaryKey constraint, got %d", attr.Constraint)
	}
}

func TestParseMultipleRelationships(t *testing.T) {
	input := `erDiagram
CUSTOMER ||--o{ ORDER : places
ORDER ||--|{ LINE_ITEM : contains
CUSTOMER }|--|{ DELIVERY_ADDRESS : uses`
	erd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(erd.Entities) != 4 {
		t.Fatalf("expected 4 entities, got %d", len(erd.Entities))
	}

	if len(erd.Relationships) != 3 {
		t.Fatalf("expected 3 relationships, got %d", len(erd.Relationships))
	}
}

func TestParseEmptyInput(t *testing.T) {
	_, err := Parse("")
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestParseNoEntities(t *testing.T) {
	_, err := Parse("erDiagram\n")
	if err == nil {
		t.Fatal("expected error for no entities")
	}
}

func TestRenderASCII(t *testing.T) {
	input := `erDiagram
CUSTOMER {
    string name
    int age
}
ORDER {
    int id
    string status
}
CUSTOMER ||--o{ ORDER : places`
	erd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	config := diagram.NewTestConfig(true, "cli")
	result, err := Render(erd, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	// Check ASCII box characters
	if !strings.Contains(result, "+") {
		t.Error("expected ASCII box corners (+)")
	}
	if !strings.Contains(result, "CUSTOMER") {
		t.Error("expected entity 'CUSTOMER' in output")
	}
	if !strings.Contains(result, "ORDER") {
		t.Error("expected entity 'ORDER' in output")
	}
	if !strings.Contains(result, "string name") {
		t.Error("expected attribute 'string name' in output")
	}
	if !strings.Contains(result, "places") {
		t.Error("expected relationship label 'places' in output")
	}
	// Verify no Unicode chars in ASCII mode
	if strings.Contains(result, "\u2502") || strings.Contains(result, "\u2500") {
		t.Error("ASCII output should not contain Unicode box-drawing characters")
	}
}

func TestRenderUnicode(t *testing.T) {
	input := `erDiagram
CUSTOMER {
    string name
}`
	erd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	config := diagram.NewTestConfig(false, "cli")
	result, err := Render(erd, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	if !strings.Contains(result, "\u250c") {
		t.Error("expected Unicode box corner in output")
	}
	if !strings.Contains(result, "CUSTOMER") {
		t.Error("expected entity name in output")
	}
}

func TestRenderCardinalities(t *testing.T) {
	input := `erDiagram
A ||--o{ B : rel1
C }|--|| D : rel2`
	erd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	config := diagram.NewTestConfig(true, "cli")
	result, err := Render(erd, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	// Check that cardinality markers appear
	if !strings.Contains(result, "||") {
		t.Error("expected exactly-one marker (||) in output")
	}
	if !strings.Contains(result, ">o") {
		t.Error("expected zero-or-many marker (>o) in output")
	}
}

func TestRenderEmptyDiagram(t *testing.T) {
	_, err := Render(nil, nil)
	if err == nil {
		t.Fatal("expected error for nil diagram")
	}
}

func TestRenderEntityWithConstraints(t *testing.T) {
	input := `erDiagram
PRODUCT {
    int id PK
    string name UK
    int category_id FK
}`
	erd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	config := diagram.NewTestConfig(true, "cli")
	result, err := Render(erd, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	if !strings.Contains(result, "PK") {
		t.Error("expected PK constraint in output")
	}
	if !strings.Contains(result, "FK") {
		t.Error("expected FK constraint in output")
	}
	if !strings.Contains(result, "UK") {
		t.Error("expected UK constraint in output")
	}
}

func TestImplicitEntityCreation(t *testing.T) {
	input := `erDiagram
CUSTOMER ||--o{ ORDER : places`
	erd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(erd.Entities) != 2 {
		t.Fatalf("expected 2 entities (implicit), got %d", len(erd.Entities))
	}
}

func TestParseCardinalityAllTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantCard Cardinality
	}{
		{"exactly one ||", "||", ExactlyOne},
		{"zero or one o|", "o|", ZeroOrOne},
		{"zero or one |o", "|o", ZeroOrOne},
		{"one or many }|", "}|", OneOrMany},
		{"one or many |{", "|{", OneOrMany},
		{"zero or many }o", "}o", ZeroOrMany},
		{"zero or many o{", "o{", ZeroOrMany},
		{"empty string default", "", ExactlyOne},
		{"whitespace only", "  ", ExactlyOne},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCardinality(tt.input)
			if got != tt.wantCard {
				t.Errorf("parseCardinality(%q) = %d, want %d", tt.input, got, tt.wantCard)
			}
		})
	}
}

func TestParseCardinalityPartialMatches(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantCard Cardinality
	}{
		{"partial }o inferred", "}o~", ZeroOrMany},
		{"partial } inferred", "}x", OneOrMany},
		{"partial o inferred", "ox", ZeroOrOne},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCardinality(tt.input)
			if got != tt.wantCard {
				t.Errorf("parseCardinality(%q) = %d, want %d", tt.input, got, tt.wantCard)
			}
		})
	}
}

func TestCardinalityStringAllConstants(t *testing.T) {
	tests := []struct {
		name     string
		card     Cardinality
		useAscii bool
		want     string
	}{
		{"ExactlyOne unicode", ExactlyOne, false, "||"},
		{"ExactlyOne ascii", ExactlyOne, true, "||"},
		{"ZeroOrOne unicode", ZeroOrOne, false, "o|"},
		{"ZeroOrOne ascii", ZeroOrOne, true, "o|"},
		{"OneOrMany unicode", OneOrMany, false, ">|"},
		{"OneOrMany ascii", OneOrMany, true, ">|"},
		{"ZeroOrMany unicode", ZeroOrMany, false, ">o"},
		{"ZeroOrMany ascii", ZeroOrMany, true, ">o"},
		{"unknown defaults to ||", Cardinality(99), false, "||"},
		{"unknown ascii defaults to ||", Cardinality(99), true, "||"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cardinalityString(tt.card, tt.useAscii)
			if got != tt.want {
				t.Errorf("cardinalityString(%d, %v) = %q, want %q", tt.card, tt.useAscii, got, tt.want)
			}
		})
	}
}

func TestParseAllCardinalityRelationships(t *testing.T) {
	input := `erDiagram
A ||--|| B : exactlyOneToOne
C o|--|| D : zeroOrOneToOne
E }|--|| F : oneOrManyToOne
G o{--|| H : zeroOrManyToOne`
	erd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(erd.Relationships) != 4 {
		t.Fatalf("expected 4 relationships, got %d", len(erd.Relationships))
	}

	expected := []struct {
		fromCard Cardinality
		toCard   Cardinality
		label    string
	}{
		{ExactlyOne, ExactlyOne, "exactlyOneToOne"},
		{ZeroOrOne, ExactlyOne, "zeroOrOneToOne"},
		{OneOrMany, ExactlyOne, "oneOrManyToOne"},
		{ZeroOrMany, ExactlyOne, "zeroOrManyToOne"},
	}
	for i, exp := range expected {
		rel := erd.Relationships[i]
		if rel.FromCardinality != exp.fromCard {
			t.Errorf("rel[%d] (%s): FromCardinality = %d, want %d", i, exp.label, rel.FromCardinality, exp.fromCard)
		}
		if rel.ToCardinality != exp.toCard {
			t.Errorf("rel[%d] (%s): ToCardinality = %d, want %d", i, exp.label, rel.ToCardinality, exp.toCard)
		}
		if rel.Label != exp.label {
			t.Errorf("rel[%d]: Label = %q, want %q", i, rel.Label, exp.label)
		}
	}
}

func TestRenderAllCardinalityMarkers(t *testing.T) {
	input := `erDiagram
A ||--o| B : rel1
C }|--o{ D : rel2`
	erd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	config := diagram.NewTestConfig(false, "cli")
	result, err := Render(erd, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	// Check all marker types appear
	if !strings.Contains(result, "||") {
		t.Error("expected ExactlyOne marker (||)")
	}
	if !strings.Contains(result, "o|") {
		t.Error("expected ZeroOrOne marker (o|)")
	}
	if !strings.Contains(result, ">|") {
		t.Error("expected OneOrMany marker (>|)")
	}
	if !strings.Contains(result, ">o") {
		t.Error("expected ZeroOrMany marker (>o)")
	}
}

func TestRenderNilConfig(t *testing.T) {
	input := `erDiagram
A ||--o{ B : test`
	erd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	result, err := Render(erd, nil)
	if err != nil {
		t.Fatalf("Render() with nil config error: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty output with nil config")
	}
	// Nil config should default to unicode
	if !strings.Contains(result, "\u250c") {
		t.Error("expected Unicode box characters with nil config (defaulting to unicode)")
	}
}
