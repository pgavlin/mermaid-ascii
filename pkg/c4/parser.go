package c4

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

var (
	// Matches C4Context, C4Container, C4Component, C4Dynamic
	c4KeywordRegex = regexp.MustCompile(`^\s*(C4Context|C4Container|C4Component|C4Dynamic)\s*$`)

	// Person(alias, "label", "description")
	personRegex = regexp.MustCompile(`^\s*Person(?:_Ext)?\(\s*(\w+)\s*,\s*"([^"]*)"\s*(?:,\s*"([^"]*)")?\s*\)`)

	// System(alias, "label", "description")
	systemRegex = regexp.MustCompile(`^\s*System(?:_Ext|_Boundary)?\(\s*(\w+)\s*,\s*"([^"]*)"\s*(?:,\s*"([^"]*)")?\s*\)`)

	// Container(alias, "label", "tech", "description")
	containerRegex = regexp.MustCompile(`^\s*Container(?:_Ext|Db|Queue|_Boundary)?\(\s*(\w+)\s*,\s*"([^"]*)"\s*(?:,\s*"([^"]*)")?\s*(?:,\s*"([^"]*)")?\s*\)`)

	// Component(alias, "label", "tech", "description")
	componentRegex = regexp.MustCompile(`^\s*Component(?:_Ext|Db|Queue)?\(\s*(\w+)\s*,\s*"([^"]*)"\s*(?:,\s*"([^"]*)")?\s*(?:,\s*"([^"]*)")?\s*\)`)

	// Boundary(alias, "label") {
	boundaryRegex = regexp.MustCompile(`^\s*(?:Enterprise_Boundary|System_Boundary|Container_Boundary|Boundary)\(\s*(\w+)\s*,\s*"([^"]*)"\s*\)\s*\{?\s*$`)

	// Rel(from, to, "label", "tech")
	relRegex = regexp.MustCompile(`^\s*(?:Rel|Rel_Back|Rel_Neighbor|Rel_Back_Neighbor|BiRel)\(\s*(\w+)\s*,\s*(\w+)\s*,\s*"([^"]*)"\s*(?:,\s*"([^"]*)")?\s*\)`)

	// end of boundary block
	endRegex = regexp.MustCompile(`^\s*\}\s*$`)
)

// C4DiagramType represents the specific C4 diagram type.
type C4DiagramType int

const (
	C4Context   C4DiagramType = iota
	C4Container
	C4Component
	C4Dynamic
)

// Element represents a C4 element (Person, System, Container, Component).
type Element struct {
	Alias       string
	Label       string
	Description string
	Technology  string
	Kind        string // "person", "system", "container", "component"
}

// Relationship represents a C4 relationship.
type Relationship struct {
	From       string
	To         string
	Label      string
	Technology string
}

// Boundary represents a C4 boundary containing elements.
type Boundary struct {
	Alias    string
	Label    string
	Elements []*Element
	Boundaries []*Boundary
}

// C4Diagram represents a parsed C4 diagram.
type C4Diagram struct {
	DiagramType  C4DiagramType
	Elements     []*Element
	Relationships []*Relationship
	Boundaries   []*Boundary
}

// IsC4Diagram returns true if the input starts with a C4 diagram keyword.
func IsC4Diagram(input string) bool {
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}
		return c4KeywordRegex.MatchString(trimmed)
	}
	return false
}

// Parse parses a C4 diagram from Mermaid-style input.
func Parse(input string) (*C4Diagram, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty input")
	}

	rawLines := diagram.SplitLines(input)
	lines := diagram.RemoveComments(rawLines)
	if len(lines) == 0 {
		return nil, fmt.Errorf("no content found")
	}

	match := c4KeywordRegex.FindStringSubmatch(strings.TrimSpace(lines[0]))
	if match == nil {
		return nil, fmt.Errorf("expected C4 diagram keyword")
	}

	var dtype C4DiagramType
	switch match[1] {
	case "C4Context":
		dtype = C4Context
	case "C4Container":
		dtype = C4Container
	case "C4Component":
		dtype = C4Component
	case "C4Dynamic":
		dtype = C4Dynamic
	}

	d := &C4Diagram{
		DiagramType: dtype,
	}

	_, err := parseC4Lines(d, lines[1:], nil)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func parseC4Lines(d *C4Diagram, lines []string, currentBoundary *Boundary) (int, error) {
	i := 0
	for i < len(lines) {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" {
			i++
			continue
		}

		// End of boundary block
		if endRegex.MatchString(trimmed) {
			return i + 1, nil
		}

		// Boundary
		if m := boundaryRegex.FindStringSubmatch(trimmed); m != nil {
			b := &Boundary{
				Alias: m[1],
				Label: m[2],
			}
			i++
			consumed, err := parseC4Lines(d, lines[i:], b)
			if err != nil {
				return 0, err
			}
			i += consumed
			if currentBoundary != nil {
				currentBoundary.Boundaries = append(currentBoundary.Boundaries, b)
			} else {
				d.Boundaries = append(d.Boundaries, b)
			}
			continue
		}

		// Person
		if m := personRegex.FindStringSubmatch(trimmed); m != nil {
			elem := &Element{Alias: m[1], Label: m[2], Description: m[3], Kind: "person"}
			addElement(d, currentBoundary, elem)
			i++
			continue
		}

		// Component (check before Container since Container regex might match)
		if m := componentRegex.FindStringSubmatch(trimmed); m != nil && strings.Contains(trimmed, "Component") {
			elem := &Element{Alias: m[1], Label: m[2], Technology: m[3], Description: m[4], Kind: "component"}
			addElement(d, currentBoundary, elem)
			i++
			continue
		}

		// Container
		if m := containerRegex.FindStringSubmatch(trimmed); m != nil && strings.Contains(trimmed, "Container") {
			elem := &Element{Alias: m[1], Label: m[2], Technology: m[3], Description: m[4], Kind: "container"}
			addElement(d, currentBoundary, elem)
			i++
			continue
		}

		// System
		if m := systemRegex.FindStringSubmatch(trimmed); m != nil {
			elem := &Element{Alias: m[1], Label: m[2], Description: m[3], Kind: "system"}
			addElement(d, currentBoundary, elem)
			i++
			continue
		}

		// Relationship
		if m := relRegex.FindStringSubmatch(trimmed); m != nil {
			rel := &Relationship{From: m[1], To: m[2], Label: m[3], Technology: m[4]}
			d.Relationships = append(d.Relationships, rel)
			i++
			continue
		}

		// Skip unrecognized lines
		i++
	}
	return i, nil
}

func addElement(d *C4Diagram, boundary *Boundary, elem *Element) {
	if boundary != nil {
		boundary.Elements = append(boundary.Elements, elem)
	} else {
		d.Elements = append(d.Elements, elem)
	}
}
