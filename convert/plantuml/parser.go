package plantuml

import (
	"fmt"
	"strings"
)

// Parser parses PlantUML source code into an AST.
type Parser struct {
	tokens []Token
	pos    int
	doc    *Document
}

// NewParser creates a new parser for the given tokens.
func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens: tokens,
		pos:    0,
		doc:    &Document{},
	}
}

// Parse parses the tokens into a Document.
func (p *Parser) Parse() (*Document, error) {
	// Skip any leading whitespace/comments
	p.skipWhitespace()

	// Check for @startuml
	if p.current().Type == TokenDirective && strings.ToLower(p.current().Value) == "@startuml" {
		p.advance()
		p.skipToNextLine()
	}

	// Detect diagram type from content
	p.doc.Type = p.detectDiagramType()

	// Parse based on diagram type
	switch p.doc.Type {
	case DiagramSequence:
		if err := p.parseSequence(); err != nil {
			return nil, err
		}
	case DiagramClass:
		if err := p.parseClass(); err != nil {
			return nil, err
		}
	case DiagramComponent:
		if err := p.parseComponent(); err != nil {
			return nil, err
		}
	default:
		// Try generic parsing
		if err := p.parseGeneric(); err != nil {
			return nil, err
		}
	}

	return p.doc, nil
}

func (p *Parser) current() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) advance() Token {
	tok := p.current()
	p.pos++
	return tok
}

func (p *Parser) skipWhitespace() {
	for p.current().Type == TokenNewline || p.current().Type == TokenWhitespace || p.current().Type == TokenComment {
		p.advance()
	}
}

func (p *Parser) skipToNextLine() {
	for p.current().Type != TokenNewline && p.current().Type != TokenEOF {
		p.advance()
	}
	if p.current().Type == TokenNewline {
		p.advance()
	}
}

func (p *Parser) detectDiagramType() DiagramType {
	// Scan through tokens looking for type indicators
	for i, tok := range p.tokens {
		if tok.Type == TokenIdent {
			lower := strings.ToLower(tok.Value)
			switch lower {
			case "participant", "actor", "boundary", "control", "entity", "database", "collections", "queue":
				return DiagramSequence
			case "class", "interface", "abstract", "annotation", "enum":
				return DiagramClass
			case "component", "package", "node", "folder", "frame", "cloud", "rectangle":
				// Check if it's followed by brackets (component) or just used in relations
				if i+1 < len(p.tokens) && (p.tokens[i+1].Type == TokenIdent || p.tokens[i+1].Type == TokenString || p.tokens[i+1].Type == TokenBraceOpen) {
					return DiagramComponent
				}
			}
		}
		if tok.Type == TokenArrow {
			// Check surrounding context for type hints
			if i > 0 && p.tokens[i-1].Type == TokenIdent {
				// Could be any diagram with relations
				continue
			}
		}
	}

	// Default to component if we see [Component] notation
	for _, tok := range p.tokens {
		if tok.Type == TokenBracketOpen {
			return DiagramComponent
		}
	}

	// Default to sequence if we see message-like arrows
	for i, tok := range p.tokens {
		if tok.Type == TokenArrow && i > 0 && i+1 < len(p.tokens) {
			if p.tokens[i-1].Type == TokenIdent && p.tokens[i+1].Type == TokenIdent {
				if strings.Contains(tok.Value, ">") || strings.Contains(tok.Value, "<") {
					return DiagramSequence
				}
			}
		}
	}

	return DiagramUnknown
}

