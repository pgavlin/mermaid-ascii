package classdiagram

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

const ClassDiagramKeyword = "classDiagram"

var (
	// classBlockRegex matches: class ClassName {
	classBlockRegex = regexp.MustCompile(`^\s*class\s+(\w+)\s*\{\s*$`)

	// classInlineRegex matches: class ClassName
	classInlineRegex = regexp.MustCompile(`^\s*class\s+(\w+)\s*$`)

	// memberRegex matches: +String name, -privateMethod(), #protectedField, ~packageField
	memberRegex = regexp.MustCompile(`^\s*([+\-#~]?)(\w[\w<>\[\],\s]*?)(?:\(([^)]*)\))?\s*(\w+)?\s*$`)

	// relationshipRegex matches relationships like:
	// Animal <|-- Dog
	// Animal *-- Leg
	// Animal o-- Lake
	// Animal ..> Water
	// Animal --> Food
	// Animal "1" --> "*" Food : eats
	relationshipRegex = regexp.MustCompile(`^\s*(\w+)\s*(?:"([^"]*)")?\s*(<\|--|<\|\.\.|\*--|\*\.\.|o--|o\.\.|-->|--|\.\.|\.\.>|--\*|--o|<--)\s*(?:"([^"]*)")?\s*(\w+)\s*(?::\s*(.+))?\s*$`)

	// closingBraceRegex matches a closing brace
	closingBraceRegex = regexp.MustCompile(`^\s*\}\s*$`)
)

// RelationType represents the type of relationship between classes.
type RelationType int

const (
	Association  RelationType = iota // -->
	Inheritance                      // <|--
	Composition                      // *--
	Aggregation                      // o--
	Dependency                       // ..>
	Realization                      // <|..
	Link                             // --
	DottedLink                       // ..
)

// Visibility represents the visibility of a class member.
type Visibility int

const (
	Public    Visibility = iota // +
	Private                    // -
	Protected                  // #
	Package                    // ~
	None                       // no prefix
)

// Member represents a class member (field or method).
type Member struct {
	Visibility Visibility
	Name       string
	Type       string
	IsMethod   bool
	Parameters string // raw parameter string for methods
}

// Class represents a class in the diagram.
type Class struct {
	Name    string
	Members []*Member
	Index   int
}

// Relationship represents a relationship between two classes.
type Relationship struct {
	From         string
	To           string
	Type         RelationType
	Label        string
	FromLabel    string // cardinality label near "from"
	ToLabel      string // cardinality label near "to"
}

// ClassDiagram represents a parsed class diagram.
type ClassDiagram struct {
	Classes       []*Class
	Relationships []*Relationship
	classMap      map[string]*Class
}

// IsClassDiagram returns true if the input begins with the classDiagram keyword.
func IsClassDiagram(input string) bool {
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}
		return strings.HasPrefix(trimmed, ClassDiagramKeyword)
	}
	return false
}

// Parse parses a class diagram from the given input string.
func Parse(input string) (*ClassDiagram, error) {
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
	if !strings.HasPrefix(first, ClassDiagramKeyword) {
		return nil, fmt.Errorf("expected %q keyword", ClassDiagramKeyword)
	}
	lines = lines[1:]

	cd := &ClassDiagram{
		Classes:       []*Class{},
		Relationships: []*Relationship{},
		classMap:      make(map[string]*Class),
	}

	i := 0
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			i++
			continue
		}

		// Check for class block: class Name {
		if match := classBlockRegex.FindStringSubmatch(line); match != nil {
			className := match[1]
			cls := cd.getOrCreateClass(className)
			i++
			// Parse members until closing brace
			for i < len(lines) {
				memberLine := strings.TrimSpace(lines[i])
				if closingBraceRegex.MatchString(memberLine) {
					i++
					break
				}
				if memberLine == "" {
					i++
					continue
				}
				member := parseMember(memberLine)
				if member != nil {
					cls.Members = append(cls.Members, member)
				}
				i++
			}
			continue
		}

		// Check for inline class declaration: class Name
		if match := classInlineRegex.FindStringSubmatch(line); match != nil {
			cd.getOrCreateClass(match[1])
			i++
			continue
		}

		// Check for relationship
		if match := relationshipRegex.FindStringSubmatch(line); match != nil {
			from := match[1]
			fromLabel := match[2]
			arrow := match[3]
			toLabel := match[4]
			to := match[5]
			label := strings.TrimSpace(match[6])

			cd.getOrCreateClass(from)
			cd.getOrCreateClass(to)

			relType := parseRelationType(arrow)
			cd.Relationships = append(cd.Relationships, &Relationship{
				From:      from,
				To:        to,
				Type:      relType,
				Label:     label,
				FromLabel: fromLabel,
				ToLabel:   toLabel,
			})
			i++
			continue
		}

		// Unknown line, skip
		i++
	}

	if len(cd.Classes) == 0 {
		return nil, fmt.Errorf("no classes found")
	}

	return cd, nil
}

func (cd *ClassDiagram) getOrCreateClass(name string) *Class {
	if cls, ok := cd.classMap[name]; ok {
		return cls
	}
	cls := &Class{
		Name:    name,
		Members: []*Member{},
		Index:   len(cd.Classes),
	}
	cd.Classes = append(cd.Classes, cls)
	cd.classMap[name] = cls
	return cls
}

func parseMember(line string) *Member {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	vis := None
	if len(line) > 0 {
		switch line[0] {
		case '+':
			vis = Public
			line = line[1:]
		case '-':
			vis = Private
			line = line[1:]
		case '#':
			vis = Protected
			line = line[1:]
		case '~':
			vis = Package
			line = line[1:]
		}
	}

	line = strings.TrimSpace(line)

	// Check if it's a method: contains ()
	if idx := strings.Index(line, "("); idx != -1 {
		endIdx := strings.LastIndex(line, ")")
		if endIdx == -1 {
			endIdx = len(line)
		}
		name := strings.TrimSpace(line[:idx])
		params := ""
		if endIdx > idx+1 {
			params = line[idx+1 : endIdx]
		}
		retType := ""
		if endIdx+1 < len(line) {
			retType = strings.TrimSpace(line[endIdx+1:])
		}
		return &Member{
			Visibility: vis,
			Name:       name,
			Type:       retType,
			IsMethod:   true,
			Parameters: params,
		}
	}

	// It's a field: Type Name or just Name
	parts := strings.Fields(line)
	if len(parts) >= 2 {
		return &Member{
			Visibility: vis,
			Name:       parts[1],
			Type:       parts[0],
			IsMethod:   false,
		}
	} else if len(parts) == 1 {
		return &Member{
			Visibility: vis,
			Name:       parts[0],
			Type:       "",
			IsMethod:   false,
		}
	}

	return nil
}

func parseRelationType(arrow string) RelationType {
	switch arrow {
	case "<|--":
		return Inheritance
	case "<|..":
		return Realization
	case "*--", "--*":
		return Composition
	case "o--", "--o":
		return Aggregation
	case "..>":
		return Dependency
	case "-->", "<--":
		return Association
	case "--":
		return Link
	case "..", "*..":
		return DottedLink
	default:
		return Association
	}
}
