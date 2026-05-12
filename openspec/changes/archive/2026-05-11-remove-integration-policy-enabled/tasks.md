## 1. Drop the model field and API mapping

- [x] 1.1 Remove `Enabled types.Bool` from `integrationPolicyModel` in
      `internal/fleet/integration_policy/models.go`.
- [x] 1.2 Remove `model.Enabled = types.BoolValue(data.Enabled)` from
      `populateFromAPI` in `internal/fleet/integration_policy/models.go`.

## 2. Bump live schema to v3 and remove the attribute

- [x] 2.1 In `internal/fleet/integration_policy/schema.go`, rename `getSchemaV2`
      → `getSchemaV3`, set `Version: 3`, and remove the `enabled` attribute.
- [x] 2.2 Update the `Schema(...)` method to call `getSchemaV3()`.

## 3. Add v2 → v3 state upgrader

- [x] 3.1 Create `internal/fleet/integration_policy/schema_v2.go` containing:
      - `integrationPolicyModelV2` (snapshot of the prior v2 model with
        `Enabled types.Bool`).
      - `getSchemaV2()` returning the prior v2 schema (with the `enabled`
        attribute and `Version: 2`).
      - `upgradeV2ToV3` that decodes prior state via the v2 schema and writes
        the same fields (minus `enabled`) into the live model.
- [x] 3.2 Register `2: {PriorSchema: getSchemaV2(), StateUpgrader: upgradeV2ToV3}`
      in `(*integrationPolicyResource).UpgradeState` in
      `internal/fleet/integration_policy/resource.go`.

## 4. Retarget v0 / v1 chains to v3

- [x] 4.1 In `internal/fleet/integration_policy/schema_v1.go`, rename `toV2` →
      `toV3` and `upgradeV1ToV2` → `upgradeV1ToV3`. Drop the
      `Enabled: m.Enabled` field assignment.
- [x] 4.2 In `internal/fleet/integration_policy/schema_v0.go`, rename
      `upgradeV0ToV2` → `upgradeV0ToV3` and update its body to call the renamed
      `toV3`.
- [x] 4.3 Update the v0 and v1 entries in
      `(*integrationPolicyResource).UpgradeState` to point at the renamed
      upgraders.

## 5. Acceptance and unit tests

- [x] 5.1 Delete `TestAccResourceIntegrationPolicyEnabled` from
      `internal/fleet/integration_policy/acc_test.go` and remove the
      `internal/fleet/integration_policy/testdata/TestAccResourceIntegrationPolicyEnabled/`
      directory.
- [x] 5.2 Drop top-level `enabled` `TestCheckResourceAttr` calls from
      `TestAccResourceIntegrationPolicy` in `acc_test.go` (per-input and
      per-stream `enabled` checks remain).
- [x] 5.3 Update `internal/fleet/integration_policy/schema_v1_test.go`:
      - Rename `v1Model.toV2(ctx)` calls to `v1Model.toV3(ctx)`.
      - Remove `assert.Equal(t, v1Model.Enabled, v2Model.Enabled)`.
- [x] 5.4 Add `internal/fleet/integration_policy/schema_v2_test.go` with a
      table-driven test that runs `upgradeV2ToV3` against representative prior
      state and asserts: (a) `enabled` is not present on the resulting model,
      (b) all other fields are carried through unchanged.

## 6. Specs and docs

- [x] 6.1 Update `openspec/specs/fleet-integration-policy/spec.md`:
      - Remove the `enabled` line from the Schema HCL block.
      - Remove `enabled` from REQ-022's mapped-fields list.
      - Update REQ-024 / REQ-025 wording to target schema version 3 and remove
        `enabled` from the carried-over field list.
      - Add a new requirement (REQ-026) covering the v2 → v3 upgrade dropping
        the `enabled` attribute.
- [x] 6.2 Add a `## [Unreleased]` ▸ breaking-change subsection to `CHANGELOG.md`
      describing the removal and the automatic v2 → v3 state upgrade.
- [x] 6.3 Regenerate provider docs: `make docs-generate` (updates
      `docs/resources/fleet_integration_policy.md`).

## 7. Validation

- [x] 7.1 `make build`.
- [x] 7.2 `go test ./internal/fleet/integration_policy/...` (unit tests).
- [x] 7.3 `make check-lint`.
- [x] 7.4 If a Stack is reachable, run
      `TF_ACC=1 go test -v -run '^TestAccResourceIntegrationPolicy$' ./internal/fleet/integration_policy/... -timeout 30m`
      to confirm the existing acceptance suite still passes (and that the now
      deleted enabled test is gone).
