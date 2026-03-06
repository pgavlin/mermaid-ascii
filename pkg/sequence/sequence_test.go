package sequence

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		wantParticipants int
		wantMessages     int
		wantErr          string
	}{
		{"empty input", "", 0, 0, "empty input"},
		{"missing sequenceDiagram keyword", "A->>B: Hello", 0, 0, "expected \"sequenceDiagram\" keyword"},
		{"only comments", "sequenceDiagram\n%% This is a comment\n%% Another comment", 0, 0, "no participants found"},
		{"no participants", "sequenceDiagram", 0, 0, "no participants found"},
		{"duplicate participant ID", "sequenceDiagram\nparticipant Alice\nparticipant Alice\nAlice->>Bob: Hi", 0, 0, "duplicate participant"},
		{"minimal diagram", "sequenceDiagram\nA->>B: Hello", 2, 1, ""},
		{"explicit participants", "sequenceDiagram\nparticipant Alice\nparticipant Bob\nAlice->>Bob: Hi", 2, 1, ""},
		{"dotted arrow", "sequenceDiagram\nA-->>B: Response", 2, 1, ""},
		{"self message", "sequenceDiagram\nA->>A: Self", 1, 1, ""},
		{"multiple messages", "sequenceDiagram\nA->>B: 1\nB->>C: 2\nC-->>A: 3", 3, 3, ""},
		{"with comments", "sequenceDiagram\n%% Comment\nA->>B: Hi %% inline comment", 2, 1, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sd, err := Parse(tt.input)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("Expected error containing %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if len(sd.Participants) != tt.wantParticipants {
				t.Errorf("Expected %d participants, got %d", tt.wantParticipants, len(sd.Participants))
			}
			if len(sd.Messages) != tt.wantMessages {
				t.Errorf("Expected %d messages, got %d", tt.wantMessages, len(sd.Messages))
			}
		})
	}
}

func TestIsSequenceDiagram(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"sequenceDiagram\nA->>B: Hello", true},
		{"graph LR\nA-->B", false},
		{"graph TD\nA-->B", false},
		{"", false},
		{"%% Just a comment", false},
	}

	for _, tt := range tests {
		if got := IsSequenceDiagram(tt.input); got != tt.want {
			t.Errorf("IsSequenceDiagram(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParticipantAlias(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantID    string
		wantLabel string
	}{
		{"simple alias", "sequenceDiagram\nparticipant A as Alice\nA->>A: Hello", "A", "Alice"},
		{"no alias defaults to id", "sequenceDiagram\nparticipant Alice\nAlice->>Alice: Hi", "Alice", "Alice"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(d.Participants) == 0 {
				t.Fatal("expected at least one participant")
			}
			p := d.Participants[0]
			if p.ID != tt.wantID {
				t.Errorf("ID = %q, want %q", p.ID, tt.wantID)
			}
			if p.Label != tt.wantLabel {
				t.Errorf("Label = %q, want %q", p.Label, tt.wantLabel)
			}
			config := diagram.DefaultConfig()
			output, err := Render(d, config)
			if err != nil {
				t.Fatalf("render error: %v", err)
			}
			if !strings.Contains(output, tt.wantLabel) {
				t.Errorf("output should contain label %q", tt.wantLabel)
			}
		})
	}
}

