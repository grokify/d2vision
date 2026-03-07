package mermaid

import (
	"strings"
	"unicode"
)

// TokenType represents the type of a lexical token.
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenNewline
	TokenWhitespace
	TokenIdent
	TokenString     // "quoted string"
	TokenText       // Text in brackets/braces
	TokenArrow      // -->, -.->. ==>, etc.
	TokenPipe       // |
	TokenColon      // :
	TokenSemicolon  // ;
	TokenOpenParen  // (
	TokenCloseParen // )
	TokenOpenBrace  // {
	TokenCloseBrace // }
	TokenOpenBracket  // [
	TokenCloseBracket // ]
	TokenComment    // %% comment
	TokenDirective  // %%{...}%%
)

// Token represents a lexical token.
type Token struct {
	Type  TokenType
	Value string
	Line  int
	Col   int
}

// Lexer tokenizes Mermaid source code.
type Lexer struct {
	input  string
	pos    int
	line   int
	col    int
	tokens []Token
}

// NewLexer creates a new lexer for the given input.
func NewLexer(input string) *Lexer {
	return &Lexer{
		input:  input,
		pos:    0,
		line:   1,
		col:    1,
		tokens: []Token{},
	}
}

// Tokenize converts the input into tokens.
func (l *Lexer) Tokenize() []Token {
	for l.pos < len(l.input) {
		l.scanToken()
	}
	l.tokens = append(l.tokens, Token{Type: TokenEOF, Line: l.line, Col: l.col})
	return l.tokens
}

func (l *Lexer) scanToken() {
	ch := l.input[l.pos]

	// Skip whitespace (except newlines)
	if ch == ' ' || ch == '\t' || ch == '\r' {
		l.advance()
		return
	}

	// Newline
	if ch == '\n' {
		l.tokens = append(l.tokens, Token{Type: TokenNewline, Value: "\n", Line: l.line, Col: l.col})
		l.line++
		l.col = 1
		l.pos++
		return
	}

	// Comment or directive
	if ch == '%' && l.peek(1) == '%' {
		if l.peek(2) == '{' {
			l.scanDirective()
		} else {
			l.scanComment()
		}
		return
	}

	// String
	if ch == '"' {
		l.scanString()
		return
	}

	// Arrow detection
	if l.isArrowStart(ch) {
		if arrow := l.scanArrow(); arrow != "" {
			l.tokens = append(l.tokens, Token{Type: TokenArrow, Value: arrow, Line: l.line, Col: l.col})
			return
		}
	}

	// Single character tokens
	switch ch {
	case '|':
		l.tokens = append(l.tokens, Token{Type: TokenPipe, Value: "|", Line: l.line, Col: l.col})
		l.advance()
		return
	case ':':
		l.tokens = append(l.tokens, Token{Type: TokenColon, Value: ":", Line: l.line, Col: l.col})
		l.advance()
		return
	case ';':
		l.tokens = append(l.tokens, Token{Type: TokenSemicolon, Value: ";", Line: l.line, Col: l.col})
		l.advance()
		return
	case '(':
		l.scanBracketedText('(', ')')
		return
	case '[':
		l.scanBracketedText('[', ']')
		return
	case '{':
		l.scanBracketedText('{', '}')
		return
	case ')':
		l.tokens = append(l.tokens, Token{Type: TokenCloseParen, Value: ")", Line: l.line, Col: l.col})
		l.advance()
		return
	case ']':
		l.tokens = append(l.tokens, Token{Type: TokenCloseBracket, Value: "]", Line: l.line, Col: l.col})
		l.advance()
		return
	case '}':
		l.tokens = append(l.tokens, Token{Type: TokenCloseBrace, Value: "}", Line: l.line, Col: l.col})
		l.advance()
		return
	}

	// Identifier
	if l.isIdentStart(ch) {
		l.scanIdent()
		return
	}

	// Skip unknown character
	l.advance()
}

func (l *Lexer) peek(offset int) byte {
	pos := l.pos + offset
	if pos >= len(l.input) {
		return 0
	}
	return l.input[pos]
}

func (l *Lexer) advance() {
	if l.pos < len(l.input) {
		l.pos++
		l.col++
	}
}

func (l *Lexer) scanComment() {
	startCol := l.col
	startLine := l.line
	var value strings.Builder

	// Skip %%
	l.advance()
	l.advance()

	// Read until end of line
	for l.pos < len(l.input) && l.input[l.pos] != '\n' {
		value.WriteByte(l.input[l.pos])
		l.advance()
	}

	l.tokens = append(l.tokens, Token{Type: TokenComment, Value: value.String(), Line: startLine, Col: startCol})
}

