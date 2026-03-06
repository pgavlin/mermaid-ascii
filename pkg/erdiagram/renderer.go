package erdiagram

import (
	"fmt"
	"strings"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

type boxChars struct {
	topLeft     rune
	topRight    rune
	bottomLeft  rune
	bottomRight rune
	horizontal  rune
	vertical    rune
	teeLeft     rune
	teeRight    rune
}

var unicodeChars = boxChars{
	topLeft: '┌', topRight: '┐',
	bottomLeft: '└', bottomRight: '┘',
	horizontal: '─', vertical: '│',
	teeLeft: '├', teeRight: '┤',
}

var asciiChars = boxChars{
	topLeft: '+', topRight: '+',
	bottomLeft: '+', bottomRight: '+',
	horizontal: '-', vertical: '|',
	teeLeft: '+', teeRight: '+',
}

// Render renders an ERDiagram to a string.
func Render(erd *ERDiagram, config *diagram.Config) (string, error) {
	if erd == nil || len(erd.Entities) == 0 {
		return "", fmt.Errorf("no entities to render")
	}
	if config == nil {
		config = diagram.DefaultConfig()
	}

	chars := unicodeChars
	if config.UseAscii {
		chars = asciiChars
	}

	var lines []string

	// Render each entity box vertically stacked
	for i, entity := range erd.Entities {
		box := renderEntityBox(entity, chars)
		lines = append(lines, box...)
		if i < len(erd.Entities)-1 {
			lines = append(lines, "")
		}
	}

	// Render relationships as annotations below the entities
	if len(erd.Relationships) > 0 {
		lines = append(lines, "")
		for _, rel := range erd.Relationships {
			relLine := renderRelationshipLine(rel, config.UseAscii)
			lines = append(lines, relLine)
		}
	}

	return strings.Join(lines, "\n") + "\n", nil
}

func renderEntityBox(entity *Entity, chars boxChars) []string {
	// Calculate width
	width := len(entity.Name) + 2 // 1 space padding on each side
	for _, attr := range entity.Attributes {
		attrStr := formatAttribute(attr)
		aw := len(attrStr) + 2
		if aw > width {
			width = aw
		}
	}
	if width < 12 {
		width = 12
	}

	var lines []string

	// Top border
	lines = append(lines, string(chars.topLeft)+strings.Repeat(string(chars.horizontal), width)+string(chars.topRight))

	// Entity name centered
	nameLen := len(entity.Name)
	pad := (width - nameLen) / 2
	nameLine := string(chars.vertical) + strings.Repeat(" ", pad) + entity.Name + strings.Repeat(" ", width-pad-nameLen) + string(chars.vertical)
	lines = append(lines, nameLine)

	// Separator
	lines = append(lines, string(chars.teeLeft)+strings.Repeat(string(chars.horizontal), width)+string(chars.teeRight))

	// Attributes
	if len(entity.Attributes) == 0 {
		lines = append(lines, string(chars.vertical)+strings.Repeat(" ", width)+string(chars.vertical))
	} else {
		for _, attr := range entity.Attributes {
			attrStr := formatAttribute(attr)
			attrLine := string(chars.vertical) + " " + attrStr + strings.Repeat(" ", width-len(attrStr)-1) + string(chars.vertical)
			lines = append(lines, attrLine)
		}
	}

	// Bottom border
	lines = append(lines, string(chars.bottomLeft)+strings.Repeat(string(chars.horizontal), width)+string(chars.bottomRight))

	return lines
}

func formatAttribute(attr *Attribute) string {
	s := attr.Type + " " + attr.Name
	if attr.Constraint != NoConstraint {
		switch attr.Constraint {
		case PrimaryKey:
			s += " PK"
		case ForeignKey:
			s += " FK"
		case UniqueKey:
			s += " UK"
		}
	}
	return s
}

func renderRelationshipLine(rel *Relationship, useAscii bool) string {
	var sb strings.Builder
	sb.WriteString(rel.From)
	sb.WriteByte(' ')
	sb.WriteString(cardinalityString(rel.FromCardinality, useAscii))
	if useAscii {
		sb.WriteString("--")
	} else {
		sb.WriteString("──")
	}
	sb.WriteString(cardinalityString(rel.ToCardinality, useAscii))
	sb.WriteByte(' ')
	sb.WriteString(rel.To)
	sb.WriteString(" : ")
	sb.WriteString(rel.Label)
	return sb.String()
}

func cardinalityString(c Cardinality, useAscii bool) string {
	switch c {
	case ExactlyOne:
		return "||"
	case ZeroOrOne:
		return "o|"
	case OneOrMany:
		if useAscii {
			return ">|"
		}
		return ">|"
	case ZeroOrMany:
		if useAscii {
			return ">o"
		}
		return ">o"
	default:
		return "||"
	}
}