func TestMessageRegex(t *testing.T) {
	tests := []struct {
		input     string
		wantFrom  string
		wantArrow string
		wantTo    string
		wantLabel string
		wantMatch bool
	}{
		{"A->>B: Hello", "A", "->>", "B", "Hello", true},
		{"A-->>B: Response", "A", "-->>", "B", "Response", true},
		{`"My Service"->>B: Test`, "My Service", "->>", "B", "Test", true},
		{"A->>B: ", "A", "->>", "B", "", true},
		{"A->B: Test", "A", "->", "B", "Test", true},
		{"A-->B: Test", "A", "-->", "B", "Test", true},
		{"A-xB: Test", "A", "-x", "B", "Test", true},
		{"A--xB: Test", "A", "--x", "B", "Test", true},
		{"A-)B: Test", "A", "-)", "B", "Test", true},
		{"A--)B: Test", "A", "--)", "B", "Test", true},
		{"A->>+B: Test", "A", "->>", "B", "Test", true},
		{"A-->>-B: Test", "A", "-->>", "B", "Test", true},
		{"A->>B", "", "", "", "", false},
	}

	for _, tt := range tests {
		match := messageRegex.FindStringSubmatch(tt.input)
		if !tt.wantMatch {
			if match != nil {
				t.Errorf("messageRegex should not match %q", tt.input)
			}
			continue
		}
		if match == nil {
			t.Fatalf("messageRegex failed to match: %q", tt.input)
		}
		gotFrom := match[2]
		if match[1] != "" {
			gotFrom = match[1]
		}
		gotArrow := match[3]
		gotTo := match[6]
		if match[5] != "" {
			gotTo = match[5]
		}
		gotLabel := match[7]

		if gotFrom != tt.wantFrom || gotArrow != tt.wantArrow || gotTo != tt.wantTo || gotLabel != tt.wantLabel {
			t.Errorf("messageRegex(%q) = (%q, %q, %q, %q), want (%q, %q, %q, %q)",
				tt.input, gotFrom, gotArrow, gotTo, gotLabel, tt.wantFrom, tt.wantArrow, tt.wantTo, tt.wantLabel)
		}
	}
}

func TestParticipantRegex(t *testing.T) {
	tests := []struct {
		input     string
		wantID    string
		wantAlias string
	}{
		{"participant Alice", "Alice", ""},
		{"participant Alice as A", "Alice", "A"},
		{`participant "My Service"`, "My Service", ""},
		{`participant "My Service" as Service`, "My Service", "Service"},
	}

	for _, tt := range tests {
		match := participantRegex.FindStringSubmatch(tt.input)
		if match == nil {
			t.Fatalf("participantRegex failed to match: %q", tt.input)
		}
		gotID := match[2]
		if match[1] != "" {
			gotID = match[1]
		}
		gotAlias := match[3]

		if gotID != tt.wantID || gotAlias != tt.wantAlias {
			t.Errorf("participantRegex(%q) = (%q, %q), want (%q, %q)",
				tt.input, gotID, gotAlias, tt.wantID, tt.wantAlias)
		}
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"A->>B: Hello", []string{"A->>B: Hello"}},
		{"line1\nline2\nline3", []string{"line1", "line2", "line3"}},
		{"line1\\nline2\\nline3", []string{"line1", "line2", "line3"}},
		{"", []string{""}},
	}

	for _, tt := range tests {
		result := diagram.SplitLines(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("SplitLines(%q) len = %d, want %d", tt.input, len(result), len(tt.expected))
		}
	}
}

func TestRemoveComments(t *testing.T) {
	tests := []struct {
		input    []string
		expected []string
	}{
		{[]string{"A->>B: Hello", "B-->>A: Hi"}, []string{"A->>B: Hello", "B-->>A: Hi"}},
		{[]string{"%% This is a comment", "A->>B: Hello"}, []string{"A->>B: Hello"}},
		{[]string{"A->>B: Hello %% inline comment", "B-->>A: Hi"}, []string{"A->>B: Hello", "B-->>A: Hi"}},
	}

	for _, tt := range tests {
		result := diagram.RemoveComments(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("RemoveComments() len = %d, want %d", len(result), len(tt.expected))
		}
	}
}

