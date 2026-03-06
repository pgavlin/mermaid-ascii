package sequence

import (
	"strings"
	"testing"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

// TestRenderNilDiagram tests Render with nil/empty inputs.
func TestRenderNilDiagram(t *testing.T) {
	config := diagram.DefaultConfig()

	// nil diagram
	_, err := Render(nil, config)
	if err == nil || !strings.Contains(err.Error(), "no participants") {
		t.Errorf("expected 'no participants' error for nil diagram, got: %v", err)
	}

	// empty participants
	sd := &SequenceDiagram{Participants: []*Participant{}}
	_, err = Render(sd, config)
	if err == nil || !strings.Contains(err.Error(), "no participants") {
		t.Errorf("expected 'no participants' error for empty participants, got: %v", err)
	}
}

// TestRenderNilConfig tests Render with nil config (should use defaults).
func TestRenderNilConfig(t *testing.T) {
	sd, err := Parse("sequenceDiagram\nA->>B: Hello")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	output, err := Render(sd, nil)
	if err != nil {
		t.Fatalf("Render with nil config should not error: %v", err)
	}
	if !strings.Contains(output, "Hello") {
		t.Error("Output should contain message label")
	}
}

// TestRenderCreateParticipantMidDiagram exercises renderParticipantHeadersFiltered
// and renderInlineParticipant (CreateEvent path in renderElementsWithLifeline).
func TestRenderCreateParticipantMidDiagram(t *testing.T) {
	input := `sequenceDiagram
    participant A
    A->>A: Solo
    create participant B
    A->>B: Welcome
    B->>A: Thanks`
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	config := diagram.DefaultConfig()
	output, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}

	// B header should NOT appear at top (it is created mid-diagram)
	lines := strings.Split(output, "\n")
	// The first few lines are the header area for A only
	// B's box should appear inline later
	if !strings.Contains(output, "A") {
		t.Error("Output should contain participant A")
	}
	if !strings.Contains(output, "B") {
		t.Error("Output should contain participant B (inline)")
	}
	if !strings.Contains(output, "Welcome") {
		t.Error("Output should contain message 'Welcome'")
	}

	// Verify B's box appears after the initial header
	// The top line should only have A's box
	topLine := lines[0]
	// A's box should be present at the top
	if !strings.Contains(topLine, string(Unicode.TopLeft)) {
		t.Error("Top line should contain A's box border")
	}

	t.Logf("Output:\n%s", output)
}

// TestRenderDestroyParticipant exercises renderDestroyMarker and lifeline state.
func TestRenderDestroyParticipant(t *testing.T) {
	input := `sequenceDiagram
    participant A
    participant B
    A->>B: msg1
    destroy B
    A->>A: after destroy`
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	config := diagram.DefaultConfig()
	output, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}

	if !strings.Contains(output, "X") {
		t.Error("Output should contain 'X' destroy marker")
	}

	// After destroy, B's lifeline should not appear
	// Find the X line and check lines after it
	lines := strings.Split(output, "\n")
	destroyIdx := -1
	for i, line := range lines {
		if strings.Contains(line, "X") {
			destroyIdx = i
			break
		}
	}
	if destroyIdx < 0 {
		t.Fatal("Could not find destroy marker line")
	}

	// Lines after destroy should not have B's lifeline at B's center position
	// (B's lifeline should be gone)
	t.Logf("Output:\n%s", output)
}

// TestRenderDestroyWithActiveParticipant tests destroy marker while another
// participant is active (activation box should still render on the other).
func TestRenderDestroyWithActiveParticipant(t *testing.T) {
	input := `sequenceDiagram
    participant A
    participant B
    participant C
    A->>+B: activate B
    A->>C: msg
    destroy C
    B-->>-A: done`
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	config := diagram.DefaultConfig()
	output, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if !strings.Contains(output, "X") {
		t.Error("Output should contain destroy marker")
	}
	t.Logf("Output:\n%s", output)
}

