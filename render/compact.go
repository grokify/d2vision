package render

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"

	"oss.terrastruct.com/d2/d2compiler"
	"oss.terrastruct.com/d2/d2exporter"
	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/d2target"
	"oss.terrastruct.com/d2/lib/log"
)

// LegendEntry maps a short numeric key back to the original node label it
// replaced during compaction.
type LegendEntry struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

// longLabelThreshold is the character length above which a leaf-node label is
// considered "long" and eligible for compaction. Labels containing a space are
// also eligible regardless of length.
const longLabelThreshold = 14

// RenderCompact compiles D2 code, replaces long leaf-node labels with short
// sequential numeric keys ("1", "2", "3", …), then lays out and renders the
// diagram. Because labels are shortened BEFORE layout, shapes are sized for the
// keys and the diagram renders much narrower. The returned legend maps each key
// to the original label text so the caller can render a legend above the image.
//
// Only leaf shapes (non-containers) with a "long" visible label are compacted.
// Container/boundary labels and edge labels are left untouched.
//
// Only SVG output is supported; PNG returns an error.
func (r *Renderer) RenderCompact(ctx context.Context, d2Code string, format Format, opts *Options) ([]byte, []LegendEntry, error) {
	if opts == nil {
		opts = DefaultOptions()
	}
	if format == FormatPNG {
		return nil, nil, fmt.Errorf("compact PNG unsupported; render SVG")
	}
	if format != FormatSVG {
		return nil, nil, fmt.Errorf("compact rendering supports SVG only, got: %s", format)
	}

	// Silence d2's debug output.
	ctx = log.With(ctx, slog.New(slog.NewTextHandler(io.Discard, nil)))

	// Apply the global font-size floor before compiling, if requested (same as
	// the normal Render path).
	if opts.FontSize > 0 {
		d2Code = ScaleFontSize(d2Code, opts.FontSize)
	}

	// Build render options.
	renderOpts := &d2svg.RenderOpts{
		Pad:     int64Ptr(opts.Pad),
		Sketch:  boolPtr(opts.Sketch),
		Center:  boolPtr(opts.Center),
		ThemeID: int64Ptr(opts.ThemeID),
	}
	if opts.Scale != 0 && opts.Scale != 1.0 {
		renderOpts.Scale = float64Ptr(opts.Scale)
	}

	// 1. Compile (parse + build graph) WITHOUT layout, so we can rewrite labels
	//    before dimensions are measured.
	g, _, err := d2compiler.Compile("", strings.NewReader(d2Code), &d2compiler.CompileOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("compiling D2: %w", err)
	}

	// 2. Rewrite long leaf-node labels to numeric keys, collecting the legend.
	legend := compactLabels(g)

	// 3. Mirror d2lib's private compile(): ApplyTheme -> SetDimensions ->
	//    layout -> Export. This repo's diagrams have no layers/scenarios/steps,
	//    so we only handle the root graph (see limitation note below).
	if err := g.ApplyTheme(*renderOpts.ThemeID); err != nil {
		return nil, nil, fmt.Errorf("applying theme: %w", err)
	}

	if len(g.Objects) > 0 {
		if err := g.SetDimensions(nil, r.ruler, nil, nil); err != nil {
			return nil, nil, fmt.Errorf("setting dimensions: %w", err)
		}

		coreLayout := func(ctx context.Context, g *d2graph.Graph) error {
			return d2dagrelayout.DefaultLayout(ctx, g)
		}
		graphInfo := d2layouts.NestedGraphInfo(g.Root)
		if err := d2layouts.LayoutNested(ctx, g, graphInfo, coreLayout, d2layouts.DefaultRouter); err != nil {
			return nil, nil, fmt.Errorf("laying out graph: %w", err)
		}
	}

	// NOTE: Layers, scenarios, and steps (g.Layers/g.Scenarios/g.Steps) are not
	// recursed here — the diagrams this is built for are single-layer. If those
	// are ever needed, mirror the recursion in d2lib.compile().

	diagram, err := d2exporter.Export(ctx, g, nil, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("exporting diagram: %w", err)
	}
	diagram.Config = &d2target.Config{
		ThemeID: renderOpts.ThemeID,
		Sketch:  renderOpts.Sketch,
	}

	// 4. Render to SVG.
	svg, err := d2svg.Render(diagram, renderOpts)
	if err != nil {
		return nil, nil, fmt.Errorf("rendering SVG: %w", err)
	}

	return svg, legend, nil
}

// compactLabels walks the graph's objects and replaces the visible label of
// every long leaf shape with the next sequential numeric key, returning the
// legend that maps keys back to the original label text. It mutates the graph
// in place. Container/boundary shapes are skipped so their labels remain
// human-readable.
func compactLabels(g *d2graph.Graph) []LegendEntry {
	var legend []LegendEntry
	next := 0
	for _, obj := range g.Objects {
		if obj.IsContainer() {
			continue
		}
		label := obj.Label.Value
		if !isLongLabel(label) {
			continue
		}
		next++
		key := strconv.Itoa(next)
		legend = append(legend, LegendEntry{Key: key, Label: label})
		obj.Label.Value = key
		// Clear any pre-measured dimensions so SetDimensions re-measures for the
		// short key rather than reusing the wide original.
		obj.LabelDimensions = d2target.TextDimensions{}
	}
	return legend
}

// isLongLabel reports whether a label should be compacted: longer than the
// threshold, or containing whitespace (multi-word labels widen shapes).
func isLongLabel(label string) bool {
	if label == "" {
		return false
	}
	if len(label) > longLabelThreshold {
		return true
	}
	return strings.ContainsRune(label, ' ')
}
