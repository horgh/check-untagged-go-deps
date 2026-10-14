# Changelog

## 1.2.0 (2026-10-14)

- Update Go in the Docker image from 1.25 to 1.27. Before, the action failed for
  modules that require Go 1.26 or later, because the image does not download a
  newer Go toolchain.

## 1.1.0 (2026-01-06)

- Add `-i` flag to include indirect dependencies (those marked with
  `// indirect` in go.mod).
- Use `golang.org/x/mod/modfile` to parse go.mod instead of manual line-by-line
  parsing.
- Use `golang.org/x/mod/module` functions (`IsPseudoVersion`,
  `PseudoVersionTime`) instead of custom regex patterns.

## 1.0.1 (2025-12-31)

- Fix Docker image to include git, which is required for `go list` to query
  module versions.

## 1.0.0 (2025-12-31)

- Initial release.
