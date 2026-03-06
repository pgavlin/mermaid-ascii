package journey

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

const JourneyKeyword = "journey"

var (
	titleRegex   = regexp.MustCompile(`^\s*title\s+(.+)$`)
	sectionRegex = regexp.MustCompile(`^\s*section\s+(.+)$`)
	taskRegex    = regexp.MustCompile(`^\s*(.+?)\s*:\s*(\d+)\s*(?::\s*(.+))?\s*$`)
)

type JourneyDiagram struct {
	Title    string
	Sections []*JourneySection
}

type JourneySection struct {
	Name  string
	Tasks []*JourneyTask
}

type JourneyTask struct {
	Name   string
	Score  int
	Actors []string
}

func IsJourneyDiagram(input string) bool {
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}
		return trimmed == JourneyKeyword
	}
	return false
}

func Parse(input string) (*JourneyDiagram, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty input")
	}

	rawLines := diagram.SplitLines(input)
	lines := diagram.RemoveComments(rawLines)
	if len(lines) == 0 {
		return nil, fmt.Errorf("no content found")
	}

	if strings.TrimSpace(lines[0]) != JourneyKeyword {
		return nil, fmt.Errorf("expected %q keyword", JourneyKeyword)
	}
	lines = lines[1:]

	jd := &JourneyDiagram{
		Sections: []*JourneySection{},
	}
	var currentSection *JourneySection

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if match := titleRegex.FindStringSubmatch(trimmed); match != nil {
			jd.Title = strings.TrimSpace(match[1])
			continue
		}

		if match := sectionRegex.FindStringSubmatch(trimmed); match != nil {
			currentSection = &JourneySection{
				Name:  strings.TrimSpace(match[1]),
				Tasks: []*JourneyTask{},
			}
			jd.Sections = append(jd.Sections, currentSection)
			continue
		}

		if match := taskRegex.FindStringSubmatch(trimmed); match != nil {
			score, _ := strconv.Atoi(match[2])
			var actors []string
			if match[3] != "" {
				for _, a := range strings.Split(match[3], ",") {
					actors = append(actors, strings.TrimSpace(a))
				}
			}
			task := &JourneyTask{
				Name:   strings.TrimSpace(match[1]),
				Score:  score,
				Actors: actors,
			}
			if currentSection != nil {
				currentSection.Tasks = append(currentSection.Tasks, task)
			} else {
				// Create default section
				currentSection = &JourneySection{
					Name:  "",
					Tasks: []*JourneyTask{task},
				}
				jd.Sections = append(jd.Sections, currentSection)
			}
			continue
		}
	}

	if len(jd.Sections) == 0 {
		return nil, fmt.Errorf("no tasks found")
	}

	return jd, nil
}
