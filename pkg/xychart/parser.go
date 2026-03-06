package xychart

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

const XYChartKeyword = "xychart-beta"

var (
	titleRegex   = regexp.MustCompile(`^\s*title\s+"([^"]+)"$`)
	xAxisRegex   = regexp.MustCompile(`^\s*x-axis\s+"([^"]+)"\s+\[(.+)\]$`)
	yAxisRegex   = regexp.MustCompile(`^\s*y-axis\s+"([^"]+)"(?:\s+(\d+)\s*-->\s*(\d+))?$`)
	barRegex     = regexp.MustCompile(`^\s*bar\s+\[(.+)\]$`)
	lineRegex    = regexp.MustCompile(`^\s*line\s+\[(.+)\]$`)
)

type XYChart struct {
	Title    string
	XLabel   string
	XValues  []string
	YLabel   string
	YMin     float64
	YMax     float64
	BarData  []float64
	LineData []float64
}

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
		XValues:  []string{},
		BarData:  []float64{},
		LineData: []float64{},
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

		if match := xAxisRegex.FindStringSubmatch(trimmed); match != nil {
			chart.XLabel = match[1]
			chart.XValues = parseStringList(match[2])
			continue
		}

		if match := yAxisRegex.FindStringSubmatch(trimmed); match != nil {
			chart.YLabel = match[1]
			if match[2] != "" {
				chart.YMin, _ = strconv.ParseFloat(match[2], 64)
			}
			if match[3] != "" {
				chart.YMax, _ = strconv.ParseFloat(match[3], 64)
			}
			continue
		}

		if match := barRegex.FindStringSubmatch(trimmed); match != nil {
			chart.BarData = parseFloatList(match[1])
			continue
		}

		if match := lineRegex.FindStringSubmatch(trimmed); match != nil {
			chart.LineData = parseFloatList(match[1])
			continue
		}
	}

	if len(chart.BarData) == 0 && len(chart.LineData) == 0 {
		return nil, fmt.Errorf("no data found")
	}

	// Auto-calculate Y range if not specified
	if chart.YMax == 0 {
		for _, v := range chart.BarData {
			if v > chart.YMax {
				chart.YMax = v
			}
		}
		for _, v := range chart.LineData {
			if v > chart.YMax {
				chart.YMax = v
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
