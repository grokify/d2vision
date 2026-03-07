package d2vision

import (
	"encoding/json"
)

// Diagram represents a parsed D2 diagram.
type Diagram struct {
	Version string `json:"version,omitempty"`
	Title   string `json:"title,omitempty"`
	ViewBox Bounds `json:"viewBox"`
	Nodes   []Node `json:"nodes"`
	Edges   []Edge `json:"edges"`
}

// NodeByID returns the node with the given ID, or nil if not found.
func (d *Diagram) NodeByID(id string) *Node {
	for i := range d.Nodes {
		if d.Nodes[i].ID == id {
			return &d.Nodes[i]
		}
	}
	return nil
}

// EdgeByID returns the edge with the given ID, or nil if not found.
func (d *Diagram) EdgeByID(id string) *Edge {
	for i := range d.Edges {
		if d.Edges[i].ID == id {
			return &d.Edges[i]
		}
	}
	return nil
}

// RootNodes returns nodes that have no parent.
func (d *Diagram) RootNodes() []Node {
	var roots []Node
	for _, n := range d.Nodes {
		if n.Parent == "" {
			roots = append(roots, n)
		}
	}
	return roots
}

// ContainerNodes returns nodes that have children.
func (d *Diagram) ContainerNodes() []Node {
	var containers []Node
	for _, n := range d.Nodes {
		if n.HasChildren() {
			containers = append(containers, n)
		}
	}
	return containers
}

// LeafNodes returns nodes that have no children.
func (d *Diagram) LeafNodes() []Node {
	var leaves []Node
	for _, n := range d.Nodes {
		if !n.HasChildren() {
			leaves = append(leaves, n)
		}
	}
	return leaves
}

// EdgesFrom returns all edges originating from the given node ID.
func (d *Diagram) EdgesFrom(nodeID string) []Edge {
	var edges []Edge
	for _, e := range d.Edges {
		if e.Source == nodeID {
			edges = append(edges, e)
		}
	}
	return edges
}

// EdgesTo returns all edges terminating at the given node ID.
func (d *Diagram) EdgesTo(nodeID string) []Edge {
	var edges []Edge
	for _, e := range d.Edges {
		if e.Target == nodeID {
			edges = append(edges, e)
		}
	}
	return edges
}

// JSON returns the diagram as JSON bytes.
func (d *Diagram) JSON() ([]byte, error) {
	return json.Marshal(d)
}

// JSONIndent returns the diagram as indented JSON bytes.
func (d *Diagram) JSONIndent(prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(d, prefix, indent)
}