func TestArrowTypeString(t *testing.T) {
	if SolidArrow.String() != "solid" {
		t.Errorf("SolidArrow.String() = %q, want \"solid\"", SolidArrow.String())
	}
	if DottedArrow.String() != "dotted" {
		t.Errorf("DottedArrow.String() = %q, want \"dotted\"", DottedArrow.String())
	}
	if SolidOpen.String() != "solid_open" {
		t.Errorf("SolidOpen.String() = %q, want \"solid_open\"", SolidOpen.String())
	}
	if DottedOpen.String() != "dotted_open" {
		t.Errorf("DottedOpen.String() = %q, want \"dotted_open\"", DottedOpen.String())
	}
	if SolidCross.String() != "solid_cross" {
		t.Errorf("SolidCross.String() = %q, want \"solid_cross\"", SolidCross.String())
	}
	if DottedCross.String() != "dotted_cross" {
		t.Errorf("DottedCross.String() = %q, want \"dotted_cross\"", DottedCross.String())
	}
	if SolidAsync.String() != "solid_async" {
		t.Errorf("SolidAsync.String() = %q, want \"solid_async\"", SolidAsync.String())
	}
	if DottedAsync.String() != "dotted_async" {
		t.Errorf("DottedAsync.String() = %q, want \"dotted_async\"", DottedAsync.String())
	}
}

// Test 2A: Additional arrow types parsing
func TestParseArrowTypes(t *testing.T) {
	tests := []struct {
		input     string
		wantArrow ArrowType
	}{
		{"sequenceDiagram\nA->>B: msg", SolidArrow},
		{"sequenceDiagram\nA-->>B: msg", DottedArrow},
		{"sequenceDiagram\nA->B: msg", SolidOpen},
		{"sequenceDiagram\nA-->B: msg", DottedOpen},
		{"sequenceDiagram\nA-xB: msg", SolidCross},
		{"sequenceDiagram\nA--xB: msg", DottedCross},
		{"sequenceDiagram\nA-)B: msg", SolidAsync},
		{"sequenceDiagram\nA--)B: msg", DottedAsync},
	}

	for _, tt := range tests {
		sd, err := Parse(tt.input)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", tt.input, err)
		}
		if len(sd.Messages) != 1 {
			t.Fatalf("Expected 1 message, got %d", len(sd.Messages))
		}
		if sd.Messages[0].ArrowType != tt.wantArrow {
			t.Errorf("Parse(%q) ArrowType = %v, want %v", tt.input, sd.Messages[0].ArrowType, tt.wantArrow)
		}
	}
}

func TestArrowTypeIsDotted(t *testing.T) {
	dotted := []ArrowType{DottedArrow, DottedOpen, DottedCross, DottedAsync}
	solid := []ArrowType{SolidArrow, SolidOpen, SolidCross, SolidAsync}

	for _, a := range dotted {
		if !a.IsDotted() {
			t.Errorf("%v.IsDotted() = false, want true", a)
		}
	}
	for _, a := range solid {
		if a.IsDotted() {
			t.Errorf("%v.IsDotted() = true, want false", a)
		}
	}
}

// Test 2B: Activation parsing
func TestParseActivation(t *testing.T) {
	// Test +/- shorthand on arrows
	input := "sequenceDiagram\nA->>+B: Request\nB-->>-A: Response"
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(sd.Messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(sd.Messages))
	}
	if !sd.Messages[0].Activate {
		t.Error("First message should have Activate=true")
	}
	if !sd.Messages[1].Deactivate {
		t.Error("Second message should have Deactivate=true")
	}
}

func TestParseActivateDeactivate(t *testing.T) {
	input := "sequenceDiagram\nA->>B: Request\nactivate B\nB-->>A: Response\ndeactivate B"
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(sd.Messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(sd.Messages))
	}
	// Check that we have activation events in elements
	activationCount := 0
	for _, elem := range sd.Elements {
		if _, ok := elem.(*ActivationEvent); ok {
			activationCount++
		}
	}
	if activationCount != 2 {
		t.Errorf("Expected 2 activation events, got %d", activationCount)
	}
}

func TestParseActivationRender(t *testing.T) {
	input := "sequenceDiagram\nA->>+B: Request\nB-->>-A: Response"
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	config := diagram.DefaultConfig()
	output, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	// Activation should show vertical bars
	if !strings.Contains(output, "│ │") {
		t.Error("Output should contain activation box characters '│ │'")
	}
}

