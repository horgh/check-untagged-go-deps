# check-untagged-go-deps

A GitHub Action that checks for updates to pseudo-versioned (commit-pinned) Go
dependencies.

## Why?

Dependabot does not support updating Go dependencies pinned to commits
(pseudo-versions like `v0.0.0-20231129151722-fdeea329fbba`). See
[dependabot-core#2028](https://github.com/dependabot/dependabot-core/issues/2028).

This action fills that gap by checking if newer commits are available on the
default branch (`main` or `master`) for any pseudo-versioned dependencies in
your `go.mod`.

## Usage

```yaml
name: Check Untagged Dependencies

on:
  push:
  pull_request:
  schedule:
    - cron: '0 14 * * *' # Daily at 14:00 UTC
  workflow_dispatch:

permissions: {}

jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: horgh/check-untagged-go-deps@v1
```

## How it works

1. Parses `go.mod` to find dependencies with pseudo-versions (versions ending
   with `YYYYMMDDHHMMSS-<commit hash>`)
2. For each pseudo-versioned dependency, queries `@main` (falling back to
   `@master`) to get the latest commit version
3. Compares the current version with the latest and reports any available
   updates
4. Exits with code 1 if updates are found, alerting you to update manually

Only direct dependencies are checked. Indirect dependencies (lines ending with
`// indirect`) are not checked.

## Example output

When updates are available:

```
Pseudo-versioned dependencies in go.mod:
  github.com/example/module
  github.com/another/module

Updates available:
  github.com/example/module: v0.0.0-20231101000000-abc123abc123 -> v0.0.0-20231201000000-def456def456
```

When no updates are available:

```
Pseudo-versioned dependencies in go.mod:
  github.com/example/module

No updates found for pseudo-versioned dependencies.
```

## License

MIT (http://opensource.org/licenses/MIT)
