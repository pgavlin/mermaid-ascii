package timeline

import (
	"fmt"
	"strings"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

func Render(td *TimelineDiagram, config *diagram.Config) (string, error) {
	if td == nil || len(td.Events) == 0 {
		return "", fmt.Errorf("no events to render")
	}
	if config == nil {
		config = diagram.DefaultConfig()
	}

	useAscii := config.UseAscii

	// Characters
	hChar := "─"
	vChar := "│"
	dotChar := "●"
	cornerTL := "┌"
	cornerTR := "┐"
	cornerBL := "└"
	cornerBR := "┘"
	teeDown := "┬"
	if useAscii {
		hChar = "-"
		vChar = "|"
		dotChar = "*"
		cornerTL = "+"
		cornerTR = "+"
		cornerBL = "+"
		cornerBR = "+"
		teeDown = "+"
	}

	var lines []string

	// Title
	if td.Title != "" {
		lines = append(lines, td.Title)
		lines = append(lines, "")
	}

	// Find widths
	periodWidth := 0
	eventWidth := 0
	for _, event := range td.Events {
		if len(event.Period) > periodWidth {
			periodWidth = len(event.Period)
		}
		for _, e := range event.Events {
			if len(e) > eventWidth {
				eventWidth = len(e)
			}
		}
	}
	periodWidth += 2
	eventWidth += 4
	if eventWidth < 10 {
		eventWidth = 10
	}

	// Render each event
	lastSection := ""
	for _, event := range td.Events {
		// Section header
		if event.Section != nil && event.Section.Name != lastSection {
			lastSection = event.Section.Name
			lines = append(lines, "")
			lines = append(lines, "  "+event.Section.Name)
			sepLine := "  " + strings.Repeat(hChar, len(event.Section.Name))
			lines = append(lines, sepLine)
		}

		// Period marker on timeline
		periodPad := strings.Repeat(" ", periodWidth-len(event.Period))
		timelineLine := "  " + event.Period + periodPad + dotChar + hChar + hChar

		if len(event.Events) == 0 {
			lines = append(lines, timelineLine)
		} else {
			// Event box
			boxWidth := eventWidth
			for _, e := range event.Events {
				if len(e)+4 > boxWidth {
					boxWidth = len(e) + 4
				}
			}

			// Top of box
			topLine := timelineLine + " " + cornerTL + strings.Repeat(hChar, boxWidth) + cornerTR
			lines = append(lines, topLine)

			// Event content
			linePrefix := strings.Repeat(" ", periodWidth+4) + " " + vChar
			for _, e := range event.Events {
				pad := strings.Repeat(" ", boxWidth-len(e)-2)
				contentLine := linePrefix + " " + e + pad + " " + vChar
				lines = append(lines, contentLine)
			}

			// Bottom of box
			bottomPrefix := strings.Repeat(" ", periodWidth+4) + " " + cornerBL
			bottomLine := bottomPrefix + strings.Repeat(hChar, boxWidth) + cornerBR
			lines = append(lines, bottomLine)
		}
	}

	// Timeline axis at the bottom
	lines = append(lines, "")
	axisLine := "  " + strings.Repeat(" ", periodWidth) + vChar
	lines = append(lines, axisLine)
	axisBottom := "  " + strings.Repeat(" ", periodWidth) + teeDown + strings.Repeat(hChar, eventWidth+2)
	lines = append(lines, axisBottom)

	_ = cornerTL
	_ = cornerTR
	_ = cornerBL
	_ = cornerBR

	return strings.Join(lines, "\n") + "\n", nil
}
