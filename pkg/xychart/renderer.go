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
		labelLine := strings.Repeat(" ", yLabelWidth+2)
		for i, label := range chart.XValues {
			if i >= dataCount {
				break
			}
			if len(label) > barWidth {
				label = label[:barWidth]
			}
			pad := (barWidth - len(label)) / 2
			labelLine += strings.Repeat(" ", pad) + label
			remaining := barWidth - pad - len(label)
			if remaining > 0 {
				labelLine += strings.Repeat(" ", remaining)
			}
		}
		lines = append(lines, strings.TrimRight(labelLine, " "))
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
