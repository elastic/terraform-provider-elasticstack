## MODIFIED Requirements

### Requirement: Metadata JSON mapping — empty-object read-side equivalence (REQ-016–REQ-017)

When the Elasticsearch API returns an empty or absent metadata map, the resource SHALL treat
the Terraform values `null` and `"{}"` as semantically equivalent. If the incoming state holds
an empty JSON object (`"{}"`), the read path SHALL preserve that value rather than overwriting it
with `null`. If the incoming state is `null`, unknown, or holds a non-empty JSON object, the read
path SHALL set `metadata` to `null` (unchanged behaviour). Drift detection is preserved: if prior
state holds non-empty metadata and the API returns empty, the resource SHALL set state to `null`.

#### Scenario: Empty metadata JSON object round-trips correctly

- GIVEN a user resource configured with `metadata = jsonencode({})`
- WHEN the resource is created or updated (sending `metadata: {}` to the Elasticsearch API)
- AND the Elasticsearch GET users API returns an empty metadata map
- THEN `metadata` in Terraform state SHALL equal `"{}"` (not `null`)
- AND no "Provider produced inconsistent result after apply" error SHALL occur

#### Scenario: null metadata preserved when API returns empty

- GIVEN a user resource configured without a `metadata` attribute (state value is `null`)
- WHEN the Elasticsearch GET users API returns an empty metadata map
- THEN `metadata` in Terraform state SHALL remain `null`

#### Scenario: Non-empty metadata drift detected

- GIVEN a user resource with `metadata = "{\"key\":\"value\"}"` in prior state
- WHEN the Elasticsearch GET users API returns an empty metadata map (server drift)
- THEN `metadata` in Terraform state SHALL be set to `null`, reflecting the server state