func (p *Parser) parseSequence() error {
	for p.current().Type != TokenEOF {
		p.skipWhitespace()

		tok := p.current()
		if tok.Type == TokenEOF {
			break
		}

		// Check for @enduml
		if tok.Type == TokenDirective && strings.ToLower(tok.Value) == "@enduml" {
			break
		}

		if tok.Type != TokenIdent {
			p.skipToNextLine()
			continue
		}

		lower := strings.ToLower(tok.Value)

		switch lower {
		case "participant", "actor", "boundary", "control", "entity", "database", "collections", "queue":
			part, err := p.parseParticipant(lower)
			if err != nil {
				p.skipToNextLine()
				continue
			}
			p.doc.Participants = append(p.doc.Participants, part)

		case "title":
			p.advance()
			p.doc.Title = p.readRestOfLine()

		case "alt", "opt", "loop", "par", "break", "critical", "group", "ref":
			group, err := p.parseGroup(lower)
			if err != nil {
				p.skipToNextLine()
				continue
			}
			p.doc.Groups = append(p.doc.Groups, group)

		case "note":
			note, err := p.parseNote()
			if err != nil {
				p.skipToNextLine()
				continue
			}
			p.doc.Notes = append(p.doc.Notes, note)

		case "activate", "deactivate", "destroy":
			p.skipToNextLine()

		case "autonumber":
			p.skipToNextLine()

		case "newpage", "skinparam", "hide":
			p.skipToNextLine()

		case "end":
			p.advance()
			p.skipToNextLine()

		default:
			// Try to parse as message
			if msg := p.tryParseMessage(); msg != nil {
				p.doc.Messages = append(p.doc.Messages, msg)
			} else {
				p.skipToNextLine()
			}
		}
	}

	return nil
}

func (p *Parser) parseParticipant(partType string) (*Participant, error) {
	line := p.current().Line
	p.advance() // skip participant keyword

	part := &Participant{
		Type: ParticipantType(partType),
		Line: line,
	}

	// Parse participant ID or string
	if p.current().Type == TokenString {
		part.ID = p.current().Value
		part.Label = part.ID
		p.advance()
	} else if p.current().Type == TokenIdent {
		part.ID = p.current().Value
		part.Label = part.ID
		p.advance()
	} else {
		return nil, fmt.Errorf("expected participant ID")
	}

	// Check for stereotype
	if p.current().Type == TokenStereo {
		part.Stereot = p.current().Value
		p.advance()
	}

	// Check for 'as' alias
	if p.current().Type == TokenIdent && strings.ToLower(p.current().Value) == "as" {
		p.advance()
		if p.current().Type == TokenIdent || p.current().Type == TokenString {
			part.ID = p.current().Value
			p.advance()
		}
	}

	p.skipToNextLine()
	return part, nil
}

func (p *Parser) tryParseMessage() *Message {
	if p.current().Type != TokenIdent {
		return nil
	}

	line := p.current().Line
	from := p.current().Value
	p.advance()

	// Expect arrow
	if p.current().Type != TokenArrow {
		return nil
	}

	arrowTok := p.advance()
	style := parseArrowStyle(arrowTok.Value)

	// Parse target
	if p.current().Type != TokenIdent {
		return nil
	}
	to := p.current().Value
	p.advance()

	// Check for label after colon
	label := ""
	if p.current().Type == TokenColon {
		p.advance()
		label = p.readRestOfLine()
	} else {
		p.skipToNextLine()
	}

	return &Message{
		From:  from,
		To:    to,
		Label: strings.TrimSpace(label),
		Style: style,
		Line:  line,
	}
}

func parseArrowStyle(arrow string) MessageStyle {
	if strings.Contains(arrow, "--") || strings.Contains(arrow, "..") {
		return MessageDashed
	}
	if strings.HasSuffix(arrow, ">>") {
		return MessageAsync
	}
	return MessageSolid
}

