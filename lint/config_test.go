package lint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	c := DefaultConfig()
	if !c.IsEnabled(CodeCornerNear) {
		t.Errorf("corner-near should be enabled by default")
	}
	if c.IsEnabled(CodeTextOverlap) {
		t.Errorf("text-overlap should be opt-in (disabled) by default")
	}
	if got := c.SeverityFor(CodeCornerNear); got != SeverityError {
		t.Errorf("default severity for corner-near = %q, want %q", got, SeverityError)
	}
}

func TestLoadConfigOverrides(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ConfigFilename)
	yaml := `rules:
  text-overlap:
    enabled: true
    settings:
      layout: elk
  deep-nesting:
    enabled: false
  corner-near:
    severity: warning
`
	if err := os.WriteFile(path, []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if !c.IsEnabled(CodeTextOverlap) {
		t.Errorf("text-overlap should be enabled by config")
	}
	if c.Setting(CodeTextOverlap, "layout", "dagre") != "elk" {
		t.Errorf("layout setting = %q, want elk", c.Setting(CodeTextOverlap, "layout", "dagre"))
	}
	if c.IsEnabled(CodeDeepNesting) {
		t.Errorf("deep-nesting should be disabled by config")
	}
	if got := c.SeverityFor(CodeCornerNear); got != SeverityWarning {
		t.Errorf("corner-near severity override = %q, want warning", got)
	}
	// A rule not mentioned in the config keeps its default.
	if !c.IsEnabled(CodeMissingGrid) {
		t.Errorf("unmentioned rule should keep default (enabled)")
	}
}

func TestFindConfig(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "a", "b")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	cfgPath := filepath.Join(dir, ConfigFilename)
	if err := os.WriteFile(cfgPath, []byte("rules: {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Discovered by walking up from a nested directory.
	if got := FindConfig(sub); got != cfgPath {
		t.Errorf("FindConfig = %q, want %q", got, cfgPath)
	}
	// None found outside the tree.
	if got := FindConfig(t.TempDir()); got != "" {
		t.Errorf("FindConfig = %q, want \"\"", got)
	}
}