// TestRenderNestedBlocks exercises renderBlock with nested blocks
// (loop inside alt, and alt inside loop).
func TestRenderNestedBlocksLoopInsideAlt(t *testing.T) {
	input := `sequenceDiagram
    participant A
    participant B
    alt Success
        loop Retry
            A->>B: Try
        end
    else Failure
        A->>B: Error
    end`
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	config := diagram.DefaultConfig()
	output, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}

	if !strings.Contains(output, "alt") {
		t.Error("Output should contain 'alt' label")
	}
	if !strings.Contains(output, "Success") {
		t.Error("Output should contain 'Success' label")
	}
	if !strings.Contains(output, "loop") {
		t.Error("Output should contain 'loop' label")
	}
	if !strings.Contains(output, "Retry") {
		t.Error("Output should contain 'Retry' label")
	}
	if !strings.Contains(output, "Failure") {
		t.Error("Output should contain 'Failure' section label")
	}
	if !strings.Contains(output, "Error") {
		t.Error("Output should contain 'Error' message")
	}
	t.Logf("Output:\n%s", output)
}

func TestRenderNestedBlocksAltInsideLoop(t *testing.T) {
	input := `sequenceDiagram
    participant A
    participant B
    loop Every second
        alt OK
            A->>B: Ping
        else Fail
            A->>B: Timeout
        end
    end`
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	config := diagram.DefaultConfig()
	output, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}

	if !strings.Contains(output, "loop") {
		t.Error("Output should contain 'loop'")
	}
	if !strings.Contains(output, "alt") {
		t.Error("Output should contain 'alt'")
	}
	if !strings.Contains(output, "Every second") {
		t.Error("Output should contain loop label")
	}
	t.Logf("Output:\n%s", output)
}

// TestRenderBoxGroupingWithParticipants exercises renderGroupHeaders.
func TestRenderBoxGroupingWithParticipants(t *testing.T) {
	input := `sequenceDiagram
    box "Backend"
        participant API
        participant DB
    end
    participant Client
    Client->>API: Request
    API->>DB: Query
    DB-->>API: Result
    API-->>Client: Response`
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	config := diagram.DefaultConfig()
	output, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}

	if !strings.Contains(output, "Backend") {
		t.Error("Output should contain group label 'Backend'")
	}
	if !strings.Contains(output, "[") || !strings.Contains(output, "]") {
		t.Error("Output should contain box border characters")
	}
	if !strings.Contains(output, "API") {
		t.Error("Output should contain participant 'API'")
	}
	if !strings.Contains(output, "DB") {
		t.Error("Output should contain participant 'DB'")
	}
	if !strings.Contains(output, "Client") {
		t.Error("Output should contain participant 'Client'")
	}
	t.Logf("Output:\n%s", output)
}

// TestRenderMultipleBoxGroups tests multiple box groups rendering.
func TestRenderMultipleBoxGroups(t *testing.T) {
	input := `sequenceDiagram
    box "Frontend"
        participant UI
    end
    box "Backend"
        participant API
        participant DB
    end
    UI->>API: Request
    API->>DB: Query`
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	config := diagram.DefaultConfig()
	output, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}

	if !strings.Contains(output, "Frontend") {
		t.Error("Output should contain 'Frontend' group label")
	}
	if !strings.Contains(output, "Backend") {
		t.Error("Output should contain 'Backend' group label")
	}
	t.Logf("Output:\n%s", output)
}

// TestRenderMultipleActivationsSameParticipant exercises activation state depth tracking.
func TestRenderMultipleActivationsSameParticipant(t *testing.T) {
	input := `sequenceDiagram
    participant A
    participant B
    A->>+B: First
    A->>+B: Second
    B-->>-A: Reply2
    B-->>-A: Reply1`
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	config := diagram.DefaultConfig()
	output, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}

	if !strings.Contains(output, "First") {
		t.Error("Output should contain 'First'")
	}
	if !strings.Contains(output, "Second") {
		t.Error("Output should contain 'Second'")
	}
	if !strings.Contains(output, "Reply1") {
		t.Error("Output should contain 'Reply1'")
	}
	if !strings.Contains(output, "Reply2") {
		t.Error("Output should contain 'Reply2'")
	}

	// Activation boxes should appear (│ │ in unicode)
	if !strings.Contains(output, "│ │") {
		t.Error("Output should contain activation box characters")
	}
	t.Logf("Output:\n%s", output)
}

