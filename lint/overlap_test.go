package lint

import "testing"

func TestBoxOverlaps(t *testing.T) {
	base := box{0, 0, 100, 20}
	tests := []struct {
		name string
		b    box
		want bool
	}{
		{"clear overlap", box{50, 5, 150, 25}, true},
		{"disjoint in x", box{200, 0, 300, 20}, false},
		{"disjoint in y", box{0, 100, 100, 120}, false},
		{"1px touch is not overlap", box{100, 0, 200, 20}, false},
		{"tiny sliver under threshold", box{99, 0, 200, 20}, false},
		{"contained", box{10, 5, 40, 15}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := base.overlaps(tt.b); got != tt.want {
				t.Errorf("overlaps = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckTextOverlap_CleanDiagram(t *testing.T) {
	// A trivial single-edge diagram lays out without label overlaps.
	findings, err := CheckTextOverlap("a -> b: hello", "dagre")
	if err != nil {
		t.Fatalf("CheckTextOverlap: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("got %d findings on a clean diagram, want 0: %+v", len(findings), findings)
	}
}

func TestCheckTextOverlap_LayoutSelection(t *testing.T) {
	if _, err := CheckTextOverlap("a -> b: hi", "elk"); err != nil {
		t.Errorf("elk layout: unexpected error: %v", err)
	}
	if _, err := CheckTextOverlap("a -> b: hi", ""); err != nil {
		t.Errorf("default layout: unexpected error: %v", err)
	}
	if _, err := CheckTextOverlap("a -> b: hi", "bogus"); err == nil {
		t.Errorf("bogus layout: expected an error")
	}
}
