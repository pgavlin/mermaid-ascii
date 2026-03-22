package graph

import (
	"testing"
)

func TestReplaceBrTags(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"br with slash", "hello<br/>world", "hello\nworld"},
		{"br without slash", "hello<br>world", "hello\nworld"},
		{"br with space and slash", "hello<br />world", "hello\nworld"},
		{"multiple br tags", "a<br/>b<br>c<br />d", "a\nb\nc\nd"},
		{"no br tags", "hello world", "hello world"},
		{"br at start", "<br/>hello", "\nhello"},
		{"br at end", "hello<br/>", "hello\n"},
		{"empty string", "", ""},
		{"consecutive br tags", "a<br/><br/>b", "a\n\nb"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := replaceBrTags(tt.input)
			if got != tt.want {
				t.Errorf("replaceBrTags(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