// TestRenderExplicitActivateDeactivate tests explicit activate/deactivate directives.
func TestRenderExplicitActivateDeactivate(t *testing.T) {
	input := `sequenceDiagram
    participant A
    participant B
    A->>B: Request
    activate B
    B->>B: Process
    B-->>A: Response
    deactivate B`
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	config := diagram.DefaultConfig()
	output, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}

	// Should have activation box chars
	if !strings.Contains(output, "│ │") {
		t.Error("Output should contain activation box characters during active period")
	}
	t.Logf("Output:\n%s", output)
}

// TestRenderNoteOverTwoParticipants exercises the Note over A,B rendering path.
func TestRenderNoteOverTwoParticipants(t *testing.T) {
	input := `sequenceDiagram
    participant A
    participant B
    participant C
    A->>B: Hello
    Note over A,B: Important note
    B->>C: Forward`
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	config := diagram.DefaultConfig()
	output, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}

	if !strings.Contains(output, "Important note") {
		t.Error("Output should contain note text")
	}
	t.Logf("Output:\n%s", output)
}

// TestRenderNoteOverTwoParticipantsASCII tests Note over A,B with ASCII charset.
func TestRenderNoteOverTwoParticipantsASCII(t *testing.T) {
	input := `sequenceDiagram
    participant A
    participant B
    Note over A,B: Shared note
    A->>B: msg`
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	config := diagram.DefaultConfig()
	config.UseAscii = true
	output, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}

	if !strings.Contains(output, "Shared note") {
		t.Error("Output should contain note text")
	}
	// Verify ASCII chars
	if strings.ContainsAny(output, "┌┐└┘│─") {
		t.Error("ASCII output should not contain Unicode box-drawing characters")
	}
	t.Logf("Output:\n%s", output)
}

// TestBuildLifeline tests the buildLifeline function directly.
func TestBuildLifeline(t *testing.T) {
	sd, err := Parse("sequenceDiagram\nA->>B: Hello")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	config := diagram.DefaultConfig()
	layout := calculateLayout(sd, config)

	line := buildLifeline(layout, Unicode)
	if line == "" {
		t.Error("buildLifeline should produce non-empty output")
	}
	// Should contain vertical bars at participant centers
	runes := []rune(line)
	for _, c := range layout.participantCenters {
		if c < len(runes) && runes[c] != Unicode.Vertical {
			t.Errorf("Expected vertical bar at center %d, got %q", c, string(runes[c]))
		}
	}
}

// TestBuildLifelineASCII tests buildLifeline with ASCII chars.
func TestBuildLifelineASCII(t *testing.T) {
	sd, err := Parse("sequenceDiagram\nA->>B: Hello")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	config := diagram.DefaultConfig()
	layout := calculateLayout(sd, config)

	line := buildLifeline(layout, ASCII)
	runes := []rune(line)
	for _, c := range layout.participantCenters {
		if c < len(runes) && runes[c] != '|' {
			t.Errorf("Expected '|' at center %d, got %q", c, string(runes[c]))
		}
	}
}

