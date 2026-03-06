package xychart

import (
	"fmt"
	"strings"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

const (
	chartHeight = 15
	chartWidth  = 50
)

// Render renders an XYChart as a formatted bar/line chart string using Unicode or ASCII characters.
func Render(chart *XYChart, config *diagram.Config) (string, error) {
	if chart == nil {
		return "", fmt.Errorf("no chart data")
	}
	if config == nil {
		config = diagram.DefaultConfig()
	}

	useAscii := config.UseAscii

	hChar := "─"
	vChar := "│"
	cornerBL := "└"
	barChar := "█"
	linePoint := "●"
	if useAscii {
		hChar = "-"
		vChar = "|"
		cornerBL = "+"
		barChar = "#"
		linePoint = "*"
	}

	yRange := chart.YMax - chart.YMin
	if yRange <= 0 {
		yRange = 1
	}

	// Determine data count
	dataCount := len(chart.BarData)
	if len(chart.LineData) > dataCount {
		dataCount = len(chart.LineData)
	}
	if dataCount == 0 {
		return "", fmt.Errorf("no data points")
	}

	// Y-axis label width
	yLabelWidth := len(fmt.Sprintf("%.0f", chart.YMax))
	if yLabelWidth < 4 {
		yLabelWidth = 4
	}

	// Calculate bar width
	barWidth := chartWidth / dataCount
	if barWidth < 1 {
		barWidth = 1
	}

	var lines []string

	// Title
	if chart.Title != "" {
		totalWidth := yLabelWidth + 2 + chartWidth
		pad := (totalWidth - len(chart.Title)) / 2
		if pad < 0 {
			pad = 0
		}
		lines = append(lines, strings.Repeat(" ", pad)+chart.Title)
		lines = append(lines, "")
	}

	// Build chart rows (top to bottom)
	for row := chartHeight - 1; row >= 0; row-- {
		yVal := chart.YMin + yRange*float64(row)/float64(chartHeight-1)
		yLabel := fmt.Sprintf("%*.0f", yLabelWidth, yVal)

		line := yLabel + " " + vChar

		for i := 0; i < dataCount; i++ {
			// Bar data
			if i < len(chart.BarData) {
				barHeight := int(float64(chartHeight) * (chart.BarData[i] - chart.YMin) / yRange)
				if row < barHeight {
					line += strings.Repeat(barChar, barWidth)
				} else {
					// Check for line data point at this position
					if i < len(chart.LineData) {
						lineHeight := int(float64(chartHeight) * (chart.LineData[i] - chart.YMin) / yRange)
						if row == lineHeight {
							pad := barWidth / 2
							line += strings.Repeat(" ", pad) + linePoint + strings.Repeat(" ", barWidth-pad-1)
						} else {
							line += strings.Repeat(" ", barWidth)
						}
					} else {
						line += strings.Repeat(" ", barWidth)
					}
				}
			} else if i < len(chart.LineData) {
				lineHeight := int(float64(chartHeight) * (chart.LineData[i] - chart.YMin) / yRange)
				if row == lineHeight {
					pad := barWidth / 2
					line += strings.Repeat(" ", pad) + linePoint + strings.Repeat(" ", barWidth-pad-1)
				} else {
					line += strings.Repeat(" ", barWidth)
				}
			}
		}

		lines = append(lines, strings.TrimRight(line, " "))
	}

	// X-axis
	xAxisLine := strings.Repeat(" ", yLabelWidth+1) + cornerBL + strings.Repeat(hChar, dataCount*barWidth)
	lines = append(lines, xAxisLine)

	// X-axis labels
	var legend []string
	if len(chart.XValues) > 0 {
		var labelLines []string
		labelLines, legend = renderXLabels(chart.XValues, dataCount, barWidth, yLabelWidth)
		lines = append(lines, labelLines...)
	}

	// Axis labels
	if chart.XLabel != "" {
		pad := (yLabelWidth + 2 + dataCount*barWidth - len(chart.XLabel)) / 2
		if pad < 0 {
			pad = 0
		}
		lines = append(lines, "")
		lines = append(lines, strings.Repeat(" ", pad)+chart.XLabel)
	}

	// Merge legend to the right of chart body if present
	if len(legend) > 0 {
		lines = mergeLegendRight(lines, legend, yLabelWidth+2+dataCount*barWidth)
	}

	return strings.Join(lines, "\n") + "\n", nil
}

// renderXLabels renders x-axis labels and optionally returns legend entries.
// Returns (label lines for below the axis, legend entries for the right side).
func renderXLabels(labels []string, dataCount, barWidth, yLabelWidth int) ([]string, []string) {
	maxLabel := 0
	for i, l := range labels {
		if i >= dataCount {
			break
		}
		if len(l) > maxLabel {
			maxLabel = len(l)
		}
	}

	prefix := strings.Repeat(" ", yLabelWidth+2)

	if maxLabel <= barWidth {
		return []string{strings.TrimRight(singleRowLabels(labels, dataCount, barWidth, prefix), " ")}, nil
	}

	if maxLabel <= barWidth*2 {
		return staggeredLabels(labels, dataCount, barWidth, prefix), nil
	}

	// Short vertical or legend with keys
	return verticalLabels(labels, dataCount, barWidth, prefix)
}

// labelKey returns a short key for index i: a-z, then aa, ab, ...
func labelKey(i int) string {
	if i < 26 {
		return string(rune('a' + i))
	}
	return string(rune('a'+i/26-1)) + string(rune('a'+i%26))
}

func singleRowLabels(labels []string, dataCount, barWidth int, prefix string) string {
	line := prefix
	for i, label := range labels {
		if i >= dataCount {
			break
		}
		if len(label) > barWidth {
			label = label[:barWidth]
		}
		pad := (barWidth - len(label)) / 2
		line += strings.Repeat(" ", pad) + label
		remaining := barWidth - pad - len(label)
		if remaining > 0 {
			line += strings.Repeat(" ", remaining)
		}
	}
	return line
}

func staggeredLabels(labels []string, dataCount, barWidth int, prefix string) []string {
	// Even-indexed labels on row 1, odd-indexed on row 2
	row1 := prefix
	row2 := prefix
	for i := 0; i < dataCount; i++ {
		label := ""
		if i < len(labels) {
			label = labels[i]
		}
		if len(label) > barWidth*2 {
			label = label[:barWidth*2-1] + "·"
		}
		if i%2 == 0 {
			// Center label within its 2-bar span starting at this position
			pad := (barWidth - len(label)) / 2
			if pad < 0 {
				pad = 0
			}
			row1 += strings.Repeat(" ", pad) + label
			remaining := barWidth - pad - len(label)
			if remaining > 0 {
				row1 += strings.Repeat(" ", remaining)
			}
			row2 += strings.Repeat(" ", barWidth)
		} else {
			pad := (barWidth - len(label)) / 2
			if pad < 0 {
				pad = 0
			}
			row2 += strings.Repeat(" ", pad) + label
			remaining := barWidth - pad - len(label)
			if remaining > 0 {
				row2 += strings.Repeat(" ", remaining)
			}
			row1 += strings.Repeat(" ", barWidth)
		}
	}
	return []string{
		strings.TrimRight(row1, " "),
		strings.TrimRight(row2, " "),
	}
}

const verticalLabelMaxRows = 5

func verticalLabels(labels []string, dataCount, barWidth int, prefix string) ([]string, []string) {
	// Determine the max label length
	maxLen := 0
	for i, l := range labels {
		if i >= dataCount {
			break
		}
		if len(l) > maxLen {
			maxLen = len(l)
		}
	}

	if maxLen <= verticalLabelMaxRows {
		// Short enough to render vertically without a legend
		return verticalLabelRows(labels, dataCount, barWidth, prefix), nil
	}

	// Use short keys on the axis with a legend to the right
	keys := make([]string, dataCount)
	for i := 0; i < dataCount; i++ {
		keys[i] = labelKey(i)
	}

	axisRow := []string{strings.TrimRight(singleRowLabels(keys, dataCount, barWidth, prefix), " ")}

	// Build legend entries
	var legend []string
	for i := 0; i < dataCount; i++ {
		label := ""
		if i < len(labels) {
			label = labels[i]
		}
		if label != "" {
			legend = append(legend, fmt.Sprintf("%s: %s", keys[i], label))
		}
	}
	return axisRow, legend
}

// mergeLegendRight places legend entries to the right of the chart body lines,
// using multiple columns if there are more entries than available rows.
func mergeLegendRight(lines []string, legend []string, chartRightEdge int) []string {
	if len(legend) == 0 {
		return lines
	}

	// Find the first chart body line (skip title/blank lines at the top)
	// Chart body lines are those starting with y-axis labels, containing the vertical bar character
	firstBody := 0
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > 0 && (trimmed[0] >= '0' && trimmed[0] <= '9' || trimmed[0] == '-') {
			firstBody = i
			break
		}
	}

	availableRows := chartHeight

	// Arrange legend into columns
	numCols := (len(legend) + availableRows - 1) / availableRows
	rowsPerCol := (len(legend) + numCols - 1) / numCols

	// Find max entry width per column for alignment
	colWidths := make([]int, numCols)
	for i, entry := range legend {
		col := i / rowsPerCol
		if len(entry) > colWidths[col] {
			colWidths[col] = len(entry)
		}
	}

	const legendGap = 2 // gap between chart and legend, and between columns

	// Build legend rows
	legendRows := make([]string, rowsPerCol)
	for row := 0; row < rowsPerCol; row++ {
		var parts []string
		for col := 0; col < numCols; col++ {
			idx := col*rowsPerCol + row
			if idx < len(legend) {
				entry := legend[idx]
				padded := entry + strings.Repeat(" ", colWidths[col]-len(entry))
				parts = append(parts, padded)
			}
		}
		legendRows[row] = strings.TrimRight(strings.Join(parts, strings.Repeat(" ", legendGap)), " ")
	}

	// Merge: pad each chart body line and append the corresponding legend row
	for i, legendRow := range legendRows {
		lineIdx := firstBody + i
		if lineIdx >= len(lines) {
			break
		}
		line := lines[lineIdx]
		// Pad line to chart right edge
		if len(line) < chartRightEdge {
			line += strings.Repeat(" ", chartRightEdge-len(line))
		}
		lines[lineIdx] = line + strings.Repeat(" ", legendGap) + legendRow
	}

	return lines
}

func verticalLabelRows(labels []string, dataCount, barWidth int, prefix string) []string {
	display := make([]string, dataCount)
	maxHeight := 0
	for i := 0; i < dataCount; i++ {
		if i < len(labels) {
			display[i] = labels[i]
		}
		if len(display[i]) > maxHeight {
			maxHeight = len(display[i])
		}
	}

	var rows []string
	for row := 0; row < maxHeight; row++ {
		line := prefix
		for i := 0; i < dataCount; i++ {
			label := display[i]
			centerPos := barWidth / 2
			if row < len(label) {
				line += strings.Repeat(" ", centerPos) + string(label[row])
				remaining := barWidth - centerPos - 1
				if remaining > 0 {
					line += strings.Repeat(" ", remaining)
				}
			} else {
				line += strings.Repeat(" ", barWidth)
			}
		}
		rows = append(rows, strings.TrimRight(line, " "))
	}
	return rows
}
