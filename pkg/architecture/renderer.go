package architecture

import (
	"fmt"
	"strings"

	"github.com/pgavlin/mermaid-ascii/pkg/canvas"
	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

type archChars struct {
	canvas.BoxChars
	Arrow string
	Line  string
}

var asciiChars = archChars{BoxChars: canvas.ASCIIBox, Arrow: "-->", Line: "---"}
var unicodeChars = archChars{BoxChars: canvas.UnicodeBox, Arrow: "──>", Line: "───"}

// Render renders an architecture diagram as ASCII/Unicode text.
func Render(d *ArchitectureDiagram, config *diagram.Config) (string, error) {
	if d == nil {
		return "", fmt.Errorf("nil diagram")
	}
	if config == nil {
		config = diagram.DefaultConfig()
	}

	chars := unicodeChars
	if config.UseAscii {
		chars = asciiChars
	}

	var lines []string

	// Render top-level services
	for _, svc := range d.Services {
		lines = append(lines, renderService(svc, chars, 0)...)
		lines = append(lines, "")
	}

	// Render groups
	for _, g := range d.Groups {
		lines = append(lines, renderGroup(g, chars, 0)...)
		lines = append(lines, "")
	}

	// Render connections
	for _, conn := range d.Connections {
		lines = append(lines, renderConnection(conn, chars))
	}

	// Remove trailing empty line if present
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return strings.Join(lines, "\n") + "\n", nil
}

func renderService(svc *Service, chars archChars, indent int) []string {
	prefix := strings.Repeat("  ", indent)
	label := svc.Label
	boxWidth := len(label) + 4

	var result []string
	result = append(result, prefix+string(chars.TopLeft)+strings.Repeat(string(chars.Horizontal), boxWidth)+string(chars.TopRight))
	pad := boxWidth - len(label)
	left := pad / 2
	right := pad - left
	result = append(result, prefix+string(chars.Vertical)+strings.Repeat(" ", left)+label+strings.Repeat(" ", right)+string(chars.Vertical))
	result = append(result, prefix+string(chars.BottomLeft)+strings.Repeat(string(chars.Horizontal), boxWidth)+string(chars.BottomRight))
	return result
}

func renderGroup(g *Group, chars archChars, indent int) []string {
	prefix := strings.Repeat("  ", indent)

	var inner []string
	for _, svc := range g.Services {
		inner = append(inner, renderService(svc, chars, indent+1)...)
	}
	for _, sub := range g.Groups {
		inner = append(inner, renderGroup(sub, chars, indent+1)...)
	}

	maxWidth := len(g.Label) + 4
	for _, l := range inner {
		stripped := strings.TrimPrefix(l, prefix)
		if len(stripped)+2 > maxWidth {
			maxWidth = len(stripped) + 2
		}
	}

	var result []string
	result = append(result, prefix+string(chars.TopLeft)+strings.Repeat(string(chars.Horizontal), maxWidth)+string(chars.TopRight))
	titlePad := maxWidth - len(g.Label)
	left := titlePad / 2
	right := titlePad - left
	result = append(result, prefix+string(chars.Vertical)+strings.Repeat(" ", left)+g.Label+strings.Repeat(" ", right)+string(chars.Vertical))
	result = append(result, prefix+string(chars.Vertical)+strings.Repeat(string(chars.Horizontal), maxWidth)+string(chars.Vertical))

	for _, l := range inner {
		stripped := strings.TrimPrefix(l, prefix)
		padNeeded := maxWidth - len(stripped)
		if padNeeded < 0 {
			padNeeded = 0
		}
		result = append(result, prefix+string(chars.Vertical)+stripped+strings.Repeat(" ", padNeeded)+string(chars.Vertical))
	}

	result = append(result, prefix+string(chars.BottomLeft)+strings.Repeat(string(chars.Horizontal), maxWidth)+string(chars.BottomRight))
	return result
}

func renderConnection(conn *Connection, chars archChars) string {
	arrow := chars.Line
	if conn.Directed {
		arrow = chars.Arrow
	}
	return fmt.Sprintf("%s %s %s", conn.From, arrow, conn.To)
}
