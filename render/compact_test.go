package render

import (
	"context"
	"sort"
	"testing"
)

// compactTestD2 has three long-labeled leaf nodes, one short-labeled node, and
// an edge with a label. The long labels should be compacted; the short node
// label and the edge label should not appear in the legend.
const compactTestD2 = `
api: API Gateway Frontend Service
db: Primary Postgres Database Cluster
cache: Distributed Redis Cache Layer
q: Queue

api -> db: writes records
api -> cache
api -> q
`

func TestRenderCompactLegend(t *testing.T) {
	r, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()

	svg, legend, err := r.RenderCompact(ctx, compactTestD2, FormatSVG, nil)
	if err != nil {
		t.Fatalf("RenderCompact: %v", err)
	}
	if len(svg) == 0 {
		t.Fatal("RenderCompact returned empty SVG")
	}

	// (a) Exactly 3 legend entries whose labels equal the originals.
	if len(legend) != 3 {
		t.Fatalf("expected 3 legend entries, got %d: %+v", len(legend), legend)
	}
	gotLabels := make([]string, len(legend))
	for i, e := range legend {
		gotLabels[i] = e.Label
		if e.Key == "" {
			t.Errorf("legend entry %d has empty key", i)
		}
	}
	sort.Strings(gotLabels)
	wantLabels := []string{
		"API Gateway Frontend Service",
		"Distributed Redis Cache Layer",
		"Primary Postgres Database Cluster",
	}
	for i, want := range wantLabels {
		if gotLabels[i] != want {
			t.Errorf("legend label[%d] = %q, want %q", i, gotLabels[i], want)
		}
	}

	// (c) The short node label and the edge label are NOT in the legend.
	for _, e := range legend {
		if e.Label == "Queue" {
			t.Errorf("short node label %q should not be compacted", e.Label)
		}
		if e.Label == "writes records" {
			t.Errorf("edge label %q should not be compacted", e.Label)
		}
	}

	// (b) The compacted SVG is narrower than the normal render of the same code.
	normalSVG, err := r.RenderSVG(ctx, compactTestD2, nil)
	if err != nil {
		t.Fatalf("RenderSVG: %v", err)
	}
	compactW, _, err := SVGDimensions(svg)
	if err != nil {
		t.Fatalf("SVGDimensions(compact): %v", err)
	}
	normalW, _, err := SVGDimensions(normalSVG)
	if err != nil {
		t.Fatalf("SVGDimensions(normal): %v", err)
	}
	if compactW >= normalW {
		t.Errorf("compact width %d not less than normal width %d", compactW, normalW)
	}
}

func TestRenderCompactPNGUnsupported(t *testing.T) {
	r, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, _, err = r.RenderCompact(context.Background(), compactTestD2, FormatPNG, nil)
	if err == nil {
		t.Fatal("expected error for PNG format, got nil")
	}
}
