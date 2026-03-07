package mermaid

import (
	"fmt"
	"regexp"
	"strings"
)

// Parser parses Mermaid source code into an AST.
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

	// Detect diagram type
	if err := p.parseDiagramType(); err != nil {
		return nil, err
	}

	// Parse based on diagram type
	switch p.doc.Type {
	case DiagramFlowchart:
		if err := p.parseFlowchart(); err != nil {
			return nil, err
		}
	case DiagramSequence:
		if err := p.parseSequence(); err != nil {
			return nil, err
		}
	case DiagramClass:
		if err := p.parseClass(); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported diagram type: %s", p.doc.Type)
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

func (p *Parser) parseDiagramType() error {
	tok := p.current()

	if tok.Type != TokenIdent {
		return fmt.Errorf("expected diagram type, got %v", tok.Type)
	}

	switch strings.ToLower(tok.Value) {
	case "graph", "flowchart":
		p.doc.Type = DiagramFlowchart
		p.advance()
		// Parse direction
		p.parseDirection()
	case "sequencediagram":
		p.doc.Type = DiagramSequence
		p.advance()
	case "classdiagram":
		p.doc.Type = DiagramClass
		p.advance()
	case "statediagram", "statediagram-v2":
		p.doc.Type = DiagramState
		p.advance()
	case "erdiagram":
		p.doc.Type = DiagramER
		p.advance()
	case "gantt":
		p.doc.Type = DiagramGantt
		p.advance()
	case "pie":
		p.doc.Type = DiagramPie
		p.advance()
	case "gitgraph":
		p.doc.Type = DiagramGitGraph
		p.advance()
	default:
		return fmt.Errorf("unknown diagram type: %s", tok.Value)
	}

	p.skipToNextLine()
	return nil
}

func (p *Parser) parseDirection() {
	tok := p.current()
	if tok.Type == TokenIdent {
		switch strings.ToUpper(tok.Value) {
		case "TB", "TD":
			p.doc.Direction = DirectionTB
		case "BT":
			p.doc.Direction = DirectionBT
		case "LR":
			p.doc.Direction = DirectionLR
		case "RL":
			p.doc.Direction = DirectionRL
		}
		p.advance()
	}
}

func (p *Parser) parseFlowchart() error {
	for p.current().Type != TokenEOF {
		p.skipWhitespace()

		tok := p.current()
		if tok.Type == TokenEOF {
			break
		}

		// Check for subgraph
		if tok.Type == TokenIdent && strings.ToLower(tok.Value) == "subgraph" {
			sg, err := p.parseSubgraph()
			if err != nil {
				return err
			}
			p.doc.Subgraphs = append(p.doc.Subgraphs, sg)
			continue
		}

		// Check for end (subgraph close)
		if tok.Type == TokenIdent && strings.ToLower(tok.Value) == "end" {
			p.advance()
			p.skipToNextLine()
			continue
		}

		// Check for direction keyword inside flowchart
		if tok.Type == TokenIdent && strings.ToLower(tok.Value) == "direction" {
			p.advance()
			p.parseDirection()
			p.skipToNextLine()
			continue
		}

		// Try to parse node or edge
		if tok.Type == TokenIdent {
			if err := p.parseNodeOrEdge(); err != nil {
				// Skip problematic line
				p.skipToNextLine()
			}
			continue
		}

		// Skip unknown content
		p.skipToNextLine()
	}

	return nil
}

func (p *Parser) parseSubgraph() (*Subgraph, error) {
	// Skip 'subgraph' keyword
	p.advance()

	sg := &Subgraph{
		Line: p.current().Line,
	}

	// Parse subgraph ID/label
	// Formats: subgraph id, subgraph id[label], subgraph "label"
	tok := p.current()

	switch tok.Type {
	case TokenIdent:
		sg.ID = tok.Value
		sg.Label = tok.Value
		p.advance()

		// Check for bracket label
		if p.current().Type == TokenOpenBracket {
			sg.Label = p.current().Value
			p.advance()
		}
	case TokenString:
		sg.Label = tok.Value
		sg.ID = sanitizeID(tok.Value)
		p.advance()
	}

	p.skipToNextLine()

	// Parse subgraph contents until 'end'
	for p.current().Type != TokenEOF {
		p.skipWhitespace()

		tok := p.current()

		// Check for end
		if tok.Type == TokenIdent && strings.ToLower(tok.Value) == "end" {
			p.advance()
			p.skipToNextLine()
			break
		}

		// Check for nested subgraph
		if tok.Type == TokenIdent && strings.ToLower(tok.Value) == "subgraph" {
			nested, err := p.parseSubgraph()
			if err != nil {
				return nil, err
			}
			sg.Subgraphs = append(sg.Subgraphs, nested)
			continue
		}

		// Check for direction
		if tok.Type == TokenIdent && strings.ToLower(tok.Value) == "direction" {
			p.advance()
			dirTok := p.current()
			if dirTok.Type == TokenIdent {
				switch strings.ToUpper(dirTok.Value) {
				case "TB", "TD":
					sg.Direction = DirectionTB
				case "BT":
					sg.Direction = DirectionBT
				case "LR":
					sg.Direction = DirectionLR
				case "RL":
					sg.Direction = DirectionRL
				}
				p.advance()
			}
			p.skipToNextLine()
			continue
		}

		// Parse node or edge
		if tok.Type == TokenIdent {
			if err := p.parseSubgraphNodeOrEdge(sg); err != nil {
				p.skipToNextLine()
			}
			continue
		}

		p.skipToNextLine()
	}

	return sg, nil
}

func (p *Parser) parseNodeOrEdge() error {
	line := p.current().Line
	firstID := p.current().Value
	p.advance()

	// Check for node shape/label
	node := p.tryParseNodeDefinition(firstID, line)

	// Check for edge
	if p.current().Type == TokenArrow {
		return p.parseEdgeChain(firstID, node, nil)
	}

	// Just a node
	if node != nil {
		p.doc.Nodes = append(p.doc.Nodes, node)
	} else {
		// Implicit node with ID as label
		p.doc.Nodes = append(p.doc.Nodes, &Node{
			ID:    firstID,
			Label: firstID,
			Shape: ShapeRectangle,
			Line:  line,
		})
	}

	p.skipToNextLine()
	return nil
}

func (p *Parser) parseSubgraphNodeOrEdge(sg *Subgraph) error {
	line := p.current().Line
	firstID := p.current().Value
	p.advance()

	// Check for node shape/label
	node := p.tryParseNodeDefinition(firstID, line)

	// Check for edge
	if p.current().Type == TokenArrow {
		return p.parseEdgeChain(firstID, node, sg)
	}

	// Just a node
	if node != nil {
		sg.Nodes = append(sg.Nodes, node)
	} else {
		sg.Nodes = append(sg.Nodes, &Node{
			ID:    firstID,
			Label: firstID,
			Shape: ShapeRectangle,
			Line:  line,
		})
	}

	p.skipToNextLine()
	return nil
}

func (p *Parser) tryParseNodeDefinition(id string, line int) *Node {
	tok := p.current()

	if tok.Type == TokenOpenBracket || tok.Type == TokenOpenParen || tok.Type == TokenOpenBrace {
		shape, label := parseShapeAndLabel(tok.Type, tok.Value)
		p.advance()

		return &Node{
			ID:    id,
			Label: label,
			Shape: shape,
			Line:  line,
		}
	}

	return nil
}

func parseShapeAndLabel(tokType TokenType, content string) (NodeShape, string) {
	// Remove any trailing brackets that might be included
	content = strings.TrimSpace(content)

	switch tokType {
	case TokenOpenBracket:
		// Check for special bracket combinations
		if strings.HasPrefix(content, "(") && strings.HasSuffix(content, ")") {
			// [(text)] - cylinder
			return ShapeCylinder, strings.Trim(content, "()")
		}
		if strings.HasPrefix(content, "/") && strings.HasSuffix(content, "/") {
			// [/text/] - parallelogram
			return ShapeParallelogram, strings.Trim(content, "/")
		}
		if strings.HasPrefix(content, "\\") && strings.HasSuffix(content, "\\") {
			// [\text\] - parallelogram alt
			return ShapeParallelogram, strings.Trim(content, "\\")
		}
		if strings.HasPrefix(content, "/") && strings.HasSuffix(content, "\\") {
			// [/text\] - trapezoid
			return ShapeTrapezoid, strings.Trim(strings.TrimPrefix(content, "/"), "\\")
		}
		if strings.HasPrefix(content, "\\") && strings.HasSuffix(content, "/") {
			// [\text/] - trapezoid alt
			return ShapeTrapezoid, strings.Trim(strings.TrimPrefix(content, "\\"), "/")
		}
		// [text] - rectangle
		return ShapeRectangle, content

	case TokenOpenParen:
		// Check for special paren combinations
		if strings.HasPrefix(content, "(") && strings.HasSuffix(content, ")") {
			// ((text)) - circle
			return ShapeCircle, strings.Trim(content, "()")
		}
		if strings.HasPrefix(content, "[") && strings.HasSuffix(content, "]") {
			// ([text]) - stadium
			return ShapeStadium, strings.Trim(content, "[]")
		}
		// (text) - rounded rectangle
		return ShapeRoundedRect, content

	case TokenOpenBrace:
		// Check for special brace combinations
		if strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}") {
			// {{text}} - hexagon
			return ShapeHexagon, strings.Trim(content, "{}")
		}
		// {text} - diamond
		return ShapeDiamond, content
	}

	return ShapeRectangle, content
}

