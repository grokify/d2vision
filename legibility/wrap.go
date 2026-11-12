package legibility

import "strings"

// Wrap breaks text into lines of at most width characters, at spaces, so a
// label makes a taller, narrower box. A word longer than width stays whole on
// its own line. Runs of spaces collapse, and existing line breaks are kept.
// A width of zero or less returns the text with its spaces collapsed.
func Wrap(text string, width int) string {
	var out []string
	for _, para := range strings.Split(text, "\n") {
		words := strings.Fields(para)
		if len(words) == 0 {
			continue
		}
		if width <= 0 {
			out = append(out, strings.Join(words, " "))
			continue
		}
		line := words[0]
		for _, w := range words[1:] {
			if len([]rune(line))+1+len([]rune(w)) > width {
				out = append(out, line)
				line = w
				continue
			}
			line += " " + w
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}
