package render

import (
	"strings"
	"testing"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

// TestDiagramTypes exercises the Detect → wrapper → Parse → Render path
// for every diagram type registered in render.go.
func TestDiagramTypes(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedType string
		wantSubstr   []string
	}{
		{
			name: "classDiagram",
			input: `classDiagram
    class Animal {
        +String name
        +eat()
    }
    class Dog {
        +bark()
    }
    Animal <|-- Dog`,
			expectedType: "classDiagram",
			wantSubstr:   []string{"Animal", "Dog"},
		},
		{
			name: "stateDiagram-v2",
			input: `stateDiagram-v2
    [*] --> Active
    Active --> Inactive
    Inactive --> [*]`,
			expectedType: "stateDiagram",
			wantSubstr:   []string{"Active", "Inactive"},
		},
		{
			name: "erDiagram",
			input: `erDiagram
    CUSTOMER ||--o{ ORDER : places
    ORDER ||--|{ LINE-ITEM : contains`,
			expectedType: "erDiagram",
			wantSubstr:   []string{"CUSTOMER", "ORDER"},
		},
		{
			name: "gantt",
			input: `gantt
    title A Gantt Chart
    dateFormat YYYY-MM-DD
    section Section
    Task1 :a1, 2024-01-01, 30d
    Task2 :after a1, 20d`,
			expectedType: "gantt",
			wantSubstr:   []string{"Task1", "Task2"},
		},
		{
			name: "pie",
			input: `pie
    title Pets
    "Dogs" : 40
    "Cats" : 30
    "Birds" : 20`,
			expectedType: "pie",
			wantSubstr:   []string{"Dogs", "Cats"},
		},
		{
			name: "mindmap",
			input: `mindmap
    root((Central))
        Topic1
            SubTopic1
        Topic2`,
			expectedType: "mindmap",
			wantSubstr:   []string{"Central"},
		},
		{
			name: "timeline",
			input: `timeline
    title Timeline of Events
    2024 : Event One
    2025 : Event Two`,
			expectedType: "timeline",
			wantSubstr:   []string{"Event"},
		},
		{
			name: "gitGraph",
			input: `gitGraph
    commit
    commit
    branch develop
    commit
    checkout main
    commit`,
			expectedType: "gitGraph",
		},
		{
			name: "journey",
			input: `journey
    title My Day
    section Morning
        Wake up: 5: Me
        Breakfast: 3: Me`,
			expectedType: "journey",
			wantSubstr:   []string{"Wake up"},
		},
		{
			name: "quadrantChart",
			input: `quadrantChart
    title Reach and Effort
    x-axis Low Reach --> High Reach
    y-axis Low Effort --> High Effort
    quadrant-1 Promote
    quadrant-2 Re-evaluate
    quadrant-3 Improve
    quadrant-4 May be improved
    Campaign A: [0.3, 0.6]
    Campaign B: [0.7, 0.4]`,
			expectedType: "quadrantChart",
			wantSubstr:   []string{"Campaign"},
		},
		{
			name: "xychart-beta",
			input: `xychart-beta
    title "Sales Revenue"
    x-axis [jan, feb, mar]
    y-axis "Revenue"
    bar [100, 200, 150]`,
			expectedType: "xychart-beta",
		},
		{
			name: "C4Context",
			input: `C4Context
    title System Context
    Person(user, "User", "A user")
    System(system, "System", "The system")
    Rel(user, system, "Uses")`,
			expectedType: "C4Context",
			wantSubstr:   []string{"User", "System"},
		},
		{
			name: "requirementDiagram",
			input: `requirementDiagram
    requirement test_req {
        id: 1
        text: the test shall do stuff
        risk: high
        verifymethod: test
    }
    element test_entity {
        type: simulation
    }
    test_entity - satisfies -> test_req`,
			expectedType: "requirementDiagram",
			wantSubstr:   []string{"test_req"},
		},
		{
			name: "block-beta",
			input: `block-beta
    columns 2
    A["Block A"]
    B["Block B"]`,
			expectedType: "block-beta",
			wantSubstr:   []string{"Block"},
		},
		{
			name: "sankey-beta",
			input: `sankey-beta
Source1,Target1,100
Source2,Target2,200`,
			expectedType: "sankey-beta",
			wantSubstr:   []string{"Source"},
		},
		{
			name: "packet-beta",
			input: `packet-beta
    0-15: "Source Port"
    16-31: "Destination Port"`,
			expectedType: "packet-beta",
			wantSubstr:   []string{"Port"},
		},
		{
			name: "kanban",
			input: `kanban
    Todo
        task1[Task One]
    InProgress
        task2[Task Two]`,
			expectedType: "kanban",
			wantSubstr:   []string{"Task"},
		},
		{
			name: "architecture-beta",
			input: `architecture-beta
    service api(Server)[API]
    service db(Database)[DB]
    api:R --> L:db`,
			expectedType: "architecture-beta",
		},
		{
			name: "zenuml",
			input: `zenuml
Client client
Server server
server.process(data)`,
			expectedType: "zenuml",
			wantSubstr:   []string{"client", "server"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Step 1: Detect should detect the correct type
			diag, err := Detect(tt.input)
			if err != nil {
				t.Fatalf("Detect error: %v", err)
			}

			// Step 2: Verify Type() returns the expected string
			if diag.Type() != tt.expectedType {
				t.Errorf("Type() = %q, want %q", diag.Type(), tt.expectedType)
			}

			// Step 3: Parse the input
			if err := diag.Parse(tt.input); err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			// Step 4: Render with config and verify non-empty output
			config := diagram.NewTestConfig(true, "cli")
			output, err := diag.Render(config)
			if err != nil {
				t.Fatalf("Render error: %v", err)
			}

			if len(strings.TrimSpace(output)) == 0 {
				t.Error("Render produced empty output")
			}

			// Check expected substrings
			for _, want := range tt.wantSubstr {
				if !strings.Contains(output, want) {
					t.Errorf("Output missing expected substring %q\nOutput:\n%s", want, output)
				}
			}
		})
	}
}
