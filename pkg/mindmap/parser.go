package mindmap

import (
	"fmt"
	"strings"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
)

const MindmapKeyword = "mindmap"

type MindmapDiagram struct {
	Root *MindmapNode
}

type MindmapNode struct {
	Text     string
	Shape    NodeShape
	Children []*MindmapNode
	Depth    int
}

type NodeShape int

const (
	ShapeDefault NodeShape = iota
	ShapeSquare      // [text]
	ShapeRounded     // (text)
	ShapeBang        // ))text((
	ShapeCloud       // )text(
	ShapeHexagon     // {{text}}
)

func IsMindmapDiagram(input string) bool {
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}
		return trimmed == MindmapKeyword
	}
	return false
}

func Parse(input string) (*MindmapDiagram, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty input")
	}

	rawLines := diagram.SplitLines(input)
	lines := diagram.RemoveComments(rawLines)
	if len(lines) == 0 {
		return nil, fmt.Errorf("no content found")
	}

	if strings.TrimSpace(lines[0]) != MindmapKeyword {
		return nil, fmt.Errorf("expected %q keyword", MindmapKeyword)
	}
	lines = lines[1:]

	if len(lines) == 0 {
		return nil, fmt.Errorf("no nodes found")
	}

	md := &MindmapDiagram{}

	// Parse indentation-based tree
	// First non-empty line is root
	var root *MindmapNode
	stack := []*MindmapNode{} // stack of parents at each depth level
	baseIndent := -1

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Measure indentation (number of leading spaces/tabs)
		indent := 0
		for _, ch := range line {
			if ch == ' ' {
				indent++
			} else if ch == '\t' {
				indent += 4
			} else {
				break
			}
		}

		text := strings.TrimSpace(line)
		if text == "" {
			continue
		}

		// Parse shape from text
		node := parseNodeText(text)

		if baseIndent < 0 {
			baseIndent = indent
		}

		// Calculate depth relative to base
		depth := 0
		if indent > baseIndent {
			depth = (indent - baseIndent + 1) / 2
			if depth == 0 {
				depth = 1
			}
		}
		node.Depth = depth

		if root == nil {
			root = node
			stack = []*MindmapNode{root}
			continue
		}

		// Pop stack until we find the right parent level
		for len(stack) > depth {
			stack = stack[:len(stack)-1]
		}

		if len(stack) == 0 {
			// Shouldn't happen, but handle gracefully
			root.Children = append(root.Children, node)
			stack = append(stack, node)
		} else {
			parent := stack[len(stack)-1]
			parent.Children = append(parent.Children, node)
			stack = append(stack, node)
		}
	}

	if root == nil {
		return nil, fmt.Errorf("no nodes found")
	}

	md.Root = root
	return md, nil
}

func parseNodeText(text string) *MindmapNode {
	node := &MindmapNode{Shape: ShapeDefault}

	// Check for shape delimiters
	if strings.HasPrefix(text, "[") && strings.HasSuffix(text, "]") {
		node.Text = text[1 : len(text)-1]
		node.Shape = ShapeSquare
	} else if strings.HasPrefix(text, "(") && strings.HasSuffix(text, ")") {
		node.Text = text[1 : len(text)-1]
		node.Shape = ShapeRounded
	} else if strings.HasPrefix(text, "{{") && strings.HasSuffix(text, "}}") {
		node.Text = text[2 : len(text)-2]
		node.Shape = ShapeHexagon
	} else if strings.HasPrefix(text, "))") && strings.HasSuffix(text, "((") {
		node.Text = text[2 : len(text)-2]
		node.Shape = ShapeBang
	} else if strings.HasPrefix(text, ")") && strings.HasSuffix(text, "(") {
		node.Text = text[1 : len(text)-1]
		node.Shape = ShapeCloud
	} else {
		node.Text = text
	}

	return node
}
