# Developer Documentation

## Releasing

1. Update `CHANGELOG.md` with the changes for the new version. Change the
   "Unreleased" header to include the version number and date.

2. Commit the changelog update.

3. Create an annotated git tag for the version:

   ```bash
   git tag -a v1.2.0 -m v1.2.0
   ```

4. Update the `v1` tag to point to the new release (this allows users to
   reference `v1` in their workflows to get the latest v1.x.x release):

   ```bash
   git tag -f v1
   ```

5. Push the commits and tags:

   ```bash
   git push && git push --tags --force
   ```

6. Create a GitHub release from the tag:
   - Go to the repository on GitHub
   - Click "Releases" in the right sidebar
   - Click "Draft a new release"
   - Select the new version tag (e.g., `v1.2.0`) from the dropdown
   - Set the release title to the version (e.g., `v1.2.0`)
   - Paste the relevant changes from `CHANGELOG.md` into the description
   - Click "Publish release"
