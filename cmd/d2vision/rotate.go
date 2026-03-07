package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var (
	rotateAngle  int
	rotateOutput string
)

var rotateCmd = &cobra.Command{
	Use:   "rotate <input.svg>",
	Short: "Rotate an SVG by 90° increments",
	Long: `Rotate an SVG diagram by 90° increments (counter-clockwise).

This is useful for converting landscape diagrams to portrait orientation
for PDF embedding with left-side binding.

Rotation angles (counter-clockwise, mathematical convention):
  90   - 90° CCW (landscape → portrait, top goes to left/binding side)
  180  - 180° (flip upside down)
  270  - 270° CCW (same as 90° clockwise)
  -90  - 90° CW (same as 270° CCW)

Examples:
  # Rotate landscape to portrait for left-binding PDF
  d2vision rotate diagram.svg --angle 90 -o portrait.svg

  # Rotate and pipe to stdout
  d2vision rotate diagram.svg --angle 90

  # Read from stdin
  cat diagram.svg | d2vision rotate - --angle 90 > portrait.svg

  # Chain with pipeline command
  d2vision pipeline spec.json --svg | d2vision rotate - --angle 90 > portrait.svg
`,
	Args: cobra.ExactArgs(1),
	RunE: runRotate,
}

func init() {
	rotateCmd.Flags().IntVarP(&rotateAngle, "angle", "a", 90, "Rotation angle in degrees (90, 180, 270, -90, -180, -270)")
	rotateCmd.Flags().StringVarP(&rotateOutput, "output", "o", "", "Output file (default: stdout)")
}

func runRotate(cmd *cobra.Command, args []string) error {
	// Normalize angle to 0, 90, 180, 270
	angle := normalizeAngle(rotateAngle)
	if angle != 0 && angle != 90 && angle != 180 && angle != 270 {
		return fmt.Errorf("angle must be a multiple of 90 (got %d)", rotateAngle)
	}

	// No rotation needed
	if angle == 0 {
		return fmt.Errorf("angle 0 results in no rotation")
	}

	// Read input
	inputPath := args[0]
	var data []byte
	var err error

	if inputPath == "-" {
		data, err = io.ReadAll(os.Stdin)
	} else {
		data, err = os.ReadFile(inputPath)
	}
	if err != nil {
		return fmt.Errorf("reading input: %w", err)
	}

	// Rotate SVG
	rotated, err := rotateSVG(data, angle)
	if err != nil {
		return fmt.Errorf("rotating SVG: %w", err)
	}

	// Write output
	if rotateOutput != "" {
		if err := os.WriteFile(rotateOutput, rotated, 0644); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
	} else {
		if _, err := os.Stdout.Write(rotated); err != nil {
			return fmt.Errorf("writing to stdout: %w", err)
		}
	}

	return nil
}

// normalizeAngle converts angle to 0, 90, 180, or 270.
// Counter-clockwise is positive (mathematical convention).
func normalizeAngle(angle int) int {
	// Normalize to 0-359
	a := angle % 360
	if a < 0 {
		a += 360
	}
	return a
}

// rotateSVG rotates an SVG by the given angle (must be 90, 180, or 270).
// Rotation is counter-clockwise.
func rotateSVG(data []byte, angle int) ([]byte, error) {
	// Parse SVG dimensions
	width, height, viewBox, err := parseSVGDimensions(data)
	if err != nil {
		return nil, err
	}

	// Calculate new dimensions
	var newWidth, newHeight float64
	var transform string

	vbParts := strings.Fields(viewBox)
	if len(vbParts) != 4 {
		return nil, fmt.Errorf("invalid viewBox: %s", viewBox)
	}
	vbWidth, _ := strconv.ParseFloat(vbParts[2], 64)
	vbHeight, _ := strconv.ParseFloat(vbParts[3], 64)

	switch angle {
	case 90: // 90° CCW
		newWidth = height
		newHeight = width
		// Rotate around origin, then translate to put in positive quadrant
		transform = fmt.Sprintf("translate(0, %.2f) rotate(-90)", vbWidth)
	case 180: // 180°
		newWidth = width
		newHeight = height
		transform = fmt.Sprintf("translate(%.2f, %.2f) rotate(180)", vbWidth, vbHeight)
	case 270: // 270° CCW (90° CW)
		newWidth = height
		newHeight = width
		transform = fmt.Sprintf("translate(%.2f, 0) rotate(90)", vbHeight)
	default:
		return nil, fmt.Errorf("unsupported angle: %d", angle)
	}

	// Build new viewBox
	var newVBWidth, newVBHeight float64
	if angle == 90 || angle == 270 {
		newVBWidth = vbHeight
		newVBHeight = vbWidth
	} else {
		newVBWidth = vbWidth
		newVBHeight = vbHeight
	}
	newViewBox := fmt.Sprintf("0 0 %.2f %.2f", newVBWidth, newVBHeight)

	// Modify SVG
	result := updateSVGDimensions(data, newWidth, newHeight, newViewBox, transform)

	return result, nil
}

