# Developer Documentation

## Tools

We use [mise](https://mise.jdx.dev/) to install the development tools.
`mise.lock` sets their versions. To install them:

```bash
mise install
```

To lint and format the code:

```bash
precious lint --all
precious tidy --all
```

To update the tools, run `mise lock --bump`. mise only selects versions that are
at least one week old.

## Releasing

1. In a PR, add a section for the new version to `CHANGELOG.md`, with the
   version number and date. Merge the PR.

2. Build and test the Docker image. CI does not build it, and users of `v1` get
   the new image when the `v1` tag moves.

   ```bash
   git checkout main && git pull
   podman build -t check-untagged-go-deps:test .
   podman run --rm -v /path/to/go/module:/github/workspace:Z \
     check-untagged-go-deps:test
   ```

   Use a Go module that has a pseudo-versioned dependency. The command must list
   the dependency and exit with code 1 when an update is available.

3. Create an annotated tag for the version, and move the `v1` tag to the same
   commit. Users reference `v1` to get the latest v1.x.x release.

   ```bash
   git tag -a v1.2.0 -m v1.2.0
   git tag -f v1
   ```

4. Push only these two tags:

   ```bash
   git push origin v1.2.0
   git push --force origin v1
   ```

5. Put the changelog section for the version in `notes.md`. Then create a GitHub
   release from the tag:

   ```bash
   gh release create v1.2.0 --verify-tag --title v1.2.0 --notes-file notes.md
   ```
