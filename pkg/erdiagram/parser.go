package erdiagram

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

const ERDiagramKeyword = "erDiagram"

var (
	// relationshipRegex matches: ENTITY1 ||--o{ ENTITY2 : label
	// Cardinality markers: ||, o|, }|, }o, |o, |{, o{
	relationshipRegex = regexp.MustCompile(`^\s*(\w+)\s+(\|?\|?[o}]?[|{]?)\s*--\s*([|o}]?[|{]?[o}]?\|?)\s+(\w+)\s*:\s*(.+)\s*$`)

	// entityBlockRegex matches: ENTITY {
	entityBlockRegex = regexp.MustCompile(`^\s*(\w+)\s*\{\s*$`)

	// attributeRegex matches: type name or type name PK/FK/UK
	attributeRegex = regexp.MustCompile(`^\s*(\w+)\s+(\w+)(?:\s+(PK|FK|UK))?\s*$`)

	// closingBraceRegex matches a closing brace
	closingBraceRegex = regexp.MustCompile(`^\s*\}\s*$`)
)

// Cardinality represents the cardinality of a relationship end.
type Cardinality int

const (
	ExactlyOne  Cardinality = iota // ||
	ZeroOrOne                      // o| or |o
	OneOrMany                      // }| or |{
	ZeroOrMany                     // }o or o{
)

// Constraint represents a column constraint.
type Constraint int

const (
	NoConstraint Constraint = iota
	PrimaryKey              // PK
	ForeignKey              // FK
	UniqueKey               // UK
)

// Attribute represents an entity attribute.
type Attribute struct {
	Type       string
	Name       string
	Constraint Constraint
}

// Entity represents an entity in the ER diagram.
type Entity struct {
	Name       string
	Attributes []*Attribute
	Index      int
}

// Relationship represents a relationship between two entities.
type Relationship struct {
	From            string
	To              string
	Label           string
	FromCardinality Cardinality
	ToCardinality   Cardinality
}

// ERDiagram represents a parsed ER diagram.
type ERDiagram struct {
	Entities      []*Entity
	Relationships []*Relationship
	entityMap     map[string]*Entity
}

// IsERDiagram returns true if the input begins with the erDiagram keyword.
func IsERDiagram(input string) bool {
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}
		return strings.HasPrefix(trimmed, ERDiagramKeyword)
	}
	return false
}

// Parse parses an ER diagram from the given input string.
func Parse(input string) (*ERDiagram, error) {
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
	if !strings.HasPrefix(first, ERDiagramKeyword) {
		return nil, fmt.Errorf("expected %q keyword", ERDiagramKeyword)
	}
	lines = lines[1:]

	erd := &ERDiagram{
		Entities:      []*Entity{},
		Relationships: []*Relationship{},
		entityMap:     make(map[string]*Entity),
	}

	i := 0
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			i++
			continue
		}

		// Check for entity block: ENTITY {
		if match := entityBlockRegex.FindStringSubmatch(line); match != nil {
			entityName := match[1]
			entity := erd.getOrCreateEntity(entityName)
			i++
			// Parse attributes until closing brace
			for i < len(lines) {
				attrLine := strings.TrimSpace(lines[i])
				if closingBraceRegex.MatchString(attrLine) {
					i++
					break
				}
				if attrLine == "" {
					i++
					continue
				}
				if attr := parseAttribute(attrLine); attr != nil {
					entity.Attributes = append(entity.Attributes, attr)
				}
				i++
			}
			continue
		}

		// Check for relationship
		if match := relationshipRegex.FindStringSubmatch(line); match != nil {
			from := match[1]
			leftCard := match[2]
			rightCard := match[3]
			to := match[4]
			label := strings.TrimSpace(match[5])

			erd.getOrCreateEntity(from)
			erd.getOrCreateEntity(to)

			fromCard := parseCardinality(leftCard)
			toCard := parseCardinality(rightCard)

			erd.Relationships = append(erd.Relationships, &Relationship{
				From:            from,
				To:              to,
				Label:           label,
				FromCardinality: fromCard,
				ToCardinality:   toCard,
			})
			i++
			continue
		}

		// Unknown line, skip
		i++
	}

	if len(erd.Entities) == 0 {
		return nil, fmt.Errorf("no entities found")
	}

	return erd, nil
}

func (erd *ERDiagram) getOrCreateEntity(name string) *Entity {
	if e, ok := erd.entityMap[name]; ok {
		return e
	}
	e := &Entity{
		Name:       name,
		Attributes: []*Attribute{},
		Index:      len(erd.Entities),
	}
	erd.Entities = append(erd.Entities, e)
	erd.entityMap[name] = e
	return e
}

func parseAttribute(line string) *Attribute {
	match := attributeRegex.FindStringSubmatch(line)
	if match == nil {
		return nil
	}

	constraint := NoConstraint
	if match[3] != "" {
		switch match[3] {
		case "PK":
			constraint = PrimaryKey
		case "FK":
			constraint = ForeignKey
		case "UK":
			constraint = UniqueKey
		}
	}

	return &Attribute{
		Type:       match[1],
		Name:       match[2],
		Constraint: constraint,
	}
}

func parseCardinality(s string) Cardinality {
	s = strings.TrimSpace(s)
	switch s {
	case "||":
		return ExactlyOne
	case "o|", "|o":
		return ZeroOrOne
	case "}|", "|{":
		return OneOrMany
	case "}o", "o{":
		return ZeroOrMany
	default:
		// Try to infer from partial matches
		if strings.Contains(s, "}") && strings.Contains(s, "o") {
			return ZeroOrMany
		}
		if strings.Contains(s, "}") {
			return OneOrMany
		}
		if strings.Contains(s, "o") {
			return ZeroOrOne
		}
		return ExactlyOne
	}
}