func (l *Lexer) scanDirective() {
	startCol := l.col
	startLine := l.line
	var value strings.Builder

	// Skip %%{
	l.advance()
	l.advance()
	l.advance()

	// Read until }%%
	for l.pos < len(l.input) {
		if l.input[l.pos] == '}' && l.peek(1) == '%' && l.peek(2) == '%' {
			l.advance()
			l.advance()
			l.advance()
			break
		}
		value.WriteByte(l.input[l.pos])
		l.advance()
	}

	l.tokens = append(l.tokens, Token{Type: TokenDirective, Value: value.String(), Line: startLine, Col: startCol})
}

func (l *Lexer) scanString() {
	startCol := l.col
	startLine := l.line
	var value strings.Builder

	// Skip opening quote
	l.advance()

	// Read until closing quote
	for l.pos < len(l.input) && l.input[l.pos] != '"' {
		if l.input[l.pos] == '\\' && l.peek(1) == '"' {
			l.advance()
		}
		value.WriteByte(l.input[l.pos])
		l.advance()
	}

	// Skip closing quote
	if l.pos < len(l.input) {
		l.advance()
	}

	l.tokens = append(l.tokens, Token{Type: TokenString, Value: value.String(), Line: startLine, Col: startCol})
}

func (l *Lexer) isArrowStart(ch byte) bool {
	return ch == '-' || ch == '=' || ch == '.' || ch == '<' || ch == '~' || ch == 'o' || ch == 'x'
}

func (l *Lexer) scanArrow() string {
	start := l.pos
	var arrow strings.Builder

	// Read arrow characters
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == '-' || ch == '=' || ch == '.' || ch == '<' || ch == '>' || ch == '~' || ch == 'o' || ch == 'x' {
			arrow.WriteByte(ch)
			l.pos++
			l.col++
		} else {
			break
		}
	}

	arrowStr := arrow.String()

	// Validate it's actually an arrow
	if l.isValidArrow(arrowStr) {
		return arrowStr
	}

	// Not an arrow, revert
	l.pos = start
	l.col -= len(arrowStr)
	return ""
}

func (l *Lexer) isValidArrow(s string) bool {
	// Common Mermaid arrow patterns
	arrows := []string{
		"-->", "---", "-.->", "-.-", "==>", "===",
		"--", "->", "=>",
		"<-->", "<-.->", "<==>",
		"--x", "-->x", "x--x",
		"--o", "-->o", "o--o",
		// Sequence diagram arrows
		"->>", "-->>", "<<-", "<<--",
		"-x", "--x", "-))", "--))",
	}

	for _, a := range arrows {
		if s == a {
			return true
		}
	}

	// Check for arrows with text: --text-->
	if strings.HasPrefix(s, "--") && strings.HasSuffix(s, "-->") {
		return true
	}
	if strings.HasPrefix(s, "-.") && strings.HasSuffix(s, ".->") {
		return true
	}
	if strings.HasPrefix(s, "==") && strings.HasSuffix(s, "==>") {
		return true
	}

	return false
}

func (l *Lexer) isIdentStart(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || ch == '_'
}

func (l *Lexer) isIdentChar(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || unicode.IsDigit(rune(ch)) || ch == '_' || ch == '-'
}

func (l *Lexer) scanIdent() {
	startCol := l.col
	startLine := l.line
	var value strings.Builder

	for l.pos < len(l.input) && l.isIdentChar(l.input[l.pos]) {
		ch := l.input[l.pos]
		// Check if '-' might be start of arrow
		if ch == '-' {
			// Look ahead to see if this is an arrow
			next := l.peek(1)
			if next == '>' || next == '-' || next == '.' {
				// This is likely an arrow, stop identifier here
				break
			}
		}
		value.WriteByte(ch)
		l.advance()
	}

	l.tokens = append(l.tokens, Token{Type: TokenIdent, Value: value.String(), Line: startLine, Col: startCol})
}

func (l *Lexer) scanBracketedText(open, close byte) {
	startCol := l.col
	startLine := l.line
	var tokenType TokenType

	switch open {
	case '(':
		tokenType = TokenOpenParen
	case '[':
		tokenType = TokenOpenBracket
	case '{':
		tokenType = TokenOpenBrace
	}

	var value strings.Builder
	depth := 1

	// Skip opening bracket
	l.advance()

	// Handle special Mermaid bracket combinations
	// [(text)] - cylinder
	// ((text)) - circle
	// {{text}} - hexagon
	// [/text/] - parallelogram
	// [\text\] - parallelogram alt
	// [/text\] - trapezoid
	// [\text/] - trapezoid alt
	// ([text]) - stadium

	// Read content
	for l.pos < len(l.input) && depth > 0 {
		ch := l.input[l.pos]
		if ch == open {
			depth++
			value.WriteByte(ch)
		} else if ch == close {
			depth--
			if depth > 0 {
				value.WriteByte(ch)
			}
		} else if ch == '\n' {
			break
		} else {
			value.WriteByte(ch)
		}
		l.advance()
	}

	l.tokens = append(l.tokens, Token{Type: tokenType, Value: value.String(), Line: startLine, Col: startCol})
}