func (p *Parser) parseGroup(groupType string) (*Group, error) {
	line := p.current().Line
	p.advance() // skip group keyword

	group := &Group{
		Type: GroupType(groupType),
		Line: line,
	}

	// Parse label (rest of line)
	group.Label = strings.TrimSpace(p.readRestOfLine())

	// Parse messages until 'end' or 'else'
	for p.current().Type != TokenEOF {
		p.skipWhitespace()

		tok := p.current()
		if tok.Type == TokenIdent {
			lower := strings.ToLower(tok.Value)
			switch lower {
			case "end":
				p.advance()
				p.skipToNextLine()
				return group, nil
			case "else":
				p.advance()
				elseLabel := strings.TrimSpace(p.readRestOfLine())
				elseBranch := &GroupElse{Label: elseLabel}

				// Parse else messages until end or another else
				for p.current().Type != TokenEOF {
					p.skipWhitespace()
					if p.current().Type == TokenIdent {
						ident := strings.ToLower(p.current().Value)
						if ident == "end" {
							p.advance()
							p.skipToNextLine()
							group.Else = append(group.Else, elseBranch)
							return group, nil
						}
						if ident == "else" {
							group.Else = append(group.Else, elseBranch)
							break
						}
					}
					if msg := p.tryParseMessage(); msg != nil {
						elseBranch.Messages = append(elseBranch.Messages, msg)
					} else {
						p.skipToNextLine()
					}
				}
			default:
				if msg := p.tryParseMessage(); msg != nil {
					group.Messages = append(group.Messages, msg)
				} else {
					p.skipToNextLine()
				}
			}
		} else {
			p.skipToNextLine()
		}
	}

	return group, nil
}

func (p *Parser) parseNote() (*Note, error) {
	line := p.current().Line
	p.advance() // skip 'note'

	note := &Note{
		Line: line,
	}

	// Parse position (left, right, over)
	if p.current().Type == TokenIdent {
		pos := strings.ToLower(p.current().Value)
		if pos == "left" || pos == "right" || pos == "over" {
			note.Position = pos
			p.advance()

			// Check for 'of' keyword
			if p.current().Type == TokenIdent && strings.ToLower(p.current().Value) == "of" {
				p.advance()
			}

			// Get target participant
			if p.current().Type == TokenIdent {
				note.Target = p.current().Value
				p.advance()
			}
		}
	}

	// Check for colon (inline note)
	if p.current().Type == TokenColon {
		p.advance()
		note.Text = strings.TrimSpace(p.readRestOfLine())
		return note, nil
	}

	// Multi-line note until "end note"
	var text strings.Builder
	p.skipToNextLine()

	for p.current().Type != TokenEOF {
		if p.current().Type == TokenIdent {
			lower := strings.ToLower(p.current().Value)
			if lower == "end" {
				p.advance()
				if p.current().Type == TokenIdent && strings.ToLower(p.current().Value) == "note" {
					p.advance()
					p.skipToNextLine()
					break
				}
			}
		}
		text.WriteString(p.readRestOfLine())
		text.WriteString("\n")
	}

	note.Text = strings.TrimSpace(text.String())
	return note, nil
}

func (p *Parser) parseClass() error {
	for p.current().Type != TokenEOF {
		p.skipWhitespace()

		tok := p.current()
		if tok.Type == TokenEOF {
			break
		}

		// Check for @enduml
		if tok.Type == TokenDirective && strings.ToLower(tok.Value) == "@enduml" {
			break
		}

		if tok.Type != TokenIdent {
			p.skipToNextLine()
			continue
		}

		lower := strings.ToLower(tok.Value)

		switch lower {
		case "class", "interface", "abstract", "annotation", "enum":
			class, err := p.parseClassDef(lower)
			if err != nil {
				p.skipToNextLine()
				continue
			}
			p.doc.Classes = append(p.doc.Classes, class)

		case "package", "namespace":
			pkg, err := p.parsePackage("package")
			if err != nil {
				p.skipToNextLine()
				continue
			}
			p.doc.Packages = append(p.doc.Packages, pkg)

		case "skinparam", "hide", "show":
			p.skipToNextLine()

		default:
			// Try to parse as relation
			if rel := p.tryParseRelation(); rel != nil {
				p.doc.Relations = append(p.doc.Relations, rel)
			} else {
				p.skipToNextLine()
			}
		}
	}

	return nil
}