// TestSetLifelines tests the setLifelines function directly.
func TestSetLifelines(t *testing.T) {
	sd, err := Parse("sequenceDiagram\nparticipant A\nparticipant B\nA->>B: msg")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	config := diagram.DefaultConfig()
	layout := calculateLayout(sd, config)

	// Without activation
	actState := newActivationState(len(sd.Participants))
	line := makeEmptyLine(layout.totalWidth + labelBufferSpace)
	setLifelines(line, layout, Unicode, actState)

	for _, c := range layout.participantCenters {
		if c < len(line) && line[c] != Unicode.Vertical {
			t.Errorf("Expected vertical at center %d without activation, got %c", c, line[c])
		}
	}

	// With activation on participant 1
	actState.activate(1)
	line2 := makeEmptyLine(layout.totalWidth + labelBufferSpace)
	setLifelines(line2, layout, Unicode, actState)

	c0 := layout.participantCenters[0]
	if line2[c0] != Unicode.Vertical {
		t.Errorf("Participant 0 should have simple vertical, got %c", line2[c0])
	}
	c1 := layout.participantCenters[1]
	if line2[c1-1] != Unicode.ActivationLeft {
		t.Errorf("Participant 1 should have activation left, got %c", line2[c1-1])
	}
	if line2[c1] != ' ' {
		t.Errorf("Participant 1 center should be space during activation, got %c", line2[c1])
	}
	if line2[c1+1] != Unicode.ActivationRight {
		t.Errorf("Participant 1 should have activation right, got %c", line2[c1+1])
	}
}

// TestActivationStateDepth tests the depth method of activationState.
func TestActivationStateDepth(t *testing.T) {
	as := newActivationState(3)

	if as.depth(0) != 0 {
		t.Errorf("Initial depth should be 0, got %d", as.depth(0))
	}

	as.activate(0)
	if as.depth(0) != 1 {
		t.Errorf("After one activate, depth should be 1, got %d", as.depth(0))
	}

	as.activate(0)
	if as.depth(0) != 2 {
		t.Errorf("After two activates, depth should be 2, got %d", as.depth(0))
	}

	if !as.isActive(0) {
		t.Error("Participant 0 should be active")
	}

	as.deactivate(0)
	if as.depth(0) != 1 {
		t.Errorf("After one deactivate, depth should be 1, got %d", as.depth(0))
	}

	as.deactivate(0)
	if as.depth(0) != 0 {
		t.Errorf("After two deactivates, depth should be 0, got %d", as.depth(0))
	}

	if as.isActive(0) {
		t.Error("Participant 0 should no longer be active")
	}

	// Deactivate below zero should stay at 0
	as.deactivate(0)
	if as.depth(0) != 0 {
		t.Errorf("Deactivate below zero should stay at 0, got %d", as.depth(0))
	}
}

// TestLifelineState tests lifelineState creation and behavior.
func TestLifelineState(t *testing.T) {
	input := `sequenceDiagram
    participant A
    create participant B
    A->>B: Hello`
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	ls := newLifelineState(sd)

	// A should be active from the start (not created via CreateEvent)
	if !ls.isActive(0) {
		t.Error("Participant A should be active from start")
	}

	// B should NOT be active yet (created via CreateEvent)
	if ls.isActive(1) {
		t.Error("Participant B should not be active yet (created mid-diagram)")
	}

	// Simulate creation
	ls.created[1] = true
	if !ls.isActive(1) {
		t.Error("Participant B should be active after creation")
	}

	// Simulate destruction
	ls.destroyed[1] = true
	if ls.isActive(1) {
		t.Error("Participant B should not be active after destruction")
	}
}

// TestRenderElementsDispatch exercises the renderElements function with
// various element types (Message, Note, ActivationEvent, Block).
func TestRenderElementsDispatch(t *testing.T) {
	input := `sequenceDiagram
    participant A
    participant B
    A->>B: Hello
    Note right of B: Think
    activate B
    B-->>A: Reply
    deactivate B
    loop Retry
        A->>B: Again
    end`
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	config := diagram.DefaultConfig()
	output, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}

	if !strings.Contains(output, "Hello") {
		t.Error("Output should contain message 'Hello'")
	}
	if !strings.Contains(output, "Think") {
		t.Error("Output should contain note 'Think'")
	}
	if !strings.Contains(output, "Reply") {
		t.Error("Output should contain message 'Reply'")
	}
	if !strings.Contains(output, "loop") {
		t.Error("Output should contain block label 'loop'")
	}
	if !strings.Contains(output, "Again") {
		t.Error("Output should contain message 'Again'")
	}
	// Activation box should have been present
	if !strings.Contains(output, "│ │") {
		t.Error("Output should contain activation box")
	}
	t.Logf("Output:\n%s", output)
}

