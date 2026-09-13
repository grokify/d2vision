package lint

import (
	"fmt"
	"os"
	"path/filepath"

	yaml "go.yaml.in/yaml/v3"
)

// ConfigFilename is the default config file name discovered by FindConfig.
const ConfigFilename = ".d2vision.yaml"

// Config declares which lint rules run and how, so one CLI invocation applies a
// whole policy (in the spirit of golangci-lint's YAML). Unset fields fall back
// to the registry defaults, so a partial config only overrides what it names.
//
// Example:
//
//	rules:
//	  text-overlap:
//	    enabled: true
//	    settings:
//	      layout: elk
//	  deep-nesting:
//	    enabled: false
//	  corner-near:
//	    severity: error
type Config struct {
	Rules map[string]RuleConfig `yaml:"rules"`
}

// RuleConfig overrides one rule. Enabled is a pointer so "unset" is
// distinguishable from an explicit false. Severity, when non-empty, overrides
// the registry severity. Settings holds rule-specific options (e.g. the
// text-overlap layout engine).
type RuleConfig struct {
	Enabled  *bool             `yaml:"enabled"`
	Severity string            `yaml:"severity"`
	Settings map[string]string `yaml:"settings"`
}

// defaultDisabled lists rules that are off unless a config enables them.
// text-overlap runs a full layout pass (slow, engine-dependent), so it is
// opt-in rather than part of the default fast, source-only lint.
var defaultDisabled = map[Code]bool{
	CodeTextOverlap: true,
}

// DefaultConfig returns the built-in policy: every registered rule enabled at
// its registry severity, except the opt-in rules in defaultDisabled.
func DefaultConfig() *Config {
	return &Config{Rules: map[string]RuleConfig{}}
}

// LoadConfig reads a YAML config from path.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	if c.Rules == nil {
		c.Rules = map[string]RuleConfig{}
	}
	return &c, nil
}

// FindConfig looks for ConfigFilename in startDir and each parent directory,
// returning the first match or "" if none is found.
func FindConfig(startDir string) string {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return ""
	}
	for {
		candidate := filepath.Join(dir, ConfigFilename)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// IsEnabled reports whether a rule runs: an explicit config value wins,
// otherwise the rule is on unless it is opt-in (defaultDisabled).
func (c *Config) IsEnabled(code Code) bool {
	if c != nil {
		if rc, ok := c.Rules[string(code)]; ok && rc.Enabled != nil {
			return *rc.Enabled
		}
	}
	return !defaultDisabled[code]
}

// SeverityFor returns the effective severity for a code: a config override if
// present, otherwise the registry severity.
func (c *Config) SeverityFor(code Code) string {
	if c != nil {
		if rc, ok := c.Rules[string(code)]; ok && rc.Severity != "" {
			return rc.Severity
		}
	}
	return SeverityOf(code)
}

// Setting returns a rule-specific setting value, or def if unset.
func (c *Config) Setting(code Code, key, def string) string {
	if c != nil {
		if rc, ok := c.Rules[string(code)]; ok {
			if v, ok := rc.Settings[key]; ok && v != "" {
				return v
			}
		}
	}
	return def
}
