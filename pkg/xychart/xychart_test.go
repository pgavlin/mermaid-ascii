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

	if len(chart.BarSeries) != 1 {
		t.Fatalf("BarSeries count = %d, want 1", len(chart.BarSeries))
	}
	if len(chart.BarSeries[0].Data) != 4 {
		t.Errorf("BarSeries[0].Data count = %d, want 4", len(chart.BarSeries[0].Data))
	}
	if chart.BarSeries[0].Data[0] != 5000 {
		t.Errorf("BarSeries[0].Data[0] = %f, want 5000", chart.BarSeries[0].Data[0])
	}
}

func TestParseNamedSeries(t *testing.T) {
	input := `xychart-beta
    title "Drift Runs"
    x-axis ["Mar 25", "Apr 25", "May 25"]
    y-axis "Runs" 0 --> 5500
    bar "Succeeded" [1533, 1577, 1745]
    bar "Failed" [519, 911, 1670]`

	chart, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(chart.BarSeries) != 2 {
		t.Fatalf("BarSeries count = %d, want 2", len(chart.BarSeries))
	}
	if chart.BarSeries[0].Name != "Succeeded" {
		t.Errorf("BarSeries[0].Name = %q, want %q", chart.BarSeries[0].Name, "Succeeded")
	}
	if chart.BarSeries[1].Name != "Failed" {
		t.Errorf("BarSeries[1].Name = %q, want %q", chart.BarSeries[1].Name, "Failed")
	}
	if len(chart.BarSeries[0].Data) != 3 {
		t.Errorf("BarSeries[0].Data count = %d, want 3", len(chart.BarSeries[0].Data))
	}
	if chart.BarSeries[0].Data[0] != 1533 {
		t.Errorf("BarSeries[0].Data[0] = %f, want 1533", chart.BarSeries[0].Data[0])
	}
}

func TestParseNamedLine(t *testing.T) {
	input := `xychart-beta
    line "Trend" [10, 20, 30]`

	chart, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(chart.LineSeries) != 1 {
		t.Fatalf("LineSeries count = %d, want 1", len(chart.LineSeries))
	}
	if chart.LineSeries[0].Name != "Trend" {
		t.Errorf("LineSeries[0].Name = %q, want %q", chart.LineSeries[0].Name, "Trend")
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
	if !strings.Contains(output, "│") {
		t.Error("expected y-axis vertical line character")
	}
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

	// Long labels should use legend with short keys to the right of the chart
	if !strings.Contains(output, "a: January") {
		t.Errorf("expected legend entry 'a: January' in output:\n%s", output)
	}
	if !strings.Contains(output, "l: December") {
		t.Errorf("expected legend entry 'l: December' in output:\n%s", output)
	}
	// Legend should appear on the same lines as chart body (to the right)
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "a: January") {
			if !strings.Contains(line, "│") {
				t.Error("legend should be on the same line as chart body")
			}
			break
		}
	}
}

func TestRenderMultiColumnLegend(t *testing.T) {
	input := `xychart-beta
    x-axis "X" ["Alpha One", "Bravo Two", "Charlie Three", "Delta Four", "Echo Five", "Foxtrot Six", "Golf Seven", "Hotel Eight", "India Nine", "Juliet Ten", "Kilo Eleven", "Lima Twelve", "Mike Thirteen", "November Fourteen", "Oscar Fifteen", "Papa Sixteen"]
    y-axis "Y" 0 --> 100
    bar [30, 50, 70, 40, 60, 80, 55, 65, 75, 45, 85, 95, 35, 55, 72, 88]`

	chart, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	output, err := Render(chart, diagram.DefaultConfig())
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// 16 entries > 15 chart rows, so should use 2 columns
	if !strings.Contains(output, "a: Alpha One") {
		t.Errorf("expected legend entry 'a: Alpha One' in output:\n%s", output)
	}
	if !strings.Contains(output, "p: Papa Sixteen") {
		t.Errorf("expected legend entry 'p: Papa Sixteen' in output:\n%s", output)
	}
	// Verify multi-column: some line should contain two legend entries
	found := false
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "a: Alpha One") && strings.Contains(line, "i: India Nine") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected multi-column legend with entries on same line:\n%s", output)
	}
}

