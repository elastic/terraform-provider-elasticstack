## Schema coverage report: Migrated ingest processor data sources (`migrate-ingest-processor-ds-pf-remaining`)

### Scope
- **Schema**: `internal/elasticsearch/ingest/processor_*_pf_data_source.go` (40 Plugin Framework processor data sources)
- **Acceptance tests**: `internal/elasticsearch/ingest/processor_*_data_source_test.go` (old SDK test files still in place)
- **Test fixtures**: `internal/elasticsearch/ingest/testdata/TestAccDataSourceIngestProcessor*/`

### Methodology
Analyzed each PF schema against its acceptance test file, recording:
- Schema attributes present in the PF data source
- Attributes configured in HCL test fixtures
- Attributes referenced in `TestCheckResourceAttr` / `TestCheckNoResourceAttr` assertions
- JSON output assertions (`CheckResourceJSON`)
- Presence/absence of multi-step tests, update steps, and absence assertions

---

### 1) Attributes with no coverage
These schema attributes/blocks are not referenced in acceptance tests (neither configured nor asserted):

- **`geoip.database_file`**: Optional, string. **Gap**: never configured in any test fixture, never asserted.
- **`geoip.properties`**: Optional, list(string). **Gap**: never configured in any test fixture, never asserted.
- **`fingerprint.salt`**: Optional, string. **Gap**: never configured or asserted. The only test step configures `fields = ["user"]` and nothing else.
- **`community_id.iana_number`**: Optional, string. **Gap**: never configured or asserted. Five test functions cover source/dest IP/port, ICMP, metadata, and on_failure, but `iana_number` is never used.

---

### 2) Attributes with poor coverage

#### 2.1 `geoip` — critical gaps
- **Schema flags**: `field` (Required), `target_field` (Optional, Computed), `database_file` (Optional), `properties` (Optional, list), `first_only` (Optional, Computed), `ignore_missing` (Optional, Computed), plus common fields `description`, `if`, `ignore_failure` (Optional, Computed), `on_failure` (Optional, list), `tag` (Optional).
- **Observed**:
  - Only **1 test step** (`read`) with minimal config: `field = "ip"`.
  - The expected JSON assertion does NOT include `ignore_failure` at all.
- **Gaps**:
  - **Zero common field coverage**: `description`, `if`, `on_failure`, `tag` are never configured or asserted. `ignore_failure` is absent from expected JSON.
  - **High-risk**: the PF implementation always emits `"ignore_failure": false` via `CommonProcessorBody` (tag `json:"ignore_failure"` without `omitempty`). The old SDK `geoip` schema did NOT have common fields, so the old expected JSON omits it. `CheckResourceJSON` uses `reflect.DeepEqual`, so **this test is expected to fail** against the PF implementation.
  - `database_file` and `properties` are not exercised.
  - No `TestCheckNoResourceAttr` absence assertions.
  - No multi-step test.
- **Suggested improvements**:
  - Add an `all_attributes` step configuring `database_file`, `properties`, `target_field`, `first_only`, and all common fields (`description`, `if`, `ignore_failure = true`, `on_failure`, `tag`).
  - Update the `read` step expected JSON to include `"ignore_failure": false`.
  - Add `TestCheckNoResourceAttr` assertions for omitted optional fields.

#### 2.2 `user_agent` — critical gaps
- **Schema flags**: `field` (Required), `target_field` (Optional), `ignore_missing` (Optional, Computed), `regex_file` (Optional), `properties` (Optional, list), `extract_device_type` (Optional), plus common fields.
- **Observed**:
  - 2 test steps (`read`, `all_attributes`). The `all_attributes` step covers processor-specific attributes well (`target_field`, `regex_file`, `properties`, `extract_device_type`, `ignore_missing`).
  - Expected JSON in both steps does NOT include `ignore_failure`.
- **Gaps**:
  - **Zero common field coverage**: `description`, `if`, `on_failure`, `tag` are never configured or asserted. Same `ignore_failure` omission issue as `geoip` — **the `read` step JSON assertion is expected to fail** against PF because PF always emits `"ignore_failure": false`.
  - No `TestCheckNoResourceAttr` absence assertions.
  - No update step.
- **Suggested improvements**:
  - Add a step with all common fields configured and asserted.
  - Update expected JSON strings to include `"ignore_failure": false`.
  - Add `TestCheckNoResourceAttr` for omitted fields in the `read` step.

#### 2.3 `fingerprint` — significant gaps
- **Schema flags**: `fields` (Required, list), `target_field` (Optional), `ignore_missing` (Optional, Computed), `salt` (Optional), `method` (Optional, with `OneOf` validator), plus common fields.
- **Observed**:
  - Only **1 test step** (`read`) with `fields = ["user"]`.
  - Expected JSON includes `"ignore_failure": false` and `"ignore_missing": false`, which is consistent with old SDK behavior.
- **Gaps**:
  - `salt` is never configured or asserted.
  - `method` is never explicitly configured or asserted (relies on default `"SHA-1"`).
  - **Common fields** (`description`, `if`, `on_failure`, `tag`) are never configured or asserted. The existing expected JSON includes `ignore_failure` because the old SDK schema already had it, but other common fields are absent.
  - No `TestCheckNoResourceAttr` absence assertions.
  - No test of the `OneOf` validator for `method` (invalid value would not be rejected in current tests).
  - Only 1 step — no update or equivalence coverage.
