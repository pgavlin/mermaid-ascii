package parser

import (
	"testing"
)

func collectTokens(input string) []Token {
	s := NewScanner(input)
	var tokens []Token
	for {
		tok := s.Next()
		tokens = append(tokens, tok)
		if tok.Kind == TokenEOF {
			break
		}
	}
	return tokens
}

func TestBasicTokens(t *testing.T) {
	tokens := collectTokens("A[hello]")
	// Should produce: Ident("A"), LBracket, Ident("hello"), RBracket, EOF
	expected := []TokenKind{TokenIdent, TokenLBracket, TokenIdent, TokenRBracket, TokenEOF}
	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d: %v", len(expected), len(tokens), tokens)
	}
	for i, kind := range expected {
		if tokens[i].Kind != kind {
			t.Errorf("token %d: expected %s, got %s", i, kind, tokens[i].Kind)
		}
	}
	if tokens[0].Text != "A" {
		t.Errorf("expected ident 'A', got %q", tokens[0].Text)
	}
	if tokens[2].Text != "hello" {
		t.Errorf("expected ident 'hello', got %q", tokens[2].Text)
	}
}

func TestComments(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []TokenKind
	}{
		{
			name:  "full line comment",
			input: "%% this is a comment\nA",
			want:  []TokenKind{TokenNewline, TokenIdent, TokenEOF},
		},
		{
			name:  "inline comment",
			input: "A %% comment",
			want:  []TokenKind{TokenIdent, TokenWhitespace, TokenEOF},
		},
		{
			name:  "comment only",
			input: "%% just a comment",
			want:  []TokenKind{TokenEOF},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := collectTokens(tt.input)
			if len(tokens) != len(tt.want) {
				t.Fatalf("expected %d tokens, got %d: %v", len(tt.want), len(tokens), tokens)
			}
			for i, kind := range tt.want {
				if tokens[i].Kind != kind {
					t.Errorf("token %d: expected %s, got %s(%q)", i, kind, tokens[i].Kind, tokens[i].Text)
				}
			}
		})
	}
}

func TestCommentInsideQuotedString(t *testing.T) {
	tokens := collectTokens(`"text %% here"`)
	// The %% inside quotes should NOT be treated as a comment
	if len(tokens) != 2 { // String + EOF
		t.Fatalf("expected 2 tokens, got %d: %v", len(tokens), tokens)
	}
	if tokens[0].Kind != TokenString {
		t.Errorf("expected String, got %s", tokens[0].Kind)
	}
	if tokens[0].Text != "text %% here" {
		t.Errorf("expected %q, got %q", "text %% here", tokens[0].Text)
	}
}

func TestQuotedStrings(t *testing.T) {
	tests := []struct {
		name string
		input string
		want  string
	}{
		{"basic", `"hello"`, "hello"},
		{"escaped quote", `"say \"hi\""`, `say "hi"`},
		{"escaped backslash", `"path\\file"`, `path\file`},
		{"unterminated", `"open`, "open"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := collectTokens(tt.input)
			if tokens[0].Kind != TokenString {
				t.Fatalf("expected String, got %s", tokens[0].Kind)
			}
			if tokens[0].Text != tt.want {
				t.Errorf("expected %q, got %q", tt.want, tokens[0].Text)
			}
		})
	}
}

func TestOperators(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"-->", "-->"},
		{"->>", "->>"},
		{"-.->", "-.->"},
		{"==>", "==>"},
		{"<-->", "<-->"},
		{"---", "---"},
		{"-.-", "-.-"},
		{"===", "==="},
		{"--x", "--x"},
		{"--o", "--o"},
		{"-)", "-)"},
		{"--)", "--)"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokens := collectTokens(tt.input)
			if tokens[0].Kind != TokenOperator {
				t.Fatalf("expected Operator, got %s(%q)", tokens[0].Kind, tokens[0].Text)
			}
			if tokens[0].Text != tt.want {
				t.Errorf("expected %q, got %q", tt.want, tokens[0].Text)
			}
		})
	}
}

func TestNumbers(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"42", "42"},
		{"3.14", "3.14"},
		{"0", "0"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokens := collectTokens(tt.input)
			if tokens[0].Kind != TokenNumber {
				t.Fatalf("expected Number, got %s", tokens[0].Kind)
			}
			if tokens[0].Text != tt.want {
				t.Errorf("expected %q, got %q", tt.want, tokens[0].Text)
			}
		})
	}
}

func TestPositionTracking(t *testing.T) {
	tokens := collectTokens("A\nB\nC")
	// A at line 1, B at line 2, C at line 3
	identTokens := []Token{}
	for _, tok := range tokens {
		if tok.Kind == TokenIdent {
			identTokens = append(identTokens, tok)
		}
	}
	if len(identTokens) != 3 {
		t.Fatalf("expected 3 idents, got %d", len(identTokens))
	}
	if identTokens[0].Pos.Line != 1 || identTokens[0].Pos.Column != 1 {
		t.Errorf("A: expected line 1 col 1, got line %d col %d", identTokens[0].Pos.Line, identTokens[0].Pos.Column)
	}
	if identTokens[1].Pos.Line != 2 || identTokens[1].Pos.Column != 1 {
		t.Errorf("B: expected line 2 col 1, got line %d col %d", identTokens[1].Pos.Line, identTokens[1].Pos.Column)
	}
	if identTokens[2].Pos.Line != 3 || identTokens[2].Pos.Column != 1 {
		t.Errorf("C: expected line 3 col 1, got line %d col %d", identTokens[2].Pos.Line, identTokens[2].Pos.Column)
	}
}

