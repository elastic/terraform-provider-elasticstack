## MODIFIED Requirements

### Requirement: Read behavior and missing-resource handling (REQ-008) — partial update

The provider SHALL apply intent-preserving null normalization to the root-level `description` attribute on read. When the Kibana API returns an empty-string `description` (`""`) and the prior Terraform plan/state had `description` as null (i.e., the practitioner did not set `description`), the provider SHALL store `null` in state — not `""`. When the prior Terraform plan/state had `description` as `""` (i.e., the practitioner explicitly set `description = ""`), the provider SHALL preserve `""` in state. When the API returns a non-empty `description`, the provider SHALL store that value unchanged.

This normalization fixes a Kibana 9.5 behavior change: previously the API omitted `description` when none was supplied; from 9.5 onward it returns `""`. Without this fix, practitioners who omit `description` see "Provider produced inconsistent result after apply" with the message `.description: was null, but now cty.StringVal("")`.

#### Scenario: Omitted description normalizes to null on read

- GIVEN a dashboard configured without `description` (null in Terraform state/plan)
- AND the Kibana API returns `description: ""`
- WHEN the provider reads the dashboard
- THEN state SHALL contain `description = null`

#### Scenario: Explicit empty description preserved on read

- GIVEN a dashboard configured with `description = ""`
- AND the Kibana API returns `description: ""`
- WHEN the provider reads the dashboard
- THEN state SHALL contain `description = ""`

#### Scenario: Non-empty description preserved unchanged

- GIVEN a dashboard configured with `description = "My dashboard"`
- AND the Kibana API returns `description: "My dashboard"`
- WHEN the provider reads the dashboard
- THEN state SHALL contain `description = "My dashboard"`

---

### Requirement: State preservation for fields Kibana omits or defaults (REQ-009) — partial update

The provider SHALL treat an API-returned `""` for `description` as semantically equivalent to an omitted field when prior plan/state had `description` null, restoring null in state rather than propagating the API-echoed empty string. This is an instance of REQ-009 null-preservation applied to the dashboard root `description`. This SHALL be consistent with the null/empty-string normalization already applied to XY chart `fitting.type`, `fitting.end_value`, and panel-level `time_range`.

#### Scenario: Empty-string description treated as null for null-intent practitioners

- GIVEN a practitioner has never set `description` on a dashboard (prior state: null)
- AND Kibana 9.5 returns `description: ""` on a subsequent read or post-apply read-back
- WHEN the provider applies REQ-009 null-preservation to `description`
- THEN state SHALL contain `description = null` and no drift SHALL be reported on the next plan
