// check-untagged-go-deps checks for updates to pseudo-versioned (commit-pinned)
// Go dependencies.
//
// Dependabot does not support updating Go dependencies pinned to commits
// (pseudo-versions like v0.0.0-20231129151722-fdeea329fbba).
// See: https://github.com/dependabot/dependabot-core/issues/2028
//
// This tool checks for updates by comparing the current version in go.mod
// with the latest commit on the default branch (@main or @master).
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/mod/modfile"
)

func main() {
	includeIndirect := flag.Bool("i", false, "include indirect dependencies")
	flag.Parse()

	gomodPath := "go.mod"
	if flag.NArg() > 0 {
		gomodPath = flag.Arg(0)
	}

	updatesFound, err := run(gomodPath, *includeIndirect)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if updatesFound {
		os.Exit(1)
	}
}

func run(gomodPath string, includeIndirect bool) (bool, error) {
	deps, updates, err := checkGoMod(context.Background(), gomodPath, includeIndirect)
	if err != nil {
		return false, err
	}

	if len(deps) == 0 {
		fmt.Println("No pseudo-versioned dependencies found in go.mod.")
		return false, nil
	}

	fmt.Println("Pseudo-versioned dependencies in go.mod:")
	for _, dep := range deps {
		fmt.Printf("  %s\n", dep.module)
	}
	fmt.Println()

	if len(updates) > 0 {
		fmt.Println("Updates available:")
		for _, u := range updates {
			fmt.Printf("  %s: %s -> %s\n", u.module, u.current, u.latest)
		}
		return true, nil
	}

	fmt.Println("No updates found for pseudo-versioned dependencies.")
	return false, nil
}

// checkGoMod finds pseudo-versioned dependencies in the given go.mod file and
// checks if updates are available for them.
func checkGoMod(
	ctx context.Context,
	gomodPath string,
	includeIndirect bool,
) ([]dependency, []update, error) {
	deps, err := findPseudoVersionedDeps(gomodPath, includeIndirect)
	if err != nil {
		return nil, nil, fmt.Errorf("reading %s: %w", gomodPath, err)
	}

	if len(deps) == 0 {
		return nil, nil, nil
	}

	updates, err := checkForUpdates(ctx, deps)
	if err != nil {
		return nil, nil, err
	}

	return deps, updates, nil
}

// dependency represents a pseudo-versioned dependency found in go.mod.
type dependency struct {
	module  string
	version string
}

// pseudoVersionRe matches Go pseudo-versions which end with a timestamp and
// commit hash, e.g.:
//   - v0.0.0-20231129151722-fdeea329fbba (no base tag)
//   - v1.1.1-0.20251215205057-2f3252140e00 (based on existing tag)
var pseudoVersionRe = regexp.MustCompile(`[0-9]{14}-[a-f0-9]{12}$`)

func findPseudoVersionedDeps(gomodPath string, includeIndirect bool) ([]dependency, error) {
	data, err := os.ReadFile(filepath.Clean(gomodPath))
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	f, err := modfile.Parse(gomodPath, data, nil)
	if err != nil {
		return nil, fmt.Errorf("parsing go.mod: %w", err)
	}

	var deps []dependency
	for _, req := range f.Require {
		if !pseudoVersionRe.MatchString(req.Mod.Version) {
			continue
		}
		if req.Indirect && !includeIndirect {
			continue
		}
		deps = append(deps, dependency{
			module:  req.Mod.Path,
			version: req.Mod.Version,
		})
	}

	return deps, nil
}

// update represents an available update for a dependency.
type update struct {
	module  string
	current string
	latest  string
}

func checkForUpdates(ctx context.Context, deps []dependency) ([]update, error) {
	var updates []update

	for _, dep := range deps {
		latest, err := getLatestVersion(ctx, dep.module)
		if err != nil {
			return nil, fmt.Errorf("checking %s: %w", dep.module, err)
		}

		if dep.version != latest {
			updates = append(updates, update{
				module:  dep.module,
				current: dep.version,
				latest:  latest,
			})
		}
	}

	return updates, nil
}

// getLatestVersion queries the Go module proxy for the latest version on the
// default branch. It queries both @main and @master and returns the one with
// the more recent timestamp (in case both exist).
func getLatestVersion(ctx context.Context, modulePath string) (string, error) {
	branches := []string{"main", "master"}

	var versions []string
	for _, branch := range branches {
		version, err := queryModuleVersion(ctx, modulePath, branch)
		if err != nil {
			// "unknown revision" means the branch doesn't exist, try next
			if strings.Contains(err.Error(), "unknown revision") {
				continue
			}
			return "", err
		}
		versions = append(versions, version)
	}

	if len(versions) == 0 {
		return "", errors.New("neither main nor master branch found")
	}

	// If we have both, return the one with the newer timestamp
	if len(versions) == 2 {
		return newerVersion(versions[0], versions[1])
	}

	return versions[0], nil
}

// moduleInfo represents the JSON output from 'go list -m -json'.
type moduleInfo struct {
	Path    string `json:"Path"`    //nolint:tagliatelle // matches go list output
	Version string `json:"Version"` //nolint:tagliatelle // matches go list output
}

// Note there are at least two cases to consider: If the repo has tagged
// versions and you're depending on a commit, then `go get -u ./...` won't
// update it even if you're on a main commit that is behind main. However if
// the repo does not have tagged versions, it will. This is mostly a
// consideration for `go get -u` but I wanted to note it somewhere.
func queryModuleVersion(
	ctx context.Context,
	modulePath,
	branch string,
) (string, error) {
	//nolint:gosec // modulePath and branch are from go.mod, intentional
	cmd := exec.CommandContext(ctx, "go", "list", "-m", "-json", modulePath+"@"+branch)
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", errors.New(strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", fmt.Errorf("running go list: %w", err)
	}

	var info moduleInfo
	if err := json.Unmarshal(output, &info); err != nil {
		return "", fmt.Errorf("parsing module info: %w", err)
	}

	return info.Version, nil
}

// newerVersion compares two pseudo-versions and returns the one with the more
// recent timestamp. Pseudo-versions contain a timestamp in YYYYMMDDHHMMSS format.
func newerVersion(a, b string) (string, error) {
	tsA, err := extractTimestamp(a)
	if err != nil {
		return "", err
	}
	tsB, err := extractTimestamp(b)
	if err != nil {
		return "", err
	}
	if tsA >= tsB {
		return a, nil
	}
	return b, nil
}

// timestampRe extracts the 14-digit timestamp from a pseudo-version.
var timestampRe = regexp.MustCompile(`[0-9]{14}`)

func extractTimestamp(version string) (string, error) {
	match := timestampRe.FindString(version)
	if match == "" {
		return "", fmt.Errorf("no timestamp found in version %q", version)
	}
	return match, nil
}