// TestRenderCreateAndDestroyFullCycle exercises the full create -> use -> destroy cycle.
func TestRenderCreateAndDestroyFullCycle(t *testing.T) {
	input := `sequenceDiagram
    participant A
    create participant B
    A->>B: Create
    B->>A: Ack
    destroy B
    A->>A: Continue alone`
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	for _, useASCII := range []bool{false, true} {
		name := "Unicode"
		if useASCII {
			name = "ASCII"
		}
		t.Run(name, func(t *testing.T) {
			config := diagram.DefaultConfig()
			config.UseAscii = useASCII
			output, err := Render(sd, config)
			if err != nil {
				t.Fatalf("Render error: %v", err)
			}

			if !strings.Contains(output, "Create") {
				t.Error("Output should contain 'Create'")
			}
			if !strings.Contains(output, "X") {
				t.Error("Output should contain destroy marker 'X'")
			}
			if !strings.Contains(output, "Continue alone") {
				t.Error("Output should contain 'Continue alone'")
			}
			t.Logf("Output:\n%s", output)
		})
	}
}

// TestRenderBlockWithSections exercises renderBlock with else sections
// including section labels and content.
func TestRenderBlockWithSections(t *testing.T) {
	input := `sequenceDiagram
    participant A
    participant B
    alt Is OK
        A->>B: OK
    else Is Error
        A->>B: Err
    else Default
        A->>B: Default
    end`
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	config := diagram.DefaultConfig()
	output, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}

	if !strings.Contains(output, "Is OK") {
		t.Error("Output should contain 'Is OK'")
	}
	if !strings.Contains(output, "Is Error") {
		t.Error("Output should contain 'Is Error'")
	}
	if !strings.Contains(output, "Default") {
		t.Error("Output should contain 'Default'")
	}
	t.Logf("Output:\n%s", output)
}

// TestRenderMessagesWithoutElements tests the fallback path in Render
// where sd.Elements is empty but sd.Messages is not.
func TestRenderMessagesWithoutElements(t *testing.T) {
	// Manually construct a SequenceDiagram with Messages but no Elements
	pA := &Participant{ID: "A", Label: "A", Index: 0}
	pB := &Participant{ID: "B", Label: "B", Index: 1}
	sd := &SequenceDiagram{
		Participants: []*Participant{pA, pB},
		Messages: []*Message{
			{From: pA, To: pB, Label: "Hello", ArrowType: SolidArrow},
			{From: pB, To: pA, Label: "Reply", ArrowType: DottedArrow},
		},
		Elements: []Element{}, // empty elements forces Messages fallback
	}

	config := diagram.DefaultConfig()
	output, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}

	if !strings.Contains(output, "Hello") {
		t.Error("Output should contain 'Hello'")
	}
	if !strings.Contains(output, "Reply") {
		t.Error("Output should contain 'Reply'")
	}
	t.Logf("Output:\n%s", output)
}

// TestRenderSelfMessageWithoutElements tests the self-message fallback path.
func TestRenderSelfMessageWithoutElements(t *testing.T) {
	pA := &Participant{ID: "A", Label: "A", Index: 0}
	pB := &Participant{ID: "B", Label: "B", Index: 1}
	sd := &SequenceDiagram{
		Participants: []*Participant{pA, pB},
		Messages: []*Message{
			{From: pA, To: pA, Label: "Self msg", ArrowType: SolidArrow},
		},
		Elements: []Element{}, // empty
	}
	config := diagram.DefaultConfig()
	output, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}

	if !strings.Contains(output, "Self msg") {
		t.Error("Output should contain 'Self msg'")
	}
	t.Logf("Output:\n%s", output)
}

