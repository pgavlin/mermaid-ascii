package classdiagram

import (
	"strings"
	"testing"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

func TestIsClassDiagram(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid", "classDiagram\n  class Animal", true},
		{"with comment", "%% comment\nclassDiagram\n  class Animal", true},
		{"invalid", "sequenceDiagram\n  A->>B: hi", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsClassDiagram(tt.input); got != tt.want {
				t.Errorf("IsClassDiagram() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseClassBlock(t *testing.T) {
	input := `classDiagram
class Animal {
  +String name
  +int age
  +isMammal() bool
  -privateMethod()
}`
	cd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(cd.Classes) != 1 {
		t.Fatalf("expected 1 class, got %d", len(cd.Classes))
	}

	cls := cd.Classes[0]
	if cls.Name != "Animal" {
		t.Errorf("expected class name 'Animal', got %q", cls.Name)
	}
	if len(cls.Members) != 4 {
		t.Fatalf("expected 4 members, got %d", len(cls.Members))
	}

	// Check first member: +String name
	m := cls.Members[0]
	if m.Visibility != Public || m.Type != "String" || m.Name != "name" || m.IsMethod {
		t.Errorf("member 0: got vis=%d type=%q name=%q method=%v", m.Visibility, m.Type, m.Name, m.IsMethod)
	}

	// Check third member: +isMammal() bool
	m = cls.Members[2]
	if m.Visibility != Public || m.Name != "isMammal" || !m.IsMethod || m.Type != "bool" {
		t.Errorf("member 2: got vis=%d name=%q method=%v type=%q", m.Visibility, m.Name, m.IsMethod, m.Type)
	}

	// Check fourth member: -privateMethod()
	m = cls.Members[3]
	if m.Visibility != Private || m.Name != "privateMethod" || !m.IsMethod {
		t.Errorf("member 3: got vis=%d name=%q method=%v", m.Visibility, m.Name, m.IsMethod)
	}
}

func TestParseRelationships(t *testing.T) {
	input := `classDiagram
class Animal
class Dog
class Leg
class Lake
class Water
class Food
Animal <|-- Dog
Animal *-- Leg
Animal o-- Lake
Animal ..> Water
Animal --> Food`
	cd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(cd.Relationships) != 5 {
		t.Fatalf("expected 5 relationships, got %d", len(cd.Relationships))
	}

	expected := []struct {
		from    string
		to      string
		relType RelationType
	}{
		{"Animal", "Dog", Inheritance},
		{"Animal", "Leg", Composition},
		{"Animal", "Lake", Aggregation},
		{"Animal", "Water", Dependency},
		{"Animal", "Food", Association},
	}

	for i, exp := range expected {
		rel := cd.Relationships[i]
		if rel.From != exp.from || rel.To != exp.to || rel.Type != exp.relType {
			t.Errorf("relationship %d: got from=%q to=%q type=%d, want from=%q to=%q type=%d",
				i, rel.From, rel.To, rel.Type, exp.from, exp.to, exp.relType)
		}
	}
}

func TestParseRelationshipWithLabel(t *testing.T) {
	input := `classDiagram
class Animal
class Food
Animal "1" --> "*" Food : eats`
	cd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(cd.Relationships) != 1 {
		t.Fatalf("expected 1 relationship, got %d", len(cd.Relationships))
	}

	rel := cd.Relationships[0]
	if rel.Label != "eats" {
		t.Errorf("expected label 'eats', got %q", rel.Label)
	}
	if rel.FromLabel != "1" {
		t.Errorf("expected fromLabel '1', got %q", rel.FromLabel)
	}
	if rel.ToLabel != "*" {
		t.Errorf("expected toLabel '*', got %q", rel.ToLabel)
	}
}

func TestParseEmptyInput(t *testing.T) {
	_, err := Parse("")
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestParseNoClasses(t *testing.T) {
	_, err := Parse("classDiagram\n")
	if err == nil {
		t.Fatal("expected error for no classes")
	}
}

func TestRenderASCII(t *testing.T) {
	input := `classDiagram
class Animal {
  +String name
  +isMammal() bool
}`
	cd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	config := diagram.NewTestConfig(true, "cli")
	result, err := Render(cd, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	// Check that the output contains ASCII box characters
	if !strings.Contains(result, "+") {
		t.Error("expected ASCII box corners (+)")
	}
	if !strings.Contains(result, "Animal") {
		t.Error("expected class name 'Animal' in output")
	}
	if !strings.Contains(result, "+String name") {
		t.Error("expected member '+String name' in output")
	}
	if !strings.Contains(result, "+isMammal() bool") {
		t.Error("expected member '+isMammal() bool' in output")
	}
}

func TestRenderUnicode(t *testing.T) {
	input := `classDiagram
class Animal {
  +String name
}`
	cd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	config := diagram.NewTestConfig(false, "cli")
	result, err := Render(cd, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	if !strings.Contains(result, "┌") {
		t.Error("expected Unicode box corner in output")
	}
	if !strings.Contains(result, "Animal") {
		t.Error("expected class name 'Animal' in output")
	}
}

func TestRenderWithRelationships(t *testing.T) {
	input := `classDiagram
class Animal {
  +String name
}
class Dog {
  +String breed
}
Animal <|-- Dog`
	cd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	config := diagram.NewTestConfig(true, "cli")
	result, err := Render(cd, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	if !strings.Contains(result, "Animal") {
		t.Error("expected 'Animal' in output")
	}
	if !strings.Contains(result, "Dog") {
		t.Error("expected 'Dog' in output")
	}
	if !strings.Contains(result, "<|--") {
		t.Error("expected inheritance arrow in output")
	}
}

func TestRenderMultipleClasses(t *testing.T) {
	input := `classDiagram
class A
class B
class C
A --> B
B --> C`
	cd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	config := diagram.NewTestConfig(true, "cli")
	result, err := Render(cd, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	// All classes should appear
	for _, name := range []string{"A", "B", "C"} {
		if !strings.Contains(result, name) {
			t.Errorf("expected class %q in output", name)
		}
	}
}

func TestRenderEmptyDiagram(t *testing.T) {
	_, err := Render(nil, nil)
	if err == nil {
		t.Fatal("expected error for nil diagram")
	}
}

func TestImplicitClassCreation(t *testing.T) {
	input := `classDiagram
Animal <|-- Dog`
	cd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(cd.Classes) != 2 {
		t.Fatalf("expected 2 classes (implicit), got %d", len(cd.Classes))
	}
}

func TestParseRelationTypeAll(t *testing.T) {
	tests := []struct {
		arrow   string
		relType RelationType
	}{
		{"<|--", Inheritance},
		{"<|..", Realization},
		{"*--", Composition},
		{"--*", Composition},
		{"o--", Aggregation},
		{"--o", Aggregation},
		{"..>", Dependency},
		{"-->", Association},
		{"<--", Association},
		{"--", Link},
		{"..", DottedLink},
		{"*..", DottedLink},
		{"unknown", Association}, // default case
	}
	for _, tt := range tests {
		t.Run(tt.arrow, func(t *testing.T) {
			got := parseRelationType(tt.arrow)
			if got != tt.relType {
				t.Errorf("parseRelationType(%q) = %d, want %d", tt.arrow, got, tt.relType)
			}
		})
	}
}

func TestRelationshipArrowASCII(t *testing.T) {
	tests := []struct {
		relType  RelationType
		expected string
	}{
		{Inheritance, "<|--"},
		{Composition, "*--"},
		{Aggregation, "o--"},
		{Dependency, "..>"},
		{Association, "-->"},
		{Realization, "<|.."},
		{Link, "--"},
		{DottedLink, ".."},
		{RelationType(99), "-->"}, // default case
	}
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := relationshipArrow(tt.relType, true)
			if got != tt.expected {
				t.Errorf("relationshipArrow(%d, true) = %q, want %q", tt.relType, got, tt.expected)
			}
		})
	}
}

func TestRelationshipArrowUnicode(t *testing.T) {
	tests := []struct {
		relType  RelationType
		expected string
	}{
		{Inheritance, "<|──"},
		{Composition, "*──"},
		{Aggregation, "o──"},
		{Dependency, "..>"},
		{Association, "──>"},
		{Realization, "<|.."},
		{Link, "──"},
		{DottedLink, ".."},
		{RelationType(99), "──>"}, // default case
	}
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := relationshipArrow(tt.relType, false)
			if got != tt.expected {
				t.Errorf("relationshipArrow(%d, false) = %q, want %q", tt.relType, got, tt.expected)
			}
		})
	}
}

func TestParseMemberVariants(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantVis    Visibility
		wantName   string
		wantType   string
		wantMethod bool
		wantParams string
	}{
		{
			name: "public field with type",
			input: "+String name", wantVis: Public, wantName: "name", wantType: "String",
		},
		{
			name: "private field",
			input: "-int count", wantVis: Private, wantName: "count", wantType: "int",
		},
		{
			name: "protected field",
			input: "#float value", wantVis: Protected, wantName: "value", wantType: "float",
		},
		{
			name: "package field",
			input: "~bool flag", wantVis: Package, wantName: "flag", wantType: "bool",
		},
		{
			name: "no visibility field",
			input: "String data", wantVis: None, wantName: "data", wantType: "String",
		},
		{
			name: "single name field no type",
			input: "+fieldOnly", wantVis: Public, wantName: "fieldOnly", wantType: "",
		},
		{
			name: "method with return type",
			input: "+getAge() int", wantVis: Public, wantName: "getAge", wantType: "int", wantMethod: true,
		},
		{
			name: "method with parameters",
			input: "+setName(name, alias)", wantVis: Public, wantName: "setName", wantType: "", wantMethod: true, wantParams: "name, alias",
		},
		{
			name: "method with params and return type",
			input: "+calculate(x, y) Result", wantVis: Public, wantName: "calculate", wantType: "Result", wantMethod: true, wantParams: "x, y",
		},
		{
			name: "private method no params",
			input: "-doStuff()", wantVis: Private, wantName: "doStuff", wantType: "", wantMethod: true,
		},
		{
			name: "protected method",
			input: "#init()", wantVis: Protected, wantName: "init", wantType: "", wantMethod: true,
		},
		{
			name: "package method",
			input: "~helper()", wantVis: Package, wantName: "helper", wantType: "", wantMethod: true,
		},
		{
			name:  "empty line",
			input: "", wantVis: None, wantName: "", wantType: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := parseMember(tt.input)
			if tt.input == "" {
				if m != nil {
					t.Error("expected nil for empty input")
				}
				return
			}
			if m == nil {
				t.Fatal("parseMember returned nil")
			}
			if m.Visibility != tt.wantVis {
				t.Errorf("Visibility = %d, want %d", m.Visibility, tt.wantVis)
			}
			if m.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", m.Name, tt.wantName)
			}
			if m.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", m.Type, tt.wantType)
			}
			if m.IsMethod != tt.wantMethod {
				t.Errorf("IsMethod = %v, want %v", m.IsMethod, tt.wantMethod)
			}
			if tt.wantParams != "" && m.Parameters != tt.wantParams {
				t.Errorf("Parameters = %q, want %q", m.Parameters, tt.wantParams)
			}
		})
	}
}

