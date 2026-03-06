package kanban

import (
	"strings"
	"testing"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

func TestIsKanbanBoard(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"kanban\nTodo\n  Task 1", true},
		{"%% comment\nkanban", true},
		{"graph LR\n  A-->B", false},
		{"", false},
	}
	for _, tt := range tests {
		got := IsKanbanBoard(tt.input)
		if got != tt.want {
			t.Errorf("IsKanbanBoard(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParse(t *testing.T) {
	input := `kanban
Todo
    Task 1
    Task 2
In Progress
    Task 3
Done
    Task 4
    Task 5`

	kb, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(kb.Columns) != 3 {
		t.Fatalf("Columns count = %d, want 3", len(kb.Columns))
	}

	if kb.Columns[0].Name != "Todo" {
		t.Errorf("Column 0 name = %q, want %q", kb.Columns[0].Name, "Todo")
	}

	if len(kb.Columns[0].Cards) != 2 {
		t.Errorf("Column 0 cards count = %d, want 2", len(kb.Columns[0].Cards))
	}
}

func TestRender(t *testing.T) {
	input := `kanban
Todo
    Task 1
    Task 2
Done
    Task 3`

	kb, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	output, err := Render(kb, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(output, "Todo") {
		t.Error("Output should contain column name")
	}
	if !strings.Contains(output, "Task 1") {
		t.Error("Output should contain card title")
	}
}

func TestRenderASCII(t *testing.T) {
	input := `kanban
Col
    Card`

	kb, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	config.UseAscii = true
	output, err := Render(kb, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if strings.Contains(output, "│") || strings.Contains(output, "─") {
		t.Error("ASCII output should not contain Unicode characters")
	}
}

func TestParseMultipleColumns(t *testing.T) {
	input := `kanban
Backlog
    Item A
    Item B
    Item C
In Progress
    Item D
Review
    Item E
    Item F
Done
    Item G`

	kb, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(kb.Columns) != 4 {
		t.Fatalf("Columns count = %d, want 4", len(kb.Columns))
	}

	expected := []struct {
		name      string
		cardCount int
	}{
		{"Backlog", 3},
		{"In Progress", 1},
		{"Review", 2},
		{"Done", 1},
	}

	for i, exp := range expected {
		if kb.Columns[i].Name != exp.name {
			t.Errorf("Column %d name = %q, want %q", i, kb.Columns[i].Name, exp.name)
		}
		if len(kb.Columns[i].Cards) != exp.cardCount {
			t.Errorf("Column %d cards = %d, want %d", i, len(kb.Columns[i].Cards), exp.cardCount)
		}
	}
}

func TestParseEmptyColumn(t *testing.T) {
	input := `kanban
Todo
    Task 1
Empty Column
Done
    Task 2`

	kb, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(kb.Columns) != 3 {
		t.Fatalf("Columns count = %d, want 3", len(kb.Columns))
	}

	if kb.Columns[1].Name != "Empty Column" {
		t.Errorf("Column 1 name = %q, want %q", kb.Columns[1].Name, "Empty Column")
	}
	if len(kb.Columns[1].Cards) != 0 {
		t.Errorf("Column 1 cards = %d, want 0", len(kb.Columns[1].Cards))
	}
}

func TestParseItemsWithMetadata(t *testing.T) {
	input := `kanban
Todo
    Task with details @user1
    Another task #tag1 #tag2
    Simple task`

	kb, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(kb.Columns[0].Cards) != 3 {
		t.Fatalf("Cards count = %d, want 3", len(kb.Columns[0].Cards))
	}

	// Cards should preserve the full text including metadata
	if kb.Columns[0].Cards[0].Title != "Task with details @user1" {
		t.Errorf("Card 0 title = %q, want %q", kb.Columns[0].Cards[0].Title, "Task with details @user1")
	}
	if kb.Columns[0].Cards[1].Title != "Another task #tag1 #tag2" {
		t.Errorf("Card 1 title = %q, want %q", kb.Columns[0].Cards[1].Title, "Another task #tag1 #tag2")
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty input", ""},
		{"wrong keyword", "flowchart\n  A-->B"},
		{"no columns", "kanban"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			if err == nil {
				t.Errorf("Parse(%q) should return error", tt.input)
			}
		})
	}
}

func TestParseCardWithoutColumn(t *testing.T) {
	// Cards before any column header should create a "Default" column
	input := `kanban
    Orphan card`

	kb, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(kb.Columns) != 1 {
		t.Fatalf("Columns count = %d, want 1", len(kb.Columns))
	}
	if kb.Columns[0].Name != "Default" {
		t.Errorf("Column name = %q, want %q", kb.Columns[0].Name, "Default")
	}
	if len(kb.Columns[0].Cards) != 1 {
		t.Errorf("Cards count = %d, want 1", len(kb.Columns[0].Cards))
	}
}

func TestRenderNilBoard(t *testing.T) {
	_, err := Render(nil, nil)
	if err == nil {
		t.Error("Render(nil, nil) should return error")
	}
}

func TestRenderEmptyColumns(t *testing.T) {
	kb := &KanbanBoard{Columns: []*Column{}}
	_, err := Render(kb, nil)
	if err == nil {
		t.Error("Render with empty columns should return error")
	}
}

func TestRenderNilConfig(t *testing.T) {
	input := `kanban
Col
    Card`

	kb, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	output, err := Render(kb, nil)
	if err != nil {
		t.Fatalf("Render with nil config should succeed, got: %v", err)
	}
	if !strings.Contains(output, "Card") {
		t.Error("Output should contain card title")
	}
}

func TestRenderMultipleColumnsOutput(t *testing.T) {
	input := `kanban
Todo
    Task A
    Task B
Done
    Task C`

	kb, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	output, err := Render(kb, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(output, "Todo") {
		t.Error("Output should contain 'Todo'")
	}
	if !strings.Contains(output, "Done") {
		t.Error("Output should contain 'Done'")
	}
	if !strings.Contains(output, "Task A") {
		t.Error("Output should contain 'Task A'")
	}
	if !strings.Contains(output, "Task C") {
		t.Error("Output should contain 'Task C'")
	}
}