func (p *Parser) parseClassDef(classType string) (*Class, error) {
	line := p.current().Line
	p.advance() // skip class keyword

	class := &Class{
		Line:      line,
		Abstract:  classType == "abstract",
		Interface: classType == "interface",
	}

	// Parse class name
	if p.current().Type == TokenString {
		class.Label = p.current().Value
		class.ID = sanitizeID(class.Label)
		p.advance()
	} else if p.current().Type == TokenIdent {
		class.ID = p.current().Value
		class.Label = class.ID
		p.advance()
	} else {
		return nil, fmt.Errorf("expected class name")
	}

	// Check for stereotype
	if p.current().Type == TokenStereo {
		class.Stereotype = p.current().Value
		p.advance()
	}

	// Check for 'as' alias
	if p.current().Type == TokenIdent && strings.ToLower(p.current().Value) == "as" {
		p.advance()
		if p.current().Type == TokenIdent {
			class.ID = p.current().Value
			p.advance()
		}
	}

	// Check for body
	if p.current().Type == TokenBraceOpen {
		p.advance()
		p.skipToNextLine()

		// Parse class body
		for p.current().Type != TokenEOF && p.current().Type != TokenBraceClose {
			p.skipWhitespace()

			if p.current().Type == TokenBraceClose {
				break
			}

			// Read member line
			member := strings.TrimSpace(p.readRestOfLine())
			if member == "" || member == "}" {
				continue
			}

			// Methods have parentheses
			if strings.Contains(member, "(") {
				class.Methods = append(class.Methods, member)
			} else {
				class.Attributes = append(class.Attributes, member)
			}
		}

		if p.current().Type == TokenBraceClose {
			p.advance()
		}
	}

	p.skipToNextLine()
	return class, nil
}

func (p *Parser) parseComponent() error {
	for p.current().Type != TokenEOF {
		p.skipWhitespace()

		tok := p.current()
		if tok.Type == TokenEOF {
			break
		}

		// Check for @enduml
		if tok.Type == TokenDirective && strings.ToLower(tok.Value) == "@enduml" {
			break
		}

		if tok.Type == TokenBracketOpen {
			// [Component] notation - could be a component or start of a relation
			// Try relation first
			if rel := p.tryParseRelation(); rel != nil {
				p.doc.Relations = append(p.doc.Relations, rel)
				continue
			}
			// Just a component definition
			comp, err := p.parseComponentBracket()
			if err != nil {
				p.skipToNextLine()
				continue
			}
			p.doc.Components = append(p.doc.Components, comp)
			continue
		}

		if tok.Type != TokenIdent {
			p.skipToNextLine()
			continue
		}

		lower := strings.ToLower(tok.Value)

		switch lower {
		case "component":
			comp, err := p.parseComponentKeyword()
			if err != nil {
				p.skipToNextLine()
				continue
			}
			p.doc.Components = append(p.doc.Components, comp)

		case "package", "node", "folder", "frame", "cloud", "database", "rectangle":
			pkg, err := p.parsePackage(lower)
			if err != nil {
				p.skipToNextLine()
				continue
			}
			p.doc.Packages = append(p.doc.Packages, pkg)

		case "skinparam", "hide", "show":
			p.skipToNextLine()

		default:
			// Try to parse as relation
			if rel := p.tryParseRelation(); rel != nil {
				p.doc.Relations = append(p.doc.Relations, rel)
			} else {
				p.skipToNextLine()
			}
		}
	}

	return nil
}

func (p *Parser) parseComponentBracket() (*Component, error) {
	line := p.current().Line

	// Skip [
	p.advance()

	comp := &Component{
		Line: line,
	}

	// Read component name until ]
	var name strings.Builder
	for p.current().Type != TokenEOF && p.current().Type != TokenBracketClose && p.current().Type != TokenNewline {
		name.WriteString(p.current().Value)
		p.advance()
	}

	comp.Label = strings.TrimSpace(name.String())
	comp.ID = sanitizeID(comp.Label)

	if p.current().Type == TokenBracketClose {
		p.advance()
	}

	// Check for stereotype
	if p.current().Type == TokenStereo {
		comp.Stereotype = p.current().Value
		p.advance()
	}

	// Check for 'as' alias
	if p.current().Type == TokenIdent && strings.ToLower(p.current().Value) == "as" {
		p.advance()
		if p.current().Type == TokenIdent {
			comp.ID = p.current().Value
			p.advance()
		}
	}

	p.skipToNextLine()
	return comp, nil
}

