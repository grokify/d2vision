// Package legibility checks whether a rendered D2 diagram can be read where it
// will be shown: on a Word page, a slide, or a screen.
//
// A diagram is scaled to fit the area it is shown in, keeping its shape, so a
// wide diagram on a narrow page shrinks and its labels with it. Analyze works
// out the size each label is shown at and compares it with the smallest size a
// reader can read comfortably on that target. When labels are too small, the
// report says by how much and which change would help: wrapping labels into
// narrower boxes, fewer columns or rows, a larger font, or a larger target.
//
// The package reads the SVG that D2 writes. It does not link the D2 engine.
package legibility

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/grokify/d2vision"
)

// pointsPerInch converts inches to typographic points.
const pointsPerInch = 72

// Target is an area a diagram is shown in, and the smallest label size that
// reads comfortably there.
type Target struct {
	// Name identifies the target, such as "docx-portrait".
	Name string `json:"name"`
	// Description says what the area is.
	Description string `json:"description"`
	// WidthIn and HeightIn are the area the diagram may fill, in inches.
	WidthIn  float64 `json:"widthIn"`
	HeightIn float64 `json:"heightIn"`
	// MinFontPt is the smallest label size, in points, that reads comfortably.
	MinFontPt float64 `json:"minFontPt"`
}

// Targets are the built-in targets. Pages are US Letter or Tabloid with 1-inch
// margins. Print text reads comfortably from 9 points; text on a projected
// slide from 18 points; text in a browser from 12 CSS pixels, which is 9
// points.
var Targets = []Target{
	{Name: "docx-portrait", Description: "Word, US Letter portrait, 1-inch margins", WidthIn: 6.5, HeightIn: 9, MinFontPt: 9},
	{Name: "docx-landscape", Description: "Word, US Letter landscape, 1-inch margins", WidthIn: 9, HeightIn: 6.5, MinFontPt: 9},
	{Name: "tabloid-landscape", Description: "Tabloid (11 x 17 inches) landscape, 1-inch margins", WidthIn: 15, HeightIn: 9, MinFontPt: 9},
	{Name: "slide-16x9", Description: "16:9 slide, 13.33 x 7.5 inches, below a title", WidthIn: 12.33, HeightIn: 5.9, MinFontPt: 18},
	{Name: "screen", Description: "Browser window 1280 x 720 CSS pixels, below a title", WidthIn: 12.5, HeightIn: 6.5, MinFontPt: 9},
}

// TargetByName returns the built-in target with the given name.
func TargetByName(name string) (Target, bool) {
	for _, t := range Targets {
		if t.Name == name {
			return t, true
		}
	}
	return Target{}, false
}

// TargetNames returns the names of the built-in targets, in order.
func TargetNames() []string {
	names := make([]string, len(Targets))
	for i, t := range Targets {
		names[i] = t.Name
	}
	return names
}

// Label is one label as shown on a target.
type Label struct {
	// ID is the element the label belongs to.
	ID string `json:"id"`
	// Text is the label, its lines joined by spaces.
	Text string `json:"text"`
	// FontPx is the font size in the SVG, in pixels.
	FontPx float64 `json:"fontPx"`
	// ShownPt is the size the label is shown at, in points.
	ShownPt float64 `json:"shownPt"`
	// LongestLine is the number of characters in the label's longest line.
	LongestLine int `json:"longestLine"`
}

// Report is the legibility of one diagram on one target.
type Report struct {
	Target Target `json:"target"`
	// Width and Height are the diagram's size in SVG pixels.
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	// Scale is the points shown per SVG pixel once the diagram fits the target.
	Scale float64 `json:"scale"`
	// LimitedBy is "width" when the target's width decides the scale, and
	// "height" when its height does.
	LimitedBy string `json:"limitedBy"`
	// Labels is the number of labels with a known font size.
	Labels int `json:"labels"`
	// SmallestPt and MedianPt are label sizes as shown, in points.
	SmallestPt float64 `json:"smallestPt"`
	MedianPt   float64 `json:"medianPt"`
	// TooSmall lists the labels shown below the target's minimum, smallest first.
	TooSmall []Label `json:"tooSmall,omitempty"`
	// Readable reports whether every label is shown at the minimum or larger.
	Readable bool `json:"readable"`
	// MaxWidth and MaxHeight are the largest the diagram can be, in SVG
	// pixels, with its smallest label still readable on the target.
	MaxWidth  float64 `json:"maxWidth"`
	MaxHeight float64 `json:"maxHeight"`
	// Shrink is how many times too large the diagram is for its smallest label
	// to be read: 1 or less when it is readable.
	Shrink float64 `json:"shrink"`
	// LongestLine is the most characters on one line of any label.
	LongestLine int `json:"longestLine"`
	// Advice says what would make the diagram readable, most effective first.
	Advice []string `json:"advice,omitempty"`
}

// Line lengths, in characters, that suit a label in a box. Longer lines make
// wide boxes, and a wide diagram shrinks on a page.
const (
	// LongLine is the longest line Analyze leaves without advice to wrap.
	LongLine = 40
	// WrapWidth is the line length Analyze suggests wrapping to.
	WrapWidth = 28
)

