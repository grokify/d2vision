package render

import "fmt"

// ScaleFontSize returns d2Code with a global font-size floor applied to every
// shape and edge label, so all diagram text renders at least `size` px. It
// prepends two D2 glob directives; elements that set their own explicit
// font-size still win. If size <= 0 the code is returned unchanged.
func ScaleFontSize(d2Code string, size int64) string {
	if size <= 0 {
		return d2Code
	}
	prefix := fmt.Sprintf("**.style.font-size: %d\n(** -> **)[*].style.font-size: %d\n\n", size, size)
	return prefix + d2Code
}
