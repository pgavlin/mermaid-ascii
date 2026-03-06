package architecture

import (
	"strings"
	"testing"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

func TestIsArchitectureDiagram(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid", "architecture-beta", true},
		{"with comment", "%% comment\narchitecture-beta", true},
		{"not arch", "graph TD", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsArchitectureDiagram(tt.input); got != tt.want {
				t.Errorf("IsArchitectureDiagram() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseServices(t *testing.T) {
	input := `architecture-beta
service api(server)[API Server]
service db(database)[Database]
api --> db
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(d.Services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(d.Services))
	}
	if d.Services[0].Label != "API Server" {
		t.Errorf("expected label 'API Server', got %q", d.Services[0].Label)
	}
	if d.Services[0].Icon != "server" {
		t.Errorf("expected icon 'server', got %q", d.Services[0].Icon)
	}
	if len(d.Connections) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(d.Connections))
	}
	if !d.Connections[0].Directed {
		t.Error("expected directed connection")
	}
}

func TestParseGroup(t *testing.T) {
	input := `architecture-beta
group cloud(cloud)[Cloud] {
  service api(server)[API]
  service worker(server)[Worker]
}
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(d.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(d.Groups))
	}
	if d.Groups[0].Label != "Cloud" {
		t.Errorf("expected group label 'Cloud', got %q", d.Groups[0].Label)
	}
	if len(d.Groups[0].Services) != 2 {
		t.Fatalf("expected 2 services in group, got %d", len(d.Groups[0].Services))
	}
}

func TestParseUndirectedConnection(t *testing.T) {
	input := `architecture-beta
service a[A]
service b[B]
a -- b
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(d.Connections) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(d.Connections))
	}
	if d.Connections[0].Directed {
		t.Error("expected undirected connection")
	}
}

func TestRender(t *testing.T) {
	input := `architecture-beta
service api(server)[API Server]
service db(database)[Database]
api --> db
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
	if !strings.Contains(result, "API Server") {
		t.Error("expected output to contain 'API Server'")
	}
	if !strings.Contains(result, "Database") {
		t.Error("expected output to contain 'Database'")
	}
}

func TestRenderASCII(t *testing.T) {
	input := `architecture-beta
service api[API]
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

func TestRenderGroupWithServices(t *testing.T) {
	input := `architecture-beta
group cloud(cloud)[Cloud] {
  service api(server)[API]
  service worker(server)[Worker]
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
	if !strings.Contains(result, "Cloud") {
		t.Error("expected output to contain 'Cloud'")
	}
	if !strings.Contains(result, "API") {
		t.Error("expected output to contain 'API'")
	}
	if !strings.Contains(result, "Worker") {
		t.Error("expected output to contain 'Worker'")
	}
	// Group should use box characters
	if !strings.Contains(result, "┌") {
		t.Error("expected unicode box characters in group rendering")
	}
}

func TestRenderGroupASCII(t *testing.T) {
	input := `architecture-beta
group cloud(cloud)[Cloud] {
  service api(server)[API]
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
		t.Error("expected ASCII box characters in group rendering")
	}
	if !strings.Contains(result, "Cloud") {
		t.Error("expected output to contain 'Cloud'")
	}
}

func TestRenderNestedGroups(t *testing.T) {
	input := `architecture-beta
group outer(cloud)[Outer] {
  group inner(server)[Inner] {
    service api(server)[API]
  }
}
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(d.Groups) != 1 {
		t.Fatalf("expected 1 top-level group, got %d", len(d.Groups))
	}
	if len(d.Groups[0].Groups) != 1 {
		t.Fatalf("expected 1 nested group, got %d", len(d.Groups[0].Groups))
	}
	if d.Groups[0].Groups[0].Label != "Inner" {
		t.Errorf("expected nested group label 'Inner', got %q", d.Groups[0].Groups[0].Label)
	}

	config := diagram.NewTestConfig(false, "cli")
	result, err := Render(d, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	if !strings.Contains(result, "Outer") {
		t.Error("expected output to contain 'Outer'")
	}
	if !strings.Contains(result, "Inner") {
		t.Error("expected output to contain 'Inner'")
	}
	if !strings.Contains(result, "API") {
		t.Error("expected output to contain 'API'")
	}
}

func TestParseJunction(t *testing.T) {
	input := `architecture-beta
service api(server)[API]
junction junc1
service db(database)[DB]
api --> junc1
junc1 --> db
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	// Junction is parsed as a service with ID as label
	if len(d.Services) != 3 {
		t.Fatalf("expected 3 services (including junction), got %d", len(d.Services))
	}
	junc := d.Services[1]
	if junc.ID != "junc1" {
		t.Errorf("expected junction ID 'junc1', got %q", junc.ID)
	}
	if junc.Label != "junc1" {
		t.Errorf("expected junction label 'junc1' (same as ID), got %q", junc.Label)
	}
	if len(d.Connections) != 2 {
		t.Fatalf("expected 2 connections, got %d", len(d.Connections))
	}

	config := diagram.NewTestConfig(false, "cli")
	result, err := Render(d, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	if !strings.Contains(result, "junc1") {
		t.Error("expected output to contain 'junc1'")
	}
}

func TestParseJunctionInGroup(t *testing.T) {
	input := `architecture-beta
group cloud(cloud)[Cloud] {
  junction junc1
  service api(server)[API]
}
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(d.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(d.Groups))
	}
	if len(d.Groups[0].Services) != 2 {
		t.Fatalf("expected 2 services in group (including junction), got %d", len(d.Groups[0].Services))
	}
	if d.Groups[0].Services[0].ID != "junc1" {
		t.Errorf("expected first service to be junction 'junc1', got %q", d.Groups[0].Services[0].ID)
	}
}

func TestParseServiceWithoutLabel(t *testing.T) {
	input := `architecture-beta
service myservice
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(d.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(d.Services))
	}
	if d.Services[0].Label != "myservice" {
		t.Errorf("expected label to default to ID 'myservice', got %q", d.Services[0].Label)
	}
	if d.Services[0].Icon != "" {
		t.Errorf("expected empty icon, got %q", d.Services[0].Icon)
	}
}

func TestParseServiceWithIconNoLabel(t *testing.T) {
	input := `architecture-beta
service api(server)
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(d.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(d.Services))
	}
	if d.Services[0].Label != "api" {
		t.Errorf("expected label to default to ID 'api', got %q", d.Services[0].Label)
	}
	if d.Services[0].Icon != "server" {
		t.Errorf("expected icon 'server', got %q", d.Services[0].Icon)
	}
}

func TestParseConnectionWithEdgePositions(t *testing.T) {
	input := `architecture-beta
service api(server)[API]
service db(database)[DB]
api:R --> db:L
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(d.Connections) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(d.Connections))
	}
	conn := d.Connections[0]
	if conn.From != "api" {
		t.Errorf("expected From 'api', got %q", conn.From)
	}
	if conn.FromEdge != "R" {
		t.Errorf("expected FromEdge 'R', got %q", conn.FromEdge)
	}
	if conn.To != "db" {
		t.Errorf("expected To 'db', got %q", conn.To)
	}
	if conn.ToEdge != "L" {
		t.Errorf("expected ToEdge 'L', got %q", conn.ToEdge)
	}
	if !conn.Directed {
		t.Error("expected directed connection")
	}
}

func TestParseUndirectedConnectionWithEdgePositions(t *testing.T) {
	input := `architecture-beta
service a[A]
service b[B]
a:T -- b:B
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(d.Connections) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(d.Connections))
	}
	conn := d.Connections[0]
	if conn.FromEdge != "T" {
		t.Errorf("expected FromEdge 'T', got %q", conn.FromEdge)
	}
	if conn.ToEdge != "B" {
		t.Errorf("expected ToEdge 'B', got %q", conn.ToEdge)
	}
	if conn.Directed {
		t.Error("expected undirected connection")
	}
}

func TestRenderNilConfig(t *testing.T) {
	input := `architecture-beta
service api(server)[API]
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	result, err := Render(d, nil)
	if err != nil {
		t.Fatalf("Render() with nil config error: %v", err)
	}
	if !strings.Contains(result, "API") {
		t.Error("expected output to contain 'API'")
	}
}

func TestRenderNilDiagram(t *testing.T) {
	config := diagram.NewTestConfig(false, "cli")
	_, err := Render(nil, config)
	if err == nil {
		t.Error("expected error for nil diagram")
	}
}

func TestRenderUndirectedConnection(t *testing.T) {
	input := `architecture-beta
service a[A]
service b[B]
a -- b
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
	if !strings.Contains(result, "───") {
		t.Error("expected unicode undirected line '───' in output")
	}
	if strings.Contains(result, "──>") {
		t.Error("did not expect directed arrow in undirected connection")
	}
}

func TestRenderUndirectedConnectionASCII(t *testing.T) {
	input := `architecture-beta
service a[A]
service b[B]
a -- b
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
	if !strings.Contains(result, "---") {
		t.Error("expected ASCII undirected line '---' in output")
	}
}

func TestRenderServicesNoConnections(t *testing.T) {
	input := `architecture-beta
service api(server)[API]
service db(database)[DB]
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
	if !strings.Contains(result, "API") {
		t.Error("expected output to contain 'API'")
	}
	if !strings.Contains(result, "DB") {
		t.Error("expected output to contain 'DB'")
	}
}

func TestRenderEmptyDiagram(t *testing.T) {
	d := &ArchitectureDiagram{}
	config := diagram.NewTestConfig(false, "cli")
	result, err := Render(d, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	// An empty diagram should produce minimal output
	trimmed := strings.TrimSpace(result)
	if trimmed != "" {
		t.Errorf("expected empty output for empty diagram, got %q", result)
	}
}

func TestParseNoContent(t *testing.T) {
	input := `%% just a comment
%% another comment
`
	_, err := Parse(input)
	if err == nil {
		t.Error("expected error for input with only comments")
	}
}

func TestParseInvalidKeyword(t *testing.T) {
	input := `graph TD
service api[API]
`
	_, err := Parse(input)
	if err == nil {
		t.Error("expected error for invalid keyword")
	}
	if !strings.Contains(err.Error(), "architecture-beta") {
		t.Errorf("expected error to mention 'architecture-beta', got: %v", err)
	}
}

func TestParseGroupWithoutLabel(t *testing.T) {
	input := `architecture-beta
group mygroup {
  service api[API]
}
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(d.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(d.Groups))
	}
	if d.Groups[0].Label != "mygroup" {
		t.Errorf("expected group label to default to ID 'mygroup', got %q", d.Groups[0].Label)
	}
}

func TestParseUnrecognizedLine(t *testing.T) {
	input := `architecture-beta
service api[API]
this is not a valid line
service db[DB]
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	// Unrecognized lines are skipped
	if len(d.Services) != 2 {
		t.Fatalf("expected 2 services (unrecognized line skipped), got %d", len(d.Services))
	}
}

func TestRenderGroupWithConnectionsBetweenServices(t *testing.T) {
	input := `architecture-beta
group cloud(cloud)[Cloud] {
  service api(server)[API]
  service db(database)[DB]
}
api:R --> db:L
`
	d, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(d.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(d.Groups))
	}
	if len(d.Connections) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(d.Connections))
	}

	config := diagram.NewTestConfig(false, "cli")
	result, err := Render(d, config)
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	if !strings.Contains(result, "Cloud") {
		t.Error("expected output to contain 'Cloud'")
	}
	if !strings.Contains(result, "api") {
		t.Error("expected connection line with 'api'")
	}
	if !strings.Contains(result, "──>") {
		t.Error("expected directed arrow in connection")
	}
}
