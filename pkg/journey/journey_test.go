package journey

import (
	"strings"
	"testing"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

func TestIsJourneyDiagram(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"journey\n  title Test", true},
		{"%% comment\njourney", true},
		{"graph LR\n  A-->B", false},
		{"", false},
	}
	for _, tt := range tests {
		got := IsJourneyDiagram(tt.input)
		if got != tt.want {
			t.Errorf("IsJourneyDiagram(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParse(t *testing.T) {
	input := `journey
    title My working day
    section Go to work
    Make tea: 5: Me
    Go upstairs: 3: Me, Cat
    section Go home
    Go downstairs: 5: Me`

	jd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if jd.Title != "My working day" {
		t.Errorf("Title = %q, want %q", jd.Title, "My working day")
	}

	if len(jd.Sections) != 2 {
		t.Errorf("Sections count = %d, want 2", len(jd.Sections))
	}

	if len(jd.Sections[0].Tasks) != 2 {
		t.Fatalf("Section 0 tasks count = %d, want 2", len(jd.Sections[0].Tasks))
	}

	task := jd.Sections[0].Tasks[0]
	if task.Name != "Make tea" {
		t.Errorf("Task name = %q, want %q", task.Name, "Make tea")
	}
	if task.Score != 5 {
		t.Errorf("Task score = %d, want 5", task.Score)
	}
	if len(task.Actors) != 1 || task.Actors[0] != "Me" {
		t.Errorf("Task actors = %v, want [Me]", task.Actors)
	}
}

func TestRender(t *testing.T) {
	input := `journey
    title My Day
    section Morning
    Wake up: 1: Me
    Eat breakfast: 4: Me
    section Work
    Code: 5: Me`

	jd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	output, err := Render(jd, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(output, "My Day") {
		t.Error("Output should contain title")
	}
	if !strings.Contains(output, "Wake up") {
		t.Error("Output should contain task name")
	}
	if !strings.Contains(output, "Morning") {
		t.Error("Output should contain section name")
	}
}

func TestRenderASCII(t *testing.T) {
	input := `journey
    section Test
    Task: 3: Actor`

	jd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	config.UseAscii = true
	output, err := Render(jd, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if strings.Contains(output, "│") || strings.Contains(output, "─") {
		t.Error("ASCII output should not contain Unicode characters")
	}
}

func TestScoreToFaceAllValues(t *testing.T) {
	tests := []struct {
		score int
		want  string
	}{
		{1, "😞"},
		{2, "🙁"},
		{3, "😐"},
		{4, "🙂"},
		{5, "😊"},
	}

	for _, tt := range tests {
		got := scoreToFace(tt.score)
		if got != tt.want {
			t.Errorf("scoreToFace(%d) = %q, want %q", tt.score, got, tt.want)
		}
	}
}

func TestScoreToFaceEdgeCases(t *testing.T) {
	// Score 0 and negative should return sad face
	if got := scoreToFace(0); got != "😞" {
		t.Errorf("scoreToFace(0) = %q, want %q", got, "😞")
	}
	if got := scoreToFace(-1); got != "😞" {
		t.Errorf("scoreToFace(-1) = %q, want %q", got, "😞")
	}
	// Score above 5 should return happy face
	if got := scoreToFace(10); got != "😊" {
		t.Errorf("scoreToFace(10) = %q, want %q", got, "😊")
	}
}

func TestRenderAllScores(t *testing.T) {
	input := `journey
    title Score Test
    section Scores
    Very Bad: 1: User
    Bad: 2: User
    Neutral: 3: User
    Good: 4: User
    Great: 5: User`

	jd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	output, err := Render(jd, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// All faces should appear
	faces := []string{"😞", "🙁", "😐", "🙂", "😊"}
	for _, face := range faces {
		if !strings.Contains(output, face) {
			t.Errorf("Output should contain face %s", face)
		}
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		s     string
		width int
		want  string
	}{
		{"abc", 6, "abc   "},
		{"hello", 5, "hello"},
		{"toolong", 4, "tool"},
		{"", 3, "   "},
		{"x", 1, "x"},
	}
	for _, tt := range tests {
		got := padRight(tt.s, tt.width)
		if got != tt.want {
			t.Errorf("padRight(%q, %d) = %q, want %q", tt.s, tt.width, got, tt.want)
		}
	}
}

func TestTaskWithVeryLongName(t *testing.T) {
	longName := strings.Repeat("A", 50)
	input := `journey
    section Test
    ` + longName + `: 3: User`

	jd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	output, err := Render(jd, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Output should render without error and contain score
	if !strings.Contains(output, "3") {
		t.Error("Output should contain the task score")
	}
}

func TestParseMultipleSections(t *testing.T) {
	input := `journey
    title Multi-Section Journey
    section Morning
    Wake up: 2: Me
    Coffee: 5: Me
    section Afternoon
    Lunch: 4: Me, Coworker
    Meetings: 1: Me
    section Evening
    Dinner: 5: Me, Family
    Sleep: 4: Me`

	jd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if jd.Title != "Multi-Section Journey" {
		t.Errorf("Title = %q, want %q", jd.Title, "Multi-Section Journey")
	}
	if len(jd.Sections) != 3 {
		t.Fatalf("Sections count = %d, want 3", len(jd.Sections))
	}

	sectionNames := []string{"Morning", "Afternoon", "Evening"}
	sectionTaskCounts := []int{2, 2, 2}
	for i, sec := range jd.Sections {
		if sec.Name != sectionNames[i] {
			t.Errorf("Section %d name = %q, want %q", i, sec.Name, sectionNames[i])
		}
		if len(sec.Tasks) != sectionTaskCounts[i] {
			t.Errorf("Section %d tasks = %d, want %d", i, len(sec.Tasks), sectionTaskCounts[i])
		}
	}

	// Check actors with multiple entries
	lunchTask := jd.Sections[1].Tasks[0]
	if len(lunchTask.Actors) != 2 {
		t.Errorf("Lunch actors count = %d, want 2", len(lunchTask.Actors))
	}
}

func TestRenderNilDiagram(t *testing.T) {
	_, err := Render(nil, nil)
	if err == nil {
		t.Error("Render(nil, nil) should return error")
	}
}

func TestRenderEmptySections(t *testing.T) {
	jd := &JourneyDiagram{Sections: []*JourneySection{}}
	_, err := Render(jd, nil)
	if err == nil {
		t.Error("Render with empty sections should return error")
	}
}

func TestRenderNilConfig(t *testing.T) {
	input := `journey
    section Test
    Task: 3: Actor`

	jd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	output, err := Render(jd, nil)
	if err != nil {
		t.Fatalf("Render with nil config should succeed, got: %v", err)
	}
	if !strings.Contains(output, "Task") {
		t.Error("Output should contain task name")
	}
}

func TestParseTaskWithoutSection(t *testing.T) {
	input := `journey
    title No Section
    Orphan task: 3: Actor`

	jd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(jd.Sections) != 1 {
		t.Fatalf("Sections count = %d, want 1 (default)", len(jd.Sections))
	}
	if jd.Sections[0].Name != "" {
		t.Errorf("Default section name = %q, want empty", jd.Sections[0].Name)
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty input", ""},
		{"wrong keyword", "gantt\n  title Test"},
		{"no tasks", "journey\n  title Empty"},
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

func TestParseTaskWithoutActors(t *testing.T) {
	input := `journey
    section Test
    Solo task: 4`

	jd, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	task := jd.Sections[0].Tasks[0]
	if task.Score != 4 {
		t.Errorf("Score = %d, want 4", task.Score)
	}
	if len(task.Actors) != 0 {
		t.Errorf("Actors count = %d, want 0", len(task.Actors))
	}
}
