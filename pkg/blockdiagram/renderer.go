package blockdiagram

import (
	"fmt"
	"strings"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

// BoxChars holds the characters used for rendering.
type BoxChars struct {
	TopLeft     rune
	TopRight    rune
	BottomLeft  rune
	BottomRight rune
	Horizontal  rune
	Vertical    rune
}

var asciiChars = BoxChars{'+', '+', '+', '+', '-', '|'}
var unicodeChars = BoxChars{'┌', '┐', '└', '┘', '─', '│'}

const defaultCellWidth = 12

// Render renders a block diagram as ASCII/Unicode text.
func Render(d *BlockDiagram, config *diagram.Config) (string, error) {
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
	lines = append(lines, renderBlocks(d.Blocks, d.Columns, chars, 0)...)

	return strings.Join(lines, "\n") + "\n", nil
}

func renderBlocks(blocks []*Block, columns int, chars BoxChars, indent int) []string {
	if columns <= 0 {
		columns = 1
	}
	prefix := strings.Repeat("  ", indent)

	var allLines []string
	row := make([]*Block, 0, columns)
	col := 0

	for _, b := range blocks {
		row = append(row, b)
		col += b.Span
		if col >= columns {
			allLines = append(allLines, renderRow(row, columns, chars, prefix)...)
			row = row[:0]
			col = 0
		}
	}
	if len(row) > 0 {
		allLines = append(allLines, renderRow(row, columns, chars, prefix)...)
	}

	return allLines
}

func renderRow(blocks []*Block, _ int, chars BoxChars, prefix string) []string {
	// Calculate cell width
	cellWidth := defaultCellWidth

	// Find max label width
	for _, b := range blocks {
		lw := len(b.Label) + 4
		spanWidth := cellWidth * b.Span
		if lw > spanWidth {
			cellWidth = (lw + b.Span - 1) / b.Span
			if cellWidth < defaultCellWidth {
				cellWidth = defaultCellWidth
			}
		}
	}

	// Build top border line
	var topLine, midLine, botLine strings.Builder
	topLine.WriteString(prefix)
	midLine.WriteString(prefix)
	botLine.WriteString(prefix)

	for i, b := range blocks {
		w := cellWidth*b.Span - 2
		if w < len(b.Label)+2 {
			w = len(b.Label) + 2
		}

		if i > 0 {
			topLine.WriteString("  ")
			midLine.WriteString("  ")
			botLine.WriteString("  ")
		}

		topLine.WriteRune(chars.TopLeft)
		topLine.WriteString(strings.Repeat(string(chars.Horizontal), w))
		topLine.WriteRune(chars.TopRight)

		label := b.Label
		pad := w - len(label)
		left := pad / 2
		right := pad - left
		midLine.WriteRune(chars.Vertical)
		midLine.WriteString(strings.Repeat(" ", left))
		midLine.WriteString(label)
		midLine.WriteString(strings.Repeat(" ", right))
		midLine.WriteRune(chars.Vertical)

		botLine.WriteRune(chars.BottomLeft)
		botLine.WriteString(strings.Repeat(string(chars.Horizontal), w))
		botLine.WriteRune(chars.BottomRight)
	}

	result := []string{topLine.String(), midLine.String(), botLine.String()}

	// Render children of nested blocks
	for _, b := range blocks {
		if len(b.Children) > 0 {
			cols := b.Columns
			if cols <= 0 {
				cols = 1
			}
			childLines := renderBlocks(b.Children, cols, chars, 1)
			result = append(result, childLines...)
		}
	}

	return result
}