// Test 2C: Notes parsing
func TestParseNoteRightOf(t *testing.T) {
	input := "sequenceDiagram\nA->>B: Hello\nNote right of B: Think"
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	noteCount := 0
	for _, elem := range sd.Elements {
		if n, ok := elem.(*Note); ok {
			noteCount++
			if n.Position != NoteRightOf {
				t.Errorf("Note position = %v, want NoteRightOf", n.Position)
			}
			if n.Text != "Think" {
				t.Errorf("Note text = %q, want %q", n.Text, "Think")
			}
			if n.Participant.ID != "B" {
				t.Errorf("Note participant = %q, want %q", n.Participant.ID, "B")
			}
		}
	}
	if noteCount != 1 {
		t.Errorf("Expected 1 note, got %d", noteCount)
	}
}

func TestParseNoteLeftOf(t *testing.T) {
	input := "sequenceDiagram\nA->>B: Hello\nNote left of A: Wait"
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	for _, elem := range sd.Elements {
		if n, ok := elem.(*Note); ok {
			if n.Position != NoteLeftOf {
				t.Errorf("Note position = %v, want NoteLeftOf", n.Position)
			}
		}
	}
}

func TestParseNoteOver(t *testing.T) {
	input := "sequenceDiagram\nNote over A: Start\nA->>B: Hello"
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	for _, elem := range sd.Elements {
		if n, ok := elem.(*Note); ok {
			if n.Position != NoteOver {
				t.Errorf("Note position = %v, want NoteOver", n.Position)
			}
			if n.EndParticipant != nil {
				t.Error("Single-participant note should not have EndParticipant")
			}
		}
	}
}

func TestParseNoteOverTwo(t *testing.T) {
	input := "sequenceDiagram\nNote over A,B: Shared\nA->>B: Hello"
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	for _, elem := range sd.Elements {
		if n, ok := elem.(*Note); ok {
			if n.Position != NoteOver {
				t.Errorf("Note position = %v, want NoteOver", n.Position)
			}
			if n.EndParticipant == nil {
				t.Error("Two-participant note should have EndParticipant")
			}
			if n.EndParticipant != nil && n.EndParticipant.ID != "B" {
				t.Errorf("EndParticipant = %q, want %q", n.EndParticipant.ID, "B")
			}
		}
	}
}

func TestNoteRender(t *testing.T) {
	input := "sequenceDiagram\nA->>B: Hello\nNote right of B: Think\nB-->>A: Reply"
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	config := diagram.DefaultConfig()
	output, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if !strings.Contains(output, "Think") {
		t.Error("Output should contain note text 'Think'")
	}
}

// Test 2D: Interaction blocks
func TestParseLoopBlock(t *testing.T) {
	input := "sequenceDiagram\nloop Every minute\n    A->>B: Ping\nend"
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	blockCount := 0
	for _, elem := range sd.Elements {
		if b, ok := elem.(*Block); ok {
			blockCount++
			if b.Type != "loop" {
				t.Errorf("Block type = %q, want %q", b.Type, "loop")
			}
			if b.Label != "Every minute" {
				t.Errorf("Block label = %q, want %q", b.Label, "Every minute")
			}
			if len(b.Elements) != 1 {
				t.Errorf("Block should have 1 element, got %d", len(b.Elements))
			}
		}
	}
	if blockCount != 1 {
		t.Errorf("Expected 1 block, got %d", blockCount)
	}
}

func TestParseAltBlock(t *testing.T) {
	input := "sequenceDiagram\nalt Success\n    A->>B: OK\nelse Failure\n    A->>B: Error\nend"
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	for _, elem := range sd.Elements {
		if b, ok := elem.(*Block); ok {
			if b.Type != "alt" {
				t.Errorf("Block type = %q, want %q", b.Type, "alt")
			}
			if len(b.Sections) != 1 {
				t.Errorf("Block should have 1 section, got %d", len(b.Sections))
			}
			if len(b.Sections) > 0 && b.Sections[0].Label != "Failure" {
				t.Errorf("Section label = %q, want %q", b.Sections[0].Label, "Failure")
			}
		}
	}
}

func TestParseOptBlock(t *testing.T) {
	input := "sequenceDiagram\nopt Optional\n    A->>B: Maybe\nend"
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	for _, elem := range sd.Elements {
		if b, ok := elem.(*Block); ok {
			if b.Type != "opt" {
				t.Errorf("Block type = %q, want %q", b.Type, "opt")
			}
		}
	}
}

