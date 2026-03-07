package format

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input   string
		want    Format
		wantErr bool
	}{
		{"toon", TOON, false},
		{"", TOON, false}, // default
		{"json", JSON, false},
		{"json-compact", JSONCompact, false},
		{"yaml", YAML, false},
		{"invalid", "", true},
		{"JSON", "", true}, // case-sensitive
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Parse(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestMarshal(t *testing.T) {
	type testStruct struct {
		Name  string `json:"name" yaml:"name"`
		Count int    `json:"count" yaml:"count"`
	}

	v := testStruct{Name: "test", Count: 42}

	t.Run("JSON", func(t *testing.T) {
		got, err := Marshal(v, JSON)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}
		want := "{\n  \"name\": \"test\",\n  \"count\": 42\n}"
		if string(got) != want {
			t.Errorf("Marshal() = %q, want %q", string(got), want)
		}
	})

	t.Run("JSONCompact", func(t *testing.T) {
		got, err := Marshal(v, JSONCompact)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}
		want := `{"name":"test","count":42}`
		if string(got) != want {
			t.Errorf("Marshal() = %q, want %q", string(got), want)
		}
	})

	t.Run("YAML", func(t *testing.T) {
		got, err := Marshal(v, YAML)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}
		s := string(got)
		if !strings.Contains(s, "name: test") {
			t.Errorf("Marshal() YAML output missing 'name: test': %q", s)
		}
		if !strings.Contains(s, "count: 42") {
			t.Errorf("Marshal() YAML output missing 'count: 42': %q", s)
		}
	})

	t.Run("TOON", func(t *testing.T) {
		got, err := Marshal(v, TOON)
		if err != nil {
			t.Fatalf("Marshal() error = %v", err)
		}
		s := string(got)
		// TOON should contain field names and values
		if !strings.Contains(s, "Name") || !strings.Contains(s, "test") {
			t.Errorf("Marshal() TOON output missing expected content: %q", s)
		}
		if !strings.Contains(s, "Count") || !strings.Contains(s, "42") {
			t.Errorf("Marshal() TOON output missing expected content: %q", s)
		}
	})
}

func TestUnmarshal(t *testing.T) {
	type testStruct struct {
		Name  string `json:"name" yaml:"name"`
		Count int    `json:"count" yaml:"count"`
	}

	t.Run("JSON", func(t *testing.T) {
		data := []byte(`{"name":"test","count":42}`)
		var got testStruct
		err := Unmarshal(data, &got, JSON)
		if err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}
		if got.Name != "test" || got.Count != 42 {
			t.Errorf("Unmarshal() = %+v, want {Name:test Count:42}", got)
		}
	})

	t.Run("YAML", func(t *testing.T) {
		data := []byte("name: test\ncount: 42\n")
		var got testStruct
		err := Unmarshal(data, &got, YAML)
		if err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}
		if got.Name != "test" || got.Count != 42 {
			t.Errorf("Unmarshal() = %+v, want {Name:test Count:42}", got)
		}
	})
}

func TestFormatString(t *testing.T) {
	tests := []struct {
		f    Format
		want string
	}{
		{TOON, "toon"},
		{JSON, "json"},
		{JSONCompact, "json-compact"},
		{YAML, "yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.f.String(); got != tt.want {
				t.Errorf("Format.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidFormats(t *testing.T) {
	formats := ValidFormats()
	if len(formats) != 4 {
		t.Errorf("ValidFormats() returned %d formats, want 4", len(formats))
	}
}
