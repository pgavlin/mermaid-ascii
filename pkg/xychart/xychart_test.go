package xychart

import (
	"strings"
	"testing"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

func TestIsXYChart(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"xychart-beta\n  bar [1,2,3]", true},
		{"%% comment\nxychart-beta", true},
		{"graph LR\n  A-->B", false},
		{"", false},
	}
	for _, tt := range tests {
		got := IsXYChart(tt.input)
		if got != tt.want {
			t.Errorf("IsXYChart(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestParse(t *testing.T) {
	input := `xychart-beta
    title "Sales Revenue"
    x-axis "Month" ["Jan", "Feb", "Mar", "Apr"]
    y-axis "Revenue" 0 --> 5000
    bar [5000, 6000, 7500, 8200]
    line [5000, 6000, 7500, 8200]`

	chart, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if chart.Title != "Sales Revenue" {
		t.Errorf("Title = %q, want %q", chart.Title, "Sales Revenue")
	}

	if chart.XLabel != "Month" {
		t.Errorf("XLabel = %q, want %q", chart.XLabel, "Month")
	}

	if len(chart.XValues) != 4 {
		t.Errorf("XValues count = %d, want 4", len(chart.XValues))
	}

	if len(chart.BarData) != 4 {
		t.Errorf("BarData count = %d, want 4", len(chart.BarData))
	}

	if chart.BarData[0] != 5000 {
		t.Errorf("BarData[0] = %f, want 5000", chart.BarData[0])
	}
}

func TestParseNoData(t *testing.T) {
	_, err := Parse("xychart-beta\n  title \"Empty\"")
	if err == nil {
		t.Error("Expected error for no data")
	}
}

func TestRender(t *testing.T) {
	input := `xychart-beta
    title "Test Chart"
    x-axis "X" ["A", "B", "C"]
    y-axis "Y" 0 --> 100
    bar [30, 60, 90]`

	chart, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	output, err := Render(chart, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(output, "Test Chart") {
		t.Error("Output should contain title")
	}
	if !strings.Contains(output, "█") {
		t.Error("Output should contain bar character")
	}
}

func TestRenderASCII(t *testing.T) {
	input := `xychart-beta
    bar [50, 100]`

	chart, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	config.UseAscii = true
	output, err := Render(chart, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if strings.Contains(output, "█") {
		t.Error("ASCII output should not contain Unicode characters")
	}
	if !strings.Contains(output, "#") {
		t.Error("ASCII output should contain # character")
	}
}

func TestRenderBarChart(t *testing.T) {
	input := `xychart-beta
    title "Monthly Sales"
    x-axis "Month" ["Jan", "Feb", "Mar"]
    y-axis "Sales" 0 --> 100
    bar [30, 60, 90]`

	chart, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	output, err := Render(chart, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(output, "Monthly Sales") {
		t.Error("expected title in output")
	}
	if !strings.Contains(output, "█") {
		t.Error("expected bar character in output")
	}
	if !strings.Contains(output, "Jan") {
		t.Error("expected x-axis label 'Jan' in output")
	}
	if !strings.Contains(output, "Month") {
		t.Error("expected x-axis label 'Month' in output")
	}
	// Should have the y-axis vertical line
	if !strings.Contains(output, "│") {
		t.Error("expected y-axis vertical line character")
	}
	// Should have x-axis bottom border
	if !strings.Contains(output, "└") {
		t.Error("expected bottom-left corner character")
	}
}

func TestRenderLineChart(t *testing.T) {
	input := `xychart-beta
    title "Temperature"
    x-axis "Day" ["Mon", "Tue", "Wed"]
    y-axis "Temp" 0 --> 40
    line [10, 25, 35]`

	chart, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	output, err := Render(chart, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(output, "Temperature") {
		t.Error("expected title in output")
	}
	if !strings.Contains(output, "●") {
		t.Error("expected line point character in output")
	}
}

func TestRenderLineChartASCII(t *testing.T) {
	input := `xychart-beta
    line [10, 30, 50]`

	chart, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	config.UseAscii = true
	output, err := Render(chart, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(output, "*") {
		t.Error("expected ASCII line point character '*' in output")
	}
	if strings.Contains(output, "●") {
		t.Error("ASCII output should not contain Unicode line point")
	}
	if !strings.Contains(output, "+") {
		t.Error("expected ASCII corner '+' in output")
	}
	if !strings.Contains(output, "|") {
		t.Error("expected ASCII vertical '|' in output")
	}
}

func TestRenderNilChart(t *testing.T) {
	config := diagram.DefaultConfig()
	_, err := Render(nil, config)
	if err == nil {
		t.Error("expected error for nil chart")
	}
}

func TestRenderNilConfig(t *testing.T) {
	input := `xychart-beta
    bar [10, 20, 30]`

	chart, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	output, err := Render(chart, nil)
	if err != nil {
		t.Fatalf("Render with nil config failed: %v", err)
	}
	if output == "" {
		t.Error("expected non-empty output with nil config")
	}
	// Nil config should default to unicode
	if !strings.Contains(output, "│") {
		t.Error("expected Unicode characters with nil config")
	}
}

func TestRenderBarAndLineChart(t *testing.T) {
	input := `xychart-beta
    title "Combined"
    x-axis "X" ["A", "B", "C"]
    y-axis "Y" 0 --> 100
    bar [40, 70, 90]
    line [30, 60, 80]`

	chart, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	output, err := Render(chart, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(output, "Combined") {
		t.Error("expected title in output")
	}
	if !strings.Contains(output, "█") {
		t.Error("expected bar character in combined chart")
	}
	// Line points may or may not be visible depending on where bars are
	// Just verify it renders without error
	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestRenderStaggeredLabels(t *testing.T) {
	input := `xychart-beta
    x-axis "Month" ["January", "February", "March", "April", "May", "June", "July", "August"]
    y-axis "Y" 0 --> 100
    bar [30, 50, 70, 40, 60, 80, 55, 65]`

	chart, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	output, err := Render(chart, diagram.DefaultConfig())
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Staggered layout should have full labels across two rows
	if !strings.Contains(output, "January") {
		t.Error("expected full label 'January' in staggered output")
	}
	if !strings.Contains(output, "February") {
		t.Error("expected full label 'February' in staggered output")
	}
}

func TestRenderLegendLabels(t *testing.T) {
	input := `xychart-beta
    x-axis "Month" ["January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"]
    y-axis "Y" 0 --> 100
    bar [30, 50, 70, 40, 60, 80, 55, 65, 75, 45, 85, 95]`

	chart, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	output, err := Render(chart, diagram.DefaultConfig())
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Long labels should use legend with short keys
	if !strings.Contains(output, "a: January") {
		t.Error("expected legend entry 'a: January' in output")
	}
	if !strings.Contains(output, "l: December") {
		t.Error("expected legend entry 'l: December' in output")
	}
}

func TestRenderNoDataPoints(t *testing.T) {
	chart := &XYChart{
		BarData:  []float64{},
		LineData: []float64{},
	}
	_, err := Render(chart, diagram.DefaultConfig())
	if err == nil {
		t.Error("expected error for chart with no data points")
	}
}

func TestRenderASCIIBarChart(t *testing.T) {
	input := `xychart-beta
    title "ASCII Bars"
    x-axis "Category" ["X", "Y"]
    y-axis "Value" 0 --> 50
    bar [25, 50]`

	chart, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	config.UseAscii = true
	output, err := Render(chart, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(output, "ASCII Bars") {
		t.Error("expected title in ASCII output")
	}
	if !strings.Contains(output, "#") {
		t.Error("expected '#' bar chars in ASCII mode")
	}
	if strings.Contains(output, "█") {
		t.Error("ASCII output should not contain Unicode bar chars")
	}
	if strings.Contains(output, "│") {
		t.Error("ASCII output should not contain Unicode vertical line")
	}
}

func TestRenderNoTitle(t *testing.T) {
	input := `xychart-beta
    bar [10, 20]`

	chart, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	output, err := Render(chart, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Should still produce valid output without a title
	if output == "" {
		t.Error("expected non-empty output without title")
	}
	if !strings.Contains(output, "│") {
		t.Error("expected y-axis in output")
	}
}