// Analyze reports how readable a diagram is on a target. A label's size is
// read from the SVG; labels without one, such as edge labels, are not counted.
func Analyze(d *d2vision.Diagram, t Target) (*Report, error) {
	if t.WidthIn <= 0 || t.HeightIn <= 0 || t.MinFontPt <= 0 {
		return nil, fmt.Errorf("legibility: target %q needs a width, a height, and a minimum font size", t.Name)
	}
	w, h := d.ViewBox.Width, d.ViewBox.Height
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("legibility: the diagram has no size")
	}
	r := &Report{Target: t, Width: w, Height: h}
	areaW, areaH := t.WidthIn*pointsPerInch, t.HeightIn*pointsPerInch
	r.Scale, r.LimitedBy = areaW/w, "width"
	if s := areaH / h; s < r.Scale {
		r.Scale, r.LimitedBy = s, "height"
	}

	labels := labelsOf(d, r.Scale)
	r.Labels = len(labels)
	if len(labels) == 0 {
		r.Readable = true
		return r, nil
	}
	sort.SliceStable(labels, func(i, j int) bool { return labels[i].ShownPt < labels[j].ShownPt })
	r.SmallestPt = round2(labels[0].ShownPt)
	r.MedianPt = round2(labels[len(labels)/2].ShownPt)
	for _, l := range labels {
		if l.ShownPt < t.MinFontPt-0.005 {
			r.TooSmall = append(r.TooSmall, l)
		}
		r.LongestLine = max(r.LongestLine, l.LongestLine)
	}
	r.Readable = len(r.TooSmall) == 0

	// The smallest label reads at the minimum when the diagram is shown at
	// this scale, so the area it fills gives the largest readable size.
	need := t.MinFontPt / labels[0].FontPx
	r.MaxWidth, r.MaxHeight = math.Floor(areaW/need), math.Floor(areaH/need)
	r.Shrink = round2(need / r.Scale)
	r.Scale = round4(r.Scale)
	if !r.Readable {
		r.Advice = advise(r, labels[0])
	}
	return r, nil
}

// AnalyzeAll reports on every target given, in order.
func AnalyzeAll(d *d2vision.Diagram, targets []Target) ([]*Report, error) {
	var out []*Report
	for _, t := range targets {
		r, err := Analyze(d, t)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}

// labelsOf collects the labels with a known font size, shown at scale.
func labelsOf(d *d2vision.Diagram, scale float64) []Label {
	var out []Label
	for _, n := range d.Nodes {
		if n.Style.FontSize <= 0 || strings.TrimSpace(n.Label) == "" {
			continue
		}
		lines := n.LabelLines
		if len(lines) == 0 {
			lines = []string{n.Label}
		}
		longest := 0
		for _, l := range lines {
			longest = max(longest, len([]rune(l)))
		}
		out = append(out, Label{
			ID:          n.ID,
			Text:        n.Label,
			FontPx:      n.Style.FontSize,
			ShownPt:     round2(n.Style.FontSize * scale),
			LongestLine: longest,
		})
	}
	return out
}

// advise says what would make an unreadable diagram readable.
func advise(r *Report, smallest Label) []string {
	var a []string
	over := r.Width / r.MaxWidth
	if r.LimitedBy == "height" {
		over = r.Height / r.MaxHeight
	}
	a = append(a, fmt.Sprintf(
		"The diagram is %.0f x %.0f pixels and %.1f times too %s: to show its %.0f-pixel labels at %.0f points on %s it must fit within %.0f x %.0f.",
		r.Width, r.Height, over, map[string]string{"width": "wide", "height": "tall"}[r.LimitedBy],
		smallest.FontPx, r.Target.MinFontPt, r.Target.Name, r.MaxWidth, r.MaxHeight))
	if r.LongestLine > LongLine {
		a = append(a, fmt.Sprintf(
			"Wrap labels to about %d characters a line: the longest line has %d, and long lines make wide boxes.",
			WrapWidth, r.LongestLine))
	}
	if r.LimitedBy == "width" {
		a = append(a, "Use fewer columns, or a layout whose shape is closer to the target's.")
	} else {
		a = append(a, "Put fewer items in each column, or use a layout whose shape is closer to the target's.")
	}
	if over <= 1.5 {
		a = append(a, fmt.Sprintf(
			"Raise the label font from %.0f to about %.0f pixels; boxes grow with it, so check again.",
			smallest.FontPx, math.Ceil(smallest.FontPx*over*1.2)))
	} else {
		a = append(a, "Split the diagram into smaller ones, such as one per group of items.")
	}
	if fit := fittingTargets(r, smallest.FontPx); len(fit) > 0 {
		a = append(a, "As it is, it is readable on: "+strings.Join(fit, ", ")+".")
	}
	return a
}

// fittingTargets lists the built-in targets on which the diagram in r is
// readable as it is, given its smallest label's size in pixels.
func fittingTargets(r *Report, fontPx float64) []string {
	var out []string
	for _, t := range Targets {
		if t.Name == r.Target.Name {
			continue
		}
		s := math.Min(t.WidthIn*pointsPerInch/r.Width, t.HeightIn*pointsPerInch/r.Height)
		if fontPx*s >= t.MinFontPt-0.005 {
			out = append(out, t.Name)
		}
	}
	return out
}

func round2(f float64) float64 { return math.Round(f*100) / 100 }
func round4(f float64) float64 { return math.Round(f*10000) / 10000 }
