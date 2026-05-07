## REMOVED Requirements

### Requirement: Generated changelog selector

**Reason**: The `generated-changelog` category is removed. Changelog-only PRs are now handled by the workflow's classify job, which skips provider CI and lets the gate succeed without special-casing the auto-approve script. The auto-approve script no longer needs to recognise this PR shape.

**Migration**: Remove the `generated-changelog` category selector, its commit-author gate, and its file-allowlist gate from `scripts/auto-approve/`. Remove all unit tests specific to this category.

### Requirement: Generated changelog commit authors

**Reason**: Removed with the `generated-changelog` category (see above).

**Migration**: See above.

### Requirement: Generated changelog file allowlist

**Reason**: Removed with the `generated-changelog` category (see above).

**Migration**: See above.
