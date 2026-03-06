package piechart

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

const PieKeyword = "pie"

var (
	titleRegex = regexp.MustCompile(`^\s*title\s+(.+)$`)
	sliceRegex = regexp.MustCompile(`^\s*"([^"]+)"\s*:\s*([\d.]+)\s*$`)
)

type PieChart struct {
	Title  string
	Slices []*Slice
	Total  float64
}

type Slice struct {
	Label      string
	Value      float64
	Percentage float64
}

func IsPieChart(input string) bool {
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}
		return trimmed == PieKeyword || strings.HasPrefix(trimmed, PieKeyword+" ")
	}
	return false
}

func Parse(input string) (*PieChart, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty input")
	}

	rawLines := diagram.SplitLines(input)
	lines := diagram.RemoveComments(rawLines)
	if len(lines) == 0 {
		return nil, fmt.Errorf("no content found")
	}

	first := strings.TrimSpace(lines[0])
	if first != PieKeyword && !strings.HasPrefix(first, PieKeyword+" ") {
		return nil, fmt.Errorf("expected %q keyword", PieKeyword)
	}

	// Check for inline title: "pie title My Title"
	pc := &PieChart{
		Slices: []*Slice{},
	}
	if strings.HasPrefix(first, PieKeyword+" title ") {
		pc.Title = strings.TrimPrefix(first, PieKeyword+" title ")
	}
	lines = lines[1:]

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if match := titleRegex.FindStringSubmatch(trimmed); match != nil {
			pc.Title = strings.TrimSpace(match[1])
			continue
		}

		if match := sliceRegex.FindStringSubmatch(trimmed); match != nil {
			value, err := strconv.ParseFloat(match[2], 64)
			if err != nil {
				continue
			}
			pc.Slices = append(pc.Slices, &Slice{
				Label: match[1],
				Value: value,
			})
			continue
		}

		// Check for "showData" keyword - ignored in ASCII
		if trimmed == "showData" {
			continue
		}
	}

	if len(pc.Slices) == 0 {
		return nil, fmt.Errorf("no data found")
	}

	// Calculate percentages
	total := 0.0
	for _, s := range pc.Slices {
		total += s.Value
	}
	pc.Total = total
	for _, s := range pc.Slices {
		s.Percentage = (s.Value / total) * 100
	}

	return pc, nil
}
