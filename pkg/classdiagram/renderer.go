package classdiagram

import (
	"fmt"
	"strings"

	"github.com/pgavlin/mermaid-ascii/pkg/canvas"
	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

// Render renders a ClassDiagram to a string.
func Render(cd *ClassDiagram, config *diagram.Config) (string, error) {
	if cd == nil || len(cd.Classes) == 0 {
		return "", fmt.Errorf("no classes to render")
	}
	if config == nil {
		config = diagram.DefaultConfig()
	}

	chars := canvas.UnicodeBox
	if config.UseAscii {
		chars = canvas.ASCIIBox
	}

	var lines []string

	// Render each class box
	classBoxes := make([][]string, len(cd.Classes))
	classWidths := make([]int, len(cd.Classes))
	for i, cls := range cd.Classes {
		box, width := renderClassBox(cls, chars)
		classBoxes[i] = box
		classWidths[i] = width
	}

	// Vertical layout: classes stacked top-to-bottom with relationships to the right
	classStartY := make([]int, len(cd.Classes))
	currentY := 0
	for i, box := range classBoxes {
		classStartY[i] = currentY
		for _, line := range box {
			lines = append(lines, line)
		}
		currentY += len(box)
		// Add spacing between classes
		if i < len(classBoxes)-1 {
			lines = append(lines, "")
			currentY++
		}
	}

	// Render relationships as annotations below the classes
	if len(cd.Relationships) > 0 {
		lines = append(lines, "")
		for _, rel := range cd.Relationships {
			relLine := renderRelationshipLine(rel, chars, config.UseAscii)
			lines = append(lines, relLine)
		}
	}

	return strings.Join(lines, "\n") + "\n", nil
}

func renderClassBox(cls *Class, chars canvas.BoxChars) ([]string, int) {
	// Calculate width: max of class name and all members
	width := len(cls.Name) + 2 // 1 space padding on each side
	for _, m := range cls.Members {
		mStr := formatMember(m)
		mw := len(mStr) + 2
		if mw > width {
			width = mw
		}
	}
	if width < 10 {
		width = 10
	}

	var lines []string

	// Top border
	lines = append(lines, string(chars.TopLeft)+strings.Repeat(string(chars.Horizontal), width)+string(chars.TopRight))

	// Class name centered
	nameLen := len(cls.Name)
	pad := (width - nameLen) / 2
	nameLine := string(chars.Vertical) + strings.Repeat(" ", pad) + cls.Name + strings.Repeat(" ", width-pad-nameLen) + string(chars.Vertical)
	lines = append(lines, nameLine)

	// Separator between name and members
	lines = append(lines, string(chars.TeeRight)+strings.Repeat(string(chars.Horizontal), width)+string(chars.TeeLeft))

	// Members
	if len(cls.Members) == 0 {
		// Empty body
		lines = append(lines, string(chars.Vertical)+strings.Repeat(" ", width)+string(chars.Vertical))
	} else {
		for _, m := range cls.Members {
			mStr := formatMember(m)
			memberLine := string(chars.Vertical) + " " + mStr + strings.Repeat(" ", width-len(mStr)-1) + string(chars.Vertical)
			lines = append(lines, memberLine)
		}
	}

	// Bottom border
	lines = append(lines, string(chars.BottomLeft)+strings.Repeat(string(chars.Horizontal), width)+string(chars.BottomRight))

	return lines, width
}

func formatMember(m *Member) string {
	var sb strings.Builder

	switch m.Visibility {
	case Public:
		sb.WriteByte('+')
	case Private:
		sb.WriteByte('-')
	case Protected:
		sb.WriteByte('#')
	case Package:
		sb.WriteByte('~')
	}

	if m.IsMethod {
		sb.WriteString(m.Name)
		sb.WriteByte('(')
		sb.WriteString(m.Parameters)
		sb.WriteByte(')')
		if m.Type != "" {
			sb.WriteByte(' ')
			sb.WriteString(m.Type)
		}
	} else {
		if m.Type != "" {
			sb.WriteString(m.Type)
			sb.WriteByte(' ')
		}
		sb.WriteString(m.Name)
	}

	return sb.String()
}

func renderRelationshipLine(rel *Relationship, chars canvas.BoxChars, useAscii bool) string {
	var sb strings.Builder
	sb.WriteString(rel.From)

	if rel.FromLabel != "" {
		sb.WriteString(fmt.Sprintf(" \"%s\"", rel.FromLabel))
	}

	sb.WriteByte(' ')
	sb.WriteString(relationshipArrow(rel.Type, useAscii))
	sb.WriteByte(' ')

	if rel.ToLabel != "" {
		sb.WriteString(fmt.Sprintf("\"%s\" ", rel.ToLabel))
	}

	sb.WriteString(rel.To)

	if rel.Label != "" {
		sb.WriteString(" : ")
		sb.WriteString(rel.Label)
	}

	return sb.String()
}

func relationshipArrow(rt RelationType, useAscii bool) string {
	if useAscii {
		switch rt {
		case Inheritance:
			return "<|--"
		case Composition:
			return "*--"
		case Aggregation:
			return "o--"
		case Dependency:
			return "..>"
		case Association:
			return "-->"
		case Realization:
			return "<|.."
		case Link:
			return "--"
		case DottedLink:
			return ".."
		default:
			return "-->"
		}
	}
	// Unicode arrows
	switch rt {
	case Inheritance:
		return "<|──"
	case Composition:
		return "*──"
	case Aggregation:
		return "o──"
	case Dependency:
		return "..>"
	case Association:
		return "──>"
	case Realization:
		return "<|.."
	case Link:
		return "──"
	case DottedLink:
		return ".."
	default:
		return "──>"
	}
}
