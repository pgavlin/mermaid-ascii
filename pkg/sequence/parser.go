package sequence

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

const (
	SequenceDiagramKeyword = "sequenceDiagram"
	SolidArrowSyntax       = "->>"
	DottedArrowSyntax      = "-->>"
)

var (
	// participantRegex matches participant/actor declarations: participant [ID] [as Label]
	participantRegex = regexp.MustCompile(`^\s*(?:participant|actor)\s+(?:"([^"]+)"|(\S+))(?:\s+as\s+(.+))?$`)

	// participantTypeRegex extracts the keyword (participant or actor)
	participantTypeRegex = regexp.MustCompile(`^\s*(participant|actor)\s+`)

	// messageRegex matches messages with all arrow types including activation modifiers:
	// [From](arrow)[+|-][To]: [Label]
	// Arrow types: ->>, -->>, ->, -->, -x, --x, -), --)
	messageRegex = regexp.MustCompile(`^\s*(?:"([^"]+)"|([^\s\-]+))\s*(-->>|->>|-->|->|--x|-x|--\)|-\))\s*(\+|-)?(?:"([^"]+)"|([^\s:]+))\s*:\s*(.*)$`)

	// autonumberRegex matches the autonumber directive
	autonumberRegex = regexp.MustCompile(`^\s*autonumber\s*$`)

	// activateRegex matches activate/deactivate directives
	activateRegex = regexp.MustCompile(`^\s*(activate|deactivate)\s+(?:"([^"]+)"|(\S+))\s*$`)

	// noteRegex matches Note directives:
	// Note right of A: text
	// Note left of A: text
	// Note over A: text
	// Note over A,B: text
	noteRegex = regexp.MustCompile(`^\s*[Nn]ote\s+(right of|left of|over)\s+(?:"([^"]+)"|([^\s,:]+))(?:\s*,\s*(?:"([^"]+)"|([^\s:]+)))?\s*:\s*(.*)$`)

	// blockStartRegex matches block start: loop text, alt text, opt text, par text, critical text, break text, rect text
	blockStartRegex = regexp.MustCompile(`^\s*(loop|alt|opt|par|critical|break|rect)\s+(.*)$`)

	// blockDividerRegex matches block dividers: else text, and text, option text
	blockDividerRegex = regexp.MustCompile(`^\s*(else|and|option)\s*(.*)$`)

	// blockEndRegex matches block end: end
	blockEndRegex = regexp.MustCompile(`^\s*end\s*$`)

	// boxRegex matches box directive: box [label] [color]
	boxRegex = regexp.MustCompile(`^\s*box\s*(?:"([^"]+)")?(?:\s+(\S+))?\s*$`)

	// boxEndRegex matches end of box grouping (same as blockEndRegex but used in box context)
	boxEndRegex = regexp.MustCompile(`^\s*end\s*$`)

	// createRegex matches create participant/actor directives
	createRegex = regexp.MustCompile(`^\s*create\s+(?:participant|actor)\s+(?:"([^"]+)"|(\S+))(?:\s+as\s+(.+))?\s*$`)

	// destroyRegex matches destroy directives
	destroyRegex = regexp.MustCompile(`^\s*destroy\s+(?:"([^"]+)"|(\S+))\s*$`)
)

// ParticipantGroup represents a box grouping of participants.
type ParticipantGroup struct {
	Label        string
	Participants []*Participant
}

// CreateEvent represents a create participant directive (participant box appears mid-diagram).
type CreateEvent struct {
	Participant *Participant
}

func (c *CreateEvent) elementType() string { return "create" }

// DestroyEvent represents a destroy directive (lifeline ends with X).
type DestroyEvent struct {
	Participant *Participant
}

func (d *DestroyEvent) elementType() string { return "destroy" }

// SequenceDiagram represents a parsed sequence diagram.
type SequenceDiagram struct {
	Participants []*Participant
	Messages     []*Message
	Autonumber   bool
	Elements     []Element // Ordered list of all elements (messages, notes, activations, blocks)
	Groups       []*ParticipantGroup
}

// Element is an interface for things that appear in the sequence diagram's vertical ordering.
type Element interface {
	elementType() string
}

type ParticipantType int

const (
	ParticipantBox   ParticipantType = iota
	ParticipantActor                         // Rendered as stick figure
)

type Participant struct {
	ID    string
	Label string
	Index int
	Type  ParticipantType
}