- **Suggested improvements**:
  - Add an `all_attributes` step configuring `salt`, `method = "SHA-256"`, and common fields.
  - Add `TestCheckNoResourceAttr` for `salt` and `target_field` in the minimal step.
  - Consider adding an invalid-value test for `method` outside the allowed set.

#### 2.4 `foreach` — significant gaps
- **Schema flags**: `field` (Required), `processor` (Optional), `ignore_missing` (Optional, Computed), plus common fields.
- **Observed**:
  - Only **1 test step** (`read`) with `field = "values"` and `processor = convert.json`.
  - Expected JSON includes `"ignore_failure": false`.
- **Gaps**:
  - Common fields (`description`, `if`, `on_failure`, `tag`) are never configured or asserted.
  - `ignore_missing` is not explicitly asserted.
  - No `TestCheckNoResourceAttr` absence assertions.
  - No step testing `processor` omitted (though it is Optional).
  - Only 1 step.
- **Suggested improvements**:
  - Add an `all_attributes` or `all_common_fields` step exercising common fields.
  - Add `TestCheckNoResourceAttr` for omitted fields.

#### 2.5 `community_id` — single attribute gap
- **Schema flags**: `source_ip`, `source_port`, `destination_ip`, `destination_port`, `iana_number`, `icmp_type`, `icmp_code`, `seed`, `transport`, `target_field`, `ignore_missing`, plus common fields.
- **Observed**:
  - 5 test functions covering core network, ICMP, metadata, and on_failure scenarios.
  - Common fields are well covered in metadata and on_failure tests.
- **Gaps**:
  - `iana_number` is never configured or asserted.
- **Suggested improvements**:
  - Add a test step or function that configures `iana_number` and asserts its presence in JSON.

---

### 3) Cross-cutting poor-coverage patterns

#### 3.1 Absence assertions (`TestCheckNoResourceAttr`) are missing for many processors
The following processors have **zero** `TestCheckNoResourceAttr` assertions, meaning optional fields that are omitted in a minimal config are never verified as absent:

`append`, `bytes`, `circle`, `community_id`, `convert`, `csv`, `date`, `date_index_name`, `dissect`, `dot_expander`, `fingerprint`, `foreach`, `geoip`, `set_security_user`, `urldecode`, `user_agent`

**Impact**: A bug that causes an omitted Optional field to be incorrectly set in state would go undetected.

**Suggested improvement**: For each processor, add `TestCheckNoResourceAttr` checks in the minimal/`read` step for at least one Optional field that is deliberately omitted.

#### 3.2 Update/change steps are sparse
Processors with **no** update/change step (only read + optionally all_attributes):

`append`, `bytes`, `circle`, `community_id`, `convert`, `csv`, `date`, `date_index_name`, `dissect`, `dot_expander`, `drop`, `enrich`, `fail`, `fingerprint`, `foreach`, `geoip`, `grok`, `html_strip`, `inference`, `join`, `network_direction`, `script`, `set`, `set_security_user`, `sort`, `split`, `uri_parts`, `urldecode`, `user_agent`

**Impact**: No verification that changing an attribute value produces a different JSON output / different state.

**Reference**: `gsub`, `kv`, `lowercase`, `pipeline`, `registered_domain`, `remove`, `rename`, `trim`, `uppercase`, and `json` have good update step patterns to follow.

#### 3.3 Validator coverage is absent
Several PF schemas include validators that are never exercised with invalid values in acceptance tests:

- `circle.shape_type` — `stringvalidator.OneOf("geo_shape", "shape")`
- `convert.type` — `stringvalidator.OneOf(...)`
- `fingerprint.method` — `stringvalidator.OneOf("MD5", "SHA-1", "SHA-256", "SHA-512", "MurmurHash3")`
- `community_id.seed` — `int64validator.Between(0, 65535)`
- `date_index_name.date_rounding` — validated string
- `grok.ecs_compatibility` — `stringvalidator.OneOf(...)`
- `json.add_to_root_conflict_strategy` — `stringvalidator.OneOf(...)`
- `sort.order` — `stringvalidator.OneOf("asc", "desc")`

Only `enrich` and `fail` have `ExpectError` tests for invalid `on_failure` JSON.

**Impact**: A regression in validator wiring would not be caught by acceptance tests.

---

### 4) Prioritized action list (smallest diffs first)

| Priority | Action | Processors affected | Rationale |
|----------|--------|---------------------|-----------|
| **P0** | Update expected JSON for `geoip` and `user_agent` `read` steps to include `"ignore_failure": false` | `geoip`, `user_agent` | Tests are expected to fail against PF without this fix |
| **P0** | Add common-field coverage steps for `geoip` and `user_agent` | `geoip`, `user_agent` | These processors newly acquired common fields in PF; requirements spec mandates coverage |
| **P1** | Add `all_attributes` step to `fingerprint` covering `salt`, `method`, and common fields | `fingerprint` | `salt` is completely untested; only 1 step exists |
| **P1** | Add `iana_number` config + assertion to `community_id` | `community_id` | Single missing attribute in an otherwise well-covered processor |
| **P1** | Add `all_attributes` / `all_common_fields` step to `foreach` | `foreach` | Common fields untested; only 1 step |
| **P2** | Add `TestCheckNoResourceAttr` assertions to minimal/`read` steps for processors that lack them | ~16 processors | Catches state pollution bugs for omitted optional fields |
| **P2** | Add update/change steps to processors with only `read`/`all_attributes` | ~29 processors | Verifies value-change behavior |
| **P3** | Add invalid-value acceptance tests for validators | ~8 processors | Guards against schema wiring regressions |

---

*Report generated: 2026-05-01*