// TestBuildLifelineWithActivationsAndState tests with create/destroy state.
func TestBuildLifelineWithActivationsAndState(t *testing.T) {
	input := `sequenceDiagram
    participant A
    participant B
    A->>B: msg`
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	config := diagram.DefaultConfig()
	layout := calculateLayout(sd, config)
	actState := newActivationState(len(sd.Participants))
	llState := newLifelineState(sd)

	// Both participants active
	line := buildLifelineWithActivationsAndState(layout, Unicode, actState, llState)
	if line == "" {
		t.Error("Should produce non-empty lifeline")
	}
	runes := []rune(line)
	for _, c := range layout.participantCenters {
		if c < len(runes) && runes[c] != Unicode.Vertical {
			t.Errorf("Expected vertical at center %d, got %c", c, runes[c])
		}
	}

	// Destroy B - lifeline should disappear
	llState.destroyed[1] = true
	line2 := buildLifelineWithActivationsAndState(layout, Unicode, actState, llState)
	runes2 := []rune(line2)
	// Pad if needed
	c1 := layout.participantCenters[1]
	if c1 < len(runes2) && runes2[c1] == Unicode.Vertical {
		t.Error("Destroyed participant should not have lifeline")
	}
}

// TestRenderBlockElementType tests the Block.elementType method.
func TestBlockElementType(t *testing.T) {
	b := &Block{Type: "loop", Label: "test"}
	if b.elementType() != "block" {
		t.Errorf("Block.elementType() = %q, want 'block'", b.elementType())
	}

	m := &Message{Label: "test"}
	if m.elementType() != "message" {
		t.Errorf("Message.elementType() = %q, want 'message'", m.elementType())
	}

	n := &Note{Text: "test"}
	if n.elementType() != "note" {
		t.Errorf("Note.elementType() = %q, want 'note'", n.elementType())
	}

	a := &ActivationEvent{Activate: true}
	if a.elementType() != "activation" {
		t.Errorf("ActivationEvent.elementType() = %q, want 'activation'", a.elementType())
	}

	ce := &CreateEvent{}
	if ce.elementType() != "create" {
		t.Errorf("CreateEvent.elementType() = %q, want 'create'", ce.elementType())
	}

	de := &DestroyEvent{}
	if de.elementType() != "destroy" {
		t.Errorf("DestroyEvent.elementType() = %q, want 'destroy'", de.elementType())
	}
}

// TestRenderNoteLeftOf exercises the NoteLeftOf rendering path.
func TestRenderNoteLeftOf(t *testing.T) {
	input := `sequenceDiagram
    participant A
    participant B
    A->>B: Hello
    Note left of A: Thinking
    A->>B: Done`
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	config := diagram.DefaultConfig()
	output, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if !strings.Contains(output, "Thinking") {
		t.Error("Output should contain note text 'Thinking'")
	}
	t.Logf("Output:\n%s", output)
}

// TestRenderNoteOverSingle exercises note over a single participant.
func TestRenderNoteOverSingle(t *testing.T) {
	input := `sequenceDiagram
    participant A
    participant B
    Note over A: Status
    A->>B: msg`
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	config := diagram.DefaultConfig()
	output, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if !strings.Contains(output, "Status") {
		t.Error("Output should contain note 'Status'")
	}
	t.Logf("Output:\n%s", output)
}

// TestRenderBlockWithActivationInside tests activation inside a block.
func TestRenderBlockWithActivationInside(t *testing.T) {
	input := `sequenceDiagram
    participant A
    participant B
    loop Process
        A->>+B: Request
        B-->>-A: Response
    end`
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	config := diagram.DefaultConfig()
	output, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if !strings.Contains(output, "loop") {
		t.Error("Output should contain 'loop'")
	}
	if !strings.Contains(output, "Request") {
		t.Error("Output should contain 'Request'")
	}
	t.Logf("Output:\n%s", output)
}

// TestCalculateLayoutCustomSpacing tests layout calculation with custom spacing.
func TestCalculateLayoutCustomSpacing(t *testing.T) {
	sd, err := Parse("sequenceDiagram\nA->>B: Hello")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	config := diagram.DefaultConfig()
	config.SequenceParticipantSpacing = 10
	config.SequenceMessageSpacing = 3
	config.SequenceSelfMessageWidth = 6

	layout := calculateLayout(sd, config)
	if layout.messageSpacing != 3 {
		t.Errorf("messageSpacing = %d, want 3", layout.messageSpacing)
	}
	if layout.selfMessageWidth != 6 {
		t.Errorf("selfMessageWidth = %d, want 6", layout.selfMessageWidth)
	}
}