func TestParseNestedBlocks(t *testing.T) {
	input := "sequenceDiagram\nloop Retry\n    alt Success\n        A->>B: OK\n    else Fail\n        A->>B: Error\n    end\nend"
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	for _, elem := range sd.Elements {
		if b, ok := elem.(*Block); ok {
			if b.Type != "loop" {
				t.Errorf("Outer block type = %q, want %q", b.Type, "loop")
			}
			// The inner alt block should be in the loop's elements
			innerBlockCount := 0
			for _, innerElem := range b.Elements {
				if inner, ok := innerElem.(*Block); ok {
					innerBlockCount++
					if inner.Type != "alt" {
						t.Errorf("Inner block type = %q, want %q", inner.Type, "alt")
					}
				}
			}
			if innerBlockCount != 1 {
				t.Errorf("Expected 1 inner block, got %d", innerBlockCount)
			}
		}
	}
}

func TestParseParBlock(t *testing.T) {
	input := "sequenceDiagram\npar Task1\n    A->>B: Do1\nand Task2\n    A->>C: Do2\nend"
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	for _, elem := range sd.Elements {
		if b, ok := elem.(*Block); ok {
			if b.Type != "par" {
				t.Errorf("Block type = %q, want %q", b.Type, "par")
			}
			if len(b.Sections) != 1 {
				t.Errorf("Block should have 1 section (and), got %d", len(b.Sections))
			}
		}
	}
}

func TestParseCriticalBlock(t *testing.T) {
	input := "sequenceDiagram\ncritical Important\n    A->>B: Do\noption Fallback\n    A->>B: Retry\nend"
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	for _, elem := range sd.Elements {
		if b, ok := elem.(*Block); ok {
			if b.Type != "critical" {
				t.Errorf("Block type = %q, want %q", b.Type, "critical")
			}
			if len(b.Sections) != 1 {
				t.Errorf("Block should have 1 section (option), got %d", len(b.Sections))
			}
		}
	}
}

func TestBlockRender(t *testing.T) {
	input := "sequenceDiagram\nloop Retry\n    A->>B: Ping\n    B-->>A: Pong\nend"
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
		t.Error("Output should contain 'loop' label")
	}
	if !strings.Contains(output, "Retry") {
		t.Error("Output should contain block label")
	}
}

// Test 2E: Actor keyword
func TestParseActor(t *testing.T) {
	input := "sequenceDiagram\nactor User\nparticipant Server\nUser->>Server: Request"
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(sd.Participants) != 2 {
		t.Fatalf("Expected 2 participants, got %d", len(sd.Participants))
	}
	if sd.Participants[0].Type != ParticipantActor {
		t.Error("User should be an actor")
	}
	if sd.Participants[1].Type != ParticipantBox {
		t.Error("Server should be a box")
	}
}

func TestParseActorWithAlias(t *testing.T) {
	input := "sequenceDiagram\nactor U as User\nU->>U: Hello"
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if sd.Participants[0].ID != "U" {
		t.Errorf("ID = %q, want %q", sd.Participants[0].ID, "U")
	}
	if sd.Participants[0].Label != "User" {
		t.Errorf("Label = %q, want %q", sd.Participants[0].Label, "User")
	}
	if sd.Participants[0].Type != ParticipantActor {
		t.Error("Should be an actor")
	}
}

func TestActorRender(t *testing.T) {
	input := "sequenceDiagram\nactor User\nparticipant Server\nUser->>Server: Request"
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	config := diagram.DefaultConfig()
	output, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	// Should contain stick figure characters
	if !strings.Contains(output, "O") {
		t.Error("Actor output should contain 'O' (head)")
	}
	if !strings.Contains(output, "/|\\") {
		t.Error("Actor output should contain '/|\\' (body)")
	}
	if !strings.Contains(output, "/ \\") {
		t.Error("Actor output should contain '/ \\' (legs)")
	}
	// Should still contain participant labels
	if !strings.Contains(output, "User") {
		t.Error("Output should contain actor label 'User'")
	}
	if !strings.Contains(output, "Server") {
		t.Error("Output should contain participant label 'Server'")
	}
}

