package graph

import (
	"testing"
)

func TestWordWrap(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxWidth int
		want     string
	}{
		{
			name:     "no wrap needed",
			text:     "hello",
			maxWidth: 10,
			want:     "hello",
		},
		{
			name:     "exact fit",
			text:     "hello",
			maxWidth: 5,
			want:     "hello",
		},
		{
			name:     "wrap at space",
			text:     "hello world",
			maxWidth: 7,
			want:     "hello\nworld",
		},
		{
			name:     "wrap long line into three",
			text:     "one two three four",
			maxWidth: 9,
			want:     "one two\nthree\nfour",
		},
		{
			name:     "no space to break at",
			text:     "abcdefghij",
			maxWidth: 5,
			want:     "abcde\nfghij",
		},
		{
			name:     "preserve existing newlines",
			text:     "line one\nline two",
			maxWidth: 20,
			want:     "line one\nline two",
		},
		{
			name:     "wrap within existing lines",
			text:     "long first line\nshort",
			maxWidth: 8,
			want:     "long\nfirst\nline\nshort",
		},
		{
			name:     "single word longer than max",
			text:     "superlongword",
			maxWidth: 5,
			want:     "super\nlongw\nord",
		},
		{
			name:     "maxWidth 1",
			text:     "ab",
			maxWidth: 1,
			want:     "a\nb",
		},
		{
			name:     "maxWidth 0 treated as 1",
			text:     "ab",
			maxWidth: 0,
			want:     "a\nb",
		},
		{
			name:     "empty string",
			text:     "",
			maxWidth: 10,
			want:     "",
		},
		{
			name:     "multiple spaces between words",
			text:     "hello  world",
			maxWidth: 7,
			want:     "hello \nworld",
		},
		{
			name:     "wrap at last possible space",
			text:     "a b c d e",
			maxWidth: 5,
			want:     "a b c\nd e",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wordWrap(tt.text, tt.maxWidth)
			if got != tt.want {
				t.Errorf("wordWrap(%q, %d) = %q, want %q", tt.text, tt.maxWidth, got, tt.want)
			}
		})
	}
}

func TestLongestWord(t *testing.T) {
	tests := []struct {
		name string
		text string
		want int
	}{
		{
			name: "single word",
			text: "hello",
			want: 5,
		},
		{
			name: "multiple words",
			text: "hi there everyone",
			want: 8,
		},
		{
			name: "multiline",
			text: "short\nsuperlongword\nmed",
			want: 13,
		},
		{
			name: "empty string",
			text: "",
			want: 1, // minimum is 1
		},
		{
			name: "single character",
			text: "x",
			want: 1,
		},
		{
			name: "with leading/trailing spaces",
			text: "  hello  world  ",
			want: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := longestWord(tt.text)
			if got != tt.want {
				t.Errorf("longestWord(%q) = %d, want %d", tt.text, got, tt.want)
			}
		})
	}
}

func TestShapeExtraWidth(t *testing.T) {
	tests := []struct {
		name            string
		shape           nodeShape
		boxBorderPadding int
		want            int
	}{
		{"rect", shapeRect, 1, 0},
		{"rounded", shapeRounded, 1, 0},
		{"stadium", shapeStadium, 1, 0},
		{"cylinder", shapeCylinder, 1, 0},
		{"diamond padding=1", shapeDiamond, 1, 4},
		{"diamond padding=2", shapeDiamond, 2, 6},
		{"hexagon", shapeHexagon, 1, 4},
		{"subroutine", shapeSubroutine, 1, 2},
		{"flag", shapeFlag, 1, 2},
		{"circle", shapeCircle, 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shapeExtraWidth(tt.shape, tt.boxBorderPadding)
			if got != tt.want {
				t.Errorf("shapeExtraWidth(%v, %d) = %d, want %d", tt.shape, tt.boxBorderPadding, got, tt.want)
			}
		})
	}
}

func TestNameLines(t *testing.T) {
	tests := []struct {
		name     string
		nodeName string
		want     []string
	}{
		{"single line", "hello", []string{"hello"}},
		{"multi line", "hello\nworld", []string{"hello", "world"}},
		{"three lines", "a\nb\nc", []string{"a", "b", "c"}},
		{"empty", "", []string{""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &node{name: tt.nodeName}
			got := n.nameLines()
			if len(got) != len(tt.want) {
				t.Fatalf("nameLines() returned %d lines, want %d", len(got), len(tt.want))
			}
			for i, line := range got {
				if line != tt.want[i] {
					t.Errorf("nameLines()[%d] = %q, want %q", i, line, tt.want[i])
				}
			}
		})
	}
}

func TestNameWidth(t *testing.T) {
	tests := []struct {
		name     string
		nodeName string
		want     int
	}{
		{"single line", "hello", 5},
		{"multi line different widths", "hi\nworld", 5},
		{"multi line first wider", "longline\nhi", 8},
		{"empty", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &node{name: tt.nodeName}
			got := n.nameWidth()
			if got != tt.want {
				t.Errorf("nameWidth() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestNameHeight(t *testing.T) {
	tests := []struct {
		name     string
		nodeName string
		want     int
	}{
		{"single line", "hello", 1},
		{"two lines", "hello\nworld", 2},
		{"three lines", "a\nb\nc", 3},
		{"empty", "", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &node{name: tt.nodeName}
			got := n.nameHeight()
			if got != tt.want {
				t.Errorf("nameHeight() = %d, want %d", got, tt.want)
			}
		})
	}
}
