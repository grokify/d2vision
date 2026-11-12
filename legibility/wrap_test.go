package legibility_test

import (
	"testing"

	"github.com/grokify/d2vision/legibility"
)

func TestWrap(t *testing.T) {
	for _, c := range []struct {
		text  string
		width int
		want  string
	}{
		{"Deliver capacity building trainings and workshops", 20, "Deliver capacity\nbuilding trainings\nand workshops"},
		{"short", 20, "short"},
		{"an extraordinarilylongword here", 10, "an\nextraordinarilylongword\nhere"},
		{"  spaces   collapse  ", 40, "spaces collapse"},
		{"kept\nbreak here please", 10, "kept\nbreak here\nplease"},
		{"no   width", 0, "no width"},
		{"", 10, ""},
	} {
		if got := legibility.Wrap(c.text, c.width); got != c.want {
			t.Errorf("Wrap(%q, %d) = %q, want %q", c.text, c.width, got, c.want)
		}
	}
}
