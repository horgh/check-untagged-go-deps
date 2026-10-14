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

2. Create an annotated tag for the version, and move the `v1` tag to the same
   commit. Users reference `v1` to get the latest v1.x.x release.

   ```bash
   git checkout main && git pull
   git tag -a v1.2.0 -m v1.2.0
   git tag -f v1
   ```

3. Push only these two tags:

   ```bash
   git push origin v1.2.0
   git push --force origin v1
   ```

4. Put the changelog section for the version in `notes.md`. Then create a GitHub
   release from the tag:

   ```bash
   gh release create v1.2.0 --verify-tag --title v1.2.0 --notes-file notes.md
   ```