type Message struct {
	From       *Participant
	To         *Participant
	Label      string
	ArrowType  ArrowType
	Number     int // Message number when autonumber is enabled (0 means no number)
	Activate   bool // +  shorthand: activate target after message
	Deactivate bool // - shorthand: deactivate source after message
}

func (m *Message) elementType() string { return "message" }

type ArrowType int

const (
	SolidArrow      ArrowType = iota // ->>  solid with filled arrowhead
	DottedArrow                       // -->> dotted with filled arrowhead
	SolidOpen                         // ->   solid with open arrowhead
	DottedOpen                        // -->  dotted with open arrowhead
	SolidCross                        // -x   solid with cross end
	DottedCross                       // --x  dotted with cross end
	SolidAsync                        // -)   solid with open arrow (async)
	DottedAsync                       // --)  dotted with open arrow (async)
)

func (a ArrowType) String() string {
	switch a {
	case SolidArrow:
		return "solid"
	case DottedArrow:
		return "dotted"
	case SolidOpen:
		return "solid_open"
	case DottedOpen:
		return "dotted_open"
	case SolidCross:
		return "solid_cross"
	case DottedCross:
		return "dotted_cross"
	case SolidAsync:
		return "solid_async"
	case DottedAsync:
		return "dotted_async"
	default:
		return fmt.Sprintf("ArrowType(%d)", a)
	}
}

// IsDotted returns true if the arrow type uses a dotted line style.
func (a ArrowType) IsDotted() bool {
	return a == DottedArrow || a == DottedOpen || a == DottedCross || a == DottedAsync
}

// NotePosition indicates where a note is placed.
type NotePosition int

const (
	NoteRightOf NotePosition = iota
	NoteLeftOf
	NoteOver
)

// Note represents a note in the sequence diagram.
type Note struct {
	Position    NotePosition
	Participant *Participant
	EndParticipant *Participant // non-nil for "Note over A,B"
	Text        string
}

func (n *Note) elementType() string { return "note" }

// ActivationEvent represents an activate/deactivate directive.
type ActivationEvent struct {
	Participant *Participant
	Activate    bool // true = activate, false = deactivate
}

func (a *ActivationEvent) elementType() string { return "activation" }

// Block represents an interaction block (loop, alt, opt, par, critical, break, rect).
type Block struct {
	Type     string  // "loop", "alt", "opt", "par", "critical", "break", "rect"
	Label    string
	Elements []Element
	Sections []*BlockSection // For alt/par/critical: additional sections (else/and/option)
}

func (b *Block) elementType() string { return "block" }

// BlockSection represents an else/and/option section within a block.
type BlockSection struct {
	Label    string
	Elements []Element
}

func IsSequenceDiagram(input string) bool {
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}
		return strings.HasPrefix(trimmed, SequenceDiagramKeyword)
	}
	return false
}

func Parse(input string) (*SequenceDiagram, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty input")
	}

	rawLines := diagram.SplitLines(input)
	lines := diagram.RemoveComments(rawLines)
	if len(lines) == 0 {
		return nil, fmt.Errorf("no content found")
	}

	if !strings.HasPrefix(strings.TrimSpace(lines[0]), SequenceDiagramKeyword) {
		return nil, fmt.Errorf("expected %q keyword", SequenceDiagramKeyword)
	}
	lines = lines[1:]

	sd := &SequenceDiagram{
		Participants: []*Participant{},
		Messages:     []*Message{},
		Elements:     []Element{},
		Autonumber:   false,
		Groups:       []*ParticipantGroup{},
	}
	participantMap := make(map[string]*Participant)

	elements, _, err := sd.parseLines(lines, participantMap, 0, false)
	if err != nil {
		return nil, err
	}
	sd.Elements = elements

	if len(sd.Participants) == 0 {
		return nil, fmt.Errorf("no participants found")
	}

	return sd, nil
}