// Test 2F: Participant Grouping (box ... end)
func TestParseBoxGrouping(t *testing.T) {
	input := "sequenceDiagram\nbox \"Internal Services\"\n    participant A\n    participant B\nend\nparticipant C\nA->>B: Hello\nB->>C: World"
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(sd.Participants) != 3 {
		t.Fatalf("Expected 3 participants, got %d", len(sd.Participants))
	}
	if len(sd.Groups) != 1 {
		t.Fatalf("Expected 1 group, got %d", len(sd.Groups))
	}
	g := sd.Groups[0]
	if g.Label != "Internal Services" {
		t.Errorf("Group label = %q, want %q", g.Label, "Internal Services")
	}
	if len(g.Participants) != 2 {
		t.Fatalf("Expected 2 participants in group, got %d", len(g.Participants))
	}
	if g.Participants[0].ID != "A" || g.Participants[1].ID != "B" {
		t.Errorf("Group participants = [%q, %q], want [A, B]", g.Participants[0].ID, g.Participants[1].ID)
	}
}

func TestParseBoxGroupingNoLabel(t *testing.T) {
	input := "sequenceDiagram\nbox\n    participant A\n    participant B\nend\nA->>B: Hello"
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(sd.Groups) != 1 {
		t.Fatalf("Expected 1 group, got %d", len(sd.Groups))
	}
	if sd.Groups[0].Label != "" {
		t.Errorf("Group label = %q, want empty", sd.Groups[0].Label)
	}
}

func TestParseBoxGroupingWithColor(t *testing.T) {
	input := "sequenceDiagram\nbox \"Backend\" #lightblue\n    participant A\n    participant B\nend\nA->>B: Hello"
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(sd.Groups) != 1 {
		t.Fatalf("Expected 1 group, got %d", len(sd.Groups))
	}
	if sd.Groups[0].Label != "Backend" {
		t.Errorf("Group label = %q, want %q", sd.Groups[0].Label, "Backend")
	}
}

func TestParseMultipleBoxGroups(t *testing.T) {
	input := "sequenceDiagram\nbox \"Group1\"\n    participant A\nend\nbox \"Group2\"\n    participant B\nend\nA->>B: Hello"
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(sd.Groups) != 2 {
		t.Fatalf("Expected 2 groups, got %d", len(sd.Groups))
	}
	if sd.Groups[0].Label != "Group1" {
		t.Errorf("Group 0 label = %q, want %q", sd.Groups[0].Label, "Group1")
	}
	if sd.Groups[1].Label != "Group2" {
		t.Errorf("Group 1 label = %q, want %q", sd.Groups[1].Label, "Group2")
	}
}

func TestBoxGroupRender(t *testing.T) {
	input := "sequenceDiagram\nbox \"Services\"\n    participant A\n    participant B\nend\nA->>B: Hello"
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	config := diagram.DefaultConfig()
	output, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	// Should contain the group label
	if !strings.Contains(output, "Services") {
		t.Error("Output should contain group label 'Services'")
	}
	// Should contain group border characters
	if !strings.Contains(output, "[") || !strings.Contains(output, "]") {
		t.Error("Output should contain group border characters '[' and ']'")
	}
	// Should still render participants and messages
	if !strings.Contains(output, "A") {
		t.Error("Output should contain participant 'A'")
	}
	if !strings.Contains(output, "B") {
		t.Error("Output should contain participant 'B'")
	}
	if !strings.Contains(output, "Hello") {
		t.Error("Output should contain message 'Hello'")
	}
}

