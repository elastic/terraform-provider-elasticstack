## 1. Version constant and documentation

- [x] 1.1 Add an exported minimum-version constant for Synthetics private location `space_id` (e.g. `MinVersionSpaceID` = `9.4.0-SNAPSHOT`) in `internal/kibana/synthetics/privatelocation/`, aligned with `internal/kibana/streams` prerelease naming.
- [x] 1.2 Extend the embedded `descriptions/space_id.md` (and/or schema `MarkdownDescription`) so practitioners see that non-default `space_id` requires Elastic Stack 9.4.0-SNAPSHOT or higher.

## 2. Runtime enforcement

- [x] 2.1 Add a small helper that decides whether the version gate applies (effective non-default space from `space_id` and composite `id`; normalize literal `default` space segment to default space if applicable).
- [x] 2.2 In `Create`, `Read`, and `Delete`, call `r.client.EnforceMinVersion(ctx, MinVersionSpaceID)` when the gate applies; on failure, append diagnostics consistent with Fleet/Streams (`Unsupported server version` / minimum version message).
- [x] 2.3 Ensure `diagutil` or existing patterns are used if the project standardizes SDK diag conversion (match neighboring resources).

## 3. Acceptance tests

- [x] 3.1 Update `TestSyntheticPrivateLocationResource_nonDefaultSpace` to use `SkipFunc: versionutils.CheckIfVersionIsUnsupported(...)` with the same version as `MinVersionSpaceID` (import the exported constant from `privatelocation` to avoid drift).
- [x] 3.2 Remove or replace the comment that refers only to Fleet 9.1+ where it is no longer the binding constraint for this test.

## 4. Verification

- [x] 4.1 Run `make build` and targeted acceptance tests when a 9.4+ stack is available; confirm older stacks skip `TestSyntheticPrivateLocationResource_nonDefaultSpace`.
- [x] 4.2 Run `make check-openspec` (or `node_modules/.bin/openspec validate` as required by the repo) so the change validates structurally.
