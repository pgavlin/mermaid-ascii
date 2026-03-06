package gantt

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

const GanttKeyword = "gantt"

var (
	titleRegex      = regexp.MustCompile(`^\s*title\s+(.+)$`)
	dateFormatRegex = regexp.MustCompile(`^\s*dateFormat\s+(.+)$`)
	sectionRegex    = regexp.MustCompile(`^\s*section\s+(.+)$`)
	excludesRegex   = regexp.MustCompile(`^\s*excludes\s+(.+)$`)
	taskRegex       = regexp.MustCompile(`^\s*(.+?)\s*:\s*(.+)$`)
)

type GanttDiagram struct {
	Title      string
	DateFormat string
	Sections   []*Section
	Tasks      []*Task
}

type Section struct {
	Name  string
	Tasks []*Task
}

type Task struct {
	Name      string
	ID        string
	Status    string // done, active, crit, or empty
	StartDate time.Time
	EndDate   time.Time
	Duration  time.Duration
	After     string // dependency
	Section   *Section
}

func IsGanttDiagram(input string) bool {
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}
		return trimmed == GanttKeyword
	}
	return false
}

func Parse(input string) (*GanttDiagram, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty input")
	}

	rawLines := diagram.SplitLines(input)
	lines := diagram.RemoveComments(rawLines)
	if len(lines) == 0 {
		return nil, fmt.Errorf("no content found")
	}

	if strings.TrimSpace(lines[0]) != GanttKeyword {
		return nil, fmt.Errorf("expected %q keyword", GanttKeyword)
	}
	lines = lines[1:]

	gd := &GanttDiagram{
		DateFormat: "YYYY-MM-DD",
		Sections:   []*Section{},
		Tasks:      []*Task{},
	}

	var currentSection *Section
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	tasksByID := make(map[string]*Task)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if match := titleRegex.FindStringSubmatch(trimmed); match != nil {
			gd.Title = strings.TrimSpace(match[1])
			continue
		}

		if match := dateFormatRegex.FindStringSubmatch(trimmed); match != nil {
			gd.DateFormat = strings.TrimSpace(match[1])
			continue
		}

		if excludesRegex.MatchString(trimmed) {
			continue // acknowledged but not rendered differently
		}

		if match := sectionRegex.FindStringSubmatch(trimmed); match != nil {
			currentSection = &Section{
				Name:  strings.TrimSpace(match[1]),
				Tasks: []*Task{},
			}
			gd.Sections = append(gd.Sections, currentSection)
			continue
		}

		if match := taskRegex.FindStringSubmatch(trimmed); match != nil {
			task, err := parseTask(match[1], match[2], baseDate, tasksByID, gd.DateFormat)
			if err != nil {
				continue // skip unparseable tasks
			}
			task.Section = currentSection
			if currentSection != nil {
				currentSection.Tasks = append(currentSection.Tasks, task)
			}
			gd.Tasks = append(gd.Tasks, task)
			if task.ID != "" {
				tasksByID[task.ID] = task
			}
			if task.EndDate.After(baseDate) {
				baseDate = task.EndDate
			}
			continue
		}
	}

	if len(gd.Tasks) == 0 {
		return nil, fmt.Errorf("no tasks found")
	}

	return gd, nil
}

func parseTask(name, spec string, baseDate time.Time, tasksByID map[string]*Task, dateFormat string) (*Task, error) {
	task := &Task{Name: strings.TrimSpace(name)}

	parts := strings.Split(spec, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	idx := 0

	// Parse optional status tags
	for idx < len(parts) {
		p := parts[idx]
		if p == "done" || p == "active" || p == "crit" {
			if task.Status != "" {
				task.Status += "," + p
			} else {
				task.Status = p
			}
			idx++
		} else {
			break
		}
	}

	// Parse optional ID
	if idx < len(parts) {
		p := parts[idx]
		if !isDateLike(p, dateFormat) && !isDuration(p) && !strings.HasPrefix(p, "after ") {
			task.ID = p
			idx++
		}
	}

	// Parse start: date, "after taskID", or implied
	if idx < len(parts) {
		p := parts[idx]
		if strings.HasPrefix(p, "after ") {
			depID := strings.TrimPrefix(p, "after ")
			task.After = depID
			if dep, ok := tasksByID[depID]; ok {
				task.StartDate = dep.EndDate
			} else {
				task.StartDate = baseDate
			}
			idx++
		} else if d, err := parseDate(p, dateFormat); err == nil {
			task.StartDate = d
			idx++
		} else {
			task.StartDate = baseDate
		}
	} else {
		task.StartDate = baseDate
	}

	// Parse duration or end date
	if idx < len(parts) {
		p := parts[idx]
		if dur, err := parseDuration(p); err == nil {
			task.Duration = dur
			task.EndDate = task.StartDate.Add(dur)
		} else if d, err := parseDate(p, dateFormat); err == nil {
			task.EndDate = d
			task.Duration = d.Sub(task.StartDate)
		} else {
			// Default 1 day
			task.Duration = 24 * time.Hour
			task.EndDate = task.StartDate.Add(task.Duration)
		}
	} else {
		task.Duration = 24 * time.Hour
		task.EndDate = task.StartDate.Add(task.Duration)
	}

	return task, nil
}

func parseDate(s string, format string) (time.Time, error) {
	goFormat := mermaidToGoDateFormat(format)
	return time.Parse(goFormat, s)
}

func mermaidToGoDateFormat(format string) string {
	r := strings.NewReplacer(
		"YYYY", "2006",
		"YY", "06",
		"MM", "01",
		"DD", "02",
		"HH", "15",
		"mm", "04",
		"ss", "05",
	)
	return r.Replace(format)
}

func isDateLike(s, format string) bool {
	_, err := parseDate(s, format)
	return err == nil
}

func isDuration(s string) bool {
	_, err := parseDuration(s)
	return err == nil
}

func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	durationRegex := regexp.MustCompile(`^(\d+)(d|h|m|w)$`)
	match := durationRegex.FindStringSubmatch(s)
	if match == nil {
		return 0, fmt.Errorf("invalid duration: %s", s)
	}
	var multiplier time.Duration
	switch match[2] {
	case "d":
		multiplier = 24 * time.Hour
	case "h":
		multiplier = time.Hour
	case "m":
		multiplier = time.Minute
	case "w":
		multiplier = 7 * 24 * time.Hour
	}
	n := 0
	fmt.Sscanf(match[1], "%d", &n)
	return time.Duration(n) * multiplier, nil
}
