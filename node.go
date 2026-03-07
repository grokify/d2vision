package d2vision

// NodeStyle contains visual styling properties for a node.
type NodeStyle struct {
	Fill        string  `json:"fill,omitempty"`
	Stroke      string  `json:"stroke,omitempty"`
	StrokeWidth float64 `json:"strokeWidth,omitempty"`
	FontSize    float64 `json:"fontSize,omitempty"`
	FontColor   string  `json:"fontColor,omitempty"`
	Opacity     float64 `json:"opacity,omitempty"`
}

// Node represents a node (shape) in a D2 diagram.
type Node struct {
	ID       string    `json:"id"`
	Label    string    `json:"label,omitempty"`
	Shape    ShapeType `json:"shape"`
	Bounds   Bounds    `json:"bounds"`
	Parent   string    `json:"parent,omitempty"`
	Children []string  `json:"children,omitempty"`
	Style    NodeStyle `json:"style,omitzero"`
}

// HasChildren returns true if the node has child nodes.
func (n Node) HasChildren() bool {
	return len(n.Children) > 0
}

// IsContainer returns true if the node is a container (has children).
func (n Node) IsContainer() bool {
	return n.HasChildren()
}

// DisplayLabel returns the label for display purposes.
// If no label is set, it returns the ID.
func (n Node) DisplayLabel() string {
	if n.Label != "" {
		return n.Label
	}
	return n.ID
}