// parseSVGDimensions extracts width, height, and viewBox from SVG.
func parseSVGDimensions(data []byte) (width, height float64, viewBox string, err error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, 0, "", err
		}

		if se, ok := tok.(xml.StartElement); ok && se.Name.Local == "svg" {
			for _, attr := range se.Attr {
				switch attr.Name.Local {
				case "width":
					width = parseLength(attr.Value)
				case "height":
					height = parseLength(attr.Value)
				case "viewBox":
					viewBox = attr.Value
				}
			}
			return width, height, viewBox, nil
		}
	}

	return 0, 0, "", fmt.Errorf("no <svg> element found")
}

// parseLength parses a CSS length value (e.g., "800", "800px", "100%").
func parseLength(s string) float64 {
	s = strings.TrimSuffix(s, "px")
	s = strings.TrimSuffix(s, "pt")
	s = strings.TrimSuffix(s, "em")
	s = strings.TrimSuffix(s, "%")
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// updateSVGDimensions modifies the SVG to have new dimensions and wraps content in a transform group.
// Only modifies the root (first) SVG element.
func updateSVGDimensions(data []byte, width, height float64, viewBox, transform string) []byte {
	content := string(data)

	// Find the first <svg tag
	svgStart := strings.Index(content, "<svg")
	if svgStart == -1 {
		return data
	}

	// Find the end of the first <svg ...> tag
	svgTagEnd := strings.Index(content[svgStart:], ">")
	if svgTagEnd == -1 {
		return data
	}
	svgTagEnd += svgStart

	// Extract the first SVG tag
	svgTag := content[svgStart : svgTagEnd+1]

	// Update attributes in the first SVG tag only
	newSvgTag := svgTag

	// Update width
	widthRe := regexp.MustCompile(`\s+width="[^"]*"`)
	if widthRe.MatchString(newSvgTag) {
		newSvgTag = widthRe.ReplaceAllString(newSvgTag, fmt.Sprintf(` width="%.2f"`, width))
	}

	// Update height
	heightRe := regexp.MustCompile(`\s+height="[^"]*"`)
	if heightRe.MatchString(newSvgTag) {
		newSvgTag = heightRe.ReplaceAllString(newSvgTag, fmt.Sprintf(` height="%.2f"`, height))
	}

	// Update viewBox
	viewBoxRe := regexp.MustCompile(`\s+viewBox="[^"]*"`)
	if viewBoxRe.MatchString(newSvgTag) {
		newSvgTag = viewBoxRe.ReplaceAllString(newSvgTag, fmt.Sprintf(` viewBox="%s"`, viewBox))
	}

	// Build result: prefix + modified svg tag + transform group + content + close group + suffix
	prefix := content[:svgStart]
	suffix := content[svgTagEnd+1:]

	// Find the last </svg> to close the transform group before it
	lastSvgClose := strings.LastIndex(suffix, "</svg>")
	if lastSvgClose == -1 {
		return data
	}

	innerContent := suffix[:lastSvgClose]
	afterClose := suffix[lastSvgClose:]

	result := prefix + newSvgTag + fmt.Sprintf(`<g transform="%s">`, transform) + innerContent + "</g>" + afterClose

	return []byte(result)
}
