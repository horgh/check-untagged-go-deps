# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A GitHub Action that checks for updates to pseudo-versioned (commit-pinned) Go dependencies. Dependabot doesn't support updating Go dependencies pinned to commits (pseudo-versions like `v0.0.0-20231129151722-fdeea329fbba`), so this tool fills that gap.

## Commands

Build:
```bash
go build
```

Run tests:
```bash
go test ./...
```

Run a single test:
```bash
go test -run TestFindPseudoVersionedDeps
```

Skip integration tests (those requiring network access):
```bash
go test -short ./...
```

## Architecture

This is a single-file Go CLI tool (`main.go`) with no external dependencies. The flow is:

1. `findPseudoVersionedDeps` - Parses go.mod to find dependencies with pseudo-versions using regex matching
2. `checkForUpdates` - For each dependency, queries the Go module proxy via `go list -m -json module@branch`
3. `getLatestVersion` - Queries both `@main` and `@master` branches, returns the version with the newer timestamp

The tool uses `go list` to query module versions, which requires git to be available (see Dockerfile).

## Key Details

- Exit code 1 when updates are found (to fail CI pipelines)
- The `-i` flag includes indirect dependencies (excluded by default)
- Integration tests (those hitting the network) are skipped with `-short`
