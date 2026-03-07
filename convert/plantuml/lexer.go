package plantuml

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
	TokenString      // "quoted string"
	TokenArrow       // ->, -->, <->, etc.
	TokenColon       // :
	TokenBracketOpen // [
	TokenBracketClose // ]
	TokenParenOpen   // (
	TokenParenClose  // )
	TokenBraceOpen   // {
	TokenBraceClose  // }
	TokenAngleOpen   // <
	TokenAngleClose  // >
	TokenComment     // ' comment or /' ... '/
	TokenDirective   // @startuml, @enduml
	TokenStereo      // <<stereotype>>
)

// Token represents a lexical token.
type Token struct {
	Type  TokenType
	Value string
	Line  int
	Col   int
}

// Lexer tokenizes PlantUML source code.
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

	// Single-line comment
	if ch == '\'' {
		l.scanSingleLineComment()
		return
	}

	// Multi-line comment
	if ch == '/' && l.peek(1) == '\'' {
		l.scanMultiLineComment()
		return
	}

	// Directive (@startuml, @enduml)
	if ch == '@' {
		l.scanDirective()
		return
	}

	// Stereotype <<...>>
	if ch == '<' && l.peek(1) == '<' {
		l.scanStereotype()
		return
	}

	// String
	if ch == '"' {
		l.scanString()
		return
	}

	// Arrow detection
	if l.isArrowChar(ch) {
		if arrow := l.scanArrow(); arrow != "" {
			l.tokens = append(l.tokens, Token{Type: TokenArrow, Value: arrow, Line: l.line, Col: l.col})
			return
		}
	}

	// Single character tokens
	switch ch {
	case ':':
		l.tokens = append(l.tokens, Token{Type: TokenColon, Value: ":", Line: l.line, Col: l.col})
		l.advance()
		return
	case '[':
		l.tokens = append(l.tokens, Token{Type: TokenBracketOpen, Value: "[", Line: l.line, Col: l.col})
		l.advance()
		return
	case ']':
		l.tokens = append(l.tokens, Token{Type: TokenBracketClose, Value: "]", Line: l.line, Col: l.col})
		l.advance()
		return
	case '(':
		l.tokens = append(l.tokens, Token{Type: TokenParenOpen, Value: "(", Line: l.line, Col: l.col})
		l.advance()
		return
	case ')':
		l.tokens = append(l.tokens, Token{Type: TokenParenClose, Value: ")", Line: l.line, Col: l.col})
		l.advance()
		return
	case '{':
		l.tokens = append(l.tokens, Token{Type: TokenBraceOpen, Value: "{", Line: l.line, Col: l.col})
		l.advance()
		return
	case '}':
		l.tokens = append(l.tokens, Token{Type: TokenBraceClose, Value: "}", Line: l.line, Col: l.col})
		l.advance()
		return
	case '<':
		l.tokens = append(l.tokens, Token{Type: TokenAngleOpen, Value: "<", Line: l.line, Col: l.col})
		l.advance()
		return
	case '>':
		l.tokens = append(l.tokens, Token{Type: TokenAngleClose, Value: ">", Line: l.line, Col: l.col})
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

func (l *Lexer) scanSingleLineComment() {
	startCol := l.col
	startLine := l.line
	var value strings.Builder

	// Skip '
	l.advance()

	// Read until end of line
	for l.pos < len(l.input) && l.input[l.pos] != '\n' {
		value.WriteByte(l.input[l.pos])
		l.advance()
	}

	l.tokens = append(l.tokens, Token{Type: TokenComment, Value: value.String(), Line: startLine, Col: startCol})
}

func (l *Lexer) scanMultiLineComment() {
	startCol := l.col
	startLine := l.line
	var value strings.Builder

	// Skip /'
	l.advance()
	l.advance()

	// Read until '/
	for l.pos < len(l.input) {
		if l.input[l.pos] == '\'' && l.peek(1) == '/' {
			l.advance()
			l.advance()
			break
		}
		if l.input[l.pos] == '\n' {
			l.line++
			l.col = 0
		}
		value.WriteByte(l.input[l.pos])
		l.advance()
	}

	l.tokens = append(l.tokens, Token{Type: TokenComment, Value: value.String(), Line: startLine, Col: startCol})
}

func (l *Lexer) scanDirective() {
	startCol := l.col
	startLine := l.line
	var value strings.Builder

	// Read until whitespace or newline
	for l.pos < len(l.input) && !unicode.IsSpace(rune(l.input[l.pos])) {
		value.WriteByte(l.input[l.pos])
		l.advance()
	}

	l.tokens = append(l.tokens, Token{Type: TokenDirective, Value: value.String(), Line: startLine, Col: startCol})
}

func (l *Lexer) scanStereotype() {
	startCol := l.col
	startLine := l.line
	var value strings.Builder

	// Skip <<
	l.advance()
	l.advance()

	// Read until >>
	for l.pos < len(l.input) {
		if l.input[l.pos] == '>' && l.peek(1) == '>' {
			l.advance()
			l.advance()
			break
		}
		value.WriteByte(l.input[l.pos])
		l.advance()
	}

	l.tokens = append(l.tokens, Token{Type: TokenStereo, Value: value.String(), Line: startLine, Col: startCol})
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

func (l *Lexer) isArrowChar(ch byte) bool {
	return ch == '-' || ch == '.' || ch == '<' || ch == '>' || ch == 'o' || ch == 'x' || ch == '*' || ch == '#'
}

func (l *Lexer) scanArrow() string {
	start := l.pos
	var arrow strings.Builder

	// Read arrow characters
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if l.isArrowChar(ch) || (ch == '|' && arrow.Len() > 0) {
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
	return ""
}

func (l *Lexer) isValidArrow(s string) bool {
	// Common PlantUML arrow patterns
	arrows := []string{
		// Basic arrows
		"->", "-->", "->>", "-->>",
		"<-", "<--", "<<-", "<<--",
		"<->", "<-->", "<<->>",
		// Dotted
		".>", "..>", ".>>", "..>>",
		"<.", "<..", "<<.", "<<..",
		// Class diagram arrows
		"--|>", "<|--", "..|>", "<|..",
		"--*", "*--", "--o", "o--",
		"--#", "#--",
		// Association
		"--", "..",
	}

	for _, a := range arrows {
		if s == a {
			return true
		}
	}

	// Allow arrows with color [#color]
	if strings.Contains(s, "[#") {
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
		value.WriteByte(l.input[l.pos])
		l.advance()
	}

	l.tokens = append(l.tokens, Token{Type: TokenIdent, Value: value.String(), Line: startLine, Col: startCol})
}
