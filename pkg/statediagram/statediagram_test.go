package statediagram

import (
	"strings"
	"testing"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

func TestIsStateDiagram(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid", "stateDiagram-v2\n  [*] --> State1", true},
		{"with comment", "%% comment\nstateDiagram-v2\n  [*] --> State1", true},
		{"invalid", "classDiagram\n  class Animal", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsStateDiagram(tt.input); got != tt.want {
				t.Errorf("IsStateDiagram() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseSimpleTransitions(t *testing.T) {
	input := `stateDiagram-v2
[*] --> State1
State1 --> State2 : trigger
State2 --> [*]`
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(sd.States) != 3 {
		t.Fatalf("expected 3 states, got %d", len(sd.States))
	}

	if len(sd.Transitions) != 3 {
		t.Fatalf("expected 3 transitions, got %d", len(sd.Transitions))
	}

	// Check [*] state
	starState := sd.stateMap["[*]"]
	if starState == nil {
		t.Fatal("expected [*] state")
	}
	if !starState.IsStart {
		t.Error("expected [*] to be a start state")
	}

	// Check trigger
	tr := sd.Transitions[1]
	if tr.Trigger != "trigger" {
		t.Errorf("expected trigger 'trigger', got %q", tr.Trigger)
	}
}

func TestParseStateAlias(t *testing.T) {
	input := `stateDiagram-v2
state "Moving forward" as s1
[*] --> s1`
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	s1 := sd.stateMap["s1"]
	if s1 == nil {
		t.Fatal("expected state s1")
	}
	if s1.Label != "Moving forward" {
		t.Errorf("expected label 'Moving forward', got %q", s1.Label)
	}
}

func TestParseCompositeState(t *testing.T) {
	input := `stateDiagram-v2
state "Main" as main {
  [*] --> Inner1
  Inner1 --> Inner2
}`
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	mainState := sd.stateMap["main"]
	if mainState == nil {
		t.Fatal("expected state 'main'")
	}
	if mainState.Label != "Main" {
		t.Errorf("expected label 'Main', got %q", mainState.Label)
	}

	// Inner states should be created
	if sd.stateMap["Inner1"] == nil {
		t.Error("expected inner state Inner1")
	}
	if sd.stateMap["Inner2"] == nil {
		t.Error("expected inner state Inner2")
	}
}

func TestParseNote(t *testing.T) {
	input := `stateDiagram-v2
[*] --> State1
note right of State1 : Important state`
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(sd.Notes) != 1 {
		t.Fatalf("expected 1 note, got %d", len(sd.Notes))
	}

	note := sd.Notes[0]
	if note.Position != NoteRight {
		t.Errorf("expected NoteRight, got %d", note.Position)
	}
	if note.Text != "Important state" {
		t.Errorf("expected text 'Important state', got %q", note.Text)
	}
	if note.State.ID != "State1" {
		t.Errorf("expected state 'State1', got %q", note.State.ID)
	}
}

func TestParseEmptyInput(t *testing.T) {
	_, err := Parse("")
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestParseNoStates(t *testing.T) {
	_, err := Parse("stateDiagram-v2\n")
	if err == nil {
		t.Fatal("expected error for no states")
	}
}

func TestRenderASCII(t *testing.T) {
	input := `stateDiagram-v2
[*] --> State1
State1 --> State2 : trigger
State2 --> [*]`
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	config := diagram.NewTestConfig(true, "cli")
	result, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	if !strings.Contains(result, "(*)") {
		t.Error("expected start state (*) in output")
	}
	if !strings.Contains(result, "State1") {
		t.Error("expected 'State1' in output")
	}
	if !strings.Contains(result, "State2") {
		t.Error("expected 'State2' in output")
	}
	if !strings.Contains(result, "trigger") {
		t.Error("expected trigger label in output")
	}
	// Check ASCII chars
	if !strings.Contains(result, "+") {
		t.Error("expected ASCII box corners (+)")
	}
}

func TestRenderUnicode(t *testing.T) {
	input := `stateDiagram-v2
[*] --> State1
State1 --> [*]`
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	config := diagram.NewTestConfig(false, "cli")
	result, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	if !strings.Contains(result, "┌") {
		t.Error("expected Unicode box corner in output")
	}
	if !strings.Contains(result, "State1") {
		t.Error("expected 'State1' in output")
	}
}

func TestRenderWithNote(t *testing.T) {
	input := `stateDiagram-v2
[*] --> State1
note right of State1 : Important`
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	config := diagram.NewTestConfig(true, "cli")
	result, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	if !strings.Contains(result, "Important") {
		t.Error("expected note text 'Important' in output")
	}
}

func TestRenderWithAlias(t *testing.T) {
	input := `stateDiagram-v2
state "Moving forward" as s1
[*] --> s1`
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	config := diagram.NewTestConfig(true, "cli")
	result, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	if !strings.Contains(result, "Moving forward") {
		t.Error("expected alias label 'Moving forward' in output")
	}
}

func TestRenderEmptyDiagram(t *testing.T) {
	_, err := Render(nil, nil)
	if err == nil {
		t.Fatal("expected error for nil diagram")
	}
}

func TestParseCompositeStateWithNestedTransitions(t *testing.T) {
	input := `stateDiagram-v2
    state "Process" as proc {
        [*] --> Running
        Running --> Done
    }
    [*] --> proc
    proc --> [*]`

	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	procState := sd.stateMap["proc"]
	if procState == nil {
		t.Fatal("expected state 'proc'")
	}
	if procState.Label != "Process" {
		t.Errorf("expected label 'Process', got %q", procState.Label)
	}

	// Inner states should exist
	if sd.stateMap["Running"] == nil {
		t.Error("expected inner state 'Running'")
	}
	if sd.stateMap["Done"] == nil {
		t.Error("expected inner state 'Done'")
	}

	// Should have transitions from inner and outer
	if len(sd.Transitions) < 4 {
		t.Errorf("expected at least 4 transitions, got %d", len(sd.Transitions))
	}
}

func TestParseNoteLeftOf(t *testing.T) {
	input := `stateDiagram-v2
    [*] --> State1
    note left of State1 : Left note`

	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if len(sd.Notes) != 1 {
		t.Fatalf("expected 1 note, got %d", len(sd.Notes))
	}

	note := sd.Notes[0]
	if note.Position != NoteLeft {
		t.Errorf("expected NoteLeft, got %d", note.Position)
	}
	if note.Text != "Left note" {
		t.Errorf("expected text 'Left note', got %q", note.Text)
	}
}

func TestRenderNilConfig(t *testing.T) {
	input := `stateDiagram-v2
    [*] --> State1
    State1 --> [*]`

	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	output, err := Render(sd, nil)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	if !strings.Contains(output, "State1") {
		t.Error("expected 'State1' in output")
	}
	// With nil config, defaults to unicode
	if !strings.Contains(output, "┌") {
		t.Error("expected Unicode box corner in output (nil config should use defaults)")
	}
}

func TestParseStateWithDescription(t *testing.T) {
	input := `stateDiagram-v2
    state "Waiting for input" as waiting
    state "Processing data" as processing
    [*] --> waiting
    waiting --> processing : submit
    processing --> [*]`

	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	waiting := sd.stateMap["waiting"]
	if waiting == nil {
		t.Fatal("expected state 'waiting'")
	}
	if waiting.Label != "Waiting for input" {
		t.Errorf("expected label 'Waiting for input', got %q", waiting.Label)
	}

	processing := sd.stateMap["processing"]
	if processing == nil {
		t.Fatal("expected state 'processing'")
	}
	if processing.Label != "Processing data" {
		t.Errorf("expected label 'Processing data', got %q", processing.Label)
	}

	config := diagram.DefaultConfig()
	output, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	if !strings.Contains(output, "Waiting for input") {
		t.Error("expected 'Waiting for input' in output")
	}
	if !strings.Contains(output, "Processing data") {
		t.Error("expected 'Processing data' in output")
	}
}

func TestParseInlineState(t *testing.T) {
	input := `stateDiagram-v2
    state MyState
    [*] --> MyState`

	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if sd.stateMap["MyState"] == nil {
		t.Error("expected state 'MyState'")
	}
}

func TestParseInvalidKeyword(t *testing.T) {
	_, err := Parse("flowchart\n  [*] --> State1")
	if err == nil {
		t.Error("expected error for invalid keyword")
	}
}

func TestRenderWithNoteAnnotation(t *testing.T) {
	input := `stateDiagram-v2
    [*] --> Active
    note right of Active : This is active
    Active --> [*]`

	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	config := diagram.DefaultConfig()
	output, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	if !strings.Contains(output, "This is active") {
		t.Error("expected note text in output")
	}
	if !strings.Contains(output, "Active") {
		t.Error("expected state name in output")
	}
}

func TestOrderStatesNoTransitions(t *testing.T) {
	input := `stateDiagram-v2
    state A
    state B`

	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	config := diagram.DefaultConfig()
	output, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	if !strings.Contains(output, "A") {
		t.Error("expected state 'A' in output")
	}
	if !strings.Contains(output, "B") {
		t.Error("expected state 'B' in output")
	}
}
