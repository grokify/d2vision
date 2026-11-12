package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/grokify/d2vision"
	"github.com/grokify/d2vision/format"
	"github.com/grokify/d2vision/legibility"
	"github.com/spf13/cobra"
)

var (
	legibilityTargets []string
	legibilityFormat  string
	legibilityCheck   bool
	legibilityWidth   float64
	legibilityHeight  float64
	legibilityMinFont float64
	legibilityShow    int
)

var legibilityCmd = &cobra.Command{
	Use:   "legibility <file.svg>",
	Short: "Check whether a diagram's labels can be read where it is shown",
	Long: `Check whether a D2-generated SVG can be read on a page, a slide, or a screen.

The diagram is scaled to fit each target, keeping its shape, and every label is
measured at the size it is shown. A label smaller than the target's minimum is
too small. For a diagram that is not readable, the report says how many times
too large it is and what would help: wrapping labels, fewer columns or rows, a
larger font, splitting it, or a target it already fits. It also lists arrows
drawn over boxes they do not join, which a layout that routes arrows avoids.

Targets (minimum label size):
  docx-portrait       Word, US Letter portrait, 6.5 x 9 in (9 pt)
  docx-landscape      Word, US Letter landscape, 9 x 6.5 in (9 pt)
  tabloid-landscape   Tabloid landscape, 15 x 9 in (9 pt)
  slide-16x9          16:9 slide below a title, 12.33 x 5.9 in (18 pt)
  screen              Browser window, 1280 x 720 px (12 px, 9 pt)

With --width, --height, and --min-font, the diagram is checked against that
area instead.

Output formats:
  - text (default): Human-readable report
  - toon: TOON format (token-efficient)
  - json: JSON format

Examples:
  # Check every built-in target
  d2vision legibility diagram.svg

  # Check a Word page, and fail if a label is too small
  d2vision legibility diagram.svg --target docx-portrait --check

  # Check an area 8 inches wide and 5 tall, reading from 10 points
  d2vision legibility diagram.svg --width 8 --height 5 --min-font 10
`,
	Args: cobra.ExactArgs(1),
	RunE: runLegibility,
}

func init() {
	legibilityCmd.Flags().StringSliceVarP(&legibilityTargets, "target", "t", nil,
		"Target to check, repeatable: "+strings.Join(legibility.TargetNames(), ", ")+" (default: all)")
	legibilityCmd.Flags().StringVarP(&legibilityFormat, "format", "f", "text", "Output format: text, toon, json")
	legibilityCmd.Flags().BoolVar(&legibilityCheck, "check", false, "Exit with status 1 if a label is too small on a target or an arrow is drawn over a box")
	legibilityCmd.Flags().Float64Var(&legibilityWidth, "width", 0, "Width of a custom target, in inches")
	legibilityCmd.Flags().Float64Var(&legibilityHeight, "height", 0, "Height of a custom target, in inches")
	legibilityCmd.Flags().Float64Var(&legibilityMinFont, "min-font", 9, "Smallest readable label on a custom target, in points")
	legibilityCmd.Flags().IntVar(&legibilityShow, "show", 5, "Number of too-small labels to list per target in text output")
}

func runLegibility(cmd *cobra.Command, args []string) error {
	targets, err := legibilityTargetList()
	if err != nil {
		return err
	}
	diagram, err := d2vision.ParseFile(args[0])
	if err != nil {
		return fmt.Errorf("parsing %s: %w", args[0], err)
	}
	reports, err := legibility.AnalyzeAll(diagram, targets)
	if err != nil {
		return err
	}

	var out string
	if legibilityFormat == "text" {
		out = formatLegibility(args[0], reports, legibilityShow)
	} else {
		f, err := format.Parse(legibilityFormat)
		if err != nil {
			return err
		}
		b, err := format.Marshal(reports, f)
		if err != nil {
			return fmt.Errorf("marshaling output: %w", err)
		}
		out = string(b) + "\n"
	}
	if _, err := io.WriteString(cmd.OutOrStdout(), out); err != nil {
		return err
	}

	if legibilityCheck {
		for _, r := range reports {
			if !r.Clear() {
				os.Exit(1)
			}
		}
	}
	return nil
}

// legibilityTargetList returns the targets the flags ask for.
func legibilityTargetList() ([]legibility.Target, error) {
	if legibilityWidth > 0 || legibilityHeight > 0 {
		if legibilityWidth <= 0 || legibilityHeight <= 0 {
			return nil, fmt.Errorf("a custom target needs both --width and --height")
		}
		if len(legibilityTargets) > 0 {
			return nil, fmt.Errorf("use --target or --width and --height, not both")
		}
		return []legibility.Target{{
			Name:        "custom",
			Description: fmt.Sprintf("%.2f x %.2f inches", legibilityWidth, legibilityHeight),
			WidthIn:     legibilityWidth,
			HeightIn:    legibilityHeight,
			MinFontPt:   legibilityMinFont,
		}}, nil
	}
	if len(legibilityTargets) == 0 {
		return legibility.Targets, nil
	}
	var out []legibility.Target
	for _, name := range legibilityTargets {
		t, ok := legibility.TargetByName(name)
		if !ok {
			return nil, fmt.Errorf("unknown target %q; targets are %s", name, strings.Join(legibility.TargetNames(), ", "))
		}
		out = append(out, t)
	}
	return out, nil
}

// formatLegibility writes the reports as text.
func formatLegibility(path string, reports []*legibility.Report, show int) string {
	if len(reports) == 0 {
		return ""
	}
	var w strings.Builder
	r0 := reports[0]
	fmt.Fprintf(&w, "%s: %.0f x %.0f pixels, %d labels, longest line %d characters\n\n",
		path, r0.Width, r0.Height, r0.Labels, r0.LongestLine)
	for _, r := range reports {
		mark := "✓ readable"
		if !r.Readable {
			mark = fmt.Sprintf("✗ %d of %d labels too small", len(r.TooSmall), r.Labels)
		}
		if n := len(r.Crossings); n == 1 {
			mark += "; ✗ 1 arrow over a box"
		} else if n > 1 {
			mark += fmt.Sprintf("; ✗ %d arrows over boxes", n)
		}
		fmt.Fprintf(&w, "%s (%s, from %.0f pt): %s\n", r.Target.Name, r.Target.Description, r.Target.MinFontPt, mark)
		fmt.Fprintf(&w, "  shown at %.0f%% (fits the %s); smallest label %.1f pt, median %.1f pt\n",
			r.Scale*100*4/3, r.LimitedBy, r.SmallestPt, r.MedianPt)
		for i, l := range r.TooSmall {
			if i == show {
				fmt.Fprintf(&w, "  … and %d more\n", len(r.TooSmall)-show)
				break
			}
			fmt.Fprintf(&w, "  - %s: %.1f pt (%q)\n", l.ID, l.ShownPt, truncate(l.Text, 50))
		}
		for i, c := range r.Crossings {
			if i == show {
				fmt.Fprintf(&w, "  … and %d more arrows\n", len(r.Crossings)-show)
				break
			}
			fmt.Fprintf(&w, "  - arrow %s -> %s over %s\n", c.From, c.To, c.Node)
		}
		for _, a := range r.Advice {
			fmt.Fprintf(&w, "  → %s\n", a)
		}
		w.WriteString("\n")
	}
	return w.String()
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}
