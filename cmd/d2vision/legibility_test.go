package main

import (
	"strings"
	"testing"

	"github.com/grokify/d2vision"
	"github.com/grokify/d2vision/legibility"
)

func TestLegibilityTargetList(t *testing.T) {
	defer func() {
		legibilityTargets, legibilityWidth, legibilityHeight = nil, 0, 0
	}()

	legibilityTargets = nil
	all, err := legibilityTargetList()
	if err != nil || len(all) != len(legibility.Targets) {
		t.Fatalf("default: %d targets, %v", len(all), err)
	}

	legibilityTargets = []string{"screen", "docx-portrait"}
	two, err := legibilityTargetList()
	if err != nil || len(two) != 2 || two[0].Name != "screen" {
		t.Fatalf("named: %+v, %v", two, err)
	}

	legibilityTargets = []string{"nope"}
	if _, err := legibilityTargetList(); err == nil || !strings.Contains(err.Error(), "unknown target") {
		t.Errorf("unknown target: %v", err)
	}

	legibilityTargets, legibilityWidth, legibilityHeight = nil, 8, 5
	custom, err := legibilityTargetList()
	if err != nil || len(custom) != 1 || custom[0].WidthIn != 8 || custom[0].MinFontPt != legibilityMinFont {
		t.Fatalf("custom: %+v, %v", custom, err)
	}

	legibilityHeight = 0
	if _, err := legibilityTargetList(); err == nil {
		t.Error("a custom target without a height was accepted")
	}
}

func TestPrintLegibility(t *testing.T) {
	d, err := d2vision.ParseFile("../../testdata/classed.svg")
	if err != nil {
		t.Fatal(err)
	}
	tg, _ := legibility.TargetByName("docx-portrait")
	wide := *d
	wide.ViewBox.Width *= 20
	reports, err := legibility.AnalyzeAll(&wide, []legibility.Target{tg})
	if err != nil {
		t.Fatal(err)
	}
	out := formatLegibility("classed.svg", reports, 1)
	for _, want := range []string{"3 labels, longest line 15", "docx-portrait", "✗ 3 of 3 labels too small", "… and 2 more", "→ The diagram is"} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
}
