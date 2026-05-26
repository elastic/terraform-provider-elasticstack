## ADDED Requirements

### Requirement: Schema — `secrets_wo` write-only attribute (REQ-WO-001)

The `elasticstack_kibana_action_connector` resource SHALL expose an optional `secrets_wo` string attribute that is:

- `Optional: true`
- `Sensitive: true`
- `WriteOnly: true` (accepted by the Terraform Plugin Framework; never persisted to state)

`secrets_wo` SHALL accept the same JSON string content as the existing `secrets` attribute. The provider SHALL apply the same JSON validation and normalization behavior to `secrets_wo` as it does to `secrets`; any JSON string accepted, rejected, or normalized for `secrets` SHALL be accepted, rejected, or normalized identically for `secrets_wo`. It is intended for practitioners who source connector secrets from ephemeral providers (e.g. Vault) and do not want the secret value to appear in the Terraform state file.

`secrets_wo` SHALL be mutually exclusive with `secrets`. Setting both simultaneously SHALL be invalid and the provider SHALL enforce this via a `ConflictsWith` validator.

#### Scenario: `secrets_wo` is accepted with an ephemeral value

- GIVEN a `secrets_wo` attribute set to a JSON string sourced from an ephemeral resource
- WHEN Terraform applies the configuration
- THEN the provider SHALL send the JSON value as the connector secrets to the Kibana API
- AND the Terraform state after apply SHALL NOT contain the `secrets_wo` value (it SHALL be null or absent in state)

#### Scenario: `secrets_wo` and `secrets` cannot both be set

- GIVEN a configuration that sets both `secrets` and `secrets_wo` to non-null values
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation diagnostic describing the mutual-exclusion constraint

---

### Requirement: Schema — `secrets_wo_version` rotation trigger (REQ-WO-002)

The `elasticstack_kibana_action_connector` resource SHALL expose an optional `secrets_wo_version` string attribute. It is persisted to state (it is not write-only) and SHALL be used by the practitioner to signal that the secret has rotated, triggering a re-send of `secrets_wo` on the next apply even when other resource attributes are unchanged.

`secrets_wo_version` SHALL require `secrets_wo` to be set (enforced via `AlsoRequires` validator). Setting `secrets_wo_version` without `secrets_wo` SHALL be invalid.

`secrets_wo` MAY be set without `secrets_wo_version`; the version attribute is optional.

#### Scenario: `secrets_wo_version` persists in state

- GIVEN a configuration with `secrets_wo = <json>` and `secrets_wo_version = "1"`
- WHEN Terraform applies the configuration
- THEN state SHALL contain `secrets_wo_version = "1"`
- AND state SHALL NOT contain `secrets_wo` (null/unknown by write-only contract)

#### Scenario: Bumping `secrets_wo_version` triggers a secret update

- GIVEN an existing connector resource with `secrets_wo_version = "1"` in state
- WHEN the configuration changes `secrets_wo_version` to `"2"`
- THEN Terraform SHALL plan an update for the connector resource
- AND the provider SHALL send the new `secrets_wo` value to the Kibana PUT API

#### Scenario: `secrets_wo_version` without `secrets_wo` is invalid

- GIVEN a configuration with `secrets_wo_version = "1"` but no `secrets_wo`
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation diagnostic

---

### Requirement: Schema — `secrets` advisory validator (REQ-WO-003)

The existing `secrets` attribute SHALL gain a `PreferWriteOnlyAttribute` validator pointing at `secrets_wo`. This causes Terraform to emit an advisory warning when `secrets` is used instead of `secrets_wo`, guiding practitioners toward the write-only path without breaking existing configurations.

The `secrets` attribute SHALL also gain a `ConflictsWith(secrets_wo)` validator (complementary to the one on `secrets_wo`).

#### Scenario: `secrets` alone remains valid

- GIVEN a configuration with `secrets` set and `secrets_wo` absent
- WHEN Terraform validates the configuration
- THEN the provider SHALL accept the configuration (no error)
- AND the provider MAY emit an advisory warning suggesting `secrets_wo`

---

### Requirement: Create path — write-only secrets sourced from config (REQ-WO-004)

