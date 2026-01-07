# Changelog

## 1.1.0 (Unreleased)

* Add `-i` flag to include indirect dependencies (those marked with
  `// indirect` in go.mod).
* Use `golang.org/x/mod/modfile` to parse go.mod instead of manual
  line-by-line parsing.
* Use `golang.org/x/mod/module` functions (`IsPseudoVersion`,
  `PseudoVersionTime`) instead of custom regex patterns.

## 1.0.1 (2025-12-31)

* Fix Docker image to include git, which is required for `go list` to query
  module versions.

## 1.0.0 (2025-12-31)

* Initial release.