func (p *Parser) parseComponentKeyword() (*Component, error) {
	line := p.current().Line
	p.advance() // skip 'component'

	comp := &Component{
		Line: line,
	}

	// Parse component name
	if p.current().Type == TokenString {
		comp.Label = p.current().Value
		comp.ID = sanitizeID(comp.Label)
		p.advance()
	} else if p.current().Type == TokenIdent {
		comp.ID = p.current().Value
		comp.Label = comp.ID
		p.advance()
	} else if p.current().Type == TokenBracketOpen {
		// component [Name] syntax
		p.advance()
		var name strings.Builder
		for p.current().Type != TokenEOF && p.current().Type != TokenBracketClose {
			name.WriteString(p.current().Value)
			p.advance()
		}
		comp.Label = strings.TrimSpace(name.String())
		comp.ID = sanitizeID(comp.Label)
		if p.current().Type == TokenBracketClose {
			p.advance()
		}
	}

	// Check for stereotype
	if p.current().Type == TokenStereo {
		comp.Stereotype = p.current().Value
		p.advance()
	}

	// Check for 'as' alias
	if p.current().Type == TokenIdent && strings.ToLower(p.current().Value) == "as" {
		p.advance()
		if p.current().Type == TokenIdent {
			comp.ID = p.current().Value
			p.advance()
		}
	}

	p.skipToNextLine()
	return comp, nil
}

func (p *Parser) parsePackage(pkgType string) (*Package, error) {
	line := p.current().Line
	p.advance() // skip package keyword

	pkg := &Package{
		Type: PackageType(pkgType),
		Line: line,
	}

	// Parse package name
	if p.current().Type == TokenString {
		pkg.Label = p.current().Value
		pkg.ID = sanitizeID(pkg.Label)
		p.advance()
	} else if p.current().Type == TokenIdent {
		pkg.ID = p.current().Value
		pkg.Label = pkg.ID
		p.advance()
	}

	// Check for 'as' alias
	if p.current().Type == TokenIdent && strings.ToLower(p.current().Value) == "as" {
		p.advance()
		if p.current().Type == TokenIdent {
			pkg.ID = p.current().Value
			p.advance()
		}
	}

	// Check for body
	if p.current().Type == TokenBraceOpen {
		p.advance()
		p.skipToNextLine()

		// Parse package contents
		for p.current().Type != TokenEOF && p.current().Type != TokenBraceClose {
			p.skipWhitespace()

			if p.current().Type == TokenBraceClose {
				break
			}

			tok := p.current()
			switch tok.Type {
			case TokenIdent:
				lower := strings.ToLower(tok.Value)
				switch lower {
				case "class", "interface", "abstract":
					class, err := p.parseClassDef(lower)
					if err == nil {
						pkg.Classes = append(pkg.Classes, class)
					}
				case "component":
					comp, err := p.parseComponentKeyword()
					if err == nil {
						pkg.Components = append(pkg.Components, comp)
					}
				case "package", "node", "folder", "frame":
					nested, err := p.parsePackage(lower)
					if err == nil {
						pkg.Packages = append(pkg.Packages, nested)
					}
				default:
					p.skipToNextLine()
				}
			case TokenBracketOpen:
				comp, err := p.parseComponentBracket()
				if err == nil {
					pkg.Components = append(pkg.Components, comp)
				}
			default:
				p.skipToNextLine()
			}
		}

		if p.current().Type == TokenBraceClose {
			p.advance()
		}
	}

	p.skipToNextLine()
	return pkg, nil
}

