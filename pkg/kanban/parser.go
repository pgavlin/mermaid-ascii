package kanban

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

const KanbanKeyword = "kanban"

var (
	columnRegex = regexp.MustCompile(`^\s*(.+)\s*$`)
	cardRegex   = regexp.MustCompile(`^\s+(.+)\s*$`)
)

type KanbanBoard struct {
	Columns []*Column
}

type Column struct {
	Name  string
	Cards []*Card
}

type Card struct {
	Title string
}

func IsKanbanBoard(input string) bool {
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}
		return trimmed == KanbanKeyword
	}
	return false
}

func Parse(input string) (*KanbanBoard, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty input")
	}

	rawLines := diagram.SplitLines(input)
	lines := diagram.RemoveComments(rawLines)
	if len(lines) == 0 {
		return nil, fmt.Errorf("no content found")
	}

	if strings.TrimSpace(lines[0]) != KanbanKeyword {
		return nil, fmt.Errorf("expected %q keyword", KanbanKeyword)
	}
	lines = lines[1:]

	kb := &KanbanBoard{Columns: []*Column{}}
	var currentColumn *Column

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// If line starts with whitespace, it's a card
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			text := strings.TrimSpace(line)
			if text == "" {
				continue
			}
			if currentColumn == nil {
				currentColumn = &Column{Name: "Default", Cards: []*Card{}}
				kb.Columns = append(kb.Columns, currentColumn)
			}
			currentColumn.Cards = append(currentColumn.Cards, &Card{Title: text})
		} else {
			// Column header
			text := strings.TrimSpace(line)
			currentColumn = &Column{Name: text, Cards: []*Card{}}
			kb.Columns = append(kb.Columns, currentColumn)
		}
	}

	if len(kb.Columns) == 0 {
		return nil, fmt.Errorf("no columns found")
	}

	return kb, nil
}
