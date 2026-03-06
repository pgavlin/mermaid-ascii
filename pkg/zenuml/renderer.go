package zenuml

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
	TeeDown     rune
	TeeRight    rune
	TeeLeft     rune
	ArrowRight  rune
	ArrowLeft   rune
	SolidLine   rune
	DottedLine  rune
}

var asciiChars = BoxChars{'+', '+', '+', '+', '-', '|', '+', '+', '+', '>', '<', '-', '.'}
var unicodeChars = BoxChars{
	TopLeft: '\u250c', TopRight: '\u2510',
	BottomLeft: '\u2514', BottomRight: '\u2518',
	Horizontal: '\u2500', Vertical: '\u2502',
	TeeDown: '\u252c', TeeRight: '\u251c', TeeLeft: '\u2524',
	ArrowRight: '\u25ba', ArrowLeft: '\u25c4',
	SolidLine: '\u2500', DottedLine: '\u2508',
}

const (
	participantSpacing = 5
	boxPadding         = 2
	boxBorder          = 2
	minBoxWidth        = 3
	labelMargin        = 2
)

type layout struct {
	widths  []int
	centers []int
	total   int
}

// Render renders a ZenUMLDiagram as ASCII/Unicode text.
func Render(d *ZenUMLDiagram, config *diagram.Config) (string, error) {
	if d == nil || len(d.Participants) == 0 {
		return "", fmt.Errorf("no participants")
	}
	if config == nil {
		config = diagram.DefaultConfig()
	}

	chars := unicodeChars
	if config.UseAscii {
		chars = asciiChars
	}

	ly := calcLayout(d)
	var lines []string

	// Draw participant headers
	lines = append(lines, renderHeaders(d, ly, chars)...)

	// Flatten and render messages (including nested)
	flat := flattenMessages(d.Messages)
	for _, msg := range flat {
		lines = append(lines, buildLifeline(ly, chars))

		switch msg.Type {
		case SyncMessage, AsyncMessage:
			lines = append(lines, renderCallMessage(msg, ly, chars)...)
		case ReturnMessage:
			lines = append(lines, renderReturnMessage(msg, ly, chars)...)
		}
	}

	// Final lifeline
	lines = append(lines, buildLifeline(ly, chars))

	return strings.Join(lines, "\n") + "\n", nil
}

// flattenMessages recursively flattens nested messages into a single slice.
func flattenMessages(msgs []*Message) []*Message {
	var flat []*Message
	for _, m := range msgs {
		flat = append(flat, m)
		if len(m.Nested) > 0 {
			flat = append(flat, flattenMessages(m.Nested)...)
		}
	}
	return flat
}

func calcLayout(d *ZenUMLDiagram) *layout {
	widths := make([]int, len(d.Participants))
	for i, p := range d.Participants {
		w := len(p.ID) + boxPadding
		if w < minBoxWidth {
			w = minBoxWidth
		}
		widths[i] = w
	}

	centers := make([]int, len(d.Participants))
	currentX := 0
	for i := range d.Participants {
		bw := widths[i] + boxBorder
		if i == 0 {
			centers[i] = bw / 2
			currentX = bw
		} else {
			currentX += participantSpacing
			centers[i] = currentX + bw/2
			currentX += bw
		}
	}

	last := len(d.Participants) - 1
	total := centers[last] + (widths[last]+boxBorder)/2

	return &layout{widths: widths, centers: centers, total: total}
}

func renderHeaders(d *ZenUMLDiagram, ly *layout, chars BoxChars) []string {
	var lines []string

	// Top border
	lines = append(lines, buildHeaderLine(d, ly, func(i int) string {
		return string(chars.TopLeft) + strings.Repeat(string(chars.Horizontal), ly.widths[i]) + string(chars.TopRight)
	}))

	// Label (use participant ID)
	lines = append(lines, buildHeaderLine(d, ly, func(i int) string {
		w := ly.widths[i]
		label := d.Participants[i].ID
		pad := (w - len(label)) / 2
		return string(chars.Vertical) + strings.Repeat(" ", pad) + label +
			strings.Repeat(" ", w-pad-len(label)) + string(chars.Vertical)
	}))

	// Bottom border with lifeline tee
	lines = append(lines, buildHeaderLine(d, ly, func(i int) string {
		w := ly.widths[i]
		return string(chars.BottomLeft) + strings.Repeat(string(chars.Horizontal), w/2) +
			string(chars.TeeDown) + strings.Repeat(string(chars.Horizontal), w-w/2-1) +
			string(chars.BottomRight)
	}))

	return lines
}

