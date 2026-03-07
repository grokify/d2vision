// Package render provides D2 diagram rendering using the d2 library.
package render

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"

	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/d2target"
	"oss.terrastruct.com/d2/d2themes/d2themescatalog"
	"oss.terrastruct.com/d2/lib/log"
	"oss.terrastruct.com/d2/lib/textmeasure"
)

// Format represents the output format for rendering.
type Format string

const (
	FormatSVG Format = "svg"
	FormatPNG Format = "png"
	FormatPDF Format = "pdf"
)

// ParseFormat parses a format string.
func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(s) {
	case "svg":
		return FormatSVG, nil
	case "png":
		return FormatPNG, nil
	case "pdf":
		return FormatPDF, nil
	default:
		return "", fmt.Errorf("unsupported format: %s (supported: svg, png, pdf)", s)
	}
}

// FormatFromPath infers the format from a file path extension.
func FormatFromPath(path string) (Format, error) {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	return ParseFormat(ext)
}

// Options configures rendering behavior.
type Options struct {
	// ThemeID is the D2 theme to use (0 = default).
	ThemeID int64

	// Pad is the padding around the diagram in pixels.
	Pad int64

	// Sketch enables sketch/hand-drawn mode.
	Sketch bool

	// Center centers the diagram in the output.
	Center bool

	// Scale is the output scale factor (default 1.0).
	Scale float64
}

// DefaultOptions returns default rendering options.
func DefaultOptions() *Options {
	return &Options{
		ThemeID: 0,
		Pad:     d2svg.DEFAULT_PADDING,
		Sketch:  false,
		Center:  false,
		Scale:   1.0,
	}
}

// Renderer renders D2 code to various output formats.
type Renderer struct {
	ruler *textmeasure.Ruler
}

// New creates a new Renderer.
func New() (*Renderer, error) {
	ruler, err := textmeasure.NewRuler()
	if err != nil {
		return nil, fmt.Errorf("creating text ruler: %w", err)
	}
	return &Renderer{ruler: ruler}, nil
}

