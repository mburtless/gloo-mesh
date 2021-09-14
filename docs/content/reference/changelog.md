
---
title: Gloo Mesh Open Source changelog
weight: 7
description: Changelog entries for Gloo Mesh Open Source
---

Review the changelog for Gloo Mesh Open Source releases. Changelog entries are categorized into the following types:
- **Dependency Bumps**: The version for a dependency in Gloo Mesh Open Source is bumped in this release. Be sure to check for any
**Breaking Change** entries that accompany a dependency bump.
- **Breaking Changes**: An API is changed in a way that is not backwards compatible, such as a changed format for an API field. Occasionally, a breaking change occurs for the process to upgrade Gloo Mesh, in which the changelog entry indicates how to use the new upgrade process.
- **Helm Changes**: The installation Helm chart is changed. If this change is not backwards compatible, an accompanying **Breaking Change** entry is indicated for the release.
- **New Features**: A new feature is implemented in the release.
- **Fixes**: A bug is resolved in this release.

{{% render_changelog enterprise="false" changelogJsonPath="community_changelog.json" %}}
