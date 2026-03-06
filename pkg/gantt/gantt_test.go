package gantt

import (
	"strings"
	"testing"
	"time"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

func TestParseTaskAfterDependency(t *testing.T) {
	input := `gantt
    dateFormat YYYY-MM-DD
    section Build
    Task A :a1, 2024-01-01, 3d
    Task B :b1, after a1, 2d
    Task C :c1, after b1, 1d`

	gd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(gd.Tasks) != 3 {
		t.Fatalf("Tasks count = %d, want 3", len(gd.Tasks))
	}

	taskA := gd.Tasks[0]
	taskB := gd.Tasks[1]
	taskC := gd.Tasks[2]

	if taskB.After != "a1" {
		t.Errorf("Task B After = %q, want %q", taskB.After, "a1")
	}
	if !taskB.StartDate.Equal(taskA.EndDate) {
		t.Errorf("Task B start = %v, want %v (end of Task A)", taskB.StartDate, taskA.EndDate)
	}

	if taskC.After != "b1" {
		t.Errorf("Task C After = %q, want %q", taskC.After, "b1")
	}
	if !taskC.StartDate.Equal(taskB.EndDate) {
		t.Errorf("Task C start = %v, want %v (end of Task B)", taskC.StartDate, taskB.EndDate)
	}
}

func TestParseTaskStatusMarkers(t *testing.T) {
	input := `gantt
    dateFormat YYYY-MM-DD
    Active Task  :active, a1, 2024-01-01, 3d
    Done Task    :done, d1, 2024-01-01, 2d
    Crit Task    :crit, c1, 2024-01-01, 4d
    Crit Done    :crit, done, cd1, 2024-01-01, 1d`

	gd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(gd.Tasks) != 4 {
		t.Fatalf("Tasks count = %d, want 4", len(gd.Tasks))
	}

	if gd.Tasks[0].Status != "active" {
		t.Errorf("Task 0 status = %q, want %q", gd.Tasks[0].Status, "active")
	}
	if gd.Tasks[0].ID != "a1" {
		t.Errorf("Task 0 ID = %q, want %q", gd.Tasks[0].ID, "a1")
	}
	if gd.Tasks[1].Status != "done" {
		t.Errorf("Task 1 status = %q, want %q", gd.Tasks[1].Status, "done")
	}
	if gd.Tasks[2].Status != "crit" {
		t.Errorf("Task 2 status = %q, want %q", gd.Tasks[2].Status, "crit")
	}
	if gd.Tasks[3].Status != "crit,done" {
		t.Errorf("Task 3 status = %q, want %q", gd.Tasks[3].Status, "crit,done")
	}
}

func TestParseDurationFormats(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantDur time.Duration
	}{
		{"days", "2d", 2 * 24 * time.Hour},
		{"weeks", "1w", 7 * 24 * time.Hour},
		{"hours", "3h", 3 * time.Hour},
		{"minutes", "30m", 30 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dur, err := parseDuration(tt.input)
			if err != nil {
				t.Fatalf("parseDuration(%q) error: %v", tt.input, err)
			}
			if dur != tt.wantDur {
				t.Errorf("parseDuration(%q) = %v, want %v", tt.input, dur, tt.wantDur)
			}
		})
	}
}

func TestParseDurationInvalid(t *testing.T) {
	invalids := []string{"abc", "5x", "", "d3"}
	for _, s := range invalids {
		_, err := parseDuration(s)
		if err == nil {
			t.Errorf("parseDuration(%q) should fail", s)
		}
	}
}

func TestParseTaskExplicitStartDate(t *testing.T) {
	input := `gantt
    dateFormat YYYY-MM-DD
    Task 1 :t1, 2024-06-15, 5d`

	gd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	task := gd.Tasks[0]
	expectedStart := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	if !task.StartDate.Equal(expectedStart) {
		t.Errorf("StartDate = %v, want %v", task.StartDate, expectedStart)
	}
	expectedEnd := expectedStart.Add(5 * 24 * time.Hour)
	if !task.EndDate.Equal(expectedEnd) {
		t.Errorf("EndDate = %v, want %v", task.EndDate, expectedEnd)
	}
	if task.ID != "t1" {
		t.Errorf("ID = %q, want %q", task.ID, "t1")
	}
}

func TestRenderNilDiagram(t *testing.T) {
	_, err := Render(nil, nil)
	if err == nil {
		t.Error("Render(nil, nil) should return error")
	}
}

func TestRenderEmptyTasks(t *testing.T) {
	gd := &GanttDiagram{Tasks: []*Task{}}
	_, err := Render(gd, nil)
	if err == nil {
		t.Error("Render with empty tasks should return error")
	}
}

func TestRenderNilConfig(t *testing.T) {
	input := `gantt
    dateFormat YYYY-MM-DD
    Task 1 :2024-01-01, 3d`

	gd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	output, err := Render(gd, nil)
	if err != nil {
		t.Fatalf("Render with nil config should succeed, got: %v", err)
	}
	if !strings.Contains(output, "Task 1") {
		t.Error("Output should contain task name")
	}
}