func buildHeaderLine(d *ZenUMLDiagram, ly *layout, draw func(int) string) string {
	var sb strings.Builder
	for i := range d.Participants {
		bw := ly.widths[i] + boxBorder
		left := ly.centers[i] - bw/2
		needed := left - sb.Len()
		if needed > 0 {
			sb.WriteString(strings.Repeat(" ", needed))
		}
		sb.WriteString(draw(i))
	}
	return sb.String()
}

func buildLifeline(ly *layout, chars BoxChars) string {
	line := make([]rune, ly.total+1)
	for i := range line {
		line[i] = ' '
	}
	for _, c := range ly.centers {
		if c < len(line) {
			line[c] = chars.Vertical
		}
	}
	return strings.TrimRight(string(line), " ")
}

func renderCallMessage(msg *Message, ly *layout, chars BoxChars) []string {
	var lines []string
	from := ly.centers[msg.From.Index]
	to := ly.centers[msg.To.Index]

	label := msg.Method + "(" + msg.Args + ")"
	if msg.Type == AsyncMessage {
		label = "(async) " + label
	}

	// Self-call
	if msg.From.Index == msg.To.Index {
		labelLine := makeLine(ly, chars, ly.total+len(label)+10)
		start := from + labelMargin
		for j, r := range label {
			if start+j < len(labelLine) {
				labelLine[start+j] = r
			}
		}
		lines = append(lines, strings.TrimRight(string(labelLine), " "))

		arrowLine := makeLine(ly, chars, ly.total+6)
		if from < len(arrowLine) {
			arrowLine[from] = chars.TeeRight
		}
		for j := 1; j < 4 && from+j < len(arrowLine); j++ {
			arrowLine[from+j] = chars.Horizontal
		}
		lines = append(lines, strings.TrimRight(string(arrowLine), " "))
		return lines
	}

	// Label line
	start := minInt(from, to) + labelMargin
	width := maxInt(ly.total+1, start+len(label)+10)
	labelLine := makeLine(ly, chars, width)
	for j, r := range label {
		if start+j < len(labelLine) {
			labelLine[start+j] = r
		}
	}
	lines = append(lines, strings.TrimRight(string(labelLine), " "))

	// Arrow line
	arrowLine := makeLine(ly, chars, ly.total+1)
	lineChar := chars.SolidLine
	if msg.Type == AsyncMessage {
		lineChar = chars.DottedLine
	}

	if from < to {
		arrowLine[from] = chars.TeeRight
		for j := from + 1; j < to; j++ {
			arrowLine[j] = lineChar
		}
		arrowLine[to-1] = chars.ArrowRight
		arrowLine[to] = chars.Vertical
	} else {
		arrowLine[to] = chars.Vertical
		arrowLine[to+1] = chars.ArrowLeft
		for j := to + 2; j < from; j++ {
			arrowLine[j] = lineChar
		}
		arrowLine[from] = chars.TeeLeft
	}
	lines = append(lines, strings.TrimRight(string(arrowLine), " "))

	return lines
}

func renderReturnMessage(msg *Message, ly *layout, chars BoxChars) []string {
	var lines []string

	// If we don't have from/to, skip
	if msg.From == nil || msg.To == nil {
		return lines
	}

	from := ly.centers[msg.From.Index]
	to := ly.centers[msg.To.Index]

	// Label
	if msg.Label != "" {
		label := "return " + msg.Label
		start := minInt(from, to) + labelMargin
		width := maxInt(ly.total+1, start+len(label)+10)
		labelLine := makeLine(ly, chars, width)
		for j, r := range label {
			if start+j < len(labelLine) {
				labelLine[start+j] = r
			}
		}
		lines = append(lines, strings.TrimRight(string(labelLine), " "))
	}

	// Dotted return arrow
	if from != to {
		arrowLine := makeLine(ly, chars, ly.total+1)
		if from < to {
			arrowLine[from] = chars.TeeRight
			for j := from + 1; j < to; j++ {
				arrowLine[j] = chars.DottedLine
			}
			arrowLine[to-1] = chars.ArrowRight
			arrowLine[to] = chars.Vertical
		} else {
			arrowLine[to] = chars.Vertical
			arrowLine[to+1] = chars.ArrowLeft
			for j := to + 2; j < from; j++ {
				arrowLine[j] = chars.DottedLine
			}
			arrowLine[from] = chars.TeeLeft
		}
		lines = append(lines, strings.TrimRight(string(arrowLine), " "))
	}

	return lines
}

// makeLine creates a rune slice with lifelines drawn.
func makeLine(ly *layout, chars BoxChars, width int) []rune {
	line := make([]rune, width)
	for i := range line {
		line[i] = ' '
	}
	for _, c := range ly.centers {
		if c < len(line) {
			line[c] = chars.Vertical
		}
	}
	return line
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