// parseLines parses a slice of lines and returns elements. When inBlock is true,
// it stops at "end" or block dividers (else/and/option).
func (sd *SequenceDiagram) parseLines(lines []string, participantMap map[string]*Participant, startLine int, inBlock bool) ([]Element, int, error) {
	var elements []Element
	var currentGroup *ParticipantGroup // tracks current box group
	i := 0
	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		lineNum := startLine + i + 2

		if trimmed == "" {
			i++
			continue
		}

		// Check for autonumber directive
		if autonumberRegex.MatchString(trimmed) {
			sd.Autonumber = true
			i++
			continue
		}

		// Check for box end (end while in a box group context, but NOT in a block)
		if currentGroup != nil && boxEndRegex.MatchString(trimmed) {
			sd.Groups = append(sd.Groups, currentGroup)
			currentGroup = nil
			i++
			continue
		}

		// Check for block end
		if inBlock && blockEndRegex.MatchString(trimmed) {
			return elements, i, nil
		}

		// Check for block dividers (else/and/option) - only valid inside a block
		if inBlock && blockDividerRegex.MatchString(trimmed) {
			return elements, i, nil
		}

		// Check for box directive (participant grouping)
		if match := boxRegex.FindStringSubmatch(trimmed); match != nil && !inBlock {
			label := match[1] // quoted label
			if label == "" {
				label = match[2] // unquoted color/label
			}
			currentGroup = &ParticipantGroup{
				Label:        label,
				Participants: []*Participant{},
			}
			i++
			continue
		}

		// Check for create participant/actor
		if match := createRegex.FindStringSubmatch(trimmed); match != nil {
			id := match[2]
			if match[1] != "" {
				id = match[1]
			}
			label := match[3]
			if label == "" {
				label = id
			}
			label = strings.Trim(label, `"`)

			// Determine participant type
			pType := ParticipantBox
			if strings.Contains(trimmed, "create actor") {
				pType = ParticipantActor
			}

			if _, exists := participantMap[id]; exists {
				return nil, 0, fmt.Errorf("line %d: duplicate participant %q", lineNum, id)
			}

			p := &Participant{
				ID:    id,
				Label: label,
				Index: len(sd.Participants),
				Type:  pType,
			}
			sd.Participants = append(sd.Participants, p)
			participantMap[id] = p

			if currentGroup != nil {
				currentGroup.Participants = append(currentGroup.Participants, p)
			}

			elements = append(elements, &CreateEvent{Participant: p})
			i++
			continue
		}

		// Check for destroy
		if match := destroyRegex.FindStringSubmatch(trimmed); match != nil {
			id := match[2]
			if match[1] != "" {
				id = match[1]
			}
			p := sd.getParticipant(id, participantMap)
			elements = append(elements, &DestroyEvent{Participant: p})
			i++
			continue
		}

		// Check for block start
		if match := blockStartRegex.FindStringSubmatch(trimmed); match != nil {
			block := &Block{
				Type:  match[1],
				Label: strings.TrimSpace(match[2]),
			}
			i++
			// Parse elements within this block
			blockElements, consumed, err := sd.parseLines(lines[i:], participantMap, startLine+i, true)
			if err != nil {
				return nil, 0, err
			}
			block.Elements = blockElements
			i += consumed

			// Parse additional sections (else/and/option)
			for i < len(lines) {
				divTrimmed := strings.TrimSpace(lines[i])
				if blockDividerRegex.MatchString(divTrimmed) {
					divMatch := blockDividerRegex.FindStringSubmatch(divTrimmed)
					section := &BlockSection{
						Label: strings.TrimSpace(divMatch[2]),
					}
					i++
					sectionElements, consumed2, err := sd.parseLines(lines[i:], participantMap, startLine+i, true)
					if err != nil {
						return nil, 0, err
					}
					section.Elements = sectionElements
					block.Sections = append(block.Sections, section)
					i += consumed2
				} else if blockEndRegex.MatchString(divTrimmed) {
					i++ // consume the "end"
					break
				} else {
					break
				}
			}
			elements = append(elements, block)
			continue
		}

		// Check for activate/deactivate
		if match := activateRegex.FindStringSubmatch(trimmed); match != nil {
			id := match[3]
			if match[2] != "" {
				id = match[2]
			}
			p := sd.getParticipant(id, participantMap)
			event := &ActivationEvent{
				Participant: p,
				Activate:    match[1] == "activate",
			}
			elements = append(elements, event)
			i++
			continue
		}

		// Check for note
		if match := noteRegex.FindStringSubmatch(trimmed); match != nil {
			posStr := match[1]
			pID := match[3]
			if match[2] != "" {
				pID = match[2]
			}
			text := strings.TrimSpace(match[6])

			p := sd.getParticipant(pID, participantMap)

			var pos NotePosition
			switch posStr {
			case "right of":
				pos = NoteRightOf
			case "left of":
				pos = NoteLeftOf
			case "over":
				pos = NoteOver
			}

			note := &Note{
				Position:    pos,
				Participant: p,
				Text:        text,
			}

			// Check for second participant in "Note over A,B"
			endID := match[5]
			if match[4] != "" {
				endID = match[4]
			}
			if endID != "" {
				note.EndParticipant = sd.getParticipant(endID, participantMap)
			}

			elements = append(elements, note)
			i++
			continue
		}

		if matched, err := sd.parseParticipant(trimmed, participantMap); err != nil {
			return nil, 0, fmt.Errorf("line %d: %w", lineNum, err)
		} else if matched {
			// If we're inside a box group, add this participant to it
			if currentGroup != nil {
				currentGroup.Participants = append(currentGroup.Participants, sd.Participants[len(sd.Participants)-1])
			}
			i++
			continue
		}

		if matched, err := sd.parseMessage(trimmed, participantMap, &elements); err != nil {
			return nil, 0, fmt.Errorf("line %d: %w", lineNum, err)
		} else if matched {
			i++
			continue
		}

		return nil, 0, fmt.Errorf("line %d: invalid syntax: %q", lineNum, trimmed)
	}

	return elements, i, nil
}

