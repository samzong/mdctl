package cmd

import (
	"runtime/debug"
	"testing"
)

func TestResolvedVersion(t *testing.T) {
	cases := []struct {
		name    string
		version string
		module  string
		ok      bool
		want    string
	}{
		{"linker", "1.2.3", "v0.1.0", true, "1.2.3"},
		{"linker tag", "v1.2.3-rc.1", "v0.1.0", true, "1.2.3-rc.1"},
		{"module", "dev", "v1.2.3", true, "1.2.3"},
		{"module prerelease", "dev", "v1.2.3-rc.1", true, "1.2.3-rc.1"},
		{"empty linker", "", "v1.2.3", true, "1.2.3"},
		{"source", "dev", "(devel)", true, "dev"},
		{"empty module", "dev", "", true, "dev"},
		{"unavailable", "dev", "", false, "dev"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := resolvedVersion(tc.version, func() (*debug.BuildInfo, bool) {
				if tc.version != "" && tc.version != "dev" {
					t.Fatal("linker version must take precedence over build information")
				}
				return &debug.BuildInfo{Main: debug.Module{Version: tc.module}}, tc.ok
			})
			if got != tc.want {
				t.Fatalf("resolvedVersion() = %q, want %q", got, tc.want)
			}
		})
	}
}
