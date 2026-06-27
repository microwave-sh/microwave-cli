package config

import "testing"

func TestIsNewer(t *testing.T) {
	cases := []struct {
		latest, current string
		want            bool
	}{
		// The reported bug: a stale/older "latest" must NOT prompt an update.
		{"1.6.0", "1.7.1", false},
		{"v1.6.0", "v1.7.1", false},
		{"1.7.1", "1.7.1", false}, // equal
		// Genuine upgrades across each component.
		{"1.7.2", "1.7.1", true},
		{"1.8.0", "1.7.9", true},
		{"2.0.0", "1.9.9", true},
		{"v1.7.2", "1.7.1", true}, // mixed v-prefix
		// Pre-release / build suffixes are ignored for the core comparison.
		{"1.7.2-rc1", "1.7.1", true},
		{"1.7.1+build5", "1.7.1", false},
		// Unparsable versions (the "dev" default) never nag.
		{"dev", "1.7.1", false},
		{"1.7.2", "dev", false},
		{"", "1.7.1", false},
		{"1.7", "1.7.1", false}, // not 3 components
	}
	for _, c := range cases {
		if got := isNewer(c.latest, c.current); got != c.want {
			t.Errorf("isNewer(%q, %q) = %v, want %v", c.latest, c.current, got, c.want)
		}
	}
}