func (sd *SequenceDiagram) parseParticipant(line string, participants map[string]*Participant) (bool, error) {
	match := participantRegex.FindStringSubmatch(line)
	if match == nil {
		return false, nil
	}

	// Determine participant type
	pType := ParticipantBox
	if typeMatch := participantTypeRegex.FindStringSubmatch(line); typeMatch != nil {
		if typeMatch[1] == "actor" {
			pType = ParticipantActor
		}
	}

	id := match[2]
	if match[1] != "" {
		id = match[1]
	}
	label := match[3]
	if label == "" {
		label = id
	}
	label = strings.Trim(label, `"`)

	if _, exists := participants[id]; exists {
		return true, fmt.Errorf("duplicate participant %q", id)
	}

	p := &Participant{
		ID:    id,
		Label: label,
		Index: len(sd.Participants),
		Type:  pType,
	}
	sd.Participants = append(sd.Participants, p)
	participants[id] = p
	return true, nil
}

func (sd *SequenceDiagram) parseMessage(line string, participants map[string]*Participant, elements *[]Element) (bool, error) {
	match := messageRegex.FindStringSubmatch(line)
	if match == nil {
		return false, nil
	}

	fromID := match[2]
	if match[1] != "" {
		fromID = match[1]
	}

	arrow := match[3]

	activationMod := match[4] // "+" or "-" or ""

	toID := match[6]
	if match[5] != "" {
		toID = match[5]
	}

	label := strings.TrimSpace(match[7])

	from := sd.getParticipant(fromID, participants)
	to := sd.getParticipant(toID, participants)

	var aType ArrowType
	switch arrow {
	case "->>":
		aType = SolidArrow
	case "-->>":
		aType = DottedArrow
	case "->":
		aType = SolidOpen
	case "-->":
		aType = DottedOpen
	case "-x":
		aType = SolidCross
	case "--x":
		aType = DottedCross
	case "-)":
		aType = SolidAsync
	case "--)":
		aType = DottedAsync
	}

	msgNumber := 0
	if sd.Autonumber {
		msgNumber = len(sd.Messages) + 1
	}

	msg := &Message{
		From:       from,
		To:         to,
		Label:      label,
		ArrowType:  aType,
		Number:     msgNumber,
		Activate:   activationMod == "+",
		Deactivate: activationMod == "-",
	}
	sd.Messages = append(sd.Messages, msg)
	*elements = append(*elements, msg)

	// Generate activation/deactivation events from +/- shorthand
	if msg.Activate {
		*elements = append(*elements, &ActivationEvent{Participant: to, Activate: true})
	}
	if msg.Deactivate {
		*elements = append(*elements, &ActivationEvent{Participant: from, Activate: false})
	}

	return true, nil
}

func (sd *SequenceDiagram) getParticipant(id string, participants map[string]*Participant) *Participant {
	if p, exists := participants[id]; exists {
		return p
	}

	p := &Participant{
		ID:    id,
		Label: id,
		Index: len(sd.Participants),
	}
	sd.Participants = append(sd.Participants, p)
	participants[id] = p
	return p
}