// Test 2G: Create/Destroy
func TestParseCreateParticipant(t *testing.T) {
	input := "sequenceDiagram\nparticipant A\ncreate participant B\nA->>B: Hello\nB->>A: Hi"
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(sd.Participants) != 2 {
		t.Fatalf("Expected 2 participants, got %d", len(sd.Participants))
	}
	// Check for CreateEvent in elements
	createCount := 0
	for _, elem := range sd.Elements {
		if ce, ok := elem.(*CreateEvent); ok {
			createCount++
			if ce.Participant.ID != "B" {
				t.Errorf("CreateEvent participant = %q, want %q", ce.Participant.ID, "B")
			}
		}
	}
	if createCount != 1 {
		t.Errorf("Expected 1 CreateEvent, got %d", createCount)
	}
}

func TestParseCreateActor(t *testing.T) {
	input := "sequenceDiagram\nparticipant A\ncreate actor B\nA->>B: Hello"
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(sd.Participants) != 2 {
		t.Fatalf("Expected 2 participants, got %d", len(sd.Participants))
	}
	if sd.Participants[1].Type != ParticipantActor {
		t.Error("Created participant should be an actor")
	}
}

func TestParseCreateWithAlias(t *testing.T) {
	input := "sequenceDiagram\nparticipant A\ncreate participant B as Bob\nA->>B: Hello"
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if sd.Participants[1].Label != "Bob" {
		t.Errorf("Created participant label = %q, want %q", sd.Participants[1].Label, "Bob")
	}
}

func TestParseDestroy(t *testing.T) {
	input := "sequenceDiagram\nparticipant A\nparticipant B\nA->>B: Hello\ndestroy B"
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	destroyCount := 0
	for _, elem := range sd.Elements {
		if de, ok := elem.(*DestroyEvent); ok {
			destroyCount++
			if de.Participant.ID != "B" {
				t.Errorf("DestroyEvent participant = %q, want %q", de.Participant.ID, "B")
			}
		}
	}
	if destroyCount != 1 {
		t.Errorf("Expected 1 DestroyEvent, got %d", destroyCount)
	}
}

func TestCreateRender(t *testing.T) {
	input := "sequenceDiagram\nparticipant A\ncreate participant B\nA->>B: Create\nB->>A: Ack"
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	config := diagram.DefaultConfig()
	output, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	// Output should contain both participants
	if !strings.Contains(output, "A") {
		t.Error("Output should contain participant 'A'")
	}
	if !strings.Contains(output, "B") {
		t.Error("Output should contain participant 'B'")
	}
	// The 'B' box should appear inline (not at the top header level)
	// Verify output is valid
	if strings.TrimSpace(output) == "" {
		t.Error("Output should not be empty")
	}
}

func TestDestroyRender(t *testing.T) {
	input := "sequenceDiagram\nparticipant A\nparticipant B\nA->>B: Hello\ndestroy B\nA->>A: Solo"
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	config := diagram.DefaultConfig()
	output, err := Render(sd, config)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	// Should contain X marker for destroy
	if !strings.Contains(output, "X") {
		t.Error("Output should contain 'X' destroy marker")
	}
	if !strings.Contains(output, "Hello") {
		t.Error("Output should contain message 'Hello'")
	}
}

func TestCreateAndDestroy(t *testing.T) {
	input := "sequenceDiagram\nparticipant A\ncreate participant B\nA->>B: Create\nB->>A: Response\ndestroy B"
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
	if !strings.Contains(output, "Create") {
		t.Error("Output should contain 'Create' message")
	}
	if strings.TrimSpace(output) == "" {
		t.Error("Output should not be empty")
	}
}

func TestDestroyQuotedParticipant(t *testing.T) {
	input := "sequenceDiagram\nparticipant A\nparticipant \"My Service\"\nA->>\"My Service\": Hello\ndestroy \"My Service\""
	sd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	destroyCount := 0
	for _, elem := range sd.Elements {
		if de, ok := elem.(*DestroyEvent); ok {
			destroyCount++
			if de.Participant.ID != "My Service" {
				t.Errorf("DestroyEvent participant = %q, want %q", de.Participant.ID, "My Service")
			}
		}
	}
	if destroyCount != 1 {
		t.Errorf("Expected 1 DestroyEvent, got %d", destroyCount)
	}
}