During `Create`, the provider SHALL read `secrets_wo` from `request.Config` (not from `request.Plan`, where write-only values are always null). If `secrets_wo` is known (non-null, non-unknown) in config, the provider SHALL use it as the connector `secrets` payload sent to the Kibana POST API. If `secrets_wo` is not set, the provider SHALL fall back to the `secrets` value from the plan.

#### Scenario: Create with `secrets_wo`

- GIVEN a configuration with `secrets_wo` set to a valid JSON string
- WHEN the provider creates the connector
- THEN the Kibana POST request body SHALL include the JSON value as the `secrets` field
- AND the apply SHALL succeed

#### Scenario: Create with `secrets` (existing path, no regression)

- GIVEN a configuration with `secrets` set and `secrets_wo` absent
- WHEN the provider creates the connector
- THEN the Kibana POST request body SHALL include the `secrets` value
- AND the apply SHALL succeed

---

### Requirement: Update path — write-only secrets re-sent from config (REQ-WO-005)

During `Update`, the provider SHALL read `secrets_wo` from `request.Config`. If `secrets_wo` is known in config, the provider SHALL include it in the Kibana PUT request body as the `secrets` field on every update, regardless of whether `secrets_wo_version` changed. This is required because the provider cannot read `secrets_wo` back from state, and the Kibana API behavior when `secrets` is omitted from an update request has not been confirmed.

Once the Kibana omit-secrets behavior is confirmed (see design.md open question), this requirement MAY be relaxed to skip re-sending `secrets_wo` when no rotation is indicated.

#### Scenario: Update with `secrets_wo` always re-sends the secret

- GIVEN an existing connector resource with `secrets_wo` set in config
- WHEN any attribute of the connector is updated (e.g. `name` changes)
- THEN the Kibana PUT request body SHALL include the `secrets_wo` value as `secrets`
- AND the ephemeral source MUST be available at apply time

#### Scenario: Update with `secrets` (existing path, no regression)

- GIVEN an existing connector resource with `secrets` set in config and `secrets_wo` absent
- WHEN the connector is updated
- THEN the Kibana PUT request body SHALL include the `secrets` value
- AND the apply SHALL succeed

---

### Requirement: Read path — `secrets_wo` absent from state (REQ-WO-006)

During `Read`, the provider SHALL NOT attempt to populate `secrets_wo` from the Kibana API response (the API never returns connector secrets). The framework write-only contract ensures `secrets_wo` remains null in state. `secrets_wo_version` SHALL be populated from state as normal (it is a regular string attribute and persists across reads).

#### Scenario: Read does not expose `secrets_wo`

- GIVEN an existing connector resource with `secrets_wo` configured
- WHEN Terraform refreshes state (terraform refresh or plan)
- THEN `secrets_wo` in state SHALL be null
- AND `secrets_wo_version` in state SHALL retain its previously written value

---

### Requirement: Acceptance test coverage (REQ-WO-007)

Acceptance tests for `elasticstack_kibana_action_connector` SHALL include coverage of the write-only secrets path, including create with `secrets_wo`, update via version bump, and absence of the secret from state.

#### Scenario: Acceptance test — create with `secrets_wo` and confirm absence from state

- GIVEN an acceptance test that configures a connector with `secrets_wo` set to a JSON secrets string and `secrets` absent
- WHEN the test applies the configuration
- THEN the test SHALL assert that `secrets_wo` is null or absent in the resulting Terraform state
- AND the test SHALL assert that the connector was created successfully in Kibana

#### Scenario: Acceptance test — rotation via version bump triggers update

- GIVEN an acceptance test that applies a connector with `secrets_wo_version = "1"`
- WHEN the test updates the configuration to `secrets_wo_version = "2"`
- THEN the test SHALL confirm that the apply succeeds (connector update is accepted by Kibana)

#### Scenario: Acceptance test — existing `secrets` path has no regression

- GIVEN an acceptance test that configures a connector with the existing `secrets` attribute
- WHEN the test applies the configuration
- THEN the apply SHALL succeed and behavior SHALL be identical to pre-change behavior