func TestPunctuation(t *testing.T) {
	tokens := collectTokens("()[]{}:,|#@&;")
	expected := []TokenKind{
		TokenLParen, TokenRParen,
		TokenLBracket, TokenRBracket,
		TokenLBrace, TokenRBrace,
		TokenColon, TokenComma,
		TokenPipe, TokenHash,
		TokenAt, TokenAmpersand,
		TokenSemicolon, TokenEOF,
	}
	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d: %v", len(expected), len(tokens), tokens)
	}
	for i, kind := range expected {
		if tokens[i].Kind != kind {
			t.Errorf("token %d: expected %s, got %s", i, kind, tokens[i].Kind)
		}
	}
}

func TestDot(t *testing.T) {
	// A lone dot should be TokenDot, not TokenOperator
	tokens := collectTokens(".")
	if tokens[0].Kind != TokenDot {
		t.Errorf("expected Dot, got %s(%q)", tokens[0].Kind, tokens[0].Text)
	}
}

func TestPeekAndNext(t *testing.T) {
	s := NewScanner("AB")
	// Peek should not advance
	tok1 := s.Peek()
	tok2 := s.Peek()
	if tok1.Text != tok2.Text {
		t.Error("Peek should return same token")
	}
	// Next should advance
	tok3 := s.Next()
	if tok3.Text != "AB" {
		t.Errorf("expected 'AB', got %q", tok3.Text)
	}
	// Next should now return EOF
	tok4 := s.Next()
	if tok4.Kind != TokenEOF {
		t.Errorf("expected EOF, got %s", tok4.Kind)
	}
}

func TestExpect(t *testing.T) {
	s := NewScanner("A")
	_, err := s.Expect(TokenIdent)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	s = NewScanner("42")
	_, err = s.Expect(TokenIdent)
	if err == nil {
		t.Error("expected error for wrong token kind")
	}
}

func TestSkipWhitespace(t *testing.T) {
	s := NewScanner("  \t A")
	s.SkipWhitespace()
	tok := s.Next()
	if tok.Kind != TokenIdent || tok.Text != "A" {
		t.Errorf("expected Ident(A) after SkipWhitespace, got %s(%q)", tok.Kind, tok.Text)
	}
}

func TestSkipNewlines(t *testing.T) {
	s := NewScanner("\n\n  \n A")
	s.SkipNewlines()
	tok := s.Next()
	if tok.Kind != TokenIdent || tok.Text != "A" {
		t.Errorf("expected Ident(A) after SkipNewlines, got %s(%q)", tok.Kind, tok.Text)
	}
}

func TestHelperParseQuotedOrIdent(t *testing.T) {
	s := NewScanner(`"hello" world`)
	text, err := ParseQuotedOrIdent(s)
	if err != nil || text != "hello" {
		t.Errorf("expected 'hello', got %q, err=%v", text, err)
	}
	s.SkipWhitespace()
	text, err = ParseQuotedOrIdent(s)
	if err != nil || text != "world" {
		t.Errorf("expected 'world', got %q, err=%v", text, err)
	}
}

func TestHelperConsumeRestOfLine(t *testing.T) {
	s := NewScanner("hello world\nnext")
	text := ConsumeRestOfLine(s)
	if text != "hello world" {
		t.Errorf("expected 'hello world', got %q", text)
	}
	// Should now be at "next"
	tok := s.Next()
	if tok.Kind != TokenIdent || tok.Text != "next" {
		t.Errorf("expected Ident(next), got %s(%q)", tok.Kind, tok.Text)
	}
}

func TestGraphLine(t *testing.T) {
	// Test a realistic graph line: A --> B
	tokens := collectTokens("A --> B")
	expected := []struct {
		kind TokenKind
		text string
	}{
		{TokenIdent, "A"},
		{TokenWhitespace, " "},
		{TokenOperator, "-->"},
		{TokenWhitespace, " "},
		{TokenIdent, "B"},
		{TokenEOF, ""},
	}
	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d: %v", len(expected), len(tokens), tokens)
	}
	for i, exp := range expected {
		if tokens[i].Kind != exp.kind {
			t.Errorf("token %d: expected %s(%q), got %s(%q)", i, exp.kind, exp.text, tokens[i].Kind, tokens[i].Text)
		}
	}
}

func TestNodeWithShape(t *testing.T) {
	// A[hello world]
	tokens := collectTokens("A[hello world]")
	kinds := []TokenKind{}
	for _, tok := range tokens {
		kinds = append(kinds, tok.Kind)
	}
	// A, [, hello, ws, world, ], EOF
	if tokens[0].Kind != TokenIdent || tokens[0].Text != "A" {
		t.Errorf("expected Ident(A), got %s(%q)", tokens[0].Kind, tokens[0].Text)
	}
	if tokens[1].Kind != TokenLBracket {
		t.Errorf("expected LBracket, got %s", tokens[1].Kind)
	}
}

func TestAtEnd(t *testing.T) {
	s := NewScanner("")
	if !s.AtEnd() {
		t.Error("expected AtEnd for empty input")
	}
	s = NewScanner("A")
	if s.AtEnd() {
		t.Error("expected not AtEnd for non-empty input")
	}
}