func TestFormatMemberVariants(t *testing.T) {
	tests := []struct {
		name     string
		member   Member
		expected string
	}{
		{
			name:     "public field with type",
			member:   Member{Visibility: Public, Name: "name", Type: "String"},
			expected: "+String name",
		},
		{
			name:     "private field with type",
			member:   Member{Visibility: Private, Name: "count", Type: "int"},
			expected: "-int count",
		},
		{
			name:     "protected field",
			member:   Member{Visibility: Protected, Name: "value", Type: "float"},
			expected: "#float value",
		},
		{
			name:     "package field",
			member:   Member{Visibility: Package, Name: "flag", Type: "bool"},
			expected: "~bool flag",
		},
		{
			name:     "no visibility field",
			member:   Member{Visibility: None, Name: "data", Type: "String"},
			expected: "String data",
		},
		{
			name:     "field without type",
			member:   Member{Visibility: Public, Name: "solo"},
			expected: "+solo",
		},
		{
			name:     "public method no params",
			member:   Member{Visibility: Public, Name: "run", IsMethod: true},
			expected: "+run()",
		},
		{
			name:     "method with return type",
			member:   Member{Visibility: Public, Name: "getAge", Type: "int", IsMethod: true},
			expected: "+getAge() int",
		},
		{
			name:     "method with params",
			member:   Member{Visibility: Public, Name: "setName", IsMethod: true, Parameters: "name, alias"},
			expected: "+setName(name, alias)",
		},
		{
			name:     "method with params and return",
			member:   Member{Visibility: Private, Name: "calc", Type: "Result", IsMethod: true, Parameters: "x"},
			expected: "-calc(x) Result",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatMember(&tt.member)
			if got != tt.expected {
				t.Errorf("formatMember() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestParseAllRelationshipTypes(t *testing.T) {
	input := `classDiagram
class A
class B
class C
class D
class E
class F
class G
class H
A <|-- B
A *-- C
A o-- D
A ..> E
A <|.. F
A -- G
A --> H`
	cd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(cd.Relationships) != 7 {
		t.Fatalf("expected 7 relationships, got %d", len(cd.Relationships))
	}

	expected := []struct {
		to      string
		relType RelationType
	}{
		{"B", Inheritance},
		{"C", Composition},
		{"D", Aggregation},
		{"E", Dependency},
		{"F", Realization},
		{"G", Link},
		{"H", Association},
	}

	for i, exp := range expected {
		rel := cd.Relationships[i]
		if rel.To != exp.to {
			t.Errorf("rel %d: To = %q, want %q", i, rel.To, exp.to)
		}
		if rel.Type != exp.relType {
			t.Errorf("rel %d: Type = %d, want %d", i, rel.Type, exp.relType)
		}
	}
}

func TestRenderAllRelationshipTypesASCII(t *testing.T) {
	input := `classDiagram
class A
class B
class C
class D
class E
class F
class G
class H
A <|-- B
A *-- C
A o-- D
A ..> E
A <|.. F
A -- G
A --> H`
	cd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	config := diagram.NewTestConfig(true, "cli")
	result, err := Render(cd, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	expectedArrows := []string{"<|--", "*--", "o--", "..>", "<|..", "--", "-->"}
	for _, arrow := range expectedArrows {
		if !strings.Contains(result, arrow) {
			t.Errorf("expected arrow %q in ASCII output", arrow)
		}
	}
}

func TestRenderAllRelationshipTypesUnicode(t *testing.T) {
	input := `classDiagram
class A
class B
class C
class D
class E
class F
A <|-- B
A *-- C
A o-- D
A ..> E
A <|.. F`
	cd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	config := diagram.NewTestConfig(false, "cli")
	result, err := Render(cd, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	expectedArrows := []string{"<|──", "*──", "o──", "..>", "<|.."}
	for _, arrow := range expectedArrows {
		if !strings.Contains(result, arrow) {
			t.Errorf("expected unicode arrow %q in output", arrow)
		}
	}
}

func TestRenderRelationshipWithLabels(t *testing.T) {
	input := `classDiagram
class A
class B
A "1" --> "*" B : contains`
	cd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	config := diagram.NewTestConfig(true, "cli")
	result, err := Render(cd, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	if !strings.Contains(result, `"1"`) {
		t.Error("expected from label \"1\" in output")
	}
	if !strings.Contains(result, `"*"`) {
		t.Error("expected to label \"*\" in output")
	}
	if !strings.Contains(result, ": contains") {
		t.Error("expected label 'contains' in output")
	}
}

func TestRenderNilConfig(t *testing.T) {
	input := `classDiagram
class A`
	cd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	// nil config should use default (Unicode)
	result, err := Render(cd, nil)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	if !strings.Contains(result, "┌") {
		t.Error("expected Unicode box drawing with nil config (default)")
	}
}

func TestRenderClassWithNoMembers(t *testing.T) {
	input := `classDiagram
class EmptyClass`
	cd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	config := diagram.NewTestConfig(true, "cli")
	result, err := Render(cd, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	if !strings.Contains(result, "EmptyClass") {
		t.Error("expected class name in output")
	}
}

func TestParseRealizationRelationship(t *testing.T) {
	input := `classDiagram
class Animal
class Flyable
Animal <|.. Flyable`
	cd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(cd.Relationships) != 1 {
		t.Fatalf("expected 1 relationship, got %d", len(cd.Relationships))
	}

	rel := cd.Relationships[0]
	if rel.Type != Realization {
		t.Errorf("expected Realization, got %d", rel.Type)
	}
}

func TestParseLinkRelationship(t *testing.T) {
	input := `classDiagram
class A
class B
A -- B`
	cd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(cd.Relationships) != 1 {
		t.Fatalf("expected 1 relationship, got %d", len(cd.Relationships))
	}

	rel := cd.Relationships[0]
	if rel.Type != Link {
		t.Errorf("expected Link, got %d", rel.Type)
	}
}

func TestRenderRelationshipNoLabel(t *testing.T) {
	// Relationship without any labels or cardinalities
	rel := &Relationship{
		From: "ClassA",
		To:   "ClassB",
		Type: Association,
	}
	result := renderRelationshipLine(rel, asciiChars, true)
	if result != "ClassA --> ClassB" {
		t.Errorf("got %q, want %q", result, "ClassA --> ClassB")
	}
}

func TestRenderRelationshipWithFromLabelOnly(t *testing.T) {
	rel := &Relationship{
		From:      "ClassA",
		To:        "ClassB",
		Type:      Association,
		FromLabel: "1",
	}
	result := renderRelationshipLine(rel, asciiChars, true)
	if !strings.Contains(result, `"1"`) {
		t.Errorf("expected fromLabel in output, got %q", result)
	}
}

func TestRenderRelationshipWithToLabelOnly(t *testing.T) {
	rel := &Relationship{
		From:    "ClassA",
		To:      "ClassB",
		Type:    Association,
		ToLabel: "*",
	}
	result := renderRelationshipLine(rel, asciiChars, true)
	if !strings.Contains(result, `"*"`) {
		t.Errorf("expected toLabel in output, got %q", result)
	}
}

func TestRenderRelationshipLineAllTypes(t *testing.T) {
	types := []struct {
		relType  RelationType
		asciiArr string
	}{
		{Inheritance, "<|--"},
		{Composition, "*--"},
		{Aggregation, "o--"},
		{Dependency, "..>"},
		{Realization, "<|.."},
		{Link, "--"},
		{DottedLink, ".."},
		{Association, "-->"},
	}

	for _, tt := range types {
		t.Run(tt.asciiArr, func(t *testing.T) {
			rel := &Relationship{From: "X", To: "Y", Type: tt.relType}
			result := renderRelationshipLine(rel, asciiChars, true)
			expected := "X " + tt.asciiArr + " Y"
			if result != expected {
				t.Errorf("got %q, want %q", result, expected)
			}
		})
	}
}
