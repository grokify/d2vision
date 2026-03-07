package d2vision

import (
	"fmt"
	"strconv"
	"strings"
)

// Point represents a 2D coordinate.
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// String returns a string representation of the point.
func (p Point) String() string {
	return fmt.Sprintf("(%.2f, %.2f)", p.X, p.Y)
}

// Bounds represents a rectangular bounding box.
type Bounds struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// Center returns the center point of the bounds.
func (b Bounds) Center() Point {
	return Point{
		X: b.X + b.Width/2,
		Y: b.Y + b.Height/2,
	}
}

// Contains returns true if the point is within the bounds.
func (b Bounds) Contains(p Point) bool {
	return p.X >= b.X && p.X <= b.X+b.Width &&
		p.Y >= b.Y && p.Y <= b.Y+b.Height
}

// String returns a string representation of the bounds.
func (b Bounds) String() string {
	return fmt.Sprintf("(%.2f, %.2f, %.2f, %.2f)", b.X, b.Y, b.Width, b.Height)
}

// ParseViewBox parses an SVG viewBox attribute string.
func ParseViewBox(s string) (Bounds, error) {
	parts := strings.Fields(s)
	if len(parts) != 4 {
		return Bounds{}, fmt.Errorf("invalid viewBox: expected 4 values, got %d", len(parts))
	}

	var vals [4]float64
	for i, part := range parts {
		v, err := strconv.ParseFloat(part, 64)
		if err != nil {
			return Bounds{}, fmt.Errorf("invalid viewBox value %q: %w", part, err)
		}
		vals[i] = v
	}

	return Bounds{
		X:      vals[0],
		Y:      vals[1],
		Width:  vals[2],
		Height: vals[3],
	}, nil
}