func TestRenderNoDataPoints(t *testing.T) {
	chart := &XYChart{}
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

func TestParseXAxisNoLabel(t *testing.T) {
	input := `xychart-beta
    x-axis ["Mar 25", "Apr 25", "May 25"]
    y-axis "Runs" 0 --> 6500
    bar [2052, 2488, 3415]`

	chart, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(chart.XValues) != 3 {
		t.Errorf("XValues count = %d, want 3", len(chart.XValues))
	}
	if chart.XValues[0] != "Mar 25" {
		t.Errorf("XValues[0] = %q, want %q", chart.XValues[0], "Mar 25")
	}
	if chart.XLabel != "" {
		t.Errorf("XLabel = %q, want empty", chart.XLabel)
	}
}

func TestParseYAxisNoLabel(t *testing.T) {
	input := `xychart-beta
    y-axis 0 --> 100
    bar [30, 60, 90]`

	chart, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if chart.YMin != 0 {
		t.Errorf("YMin = %f, want 0", chart.YMin)
	}
	if chart.YMax != 100 {
		t.Errorf("YMax = %f, want 100", chart.YMax)
	}
	if chart.YLabel != "" {
		t.Errorf("YLabel = %q, want empty", chart.YLabel)
	}
}

func TestRenderXAxisNoLabel(t *testing.T) {
	input := `xychart-beta
    title "Drift Detection Runs per Month"
    x-axis ["Mar 25", "Apr 25", "May 25"]
    y-axis "Runs" 0 --> 6500
    bar [2052, 2488, 3415]`

	chart, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	output, err := Render(chart, diagram.DefaultConfig())
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if !strings.Contains(output, "Mar 25") {
		t.Errorf("expected x-axis label 'Mar 25' in output:\n%s", output)
	}
	if !strings.Contains(output, "May 25") {
		t.Errorf("expected x-axis label 'May 25' in output:\n%s", output)
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

	if output == "" {
		t.Error("expected non-empty output without title")
	}
	if !strings.Contains(output, "│") {
		t.Error("expected y-axis in output")
	}
}

func TestRenderMultipleBarSeries(t *testing.T) {
	input := `xychart-beta
    title "Drift Runs: Execution Succeeded vs. Failed"
    x-axis ["Mar 25", "Apr 25", "May 25"]
    y-axis "Runs" 0 --> 5500
    bar "Succeeded" [1533, 1577, 1745]
    bar "Failed" [519, 911, 1670]`

	chart, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	output, err := Render(chart, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Should contain both bar characters for the two series.
	if !strings.Contains(output, "█") {
		t.Error("expected first bar character █")
	}
	if !strings.Contains(output, "▓") {
		t.Error("expected second bar character ▓")
	}
	// Should contain series legend.
	if !strings.Contains(output, "██ Succeeded") {
		t.Errorf("expected series legend for Succeeded in output:\n%s", output)
	}
	if !strings.Contains(output, "▓▓ Failed") {
		t.Errorf("expected series legend for Failed in output:\n%s", output)
	}
}

func TestRenderMultipleBarSeriesFull(t *testing.T) {
	input := `xychart-beta
    title "Drift Runs: Execution Succeeded vs. Failed"
    x-axis ["Mar 25", "Apr 25", "May 25", "Jun 25", "Jul 25", "Aug 25", "Sep 25", "Oct 25", "Nov 25", "Dec 25", "Jan 26", "Feb 26"]
    y-axis "Runs" 0 --> 5500
    bar "Succeeded" [1533, 1577, 1745, 1655, 1945, 2016, 1887, 3008, 4885, 4218, 3609, 3217]
    bar "Failed" [519, 911, 1670, 1708, 1811, 1312, 1286, 929, 1040, 1519, 1673, 1822]`

	chart, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	config := diagram.DefaultConfig()
	output, err := Render(chart, config)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	t.Logf("Rendered output:\n%s", output)

	if !strings.Contains(output, "Drift Runs") {
		t.Error("expected title in output")
	}
	if !strings.Contains(output, "█") {
		t.Error("expected first series bar character")
	}
	if !strings.Contains(output, "▓") {
		t.Error("expected second series bar character")
	}
}
