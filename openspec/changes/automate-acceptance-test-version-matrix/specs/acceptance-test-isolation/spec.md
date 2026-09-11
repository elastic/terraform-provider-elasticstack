## MODIFIED Requirements

### Requirement: Elastic Stack 9.4.2 included in acceptance-test CI matrix (REQ-ACC-002)

The stack-version list pinned at `.github/versions/acceptance-test-matrix.json` (consumed by `.github/workflows/provider.yml`'s acceptance-test matrix) SHALL include exactly one `9.4.x` entry — the latest published `9.4` patch — immediately before the newest snapshot-labeled entry once the list is sorted ascending, and SHALL NOT include more than one `9.4.x` patch at a time. This requirement is expressed as a version-range invariant, not a pin to a specific patch string, so that an automated patch bump for the `9.4` line (for example, `9.4.2` → `9.4.6`) satisfies this requirement without a spec change.

#### Scenario: Entity-store tests pass in CI against 9.4.2

- GIVEN `"9.4.2"` was the pinned `9.4.x` entry and per-test space isolation applied to all five packages
- WHEN the acceptance-test jobs ran for the `9.4.2` matrix entry
- THEN all entity-store family tests passed
- AND the issue #3952 closure gate ("tests run successfully against that version") was satisfied

#### Scenario: 9.4.2 replaces 9.4.0

- GIVEN the pinned versions artifact historically included `"9.4.0"` followed by `"9.4.2"` and a `9.5.x`-line snapshot entry
- WHEN the redundant `9.4.x` entry was removed and CI ran the full matrix
- THEN `"9.4.2"` remained immediately before the newest snapshot entry
- AND `"9.4.0"` was no longer included

#### Scenario: Automated patch bump keeps exactly one 9.4.x entry

- GIVEN the pinned versions artifact contains a `9.4.x` entry followed by a newer snapshot-labeled entry
- WHEN the version-matrix generator (`ci-version-matrix-generation` capability) updates the `9.4.x` entry to a newer published patch of the same minor (for example, `9.4.2` → `9.4.6`)
- THEN the pinned versions artifact SHALL still contain exactly one `9.4.x` entry, immediately before the newest snapshot-labeled entry
- AND all other version entries SHALL continue to behave as before

#### Scenario: A stale 9.4.x patch and the current 9.4.x patch cannot both be present

- GIVEN the pinned versions artifact is inspected at any point in time
- WHEN more than one `9.4.x` version string is present
- THEN this requirement is violated, regardless of which specific patch numbers are involved