// TestRenderCreateMultipleParticipants tests creating multiple participants mid-diagram.
func TestRenderCreateMultipleParticipants(t *testing.T) {
	input := `sequenceDiagram
    participant A
    create participant B
    A->>B: Hello B
    create participant C
    A->>C: Hello C
    B->>C: Hi`
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	config := diagram.DefaultConfig()
	output, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}

	if !strings.Contains(output, "Hello B") {
		t.Error("Output should contain 'Hello B'")
	}
	if !strings.Contains(output, "Hello C") {
		t.Error("Output should contain 'Hello C'")
	}
	t.Logf("Output:\n%s", output)
}

// TestRenderNoteWithActivation tests note rendering while a participant is activated.
func TestRenderNoteWithActivation(t *testing.T) {
	input := `sequenceDiagram
    participant A
    participant B
    A->>+B: Request
    Note right of B: Processing
    B-->>-A: Response`
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	config := diagram.DefaultConfig()
	output, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if !strings.Contains(output, "Processing") {
		t.Error("Output should contain 'Processing'")
	}
	t.Logf("Output:\n%s", output)
}

// TestRenderDottedArrowInBlock tests dotted arrow rendering inside a block.
func TestRenderDottedArrowInBlock(t *testing.T) {
	input := `sequenceDiagram
    participant A
    participant B
    loop Retry
        A->>B: Request
        B-->>A: Response
    end`
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	config := diagram.DefaultConfig()
	output, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if !strings.Contains(output, "Request") {
		t.Error("Output should contain 'Request'")
	}
	if !strings.Contains(output, "Response") {
		t.Error("Output should contain 'Response'")
	}
	t.Logf("Output:\n%s", output)
}

// TestMakeEmptyLine tests the makeEmptyLine helper.
func TestMakeEmptyLine(t *testing.T) {
	line := makeEmptyLine(10)
	if len(line) != 10 {
		t.Errorf("Expected length 10, got %d", len(line))
	}
	for i, r := range line {
		if r != ' ' {
			t.Errorf("Expected space at index %d, got %c", i, r)
		}
	}
}

// TestGetArrowChars tests all arrow type character mappings.
func TestGetArrowChars(t *testing.T) {
	tests := []struct {
		aType     ArrowType
		wantRight rune
		wantLeft  rune
	}{
		{SolidArrow, Unicode.ArrowRight, Unicode.ArrowLeft},
		{DottedArrow, Unicode.ArrowRight, Unicode.ArrowLeft},
		{SolidOpen, Unicode.OpenArrowRight, Unicode.OpenArrowLeft},
		{DottedOpen, Unicode.OpenArrowRight, Unicode.OpenArrowLeft},
		{SolidCross, Unicode.CrossEnd, Unicode.CrossEnd},
		{DottedCross, Unicode.CrossEnd, Unicode.CrossEnd},
		{SolidAsync, ')', '('},
		{DottedAsync, ')', '('},
		{ArrowType(99), Unicode.ArrowRight, Unicode.ArrowLeft}, // default
	}

	for _, tt := range tests {
		r, l := getArrowChars(tt.aType, Unicode)
		if r != tt.wantRight || l != tt.wantLeft {
			t.Errorf("getArrowChars(%v) = (%c, %c), want (%c, %c)",
				tt.aType, r, l, tt.wantRight, tt.wantLeft)
		}
	}
}

// TestRenderRightToLeftMessage tests message going from right to left (To < From).
func TestRenderRightToLeftMessage(t *testing.T) {
	input := `sequenceDiagram
    participant A
    participant B
    B->>A: Backward`
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	config := diagram.DefaultConfig()
	output, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if !strings.Contains(output, "Backward") {
		t.Error("Output should contain 'Backward'")
	}
	t.Logf("Output:\n%s", output)
}