func (p *Parser) parseEdgeChain(fromID string, fromNode *Node, sg *Subgraph) error {
	line := p.current().Line
	currentID := fromID

	// Register first node if it has a definition
	if fromNode != nil {
		if sg != nil {
			sg.Nodes = append(sg.Nodes, fromNode)
		} else {
			p.doc.Nodes = append(p.doc.Nodes, fromNode)
		}
	}

	for p.current().Type == TokenArrow {
		arrowTok := p.advance()
		edge := &Edge{
			From:  currentID,
			Style: parseArrowStyle(arrowTok.Value),
			Line:  line,
		}

		// Check for edge label before target: A --text--> B or A -->|text| B
		edgeLabel := ""
		if p.current().Type == TokenPipe {
			p.advance()
			// Read until next pipe
			for p.current().Type != TokenPipe && p.current().Type != TokenNewline && p.current().Type != TokenEOF {
				edgeLabel += p.current().Value
				p.advance()
			}
			if p.current().Type == TokenPipe {
				p.advance()
			}
		}

		// Parse target node
		if p.current().Type != TokenIdent {
			return fmt.Errorf("expected node ID after arrow at line %d", line)
		}

		targetID := p.current().Value
		p.advance()

		// Check for target node shape/label
		targetNode := p.tryParseNodeDefinition(targetID, p.current().Line)
		if targetNode != nil {
			if sg != nil {
				sg.Nodes = append(sg.Nodes, targetNode)
			} else {
				p.doc.Nodes = append(p.doc.Nodes, targetNode)
			}
		}

		edge.To = targetID
		edge.Label = edgeLabel

		if sg != nil {
			sg.Edges = append(sg.Edges, edge)
		} else {
			p.doc.Edges = append(p.doc.Edges, edge)
		}

		currentID = targetID
	}

	p.skipToNextLine()
	return nil
}

