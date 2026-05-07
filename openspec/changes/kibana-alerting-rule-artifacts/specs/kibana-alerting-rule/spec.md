## ADDED Requirements

### Requirement: Schema — `artifacts` block (REQ-045)

The `elasticstack_kibana_alerting_rule` resource SHALL expose an optional **`artifacts`** single nested block at the rule level. When configured, it MAY contain:

- A **`dashboards`** list nested block (zero or more entries, each with a required `id` string attribute).
- An **`investigation_guide`** single nested block with optional `content` (string), optional `content_path` (string), and computed `checksum` (string).

The `artifacts` block SHALL be entirely optional. When absent from configuration, the provider SHALL treat it as unconfigured and SHALL omit the `artifacts` key from create and update request bodies (allowing Kibana to preserve any previously stored artifact values).

#### Scenario: Rule with dashboards

- GIVEN an `artifacts` block with one or more `dashboards` entries each specifying an `id`
- WHEN Terraform validates configuration
- THEN the provider SHALL accept the configuration and store each `id` in state

#### Scenario: Rule without artifacts

- GIVEN no `artifacts` block in configuration
- WHEN the provider executes create or update
- THEN the request body SHALL NOT include an `artifacts` key

### Requirement: Validation — `investigation_guide` mutual exclusion (REQ-046)

When the practitioner configures an `investigation_guide` block, **exactly one** of `content` or `content_path` MUST be set. Setting both simultaneously, or neither, SHALL be invalid. The provider SHALL enforce this at plan/validate time.

`checksum` is computed by the provider and SHALL NOT be set by the practitioner; it is only meaningful when `content_path` is used.

#### Scenario: Both content and content_path set

- GIVEN an `investigation_guide` block with both `content` and `content_path` set to non-null values
- WHEN Terraform validates configuration
- THEN the provider SHALL return a validation diagnostic describing the mutual exclusion constraint

#### Scenario: Neither content nor content_path set

- GIVEN an `investigation_guide` block with both `content` and `content_path` null or absent
- WHEN Terraform validates configuration
- THEN the provider SHALL return a validation diagnostic

#### Scenario: Only content set

- GIVEN an `investigation_guide` block with `content` set to a non-empty string and `content_path` absent
- WHEN Terraform validates configuration
- THEN the provider SHALL accept the configuration

#### Scenario: Only content_path set

- GIVEN an `investigation_guide` block with `content_path` set to a non-empty string and `content` absent
- WHEN Terraform validates configuration
- THEN the provider SHALL accept the configuration

### Requirement: Write path — `artifacts` on create and update (REQ-047)

When the practitioner configures `artifacts`, the create and update request bodies sent to Kibana SHALL include an `artifacts` JSON object whose structure mirrors the configured values:

- `artifacts.dashboards`: JSON array of `{"id": "<id>"}` objects, one per configured `dashboards` entry.
- `artifacts.investigation_guide.blob`: the investigation guide content. When `content` is configured, `blob` SHALL equal the `content` string. When `content_path` is configured, the provider SHALL read the file at that path and send its content as `blob`.

When the practitioner does **not** configure `artifacts`, the provider SHALL **omit** the `artifacts` key from the update request body so Kibana does not alter existing rule-level artifact state for that field.

#### Scenario: Create with dashboards and investigation guide (content)

- GIVEN a configured `artifacts` block with one dashboard id and `investigation_guide.content` set
- WHEN create runs
- THEN the create request body SHALL include `artifacts.dashboards[0].id` and `artifacts.investigation_guide.blob` equal to the configured content

#### Scenario: Update omits artifacts when unset

- GIVEN an update where `artifacts` is not configured in Terraform
- WHEN update runs
- THEN the update request body SHALL NOT include an `artifacts` key

#### Scenario: Create with content_path sends file content as blob

- GIVEN a configured `investigation_guide` with `content_path` pointing to a readable file
- WHEN create runs
- THEN the create request body SHALL include `artifacts.investigation_guide.blob` equal to the file's contents

### Requirement: Read path — `artifacts` state mapping (REQ-048)

After a successful create or update (and on refresh reads), if the Kibana API response includes `artifacts`, the provider SHALL populate `artifacts` in state as follows:

- `artifacts.dashboards`: each element's `id` reflects the corresponding API dashboard id.
- `artifacts.investigation_guide.content`: if prior state used `content` (i.e., `content_path` was null in prior state), the provider SHALL set `content` in state from the API-returned `blob`. If prior state used `content_path` (i.e., `content` was null), the provider SHALL leave `content` null and SHALL NOT overwrite `content_path` from the API response.
- `artifacts.investigation_guide.checksum`: NOT updated from the API on read; it is managed exclusively by the plan modifier (see REQ-049).

