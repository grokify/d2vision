package export

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// SVGInfo is what capture needs to know about an SVG document.
type SVGInfo struct {
	Width, Height int     // CSS pixels
	Loop          float64 // seconds; 0 if no animation duration was found
}

var (
	// SMIL: dur="2s", dur="750ms", dur="1.5". "indefinite" is skipped.
	rxSMILDur = regexp.MustCompile(`\bdur="\s*([0-9.]+)\s*(ms|s)?\s*"`)
	// CSS: animation-duration: 1.2s, or the shorthand "animation: name 4.9s linear
	// infinite" that D2's animated edges use.
	rxCSSDur = regexp.MustCompile(`animation(?:-duration)?:[^;}]*?\b([0-9.]+)(ms|s)\b`)
)

// InspectSVG reads the pixel size from the root element (width/height, falling
// back to the viewBox) and the animation loop as the longest animation
// duration found. A document whose animations have different durations only
// repeats seamlessly at their common multiple, so callers should pass an
// explicit loop when in doubt.
func InspectSVG(svg []byte) (SVGInfo, error) {
	var info SVGInfo
	dec := xml.NewDecoder(strings.NewReader(string(svg)))
	dec.Strict = false
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			return info, errors.New("export: no <svg> element found")
		}
		if err != nil {
			return info, fmt.Errorf("export: parse svg: %w", err)
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "svg" {
			continue
		}
		var w, h float64
		var vb []float64
		for _, a := range se.Attr {
			switch a.Name.Local {
			case "width":
				w = cssLength(a.Value)
			case "height":
				h = cssLength(a.Value)
			case "viewBox":
				for _, f := range strings.Fields(strings.ReplaceAll(a.Value, ",", " ")) {
					v, _ := strconv.ParseFloat(f, 64)
					vb = append(vb, v)
				}
			}
		}
		if (w <= 0 || h <= 0) && len(vb) == 4 {
			w, h = vb[2], vb[3]
		}
		if w <= 0 || h <= 0 {
			return info, errors.New("export: svg has no usable width/height or viewBox; pass them explicitly")
		}
		info.Width, info.Height = int(math.Round(w)), int(math.Round(h))
		break
	}

	for _, rx := range []*regexp.Regexp{rxSMILDur, rxCSSDur} {
		for _, m := range rx.FindAllStringSubmatch(string(svg), -1) {
			v, err := strconv.ParseFloat(m[1], 64)
			if err != nil || v <= 0 {
				continue
			}
			if m[2] == "ms" {
				v /= 1000
			}
			info.Loop = max(info.Loop, v)
		}
	}
	return info, nil
}

// cssLength parses "800", "800px" or "12.5"; percentages and other units are
// not meaningful without a container and yield 0.
func cssLength(s string) float64 {
	s = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(s), "px"))
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}