func parseArrowStyle(arrow string) EdgeStyle {
	style := EdgeStyle{
		ArrowType: ArrowNormal,
	}

	// Dashed arrow: -.-> or -.-
	if strings.Contains(arrow, ".") {
		style.Dashed = true
	}

	// Thick arrow: ==> or ===
	if strings.Contains(arrow, "=") {
		style.Thick = true
	}

	// Arrow type based on ending
	if strings.HasSuffix(arrow, "x") {
		style.ArrowType = ArrowCross
	} else if strings.HasSuffix(arrow, "o") {
		style.ArrowType = ArrowCircle
	} else if !strings.HasSuffix(arrow, ">") {
		style.ArrowType = ArrowNone
	}

	// Bidirectional
	if strings.HasPrefix(arrow, "<") {
		style.ArrowType = ArrowBidir
	}

	return style
}

func (p *Parser) parseSequence() error {
	for p.current().Type != TokenEOF {
		p.skipWhitespace()

		tok := p.current()
		if tok.Type == TokenEOF {
			break
		}

		if tok.Type != TokenIdent {
			p.skipToNextLine()
			continue
		}

		switch strings.ToLower(tok.Value) {
		case "participant", "actor":
			actor, err := p.parseActor(tok.Value)
			if err != nil {
				p.skipToNextLine()
				continue
			}
			p.doc.Actors = append(p.doc.Actors, actor)

		case "alt", "opt", "loop", "par", "critical", "break":
			group, err := p.parseMessageGroup(tok.Value)
			if err != nil {
				p.skipToNextLine()
				continue
			}
			p.doc.Groups = append(p.doc.Groups, group)

		case "note":
			// Skip notes for now
			p.skipToNextLine()

		case "activate", "deactivate":
			// Skip activation for now
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

func (p *Parser) parseActor(actorType string) (*Actor, error) {
	line := p.current().Line
	p.advance() // skip participant/actor

	actor := &Actor{
		Shape: actorType,
		Line:  line,
	}

	// Parse actor ID
	if p.current().Type != TokenIdent {
		return nil, fmt.Errorf("expected actor ID")
	}
	actor.ID = p.current().Value
	actor.Label = actor.ID
	p.advance()

	// Check for 'as' alias
	if p.current().Type == TokenIdent && strings.ToLower(p.current().Value) == "as" {
		p.advance()
		// Read label
		if p.current().Type == TokenIdent {
			actor.Label = p.current().Value
			p.advance()
		} else if p.current().Type == TokenString {
			actor.Label = p.current().Value
			p.advance()
		}
	}

	p.skipToNextLine()
	return actor, nil
}

func (p *Parser) tryParseMessage() *Message {
	if p.current().Type != TokenIdent {
		return nil
	}

	// Save position for backtracking
	startPos := p.pos

	line := p.current().Line
	from := p.current().Value
	p.advance()

	// Expect arrow
	if p.current().Type != TokenArrow {
		// Restore position
		p.pos = startPos
		return nil
	}

	arrowTok := p.advance()
	style := MessageSolid
	if strings.Contains(arrowTok.Value, "--") {
		style = MessageDashed
	}
	if strings.HasSuffix(arrowTok.Value, ">>") {
		style = MessageAsync
	}

	// Parse target
	if p.current().Type != TokenIdent {
		// Restore position
		p.pos = startPos
		return nil
	}
	to := p.current().Value
	p.advance()

	// Check for label after colon
	label := ""
	if p.current().Type == TokenColon {
		p.advance()
		// Read rest of line as label
		var labelBuilder strings.Builder
		for p.current().Type != TokenNewline && p.current().Type != TokenEOF {
			labelBuilder.WriteString(p.current().Value)
			labelBuilder.WriteString(" ")
			p.advance()
		}
		label = strings.TrimSpace(labelBuilder.String())
	}

	p.skipToNextLine()

	return &Message{
		From:  from,
		To:    to,
		Label: label,
		Style: style,
		Line:  line,
	}
}

func (p *Parser) parseMessageGroup(groupType string) (*MessageGroup, error) {
	line := p.current().Line
	p.advance() // skip group keyword

	group := &MessageGroup{
		Type: GroupType(strings.ToLower(groupType)),
		Line: line,
	}

	// Parse label (rest of line)
	label := ""
	for p.current().Type != TokenNewline && p.current().Type != TokenEOF {
		label += p.current().Value + " "
		p.advance()
	}
	group.Label = strings.TrimSpace(label)
	p.skipToNextLine()

	// Parse messages until 'end' or 'else'
	for p.current().Type != TokenEOF {
		p.skipWhitespace()

		tok := p.current()
		if tok.Type == TokenIdent {
			switch strings.ToLower(tok.Value) {
			case "end":
				p.advance()
				p.skipToNextLine()
				return group, nil
			case "else":
				p.advance()
				p.skipToNextLine()
				// Parse else messages
				for p.current().Type != TokenEOF {
					p.skipWhitespace()
					if p.current().Type == TokenIdent && strings.ToLower(p.current().Value) == "end" {
						p.advance()
						p.skipToNextLine()
						return group, nil
					}
					if msg := p.tryParseMessage(); msg != nil {
						group.Else = append(group.Else, msg)
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

func (p *Parser) parseClass() error {
	for p.current().Type != TokenEOF {
		p.skipWhitespace()

		tok := p.current()
		if tok.Type == TokenEOF {
			break
		}

		if tok.Type != TokenIdent {
			p.skipToNextLine()
			continue
		}

		switch strings.ToLower(tok.Value) {
		case "class":
			class, err := p.parseClassDefinition()
			if err != nil {
				p.skipToNextLine()
				continue
			}
			p.doc.Classes = append(p.doc.Classes, class)

		default:
			// Try to parse as relationship
			p.parseClassRelationship()
		}
	}

	return nil
}

func (p *Parser) parseClassDefinition() (*Class, error) {
	line := p.current().Line
	p.advance() // skip 'class'

	class := &Class{
		Line: line,
	}

	// Parse class name
	if p.current().Type != TokenIdent {
		return nil, fmt.Errorf("expected class name")
	}
	class.ID = p.current().Value
	class.Label = class.ID
	p.advance()

	// Check for class body
	if p.current().Type == TokenOpenBrace {
		content := p.current().Value
		p.advance()

		// Parse attributes and methods
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if strings.Contains(line, "(") {
				class.Methods = append(class.Methods, line)
			} else {
				class.Attributes = append(class.Attributes, line)
			}
		}
	}

	p.skipToNextLine()
	return class, nil
}

func (p *Parser) parseClassRelationship() {
	line := p.current().Line
	fromID := p.current().Value
	p.advance()

	// Check for arrow
	if p.current().Type != TokenArrow {
		p.skipToNextLine()
		return
	}

	arrowTok := p.advance()

	// Parse target
	if p.current().Type != TokenIdent {
		p.skipToNextLine()
		return
	}
	toID := p.current().Value
	p.advance()

	// Check for label
	label := ""
	if p.current().Type == TokenColon {
		p.advance()
		for p.current().Type != TokenNewline && p.current().Type != TokenEOF {
			label += p.current().Value + " "
			p.advance()
		}
		label = strings.TrimSpace(label)
	}

	edge := &Edge{
		From:  fromID,
		To:    toID,
		Label: label,
		Style: parseArrowStyle(arrowTok.Value),
		Line:  line,
	}

	p.doc.Edges = append(p.doc.Edges, edge)
	p.skipToNextLine()
}

// sanitizeID converts a label into a valid ID.
func sanitizeID(label string) string {
	// Replace spaces and special characters with underscores
	re := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	return re.ReplaceAllString(label, "_")
}
