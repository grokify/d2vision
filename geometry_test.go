package d2vision

import (
	"testing"
)

func TestPointString(t *testing.T) {
	p := Point{X: 10.5, Y: 20.25}
	expected := "(10.50, 20.25)"
	if got := p.String(); got != expected {
		t.Errorf("Point.String() = %q, want %q", got, expected)
	}
}

func TestBoundsCenter(t *testing.T) {
	b := Bounds{X: 0, Y: 0, Width: 100, Height: 50}
	center := b.Center()
	if center.X != 50 || center.Y != 25 {
		t.Errorf("Bounds.Center() = %v, want (50, 25)", center)
	}
}

func TestBoundsContains(t *testing.T) {
	b := Bounds{X: 10, Y: 10, Width: 100, Height: 100}

	tests := []struct {
		p        Point
		contains bool
	}{
		{Point{X: 50, Y: 50}, true},   // center
		{Point{X: 10, Y: 10}, true},   // top-left corner
		{Point{X: 110, Y: 110}, true}, // bottom-right corner
		{Point{X: 0, Y: 0}, false},    // outside top-left
		{Point{X: 120, Y: 50}, false}, // outside right
	}

	for _, tt := range tests {
		got := b.Contains(tt.p)
		if got != tt.contains {
			t.Errorf("Bounds.Contains(%v) = %v, want %v", tt.p, got, tt.contains)
		}
	}
}

func TestBoundsString(t *testing.T) {
	b := Bounds{X: 10, Y: 20, Width: 100, Height: 50}
	expected := "(10.00, 20.00, 100.00, 50.00)"
	if got := b.String(); got != expected {
		t.Errorf("Bounds.String() = %q, want %q", got, expected)
	}
}

func TestParseViewBox(t *testing.T) {
	tests := []struct {
		input   string
		want    Bounds
		wantErr bool
	}{
		{
			input:   "0 0 100 200",
			want:    Bounds{X: 0, Y: 0, Width: 100, Height: 200},
			wantErr: false,
		},
		{
			input:   "10 20 300 400",
			want:    Bounds{X: 10, Y: 20, Width: 300, Height: 400},
			wantErr: false,
		},
		{
			input:   "-10 -20 100 100",
			want:    Bounds{X: -10, Y: -20, Width: 100, Height: 100},
			wantErr: false,
		},
		{
			input:   "0 0 100",
			want:    Bounds{},
			wantErr: true, // not enough values
		},
		{
			input:   "0 0 100 abc",
			want:    Bounds{},
			wantErr: true, // invalid number
		},
	}

	for _, tt := range tests {
		got, err := ParseViewBox(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseViewBox(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && got != tt.want {
			t.Errorf("ParseViewBox(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
