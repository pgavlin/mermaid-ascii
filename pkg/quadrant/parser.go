package quadrant

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

const QuadrantKeyword = "quadrantChart"

var (
	titleRegex    = regexp.MustCompile(`^\s*title\s+(.+)$`)
	xAxisRegex    = regexp.MustCompile(`^\s*x-axis\s+(.+?)(?:\s*-->\s*(.+))?$`)
	yAxisRegex    = regexp.MustCompile(`^\s*y-axis\s+(.+?)(?:\s*-->\s*(.+))?$`)
	quadrant1Regex = regexp.MustCompile(`^\s*quadrant-1\s+(.+)$`)
	quadrant2Regex = regexp.MustCompile(`^\s*quadrant-2\s+(.+)$`)
	quadrant3Regex = regexp.MustCompile(`^\s*quadrant-3\s+(.+)$`)
	quadrant4Regex = regexp.MustCompile(`^\s*quadrant-4\s+(.+)$`)
	pointRegex    = regexp.MustCompile(`^\s*(.+?)\s*:\s*\[\s*([\d.]+)\s*,\s*([\d.]+)\s*\]\s*$`)
)

type QuadrantChart struct {
	Title      string
	XAxisLeft  string
	XAxisRight string
	YAxisBottom string
	YAxisTop    string
	Quadrant1  string // top-right
	Quadrant2  string // top-left
	Quadrant3  string // bottom-left
	Quadrant4  string // bottom-right
	Points     []*DataPoint
}

type DataPoint struct {
	Label string
	X     float64
	Y     float64
}

func IsQuadrantChart(input string) bool {
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}
		return trimmed == QuadrantKeyword
	}
	return false
}

func Parse(input string) (*QuadrantChart, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty input")
	}

	rawLines := diagram.SplitLines(input)
	lines := diagram.RemoveComments(rawLines)
	if len(lines) == 0 {
		return nil, fmt.Errorf("no content found")
	}

	if strings.TrimSpace(lines[0]) != QuadrantKeyword {
		return nil, fmt.Errorf("expected %q keyword", QuadrantKeyword)
	}
	lines = lines[1:]

	qc := &QuadrantChart{
		XAxisLeft:  "Low",
		XAxisRight: "High",
		YAxisBottom: "Low",
		YAxisTop:    "High",
		Points:     []*DataPoint{},
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if match := titleRegex.FindStringSubmatch(trimmed); match != nil {
			qc.Title = strings.TrimSpace(match[1])
			continue
		}

		if match := xAxisRegex.FindStringSubmatch(trimmed); match != nil {
			qc.XAxisLeft = strings.TrimSpace(match[1])
			if match[2] != "" {
				qc.XAxisRight = strings.TrimSpace(match[2])
			}
			continue
		}

		if match := yAxisRegex.FindStringSubmatch(trimmed); match != nil {
			qc.YAxisBottom = strings.TrimSpace(match[1])
			if match[2] != "" {
				qc.YAxisTop = strings.TrimSpace(match[2])
			}
			continue
		}

		if match := quadrant1Regex.FindStringSubmatch(trimmed); match != nil {
			qc.Quadrant1 = strings.TrimSpace(match[1])
			continue
		}
		if match := quadrant2Regex.FindStringSubmatch(trimmed); match != nil {
			qc.Quadrant2 = strings.TrimSpace(match[1])
			continue
		}
		if match := quadrant3Regex.FindStringSubmatch(trimmed); match != nil {
			qc.Quadrant3 = strings.TrimSpace(match[1])
			continue
		}
		if match := quadrant4Regex.FindStringSubmatch(trimmed); match != nil {
			qc.Quadrant4 = strings.TrimSpace(match[1])
			continue
		}

		if match := pointRegex.FindStringSubmatch(trimmed); match != nil {
			x, _ := strconv.ParseFloat(match[2], 64)
			y, _ := strconv.ParseFloat(match[3], 64)
			qc.Points = append(qc.Points, &DataPoint{
				Label: strings.TrimSpace(match[1]),
				X:     x,
				Y:     y,
			})
			continue
		}
	}

	return qc, nil
}
