package card

import "strings"

// textWidth estimates the rendered width of s in a proportional sans-serif at
// the given size. It is deliberately conservative (slightly wide) so wrapped
// text errs toward fitting; SVG has no text measurement of its own.
func textWidth(s string, size float64, bold bool) float64 {
	var w float64
	for _, r := range s {
		switch {
		case strings.ContainsRune("iljI.,:;'|!()[]/ ", r):
			w += 0.30
		case strings.ContainsRune("mwMW@", r):
			w += 0.86
		case r >= 'A' && r <= 'Z':
			w += 0.68
		case r >= '0' && r <= '9':
			w += 0.58
		case r > 0x2000: // emoji and symbols
			w += 1.1
		default:
			w += 0.55
		}
	}
	if bold {
		w *= 1.07
	}
	return w * size
}

// wrap breaks text into lines no wider than maxWidth. "\n" forces a break.
func wrap(text string, size, maxWidth float64, bold bool) []string {
	var lines []string
	for _, para := range strings.Split(text, "\n") {
		words := strings.Fields(para)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}
		cur := words[0]
		for _, w := range words[1:] {
			if textWidth(cur+" "+w, size, bold) <= maxWidth {
				cur += " " + w
				continue
			}
			lines = append(lines, cur)
			cur = w
		}
		lines = append(lines, cur)
	}
	return lines
}
