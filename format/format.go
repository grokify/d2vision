// Package format provides output format abstraction for CLI commands.
// TOON (Token-Oriented Object Notation) is the default format, optimized
// for LLM consumption with ~40% fewer tokens than JSON.
package format

import (
	"encoding/json"
	"fmt"

	toon "github.com/toon-format/toon-go"
	"gopkg.in/yaml.v3"
)

// Format represents an output format type.
type Format string

// Supported output formats.
const (
	TOON        Format = "toon"
	JSON        Format = "json"
	JSONCompact Format = "json-compact"
	YAML        Format = "yaml"
)

// Parse parses a format string into a Format type.
// Empty string defaults to TOON.
func Parse(s string) (Format, error) {
	switch s {
	case "toon", "":
		return TOON, nil
	case "json":
		return JSON, nil
	case "json-compact":
		return JSONCompact, nil
	case "yaml":
		return YAML, nil
	default:
		return "", fmt.Errorf("unknown format %q: use toon, json, json-compact, or yaml", s)
	}
}

// Marshal serializes v to the specified format.
func Marshal(v any, f Format) ([]byte, error) {
	switch f {
	case TOON:
		return toon.Marshal(v)
	case JSON:
		return json.MarshalIndent(v, "", "  ")
	case JSONCompact:
		return json.Marshal(v)
	case YAML:
		return yaml.Marshal(v)
	default:
		return toon.Marshal(v)
	}
}

// Unmarshal deserializes data in the specified format into v.
func Unmarshal(data []byte, v any, f Format) error {
	switch f {
	case TOON:
		return toon.Unmarshal(data, v)
	case JSON, JSONCompact:
		return json.Unmarshal(data, v)
	case YAML:
		return yaml.Unmarshal(data, v)
	default:
		return toon.Unmarshal(data, v)
	}
}

// String returns the string representation of the format.
func (f Format) String() string {
	return string(f)
}

// ValidFormats returns a list of all valid format strings.
func ValidFormats() []string {
	return []string{"toon", "json", "json-compact", "yaml"}
}