If the API response omits `artifacts` and the prior state value was null, the provider SHALL set `artifacts` to null. If the API omits `artifacts` and the prior state had a known non-null value, the provider SHALL keep the prior known value (consistent with the preserve-on-partial-response pattern used for `scheduled_task_id` and `alert_delay`).

#### Scenario: Read maps blob to content when prior state used content

- GIVEN prior state has `investigation_guide.content` set and `investigation_guide.content_path` null
- WHEN the provider reads the rule from Kibana
- THEN `investigation_guide.content` SHALL be set from the API `blob` value

#### Scenario: Read preserves content_path when prior state used content_path

- GIVEN prior state has `investigation_guide.content_path` set and `investigation_guide.content` null
- WHEN the provider reads the rule from Kibana
- THEN `investigation_guide.content_path` SHALL remain unchanged and `investigation_guide.content` SHALL remain null in state

### Requirement: `content_path` checksum and drift detection (REQ-049)

When `investigation_guide.content_path` is configured, the provider SHALL implement a `ModifyPlan` hook that:

1. At plan time, reads the file at the path given by `content_path`.
2. Computes the SHA-256 hex digest of the file's contents.
3. Compares the computed digest to the prior state value of `checksum`.
4. If the digests differ (or `checksum` has no prior value), marks `checksum` as unknown so Terraform shows a non-empty plan and triggers an apply.

After a successful create or update, the provider SHALL write the SHA-256 digest to `checksum` in state.

`checksum` is a **computed** attribute and SHALL NOT be set by practitioners. If `content_path` is unknown at plan time (e.g. its value comes from another resource not yet applied), the provider SHALL NOT attempt to read the file; it SHALL leave `checksum` as unknown.

#### Scenario: File unchanged between plans

- GIVEN `content_path` pointing to a file and `checksum` in state equal to the SHA-256 of that file
- WHEN `terraform plan` runs
- THEN the plan SHALL be empty (no changes for `artifacts`)

#### Scenario: File changed between plans

- GIVEN `content_path` pointing to a file whose contents have changed since the last apply, so the SHA-256 no longer matches the stored `checksum`
- WHEN `terraform plan` runs
- THEN the plan SHALL show a non-empty diff for `artifacts.investigation_guide.checksum` (marked unknown)

#### Scenario: Unknown content_path skips checksum computation

- GIVEN `content_path` is an unknown value at plan time
- WHEN the plan modifier runs
- THEN the provider SHALL NOT attempt to open or read any file and SHALL leave `checksum` as unknown

### Requirement: Acceptance tests — `artifacts` (REQ-050)

The acceptance test suite for `elasticstack_kibana_alerting_rule` SHALL include:

1. At least one test case that configures **`artifacts.dashboards`** with one or more dashboard IDs and asserts that create and update succeed and state reflects the configured IDs.
2. At least one test case that configures **`artifacts.investigation_guide`** with inline **`content`**, asserts create and update succeed, and asserts state stores the text.
3. At least one test case that configures **`artifacts.investigation_guide`** with **`content_path`**, asserts that `checksum` is set in state after create, modifies the file, runs plan, and asserts a non-empty plan is produced.
4. At least one test case or step that removes the `artifacts` block from configuration and verifies that a subsequent plan shows the expected behaviour (artifacts persisted by Kibana, drift visible in plan if Kibana returns them).

If the minimum Kibana version for `artifacts` is confirmed to be above the CI default, test cases SHALL be gated with an appropriate `SkipFunc` aligned with that minimum version.

#### Scenario: Dashboard IDs round-trip

- GIVEN a rule created with `artifacts.dashboards` listing one or more IDs
- WHEN state is read after apply
- THEN `artifacts.dashboards[*].id` SHALL match the configured IDs

#### Scenario: Inline content round-trip

- GIVEN a rule created with `artifacts.investigation_guide.content` set to a known string
- WHEN state is read after apply
- THEN `artifacts.investigation_guide.content` SHALL equal the configured string

#### Scenario: File-based content triggers re-apply on change

- GIVEN a rule applied with `artifacts.investigation_guide.content_path` pointing to a file
- WHEN the file's content is modified and `terraform plan` runs
- THEN the plan SHALL show a diff for the `artifacts` block
