package d2vision

// EdgeStyle contains visual styling properties for an edge.
type EdgeStyle struct {
	Stroke       string  `json:"stroke,omitempty"`
	StrokeWidth  float64 `json:"strokeWidth,omitempty"`
	StrokeDash   string  `json:"strokeDash,omitempty"`
	Opacity      float64 `json:"opacity,omitempty"`
	Animated     bool    `json:"animated,omitempty"`
	LabelFontSize float64 `json:"labelFontSize,omitempty"`
}

// Edge represents a connection between two nodes in a D2 diagram.
type Edge struct {
	ID          string    `json:"id"`
	Source      string    `json:"source"`
	Target      string    `json:"target"`
	Label       string    `json:"label,omitempty"`
	SourceArrow ArrowType `json:"sourceArrow,omitempty"`
	TargetArrow ArrowType `json:"targetArrow"`
	Path        []Point   `json:"path,omitempty"`
	Style       EdgeStyle `json:"style,omitzero"`
}

// IsBidirectional returns true if the edge has arrows on both ends.
func (e Edge) IsBidirectional() bool {
	return e.SourceArrow != "" && e.SourceArrow != ArrowNone &&
		e.TargetArrow != "" && e.TargetArrow != ArrowNone
}

// DisplayLabel returns the label for display purposes.
func (e Edge) DisplayLabel() string {
	return e.Label
}

// ConnectionDescription returns a natural language description of the connection.
func (e Edge) ConnectionDescription() string {
	if e.IsBidirectional() {
		return "connects to"
	}
	return "connects to"
}