func FuzzParseSequenceDiagram(f *testing.F) {
	f.Add("sequenceDiagram\nA->>B: Hello")
	f.Add("sequenceDiagram\nparticipant Alice\nAlice->>Bob: Hi")
	f.Add("sequenceDiagram\nA-->>B: Response")
	f.Add("sequenceDiagram\nA->>A: Self")

	f.Fuzz(func(t *testing.T, input string) {
		sd, err := Parse(input)
		if err != nil {
			return
		}

		for i, p := range sd.Participants {
			if p.Index != i {
				t.Errorf("Participant %q has incorrect index: got %d, expected %d", p.ID, p.Index, i)
			}
			if p.ID == "" {
				t.Errorf("Participant at index %d has empty ID", i)
			}
			if p.Label == "" {
				t.Errorf("Participant %q has empty label", p.ID)
			}
		}

		for i, msg := range sd.Messages {
			if msg.From == nil || msg.To == nil {
				t.Errorf("Message %d has nil participant", i)
			}
		}

		seen := make(map[string]bool)
		for _, p := range sd.Participants {
			if seen[p.ID] {
				t.Errorf("Duplicate participant ID: %q", p.ID)
			}
			seen[p.ID] = true
		}

		config := diagram.DefaultConfig()
		_, _ = Render(sd, config)
	})
}

func FuzzRenderSequenceDiagram(f *testing.F) {
	seeds := []string{
		"sequenceDiagram\nA->>B: Test",
		"sequenceDiagram\nA->>A: Self",
		"sequenceDiagram\nA->>B: 1\nB->>C: 2\nC->>A: 3",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		sd, err := Parse(input)
		if err != nil {
			return
		}

		for _, useAscii := range []bool{true, false} {
			config := diagram.DefaultConfig()
			config.UseAscii = useAscii

			output, err := Render(sd, config)
			if err != nil {
				return
			}

			if strings.TrimSpace(output) == "" {
				t.Error("Renderer produced empty output for valid diagram")
			}

			for _, p := range sd.Participants {
				if !strings.Contains(output, p.Label) {
					t.Errorf("Rendered output missing participant label: %q", p.Label)
				}
			}

			if !utf8.ValidString(output) {
				t.Error("Rendered output contains invalid UTF-8")
			}
		}
	})
}

func BenchmarkParse(b *testing.B) {
	tests := []struct {
		name         string
		participants int
		messages     int
	}{
		{"small_2p_5m", 2, 5},
		{"medium_5p_20m", 5, 20},
		{"large_10p_50m", 10, 50},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			input := generateDiagram(tt.participants, tt.messages)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := Parse(input)
				if err != nil {
					b.Fatalf("parse failed: %v", err)
				}
			}
		})
	}
}

func BenchmarkRender(b *testing.B) {
	tests := []struct {
		name         string
		participants int
		messages     int
	}{
		{"small_2p_5m", 2, 5},
		{"medium_5p_20m", 5, 20},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			input := generateDiagram(tt.participants, tt.messages)
			sd, err := Parse(input)
			if err != nil {
				b.Fatalf("parse failed: %v", err)
			}
			config := diagram.DefaultConfig()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, renderErr := Render(sd, config)
				if renderErr != nil {
					b.Fatalf("render error: %v", renderErr)
				}
			}
		})
	}
}

func generateDiagram(numParticipants, numMessages int) string {
	var sb strings.Builder
	sb.WriteString("sequenceDiagram\n")
	for i := 0; i < numParticipants; i++ {
		sb.WriteString("    participant P")
		sb.WriteString(string(rune('0' + i)))
		sb.WriteString("\n")
	}
	for i := 0; i < numMessages; i++ {
		from := i % numParticipants
		to := (i + 1) % numParticipants
		arrow := "-"
		if i%2 == 0 {
			arrow = "--"
		}
		sb.WriteString("    P")
		sb.WriteString(string(rune('0' + from)))
		sb.WriteString(arrow)
		sb.WriteString(">>P")
		sb.WriteString(string(rune('0' + to)))
		sb.WriteString(": Message\n")
	}
	return sb.String()
}