func TestMinInt(t *testing.T) {
	tests := []struct{ a, b, want int }{
		{1, 2, 1}, {5, 3, 3}, {4, 4, 4}, {-1, 0, -1},
	}
	for _, tt := range tests {
		if got := minInt(tt.a, tt.b); got != tt.want {
			t.Errorf("minInt(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestMaxInt(t *testing.T) {
	tests := []struct{ a, b, want int }{
		{1, 2, 2}, {5, 3, 5}, {4, 4, 4}, {-1, 0, 0},
	}
	for _, tt := range tests {
		if got := maxInt(tt.a, tt.b); got != tt.want {
			t.Errorf("maxInt(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{3 * 24 * time.Hour, "3d"},
		{7 * 24 * time.Hour, "7d"},
		{2 * time.Hour, "2h0m0s"},
		{30 * time.Minute, "30m0s"},
	}
	for _, tt := range tests {
		if got := formatDuration(tt.d); got != tt.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestParseExcludes(t *testing.T) {
	input := `gantt
    dateFormat YYYY-MM-DD
    excludes weekends
    Task 1 :2024-01-01, 3d`

	gd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(gd.Tasks) != 1 {
		t.Errorf("Tasks count = %d, want 1", len(gd.Tasks))
	}
}

func TestParseWrongKeyword(t *testing.T) {
	_, err := Parse("flowchart\n  A-->B")
	if err == nil {
		t.Error("Expected error for wrong keyword")
	}
}

func TestParseNoContent(t *testing.T) {
	_, err := Parse("%% only comments")
	if err == nil {
		t.Error("Expected error for no content")
	}
}

func TestIsGanttDiagram(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"gantt\n  title Test", true},
		{"%% comment\ngantt", true},
		{"graph LR\n  A-->B", false},
		{"sequenceDiagram\n  A->>B: hi", false},
		{"", false},
	}
	for _, tt := range tests {
		got := IsGanttDiagram(tt.input)
		if got != tt.want {
			t.Errorf("IsGanttDiagram(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParse(t *testing.T) {
	input := `gantt
    title A Gantt Diagram
    dateFormat YYYY-MM-DD
    section Section 1
    Task 1           :a1, 2024-01-01, 3d
    Task 2           :after a1, 2d
    section Section 2
    Task 3           :2024-01-02, 5d`

	gd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if gd.Title != "A Gantt Diagram" {
		t.Errorf("Title = %q, want %q", gd.Title, "A Gantt Diagram")
	}

	if len(gd.Sections) != 2 {
		t.Errorf("Sections count = %d, want 2", len(gd.Sections))
	}

	if len(gd.Tasks) != 3 {
		t.Errorf("Tasks count = %d, want 3", len(gd.Tasks))
	}

	// Task 2 should start after Task 1 ends
	if gd.Tasks[1].After != "a1" {
		t.Errorf("Task 2 After = %q, want %q", gd.Tasks[1].After, "a1")
	}
}

func TestParseEmptyInput(t *testing.T) {
	_, err := Parse("")
	if err == nil {
		t.Error("Expected error for empty input")
	}
}

func TestParseNoTasks(t *testing.T) {
	_, err := Parse("gantt\n  title Empty")
	if err == nil {
		t.Error("Expected error for no tasks")
	}
}

func TestRender(t *testing.T) {
	input := `gantt
    title Build Process
    dateFormat YYYY-MM-DD
    section Backend
    API Development  :a1, 2024-01-01, 5d
    Database Setup   :a2, 2024-01-03, 3d
    section Frontend
    UI Design        :b1, 2024-01-01, 4d`

	gd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	output, err := Render(gd, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Check that output contains key elements
	if !strings.Contains(output, "Build Process") {
		t.Error("Output should contain title")
	}
	if !strings.Contains(output, "Backend") {
		t.Error("Output should contain section name")
	}
	if !strings.Contains(output, "API Development") {
		t.Error("Output should contain task name")
	}
	if !strings.Contains(output, "█") {
		t.Error("Unicode output should contain bar fill character")
	}
}

func TestRenderASCII(t *testing.T) {
	input := `gantt
    dateFormat YYYY-MM-DD
    Task 1 :2024-01-01, 3d`

	gd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	config.UseAscii = true
	output, err := Render(gd, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if strings.Contains(output, "█") {
		t.Error("ASCII output should not contain Unicode bar characters")
	}
	if !strings.Contains(output, "#") {
		t.Error("ASCII output should contain # bar fill character")
	}
}

func TestRenderActiveAndCritTasks(t *testing.T) {
	input := `gantt
    dateFormat YYYY-MM-DD
    Active Task  :active, 2024-01-01, 3d
    Critical Task :crit, 2024-01-01, 3d
    Done Task    :done, 2024-01-01, 3d`

	gd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	output, err := Render(gd, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(output, "▓") {
		t.Error("Output should contain active bar character")
	}
	if !strings.Contains(output, "▒") {
		t.Error("Output should contain critical bar character")
	}
}