func (p *Parser) tryParseRelation() *Relation {
	if p.current().Type != TokenIdent && p.current().Type != TokenBracketOpen {
		return nil
	}

	// Save position for backtracking
	startPos := p.pos

	line := p.current().Line
	from := ""

	if p.current().Type == TokenBracketOpen {
		// [Component] notation
		p.advance()
		var name strings.Builder
		for p.current().Type != TokenEOF && p.current().Type != TokenBracketClose {
			name.WriteString(p.current().Value)
			p.advance()
		}
		from = sanitizeID(strings.TrimSpace(name.String()))
		if p.current().Type == TokenBracketClose {
			p.advance()
		}
	} else {
		from = p.current().Value
		p.advance()
	}

	// Expect arrow
	if p.current().Type != TokenArrow {
		// Restore position
		p.pos = startPos
		return nil
	}

	arrowTok := p.advance()
	relType := parseRelationType(arrowTok.Value)

	// Parse target
	to := ""
	if p.current().Type == TokenBracketOpen {
		p.advance()
		var name strings.Builder
		for p.current().Type != TokenEOF && p.current().Type != TokenBracketClose {
			name.WriteString(p.current().Value)
			p.advance()
		}
		to = sanitizeID(strings.TrimSpace(name.String()))
		if p.current().Type == TokenBracketClose {
			p.advance()
		}
	} else if p.current().Type == TokenIdent {
		to = p.current().Value
		p.advance()
	} else {
		// Restore position
		p.pos = startPos
		return nil
	}

	// Check for label after colon
	label := ""
	if p.current().Type == TokenColon {
		p.advance()
		label = strings.TrimSpace(p.readRestOfLine())
	} else {
		p.skipToNextLine()
	}

	return &Relation{
		From:  from,
		To:    to,
		Label: label,
		Type:  relType,
		Line:  line,
	}
}

func parseRelationType(arrow string) RelationType {
	if strings.Contains(arrow, "|>") || strings.Contains(arrow, "<|") {
		if strings.Contains(arrow, "..") {
			return RelationRealization
		}
		return RelationInheritance
	}
	if strings.Contains(arrow, "*") {
		return RelationComposition
	}
	if strings.Contains(arrow, "o") {
		return RelationAggregation
	}
	if strings.Contains(arrow, "..") {
		return RelationDependency
	}
	return RelationAssociation
}

func (p *Parser) parseGeneric() error {
	// Try to parse as a mix of messages and relations
	for p.current().Type != TokenEOF {
		p.skipWhitespace()

		tok := p.current()
		if tok.Type == TokenEOF {
			break
		}

		// Check for @enduml
		if tok.Type == TokenDirective && strings.ToLower(tok.Value) == "@enduml" {
			break
		}

		if tok.Type == TokenIdent {
			// Try message first
			if msg := p.tryParseMessage(); msg != nil {
				p.doc.Messages = append(p.doc.Messages, msg)
				continue
			}
			// Then try relation
			if rel := p.tryParseRelation(); rel != nil {
				p.doc.Relations = append(p.doc.Relations, rel)
				continue
			}
		}

		p.skipToNextLine()
	}

	return nil
}

func (p *Parser) readRestOfLine() string {
	var text strings.Builder
	for p.current().Type != TokenNewline && p.current().Type != TokenEOF {
		text.WriteString(p.current().Value)
		if p.current().Type == TokenIdent || p.current().Type == TokenString {
			text.WriteString(" ")
		}
		p.advance()
	}
	if p.current().Type == TokenNewline {
		p.advance()
	}
	return strings.TrimSpace(text.String())
}

func sanitizeID(label string) string {
	// Replace spaces and special characters
	result := strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' {
			return r
		}
		return '_'
	}, label)

	// Ensure it starts with a letter
	if len(result) > 0 && result[0] >= '0' && result[0] <= '9' {
		result = "_" + result
	}

	return result
}
