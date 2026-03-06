package statediagram

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

const StateDiagramKeyword = "stateDiagram-v2"

var (
	// transitionRegex matches: State1 --> State2 or State1 --> State2 : trigger
	// Also matches [*] as the start/end pseudo-state
	transitionRegex = regexp.MustCompile(`^\s*(\[\*\]|[\w]+)\s*-->\s*(\[\*\]|[\w]+)\s*(?::\s*(.+))?\s*$`)

	// stateAliasRegex matches: state "Description" as s1
	stateAliasRegex = regexp.MustCompile(`^\s*state\s+"([^"]+)"\s+as\s+(\w+)\s*$`)

	// stateBlockRegex matches: state "name" as s1 { or state s1 {
	stateBlockRegex = regexp.MustCompile(`^\s*state\s+(?:"([^"]+)"\s+as\s+)?(\w+)\s*\{\s*$`)

	// noteRegex matches: note right of State1 : text or note left of State1 : text
	noteRegex = regexp.MustCompile(`^\s*note\s+(right|left)\s+of\s+(\w+)\s*:\s*(.+)\s*$`)

	// closingBraceRegex matches a closing brace
	closingBraceRegex = regexp.MustCompile(`^\s*\}\s*$`)

	// stateInlineRegex matches: state StateName
	stateInlineRegex = regexp.MustCompile(`^\s*state\s+(\w+)\s*$`)
)

// State represents a state in the diagram.
type State struct {
	ID          string
	Label       string
	IsStart     bool
	IsEnd       bool
	Children    []*State
	Index       int
}

// Transition represents a transition between states.
type Transition struct {
	From    *State
	To      *State
	Trigger string
}

// NotePosition represents where a note is placed.
type NotePosition int

const (
	NoteRight NotePosition = iota
	NoteLeft
)

// Note represents a note attached to a state.
type Note struct {
	Position NotePosition
	State    *State
	Text     string
}

// StateDiagram represents a parsed state diagram.
type StateDiagram struct {
	States      []*State
	Transitions []*Transition
	Notes       []*Note
	stateMap    map[string]*State
}

// IsStateDiagram returns true if the input begins with the stateDiagram-v2 keyword.
func IsStateDiagram(input string) bool {
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}
		return strings.HasPrefix(trimmed, StateDiagramKeyword)
	}
	return false
}

// Parse parses a state diagram from the given input string.
func Parse(input string) (*StateDiagram, error) {
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
	if !strings.HasPrefix(first, StateDiagramKeyword) {
		return nil, fmt.Errorf("expected %q keyword", StateDiagramKeyword)
	}
	lines = lines[1:]

	sd := &StateDiagram{
		States:      []*State{},
		Transitions: []*Transition{},
		Notes:       []*Note{},
		stateMap:    make(map[string]*State),
	}

	_, err := sd.parseLines(lines)
	if err != nil {
		return nil, err
	}

	if len(sd.States) == 0 {
		return nil, fmt.Errorf("no states found")
	}

	return sd, nil
}

func (sd *StateDiagram) parseLines(lines []string) (int, error) {
	i := 0
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			i++
			continue
		}

		// Check for closing brace (end of composite state)
		if closingBraceRegex.MatchString(line) {
			return i, nil
		}

		// Check for state block: state "label" as id { or state id {
		if match := stateBlockRegex.FindStringSubmatch(line); match != nil {
			label := match[1]
			id := match[2]
			if label == "" {
				label = id
			}
			state := sd.getOrCreateState(id)
			state.Label = label
			i++
			// Parse children
			consumed, err := sd.parseLines(lines[i:])
			if err != nil {
				return 0, err
			}
			i += consumed
			// Skip closing brace
			if i < len(lines) && closingBraceRegex.MatchString(strings.TrimSpace(lines[i])) {
				i++
			}
			continue
		}

		// Check for state alias: state "Description" as s1
		if match := stateAliasRegex.FindStringSubmatch(line); match != nil {
			label := match[1]
			id := match[2]
			state := sd.getOrCreateState(id)
			state.Label = label
			i++
			continue
		}

		// Check for inline state: state StateName
		if match := stateInlineRegex.FindStringSubmatch(line); match != nil {
			sd.getOrCreateState(match[1])
			i++
			continue
		}

		// Check for note
		if match := noteRegex.FindStringSubmatch(line); match != nil {
			posStr := match[1]
			stateID := match[2]
			text := strings.TrimSpace(match[3])

			state := sd.getOrCreateState(stateID)
			pos := NoteRight
			if posStr == "left" {
				pos = NoteLeft
			}
			sd.Notes = append(sd.Notes, &Note{
				Position: pos,
				State:    state,
				Text:     text,
			})
			i++
			continue
		}

		// Check for transition
		if match := transitionRegex.FindStringSubmatch(line); match != nil {
			fromID := match[1]
			toID := match[2]
			trigger := strings.TrimSpace(match[3])

			from := sd.getOrCreateState(fromID)
			to := sd.getOrCreateState(toID)
			sd.Transitions = append(sd.Transitions, &Transition{
				From:    from,
				To:      to,
				Trigger: trigger,
			})
			i++
			continue
		}

		// Unknown line, skip
		i++
	}

	return i, nil
}

func (sd *StateDiagram) getOrCreateState(id string) *State {
	if s, ok := sd.stateMap[id]; ok {
		return s
	}

	isStart := false
	isEnd := false
	label := id

	// [*] is the start/end pseudo-state
	if id == "[*]" {
		isStart = true
		isEnd = true
		label = "[*]"
	}

	s := &State{
		ID:      id,
		Label:   label,
		IsStart: isStart,
		IsEnd:   isEnd,
		Index:   len(sd.States),
	}
	sd.States = append(sd.States, s)
	sd.stateMap[id] = s
	return s
}
