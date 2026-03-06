package blockdiagram

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

const BlockBetaKeyword = "block-beta"

var (
	columnsRegex   = regexp.MustCompile(`^\s*columns\s+(\d+)\s*$`)
	blockStartRegex = regexp.MustCompile(`^\s*block\s*(?::(\S+))?\s*$`)
	blockEndRegex   = regexp.MustCompile(`^\s*end\s*$`)
	// Block name with optional label: name["label"] or name("label") or just name
	blockNameRegex = regexp.MustCompile(`^\s*(\S+?)(?:\["([^"]+)"\]|\("([^"]+)"\))?\s*(?::(\d+))?\s*$`)
)

// Block represents a single block in the diagram.
type Block struct {
	ID       string
	Label    string
	Children []*Block
	Columns  int // number of columns for this container block
	Span     int // how many columns this block spans
}

// BlockDiagram represents a parsed block diagram.
type BlockDiagram struct {
	Columns int
	Blocks  []*Block
}

// IsBlockDiagram returns true if the input starts with block-beta keyword.
func IsBlockDiagram(input string) bool {
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}
		return strings.HasPrefix(trimmed, BlockBetaKeyword)
	}
	return false
}

// Parse parses a block diagram.
func Parse(input string) (*BlockDiagram, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty input")
	}

	rawLines := diagram.SplitLines(input)
	lines := diagram.RemoveComments(rawLines)
	if len(lines) == 0 {
		return nil, fmt.Errorf("no content found")
	}

	if !strings.HasPrefix(strings.TrimSpace(lines[0]), BlockBetaKeyword) {
		return nil, fmt.Errorf("expected %q keyword", BlockBetaKeyword)
	}

	d := &BlockDiagram{Columns: 1}

	_, err := parseBlockLines(d, lines[1:], nil)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func parseBlockLines(d *BlockDiagram, lines []string, parent *Block) (int, error) {
	i := 0
	for i < len(lines) {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" {
			i++
			continue
		}

		// columns directive
		if m := columnsRegex.FindStringSubmatch(trimmed); m != nil {
			cols, _ := strconv.Atoi(m[1])
			if parent != nil {
				parent.Columns = cols
			} else {
				d.Columns = cols
			}
			i++
			continue
		}

		// end of block
		if blockEndRegex.MatchString(trimmed) {
			return i + 1, nil
		}

		// block start
		if m := blockStartRegex.FindStringSubmatch(trimmed); m != nil {
			b := &Block{
				ID:      m[1],
				Label:   m[1],
				Columns: 1,
				Span:    1,
			}
			if b.ID == "" {
				b.ID = fmt.Sprintf("block_%d", i)
				b.Label = ""
			}
			i++
			consumed, err := parseBlockLines(d, lines[i:], b)
			if err != nil {
				return 0, err
			}
			i += consumed
			if parent != nil {
				parent.Children = append(parent.Children, b)
			} else {
				d.Blocks = append(d.Blocks, b)
			}
			continue
		}

		// Simple block name
		if m := blockNameRegex.FindStringSubmatch(trimmed); m != nil {
			id := m[1]
			label := id
			if m[2] != "" {
				label = m[2]
			} else if m[3] != "" {
				label = m[3]
			}
			span := 1
			if m[4] != "" {
				span, _ = strconv.Atoi(m[4])
			}
			b := &Block{
				ID:    id,
				Label: label,
				Span:  span,
			}
			if parent != nil {
				parent.Children = append(parent.Children, b)
			} else {
				d.Blocks = append(d.Blocks, b)
			}
			i++
			continue
		}

		i++
	}
	return i, nil
}
