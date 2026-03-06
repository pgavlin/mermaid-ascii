package zenuml

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

const ZenUMLKeyword = "zenuml"

var (
	// participantDeclRegex matches "Type Name" declarations like "Client client"
	participantDeclRegex = regexp.MustCompile(`^\s*(\w+)\s+(\w+)\s*$`)

	// syncMessageRegex matches "target.method(args)" without trailing brace
	syncMessageRegex = regexp.MustCompile(`^\s*(\w+)\.(\w+)\(([^)]*)\)\s*$`)

	// asyncMessageRegex matches "target.method(args) {" (async block start)
	asyncMessageRegex = regexp.MustCompile(`^\s*(\w+)\.(\w+)\(([^)]*)\)\s*\{\s*$`)

	// returnRegex matches "return value"
	returnRegex = regexp.MustCompile(`^\s*return\s+(.*?)\s*$`)

	// closeBraceRegex matches a closing brace
	closeBraceRegex = regexp.MustCompile(`^\s*\}\s*$`)
)

// MessageType distinguishes sync, async, and return messages.
type MessageType int

const (
	SyncMessage  MessageType = iota
	AsyncMessage
	ReturnMessage
)

// Participant represents a participant in the ZenUML diagram.
type Participant struct {
	TypeName string // declared type, e.g. "Client"
	ID       string // identifier, e.g. "client"
	Index    int
}

// Message represents a message/call in the ZenUML diagram.
type Message struct {
	From   *Participant
	To     *Participant
	Method string
	Args   string
	Type   MessageType
	Label  string     // used for return value text
	Nested []*Message // nested messages inside async blocks
}

// ZenUMLDiagram represents a parsed ZenUML diagram.
type ZenUMLDiagram struct {
	Participants []*Participant
	Messages     []*Message
}

// IsZenUML returns true if the input starts with the zenuml keyword.
func IsZenUML(input string) bool {
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}
		return strings.HasPrefix(trimmed, ZenUMLKeyword)
	}
	return false
}

// Parse parses ZenUML input text into a ZenUMLDiagram.
func Parse(input string) (*ZenUMLDiagram, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty input")
	}

	rawLines := diagram.SplitLines(input)
	lines := diagram.RemoveComments(rawLines)
	if len(lines) == 0 {
		return nil, fmt.Errorf("no content found")
	}

	if !strings.HasPrefix(strings.TrimSpace(lines[0]), ZenUMLKeyword) {
		return nil, fmt.Errorf("expected %q keyword", ZenUMLKeyword)
	}
	lines = lines[1:]

	d := &ZenUMLDiagram{
		Participants: []*Participant{},
		Messages:     []*Message{},
	}
	participantMap := make(map[string]*Participant)

	messages, _, err := parseLines(lines, d, participantMap, false)
	if err != nil {
		return nil, err
	}
	d.Messages = messages

	if len(d.Participants) == 0 {
		return nil, fmt.Errorf("no participants found")
	}

	return d, nil
}

// parseLines parses lines into messages. When inBlock is true, parsing stops at '}'.
func parseLines(lines []string, d *ZenUMLDiagram, pMap map[string]*Participant, inBlock bool) ([]*Message, int, error) {
	var messages []*Message
	i := 0
	for i < len(lines) {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" {
			i++
			continue
		}

		// Block end
		if inBlock && closeBraceRegex.MatchString(trimmed) {
			return messages, i, nil
		}

		// Return statement
		if match := returnRegex.FindStringSubmatch(trimmed); match != nil {
			// Return goes from the last callee back to the caller
			var from, to *Participant
			if len(messages) > 0 {
				last := messages[len(messages)-1]
				from = last.To
				to = last.From
			} else if len(d.Participants) > 0 {
				from = d.Participants[0]
			}
			msg := &Message{
				From:  from,
				To:    to,
				Label: strings.TrimSpace(match[1]),
				Type:  ReturnMessage,
			}
			messages = append(messages, msg)
			i++
			continue
		}

		// Async message: target.method(args) {
		if match := asyncMessageRegex.FindStringSubmatch(trimmed); match != nil {
			target := match[1]
			method := match[2]
			args := strings.TrimSpace(match[3])

			to := getOrCreateParticipant(target, d, pMap)
			from := inferCaller(d, to)

			i++
			nested, consumed, err := parseLines(lines[i:], d, pMap, true)
			if err != nil {
				return nil, 0, err
			}
			i += consumed
			if i < len(lines) {
				i++ // consume '}'
			}

			msg := &Message{
				From:   from,
				To:     to,
				Method: method,
				Args:   args,
				Type:   AsyncMessage,
				Nested: nested,
			}
			messages = append(messages, msg)
			continue
		}

		// Sync message: target.method(args)
		if match := syncMessageRegex.FindStringSubmatch(trimmed); match != nil {
			target := match[1]
			method := match[2]
			args := strings.TrimSpace(match[3])

			to := getOrCreateParticipant(target, d, pMap)
			from := inferCaller(d, to)

			msg := &Message{
				From:   from,
				To:     to,
				Method: method,
				Args:   args,
				Type:   SyncMessage,
			}
			messages = append(messages, msg)
			i++
			continue
		}

		// Participant declaration: Type Name
		if match := participantDeclRegex.FindStringSubmatch(trimmed); match != nil {
			typeName := match[1]
			id := match[2]

			if isReservedWord(typeName) {
				return nil, 0, fmt.Errorf("unexpected line: %q", trimmed)
			}

			if _, exists := pMap[id]; !exists {
				p := &Participant{
					TypeName: typeName,
					ID:       id,
					Index:    len(d.Participants),
				}
				d.Participants = append(d.Participants, p)
				pMap[id] = p
			}
			i++
			continue
		}

		return nil, 0, fmt.Errorf("unexpected line: %q", trimmed)
	}

	return messages, i, nil
}

// inferCaller returns the first declared participant as the default caller,
// as long as it is not the same as the target. If only one participant exists,
// it returns that participant (self-call).
func inferCaller(d *ZenUMLDiagram, target *Participant) *Participant {
	if len(d.Participants) == 0 {
		return nil
	}
	// Use first participant as default caller
	return d.Participants[0]
}

func getOrCreateParticipant(id string, d *ZenUMLDiagram, pMap map[string]*Participant) *Participant {
	if p, exists := pMap[id]; exists {
		return p
	}
	p := &Participant{
		TypeName: id,
		ID:       id,
		Index:    len(d.Participants),
	}
	d.Participants = append(d.Participants, p)
	pMap[id] = p
	return p
}

func isReservedWord(s string) bool {
	switch strings.ToLower(s) {
	case "return", "zenuml":
		return true
	}
	return false
}
