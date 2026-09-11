## ADDED Requirements

### Requirement: Snapshot-to-GA version promotion

When the Elastic Stack release tracked by the acceptance matrix's snapshot-labeled entry
(`<version>-SNAPSHOT`) reaches general availability, the workflow SHALL replace that matrix entry
with the released version string rather than adding a separate, additional matrix entry for the same
stack line. The promoted entry SHALL be added to every per-version step condition (such as Fleet
setup) that had matched the snapshot label only via the `-SNAPSHOT` suffix, so the promoted version
does not lose step coverage it received while labeled as a snapshot. The promoted entry SHALL no
longer match `endsWith(matrix.version, '-SNAPSHOT')` and SHALL therefore be treated as blocking
(`continue-on-error: false`) like every other non-snapshot matrix entry, and SHALL NOT trigger the
snapshot-failure PR warning comment.

#### Scenario: Snapshot entry is promoted to its GA release

- **GIVEN** the acceptance matrix contains a snapshot-labeled entry `X.Y.0-SNAPSHOT` tracking an
  in-development stack line
- **AND** that stack line reaches general availability as `X.Y.0`
- **WHEN** the matrix is updated for the release
- **THEN** the `X.Y.0-SNAPSHOT` entry SHALL be rewritten to `X.Y.0`
- **AND** no additional matrix entry SHALL be added for the same `X.Y` stack line

#### Scenario: Promoted entry keeps per-version step coverage

- **GIVEN** a per-version step condition that previously matched a snapshot entry only via
  `endsWith(matrix.version, '-SNAPSHOT')` (for example, Fleet setup)
- **WHEN** that snapshot entry is promoted to its GA version string
- **THEN** the promoted version string SHALL be added explicitly to that step's condition
- **AND** the step SHALL continue to run for the promoted version exactly as it did while the entry
  was labeled as a snapshot

#### Scenario: Promoted entry becomes blocking

- **GIVEN** a matrix entry that was promoted from a snapshot label to its GA version string
- **WHEN** the acceptance test step (`make testacc`) fails for that entry
- **THEN** `continue-on-error` SHALL NOT apply to that failure
- **AND** the snapshot-failure PR warning comment step SHALL NOT fire for that entry
