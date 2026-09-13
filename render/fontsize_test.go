package render

import (
	"context"
	"strings"
	"testing"
)

func TestScaleFontSize(t *testing.T) {
	if got := ScaleFontSize("a -> b", 0); got != "a -> b" {
		t.Fatalf("size<=0 should return code unchanged, got %q", got)
	}
	if got := ScaleFontSize("a -> b", -5); got != "a -> b" {
		t.Fatalf("negative size should return code unchanged, got %q", got)
	}
	got := ScaleFontSize("a -> b", 28)
	if !strings.HasPrefix(got, "**.style.font-size: 28\n(** -> **)[*].style.font-size: 28\n\n") {
		t.Fatalf("unexpected prefix: %q", got)
	}
	if !strings.HasSuffix(got, "a -> b") {
		t.Fatalf("original code should be preserved as suffix: %q", got)
	}
}

func TestRenderSVGFontSize(t *testing.T) {
	r, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ctx := context.Background()
	const code = "a -> b: hello"

	scaled, err := r.RenderSVG(ctx, code, &Options{FontSize: 28})
	if err != nil {
		t.Fatalf("RenderSVG scaled: %v", err)
	}
	if !strings.Contains(string(scaled), "font-size:28px") {
		t.Fatalf("scaled SVG should contain font-size:28px")
	}

	def, err := r.RenderSVG(ctx, code, &Options{FontSize: 0})
	if err != nil {
		t.Fatalf("RenderSVG default: %v", err)
	}
	if strings.Contains(string(def), "font-size:28px") {
		t.Fatalf("default SVG should not contain font-size:28px")
	}
	if !strings.Contains(string(def), "font-size:16px") {
		t.Fatalf("default SVG should contain the default font-size:16px")
	}
}
