package lint

import "testing"

func TestCheckCornerNear(t *testing.T) {
	tests := []struct {
		name      string
		src       string
		wantCount int
		wantNear  string
	}{
		{
			name: "corner near is flagged",
			src: `a -> b
legend: Legend {
  near: top-left
}`,
			wantCount: 1,
			wantNear:  "top-left",
		},
		{
			name: "each corner is flagged",
			src: `l1: A { near: top-right }
l2: B { near: bottom-left }
l3: C { near: bottom-right }`,
			wantCount: 3,
		},
		{
			name: "edge-center near is not flagged",
			src: `a -> b
legend: Legend {
  near: bottom-center
}`,
			wantCount: 0,
		},
		{
			name:      "no near is not flagged",
			src:       `a -> b`,
			wantCount: 0,
		},
		{
			name: "near referencing a real object is not flagged",
			src: `anchor: Anchor
note: Note {
  near: anchor
}`,
			wantCount: 0,
		},
		{
			name: "an object literally named like a corner is not a constant near",
			src: `top-left: Real Node
note: Note {
  near: top-left
}`,
			// near: top-left now resolves to the real object "top-left",
			// so it is NOT a constant near and must not be flagged.
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			findings, err := CheckCornerNear(tt.src)
			if err != nil {
				t.Fatalf("CheckCornerNear: %v", err)
			}
			if len(findings) != tt.wantCount {
				t.Fatalf("got %d findings, want %d: %+v", len(findings), tt.wantCount, findings)
			}
			if tt.wantNear != "" {
				if findings[0].Near != tt.wantNear {
					t.Errorf("near = %q, want %q", findings[0].Near, tt.wantNear)
				}
				if findings[0].Code != CodeCornerNear {
					t.Errorf("code = %q, want %q", findings[0].Code, CodeCornerNear)
				}
				if findings[0].Severity != SeverityError {
					t.Errorf("severity = %q, want %q", findings[0].Severity, SeverityError)
				}
				if findings[0].Suggestion == "" {
					t.Errorf("suggestion is empty; expected registry text")
				}
				if !HasErrors(findings) {
					t.Errorf("HasErrors = false, want true")
				}
			}
		})
	}
}

func TestRegistry(t *testing.T) {
	// Every finding a check can emit must be a registered code.
	if _, ok := Lookup(CodeCornerNear); !ok {
		t.Fatalf("CodeCornerNear not registered")
	}
	if got := Text(CodeCornerNear); got == "" {
		t.Errorf("Text(CodeCornerNear) is empty")
	}
	if got := SeverityOf(CodeCornerNear); got != SeverityError {
		t.Errorf("SeverityOf(CodeCornerNear) = %q, want %q", got, SeverityError)
	}
	if got := Explain(CodeCornerNear); got == "" {
		t.Errorf("Explain(CodeCornerNear) is empty")
	}

	// Unknown codes look up cleanly (net/http.StatusText-style: "" not panic).
	if _, ok := Lookup(Code("nope")); ok {
		t.Errorf("Lookup(nope) reported ok")
	}
	if got := Text(Code("nope")); got != "" {
		t.Errorf("Text(nope) = %q, want \"\"", got)
	}
	if got := Explain(Code("nope")); got != "" {
		t.Errorf("Explain(nope) = %q, want \"\"", got)
	}

	// Registry must be internally consistent: each rule's Code matches its key
	// and carries the required text.
	for _, r := range Rules() {
		if r.Severity == "" || r.Title == "" || r.Suggestion == "" || r.Remediation == "" {
			t.Errorf("rule %q missing required text: %+v", r.Code, r)
		}
	}
}
