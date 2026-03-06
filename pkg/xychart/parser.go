// Package xychart provides parsing and rendering of Mermaid xychart-beta diagrams.
package xychart

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

// XYChartKeyword is the keyword that identifies an XY chart diagram in Mermaid syntax.
const XYChartKeyword = "xychart-beta"

var (
	titleRegex          = regexp.MustCompile(`^\s*title\s+"([^"]+)"$`)
	xAxisLabelRegex     = regexp.MustCompile(`^\s*x-axis\s+"([^"]+)"\s+\[(.+)\]$`)
	xAxisNoLabelRegex   = regexp.MustCompile(`^\s*x-axis\s+\[(.+)\]$`)
	yAxisLabelRegex     = regexp.MustCompile(`^\s*y-axis\s+"([^"]+)"(?:\s+(\d+(?:\.\d+)?)\s*-->\s*(\d+(?:\.\d+)?))?$`)
	yAxisNoLabelRegex   = regexp.MustCompile(`^\s*y-axis\s+(\d+(?:\.\d+)?)\s*-->\s*(\d+(?:\.\d+)?)$`)
	barRegex            = regexp.MustCompile(`^\s*bar\s+\[(.+)\]$`)
	barNamedRegex       = regexp.MustCompile(`^\s*bar\s+"([^"]+)"\s+\[(.+)\]$`)
	lineRegex           = regexp.MustCompile(`^\s*line\s+\[(.+)\]$`)
	lineNamedRegex      = regexp.MustCompile(`^\s*line\s+"([^"]+)"\s+\[(.+)\]$`)
)

// DataSeries represents a named data series in a chart.
type DataSeries struct {
	Name string
	Data []float64
}

// XYChart represents a parsed XY chart with optional bar and line data series.
type XYChart struct {
	Title      string
	XLabel     string
	XValues    []string
	YLabel     string
	YMin       float64
	YMax       float64
	BarSeries  []DataSeries
	LineSeries []DataSeries
}

// IsXYChart reports whether the input text is an XY chart diagram.
func IsXYChart(input string) bool {
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}
		return trimmed == XYChartKeyword
	}
	return false
}

// Parse parses Mermaid xychart-beta text into an XYChart.
func Parse(input string) (*XYChart, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty input")
	}

	rawLines := diagram.SplitLines(input)
	lines := diagram.RemoveComments(rawLines)
	if len(lines) == 0 {
		return nil, fmt.Errorf("no content found")
	}

	if strings.TrimSpace(lines[0]) != XYChartKeyword {
		return nil, fmt.Errorf("expected %q keyword", XYChartKeyword)
	}
	lines = lines[1:]

	chart := &XYChart{
		XValues: []string{},
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if match := titleRegex.FindStringSubmatch(trimmed); match != nil {
			chart.Title = match[1]
			continue
		}

		if match := xAxisLabelRegex.FindStringSubmatch(trimmed); match != nil {
			chart.XLabel = match[1]
			chart.XValues = parseStringList(match[2])
			continue
		}

		if match := xAxisNoLabelRegex.FindStringSubmatch(trimmed); match != nil {
			chart.XValues = parseStringList(match[1])
			continue
		}

		if match := yAxisLabelRegex.FindStringSubmatch(trimmed); match != nil {
			chart.YLabel = match[1]
			if match[2] != "" {
				chart.YMin, _ = strconv.ParseFloat(match[2], 64)
			}
			if match[3] != "" {
				chart.YMax, _ = strconv.ParseFloat(match[3], 64)
			}
			continue
		}

		if match := yAxisNoLabelRegex.FindStringSubmatch(trimmed); match != nil {
			chart.YMin, _ = strconv.ParseFloat(match[1], 64)
			chart.YMax, _ = strconv.ParseFloat(match[2], 64)
			continue
		}

		if match := barNamedRegex.FindStringSubmatch(trimmed); match != nil {
			chart.BarSeries = append(chart.BarSeries, DataSeries{
				Name: match[1],
				Data: parseFloatList(match[2]),
			})
			continue
		}

		if match := barRegex.FindStringSubmatch(trimmed); match != nil {
			chart.BarSeries = append(chart.BarSeries, DataSeries{
				Data: parseFloatList(match[1]),
			})
			continue
		}

		if match := lineNamedRegex.FindStringSubmatch(trimmed); match != nil {
			chart.LineSeries = append(chart.LineSeries, DataSeries{
				Name: match[1],
				Data: parseFloatList(match[2]),
			})
			continue
		}

		if match := lineRegex.FindStringSubmatch(trimmed); match != nil {
			chart.LineSeries = append(chart.LineSeries, DataSeries{
				Data: parseFloatList(match[1]),
			})
			continue
		}
	}

	if len(chart.BarSeries) == 0 && len(chart.LineSeries) == 0 {
		return nil, fmt.Errorf("no data found")
	}

	// Auto-calculate Y range if not specified
	if chart.YMax == 0 {
		for _, s := range chart.BarSeries {
			for _, v := range s.Data {
				if v > chart.YMax {
					chart.YMax = v
				}
			}
		}
		for _, s := range chart.LineSeries {
			for _, v := range s.Data {
				if v > chart.YMax {
					chart.YMax = v
				}
			}
		}
	}

	return chart, nil
}

func parseStringList(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.Trim(p, `"`)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func parseFloatList(s string) []float64 {
	parts := strings.Split(s, ",")
	result := make([]float64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		v, err := strconv.ParseFloat(p, 64)
		if err == nil {
			result = append(result, v)
		}
	}
	return result
}
