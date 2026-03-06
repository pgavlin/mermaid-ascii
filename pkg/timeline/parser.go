package timeline

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

const TimelineKeyword = "timeline"

var (
	titleRegex   = regexp.MustCompile(`^\s*title\s+(.+)$`)
	sectionRegex = regexp.MustCompile(`^\s*section\s+(.+)$`)
	eventRegex   = regexp.MustCompile(`^\s*(.+?)\s*:\s*(.+)$`)
)

type TimelineDiagram struct {
	Title    string
	Sections []*TimelineSection
	Events   []*TimelineEvent
}

type TimelineSection struct {
	Name   string
	Events []*TimelineEvent
}

type TimelineEvent struct {
	Period  string
	Events  []string
	Section *TimelineSection
}

func IsTimelineDiagram(input string) bool {
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}
		return trimmed == TimelineKeyword
	}
	return false
}

func Parse(input string) (*TimelineDiagram, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty input")
	}

	rawLines := diagram.SplitLines(input)
	lines := diagram.RemoveComments(rawLines)
	if len(lines) == 0 {
		return nil, fmt.Errorf("no content found")
	}

	if strings.TrimSpace(lines[0]) != TimelineKeyword {
		return nil, fmt.Errorf("expected %q keyword", TimelineKeyword)
	}
	lines = lines[1:]

	td := &TimelineDiagram{
		Sections: []*TimelineSection{},
		Events:   []*TimelineEvent{},
	}
	var currentSection *TimelineSection

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if match := titleRegex.FindStringSubmatch(trimmed); match != nil {
			td.Title = strings.TrimSpace(match[1])
			continue
		}

		if match := sectionRegex.FindStringSubmatch(trimmed); match != nil {
			currentSection = &TimelineSection{
				Name:   strings.TrimSpace(match[1]),
				Events: []*TimelineEvent{},
			}
			td.Sections = append(td.Sections, currentSection)
			continue
		}

		if match := eventRegex.FindStringSubmatch(trimmed); match != nil {
			period := strings.TrimSpace(match[1])
			eventTexts := strings.Split(match[2], ":")
			events := make([]string, 0, len(eventTexts))
			for _, e := range eventTexts {
				e = strings.TrimSpace(e)
				if e != "" {
					events = append(events, e)
				}
			}

			event := &TimelineEvent{
				Period:  period,
				Events:  events,
				Section: currentSection,
			}
			if currentSection != nil {
				currentSection.Events = append(currentSection.Events, event)
			}
			td.Events = append(td.Events, event)
			continue
		}

		// Bare line is a period without events
		event := &TimelineEvent{
			Period:  trimmed,
			Events:  []string{},
			Section: currentSection,
		}
		if currentSection != nil {
			currentSection.Events = append(currentSection.Events, event)
		}
		td.Events = append(td.Events, event)
	}

	if len(td.Events) == 0 {
		return nil, fmt.Errorf("no events found")
	}

	return td, nil
}
