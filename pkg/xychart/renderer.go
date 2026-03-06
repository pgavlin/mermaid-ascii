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
	if len(chart.XValues) > 0 {
		lines = append(lines, renderXLabels(chart.XValues, dataCount, barWidth, yLabelWidth)...)
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

	return strings.Join(lines, "\n") + "\n", nil
}

// renderXLabels renders x-axis labels using one of three strategies:
//  1. Single row — if every label fits within its bar width
//  2. Staggered two rows — if labels fit within 2x bar width
//  3. Vertical — labels written top-to-bottom, one character per row
func renderXLabels(labels []string, dataCount, barWidth, yLabelWidth int) []string {
	// Determine the max label length (capped for display)
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
		// Strategy 1: single horizontal row
		return []string{strings.TrimRight(singleRowLabels(labels, dataCount, barWidth, prefix), " ")}
	}

	if maxLabel <= barWidth*2 {
		// Strategy 2: staggered two rows
		return staggeredLabels(labels, dataCount, barWidth, prefix)
	}

	// Strategy 3: short vertical or legend with keys
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

func verticalLabels(labels []string, dataCount, barWidth int, prefix string) []string {
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
		return verticalLabelRows(labels, dataCount, barWidth, prefix)
	}

	// Use short keys on the axis with a legend below
	keys := make([]string, dataCount)
	for i := 0; i < dataCount; i++ {
		keys[i] = labelKey(i)
	}

	// Render the short keys as a single horizontal row (they always fit)
	rows := []string{strings.TrimRight(singleRowLabels(keys, dataCount, barWidth, prefix), " ")}

	// Append legend
	rows = append(rows, "")
	for i := 0; i < dataCount; i++ {
		label := ""
		if i < len(labels) {
			label = labels[i]
		}
		if label != "" {
			rows = append(rows, fmt.Sprintf("%s%s: %s", prefix, keys[i], label))
		}
	}
	return rows
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
