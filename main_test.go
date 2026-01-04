package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckGoMod(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Create a go.mod with intentionally old pseudo-versions.
	// These are real old commits that should have newer versions available.
	content := `module test

go 1.25

require (
	github.com/maxmind/mmdbwriter v1.1.1-0.20240104181157-4f07c5502982
	go4.org/netipx v0.0.0-20220925034521-797b0c90d8ab
)
`
	dir := t.TempDir()
	gomodPath := filepath.Join(dir, "go.mod")
	if err := os.WriteFile(gomodPath, []byte(content), 0o644); err != nil {
		t.Fatalf("writing go.mod: %v", err)
	}

	ctx := t.Context()
	deps, updates, err := checkGoMod(ctx, gomodPath, false)
	if err != nil {
		t.Fatalf("checkGoMod: %v", err)
	}

	// Should find 2 pseudo-versioned deps
	if len(deps) != 2 {
		t.Errorf("got %d deps, want 2", len(deps))
	}

	// Should find updates for both (since we used old versions)
	if len(updates) != 2 {
		t.Errorf("got %d updates, want 2", len(updates))
	}

	// Verify updates are for the expected modules
	foundModules := map[string]bool{}
	for _, u := range updates {
		foundModules[u.module] = true

		// The latest version should be newer (different) than current
		if u.current == u.latest {
			t.Errorf("expected update for %s, but current == latest (%s)", u.module, u.current)
		}

		// Latest should also be a pseudo-version
		if !pseudoVersionRe.MatchString(u.latest) {
			t.Errorf("expected latest to be pseudo-version, got %q", u.latest)
		}
	}

	if !foundModules["github.com/maxmind/mmdbwriter"] {
		t.Error("expected update for github.com/maxmind/mmdbwriter")
	}
	if !foundModules["go4.org/netipx"] {
		t.Error("expected update for go4.org/netipx")
	}
}

func TestFindPseudoVersionedDeps(t *testing.T) {
	gomodContent := `module test

go 1.25

require (
	github.com/maxmind/mmdbwriter v1.1.1-0.20251215205057-2f3252140e00
	github.com/oschwald/maxminddb-golang/v2 v2.1.1
	go4.org/netipx v0.0.0-20231129151722-fdeea329fbba
)

require (
	github.com/example/indirect v0.0.0-20231129151722-abcdef123456 // indirect
)
`

	tests := []struct {
		name            string
		includeIndirect bool
		want            map[string]string
	}{
		{
			name:            "exclude indirect",
			includeIndirect: false,
			want: map[string]string{
				"github.com/maxmind/mmdbwriter": "v1.1.1-0.20251215205057-2f3252140e00",
				"go4.org/netipx":                "v0.0.0-20231129151722-fdeea329fbba",
			},
		},
		{
			name:            "include indirect",
			includeIndirect: true,
			want: map[string]string{
				"github.com/maxmind/mmdbwriter": "v1.1.1-0.20251215205057-2f3252140e00",
				"go4.org/netipx":                "v0.0.0-20231129151722-fdeea329fbba",
				"github.com/example/indirect":   "v0.0.0-20231129151722-abcdef123456",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			gomodPath := filepath.Join(dir, "go.mod")
			if err := os.WriteFile(gomodPath, []byte(gomodContent), 0o644); err != nil {
				t.Fatalf("writing go.mod: %v", err)
			}

			deps, err := findPseudoVersionedDeps(gomodPath, tt.includeIndirect)
			if err != nil {
				t.Fatalf("findPseudoVersionedDeps: %v", err)
			}

			if len(deps) != len(tt.want) {
				t.Errorf("got %d deps, want %d", len(deps), len(tt.want))
			}

			for _, dep := range deps {
				expectedVersion, ok := tt.want[dep.module]
				if !ok {
					t.Errorf("unexpected module: %s", dep.module)
					continue
				}
				if dep.version != expectedVersion {
					t.Errorf(
						"module %s: got version %s, want %s",
						dep.module,
						dep.version,
						expectedVersion,
					)
				}
			}
		})
	}
}

func TestPseudoVersionRe(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		matches bool
	}{
		{
			name:    "pseudo-version without base tag",
			input:   "v0.0.0-20231129151722-fdeea329fbba",
			matches: true,
		},
		{
			name:    "pseudo-version with base tag",
			input:   "v1.1.1-0.20251215205057-2f3252140e00",
			matches: true,
		},
		{
			name:    "tagged version",
			input:   "v1.2.3",
			matches: false,
		},
		{
			name:    "tagged version with prerelease",
			input:   "v1.2.3-beta.1",
			matches: false,
		},
		{
			name:    "go.mod line with pseudo-version",
			input:   "	github.com/maxmind/mmdbwriter v1.1.1-0.20251215205057-2f3252140e00",
			matches: true,
		},
		{
			name:    "go.mod line with indirect dependency",
			input:   "	github.com/example/module v0.0.0-20231129151722-fdeea329fbba // indirect",
			matches: true, // regex matches, but findPseudoVersionedDeps filters these out by default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pseudoVersionRe.MatchString(tt.input)
			if got != tt.matches {
				t.Errorf("pseudoVersionRe.MatchString(%q) = %v, want %v", tt.input, got, tt.matches)
			}
		})
	}
}

func TestGetLatestVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tests := []struct {
		name   string
		module string
	}{
		{
			name:   "netipx has no tagged versions",
			module: "go4.org/netipx",
		},
		{
			name:   "mmdbwriter has tagged versions",
			module: "github.com/maxmind/mmdbwriter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()

			version, err := getLatestVersion(ctx, tt.module)
			if err != nil {
				t.Fatalf("getLatestVersion: %v", err)
			}

			if !pseudoVersionRe.MatchString(version) {
				t.Errorf("expected pseudo-version, got %q", version)
			}
		})
	}
}

func TestQueryModuleVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tests := []struct {
		name        string
		module      string
		branch      string
		wantPseudo  bool
		wantErr     bool
		errContains string
	}{
		{
			name:       "netipx has no tagged versions",
			module:     "go4.org/netipx",
			branch:     "main",
			wantPseudo: true,
		},
		{
			name:       "mmdbwriter has tagged versions but main returns pseudo",
			module:     "github.com/maxmind/mmdbwriter",
			branch:     "main",
			wantPseudo: true,
		},
		{
			name:        "nonexistent branch returns error",
			module:      "go4.org/netipx",
			branch:      "nonexistent-branch",
			wantErr:     true,
			errContains: "unknown revision",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()

			version, err := queryModuleVersion(ctx, tt.module, tt.branch)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error to contain %q, got %q", tt.errContains, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("queryModuleVersion: %v", err)
			}

			if tt.wantPseudo && !pseudoVersionRe.MatchString(version) {
				t.Errorf("expected pseudo-version, got %q", version)
			}
		})
	}
}

func TestNewerVersion(t *testing.T) {
	tests := []struct {
		name    string
		a       string
		b       string
		want    string
		wantErr bool
	}{
		{
			name: "a is newer",
			a:    "v0.0.0-20231201000000-aaaaaaaaaaaa",
			b:    "v0.0.0-20231101000000-bbbbbbbbbbbb",
			want: "v0.0.0-20231201000000-aaaaaaaaaaaa",
		},
		{
			name: "b is newer",
			a:    "v0.0.0-20231101000000-aaaaaaaaaaaa",
			b:    "v0.0.0-20231201000000-bbbbbbbbbbbb",
			want: "v0.0.0-20231201000000-bbbbbbbbbbbb",
		},
		{
			name: "same timestamp returns a",
			a:    "v0.0.0-20231201000000-aaaaaaaaaaaa",
			b:    "v0.0.0-20231201000000-bbbbbbbbbbbb",
			want: "v0.0.0-20231201000000-aaaaaaaaaaaa",
		},
		{
			name:    "invalid version a returns error",
			a:       "v1.2.3",
			b:       "v0.0.0-20231201000000-bbbbbbbbbbbb",
			wantErr: true,
		},
		{
			name:    "invalid version b returns error",
			a:       "v0.0.0-20231201000000-aaaaaaaaaaaa",
			b:       "v1.2.3",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newerVersion(tt.a, tt.b)
			if tt.wantErr {
				if err == nil {
					t.Errorf("newerVersion(%q, %q) expected error, got nil", tt.a, tt.b)
				}
				return
			}
			if err != nil {
				t.Errorf("newerVersion(%q, %q) unexpected error: %v", tt.a, tt.b, err)
				return
			}
			if got != tt.want {
				t.Errorf("newerVersion(%q, %q) = %q, want %q", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestExtractTimestamp(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
		wantErr bool
	}{
		{
			name:    "pseudo-version without base tag",
			version: "v0.0.0-20231129151722-fdeea329fbba",
			want:    "20231129151722",
		},
		{
			name:    "pseudo-version with base tag",
			version: "v1.1.1-0.20251215205057-2f3252140e00",
			want:    "20251215205057",
		},
		{
			name:    "tagged version returns error",
			version: "v1.2.3",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractTimestamp(tt.version)
			if tt.wantErr {
				if err == nil {
					t.Errorf("extractTimestamp(%q) expected error, got nil", tt.version)
				}
				return
			}
			if err != nil {
				t.Errorf("extractTimestamp(%q) unexpected error: %v", tt.version, err)
				return
			}
			if got != tt.want {
				t.Errorf("extractTimestamp(%q) = %q, want %q", tt.version, got, tt.want)
			}
		})
	}
}