// Render compiles and renders D2 code to the specified format.
func (r *Renderer) Render(ctx context.Context, d2Code string, format Format, opts *Options) ([]byte, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	// Add a silent logger to suppress d2's debug output
	ctx = log.With(ctx, slog.New(slog.NewTextHandler(io.Discard, nil)))

	// Build render options
	renderOpts := &d2svg.RenderOpts{
		Pad:     int64Ptr(opts.Pad),
		Sketch:  boolPtr(opts.Sketch),
		Center:  boolPtr(opts.Center),
		ThemeID: int64Ptr(opts.ThemeID),
	}
	if opts.Scale != 0 && opts.Scale != 1.0 {
		renderOpts.Scale = float64Ptr(opts.Scale)
	}

	// Compile D2 code to diagram
	diagram, _, err := d2lib.Compile(ctx, d2Code, &d2lib.CompileOptions{
		Ruler: r.ruler,
		LayoutResolver: func(engine string) (d2graph.LayoutGraph, error) {
			return d2dagrelayout.DefaultLayout, nil
		},
	}, renderOpts)
	if err != nil {
		return nil, fmt.Errorf("compiling D2: %w", err)
	}

	// Render to SVG
	svg, err := d2svg.Render(diagram, renderOpts)
	if err != nil {
		return nil, fmt.Errorf("rendering SVG: %w", err)
	}

	switch format {
	case FormatSVG:
		return svg, nil
	case FormatPNG:
		return r.svgToPNG(ctx, svg, opts.Scale)
	case FormatPDF:
		return r.svgToPDF(ctx, svg, diagram)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// RenderSVG is a convenience method for rendering to SVG.
func (r *Renderer) RenderSVG(ctx context.Context, d2Code string, opts *Options) ([]byte, error) {
	return r.Render(ctx, d2Code, FormatSVG, opts)
}

// RenderPNG is a convenience method for rendering to PNG.
func (r *Renderer) RenderPNG(ctx context.Context, d2Code string, opts *Options) ([]byte, error) {
	return r.Render(ctx, d2Code, FormatPNG, opts)
}

// RenderPDF is a convenience method for rendering to PDF.
func (r *Renderer) RenderPDF(ctx context.Context, d2Code string, opts *Options) ([]byte, error) {
	return r.Render(ctx, d2Code, FormatPDF, opts)
}

// svgToPNG converts SVG to PNG using the d2 library's PNG renderer.
func (r *Renderer) svgToPNG(ctx context.Context, svg []byte, scale float64) ([]byte, error) {
	// PNG rendering requires playwright/chromium
	// For now, return an error with instructions
	return nil, fmt.Errorf("PNG rendering requires playwright; run: npx playwright install chromium")
}

// svgToPDF converts SVG to PDF using gofpdf.
func (r *Renderer) svgToPDF(ctx context.Context, svg []byte, diagram *d2target.Diagram) ([]byte, error) {
	// PDF rendering is complex and requires additional setup
	return nil, fmt.Errorf("PDF rendering not yet implemented; use SVG output and convert with external tools")
}

// Helper functions for pointer conversion
func int64Ptr(i int64) *int64       { return &i }
func boolPtr(b bool) *bool          { return &b }
func float64Ptr(f float64) *float64 { return &f }

// ThemeNames returns a list of available theme names.
func ThemeNames() []string {
	var names []string
	for _, t := range d2themescatalog.LightCatalog {
		names = append(names, t.Name)
	}
	for _, t := range d2themescatalog.DarkCatalog {
		names = append(names, t.Name)
	}
	return names
}

// ThemeID returns the theme ID for a theme name.
func ThemeID(name string) (int64, error) {
	name = strings.ToLower(name)
	for _, t := range d2themescatalog.LightCatalog {
		if strings.ToLower(t.Name) == name {
			return t.ID, nil
		}
	}
	for _, t := range d2themescatalog.DarkCatalog {
		if strings.ToLower(t.Name) == name {
			return t.ID, nil
		}
	}
	return 0, fmt.Errorf("unknown theme: %s", name)
}

// Compile parses D2 code and returns the compiled diagram (for validation/inspection).
func Compile(ctx context.Context, d2Code string) (*d2target.Diagram, error) {
	ruler, err := textmeasure.NewRuler()
	if err != nil {
		return nil, fmt.Errorf("creating text ruler: %w", err)
	}

	// Add a silent logger to suppress d2's debug output
	ctx = log.With(ctx, slog.New(slog.NewTextHandler(io.Discard, nil)))

	diagram, _, err := d2lib.Compile(ctx, d2Code, &d2lib.CompileOptions{
		Ruler: ruler,
		LayoutResolver: func(engine string) (d2graph.LayoutGraph, error) {
			return d2dagrelayout.DefaultLayout, nil
		},
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("compiling D2: %w", err)
	}

	return diagram, nil
}

// Validate checks if D2 code is valid without rendering.
func Validate(ctx context.Context, d2Code string) error {
	_, err := Compile(ctx, d2Code)
	return err
}

// MustRender renders D2 code and panics on error. Useful for tests.
func MustRender(d2Code string, format Format) []byte {
	r, err := New()
	if err != nil {
		panic(err)
	}
	data, err := r.Render(context.Background(), d2Code, format, nil)
	if err != nil {
		panic(err)
	}
	return data
}

// Quick renders D2 code to SVG with default options. Returns error on failure.
func Quick(d2Code string) ([]byte, error) {
	r, err := New()
	if err != nil {
		return nil, err
	}
	return r.RenderSVG(context.Background(), d2Code, nil)
}

// QuickToString renders D2 code to SVG string with default options.
func QuickToString(d2Code string) (string, error) {
	data, err := Quick(d2Code)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SVGDimensions extracts width and height from rendered SVG.
func SVGDimensions(svg []byte) (width, height int, err error) {
	// Simple regex-free extraction
	s := string(svg)

	// Look for viewBox
	vbStart := strings.Index(s, `viewBox="`)
	if vbStart == -1 {
		return 0, 0, fmt.Errorf("no viewBox found")
	}
	vbStart += len(`viewBox="`)
	vbEnd := strings.Index(s[vbStart:], `"`)
	if vbEnd == -1 {
		return 0, 0, fmt.Errorf("malformed viewBox")
	}
	viewBox := s[vbStart : vbStart+vbEnd]

	// Parse "x y width height"
	var x, y float64
	var w, h float64
	_, err = fmt.Sscanf(viewBox, "%f %f %f %f", &x, &y, &w, &h)
	if err != nil {
		return 0, 0, fmt.Errorf("parsing viewBox: %w", err)
	}

	return int(w), int(h), nil
}

// Buffer is a convenience wrapper for rendering to a bytes.Buffer.
type Buffer struct {
	bytes.Buffer
}
