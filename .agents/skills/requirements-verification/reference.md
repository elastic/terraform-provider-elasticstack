# Requirements Verification — Consistency, Implementation Mapping, Test Patterns

## Consistency checks

Use these when assessing whether the requirements document is internally consistent.

### Identity and Import

- Every **Identity** requirement that defines an id format (e.g. `<cluster_uuid>/<role_name>`) must match the **Import** requirement that specifies the accepted import id format. If one says “composite id” and another “name only”, that’s inconsistent.
- **Import** “persist to state” and **Identity** “expose computed id” must align: the imported value should match the format the resource would set on create/update.

### Schema vs requirements

- If a requirement says “when [attribute] is configured” or “when [attribute] is non-null”, that attribute must be optional or optional+computed in the schema (not required-only with no default).
- If a requirement says “the resource shall set [attribute] in state to …”, that attribute must be computed (or optional+computed) in the schema.
- **Compatibility** requirements that gate on “when [X] is configured” imply X exists in the schema; **Lifecycle** “when [X] changes” implies X is in the schema and typically has a plan modifier.

### Lifecycle

- No pair of requirements such that one says “when X changes, resource shall require replacement” and another says “when X changes, resource shall update in place”. Same attribute X cannot trigger both.

### Compatibility

- Same feature (e.g. “description”) should not have two different minimum versions in two requirements (e.g. 8.14 and 8.15).
- Version checks (“server version at least X”) should reference the same product (Elasticsearch vs Kibana) as the implementation.

### State and Plan/State

- For a given attribute, avoid one requirement saying “preserve null vs empty” (store as null when previously null) and another saying “store empty list” when the API returns empty.
- **Plan/State** “preserve prior state when unknown” should refer to attributes that are optional+computed or computed, and that have a plan modifier in the schema.

### StateUpgrade

- “Support upgrading prior state schema version N to M” implies the schema has a version ≥ M and that an upgrader for N→M exists in code.
- Multiple StateUpgrade requirements for the same transition (e.g. v0→v1) should not contradict (e.g. “remove attribute” vs “keep attribute” for the same key).

### API

- Create/Update requirements should reference the same API (or compatible APIs) that the resource actually uses (e.g. one “Put” for both create and update). Read and Delete should reference the correct get/delete endpoints.

---

## Requirement → implementation mapping

Use this to find where each requirement category is implemented so you can mark Met / Not met / Unclear.

| Category        | Typical implementation location | What to check |
|-----------------|----------------------------------|----------------|
| **API**         | `create.go`, `update.go`, `read.go`, `delete.go`; client wrappers in `internal/clients/elasticsearch/` or `internal/clients/kibanaoapi/` | Correct API call (e.g. PutRole, GetRole, DeleteRole); error handling (diagnostics on non-success, 404 on read → remove from state). |
| **Identity**    | `create.go` / `update.go` (where id is set); `resource.go` or read (where id is computed) | How `id` is built (e.g. `clients.CompositeID`, cluster UUID + name); stored in state. |
| **Import**      | `resource.go` — `ImportState`; possibly custom import logic that validates id | ImportStatePassthroughID or custom; validation of id format and error diagnostic. |
| **Lifecycle**  | `schema.go` — plan modifier `RequiresReplace()` on attributes | Which attributes have RequiresReplace. |
| **Connection**  | `resource.go` Configure; create/read/update/delete using `MaybeNewAPIClientFromFrameworkResource` or similar with connection block | Default client from provider; override when block is set. |
| **Compatibility** | `update.go` / `create.go` / validators — version checks before API call | `version.Compare`, MinSupported*Version; “Unsupported Feature” or equivalent diagnostic. |
| **Create/Update** | `create.go`, `update.go` | Put API call; read-back after create/update; error if not found after update. |
| **Read**        | `read.go` | Parse id; call get API; if 404 remove from state; set state (e.g. name from id). |
| **Delete**      | `delete.go` | Parse id from state; call delete API. |
| **Mapping**     | `models.go` (toAPIModel, fromAPIModel); schema validators | JSON parse with “Invalid JSON” diagnostic; omit unset fields in API payload. |
| **Plan/State**  | `schema.go` — plan modifiers (e.g. UseStateForUnknown) | Preserve prior state for unknown; which attributes. |
| **State**       | `read.go`, `models.go` (fromAPIModel) | Empty vs null handling; normalized JSON; null for absent blocks. |
| **StateUpgrade** | `resource.go` — `UpgradeState()`; upgrader functions | Map of version → upgrader; v0→v1 logic (remove empty, convert list→object, etc.); error on parse failure. |

For **data sources**, only Read, API, Connection, Compatibility, Mapping, State apply; no Create/Update/Delete, Import, Lifecycle, or StateUpgrade.

---

## Test opportunity patterns

Use these to suggest unit or acceptance tests that verify requirements.

### Verifiable by unit test (no live API)

| Requirement type   | Example test | Verifies |
|--------------------|--------------|----------|
| **StateUpgrade**   | Table-driven TestV0ToV1: input state JSON, expected output, expectError | Upgrade logic (remove empty, convert field_security list→object); parse failure returns error. |
| **Identity** (id format) | Unit test: build id from name/clusterUUID, parse back, assert equality | Composite id format and parsing round-trip. |
| **Import** (validation) | Unit test: ImportState with invalid id, expect error diagnostic | Validation logic and error message. |
| **Mapping** (JSON) | Unit test: toAPIModel with invalid JSON for attribute X, expect “Invalid JSON” diagnostic | JSON parse failure path. |
| **Plan/State**      | Schema/plan modifier test: plan with unknown for attribute X, assert state value preserved | UseStateForUnknown or equivalent. |

### Verifiable by acceptance test (live API or mocked provider)

| Requirement type   | Example test | Verifies |
|--------------------|--------------|----------|
| **API (create/read/update/delete)** | TestStep: Config creating resource → Check resource attributes; Config update → Check updated attributes; Destroy | Create uses Put; Read returns correct state; Update uses Put; Delete removes resource. |
| **API (errors)**   | TestStep: Config that triggers API error (e.g. invalid payload), ExpectError or ExpectDiagnostic | Non-success API response surfaced as diagnostic. |
| **Read (404)**     | TestStep: Create resource, delete it outside Terraform, run refresh → resource removed from state | Remove from state when not found. |
| **Import**         | TestStep: Create resource, ImportState with id, ImportStateVerify | Import accepts id format and state matches. |
| **Import (invalid)** | TestStep: ImportState with invalid id format, ExpectError | Error diagnostic for bad id. |
| **Connection**     | TestStep: Config with resource-level connection block, Check resource created/read | Override client used. |
| **Compatibility**  | TestStep: Config with version-gated attribute, run against older server (or mock), ExpectError “Unsupported Feature” | Version check and diagnostic. |
| **Lifecycle**      | TestStep: Create, then Config with changed name (or ForceNew attribute), expect replace (destroy+create) | RequiresReplace behavior. |
| **State (empty vs null)** | TestStep: Config with empty set, apply, refresh; Check attribute absent or `#` = 0 as required by requirement | No drift from empty vs null. |

### Suggested test wording for report

For each requirement with no or weak coverage, suggest in the form:

- **REQ-xxx**: **Unit test** — “TestImportStateInvalidID: call ImportState with id `badformat`, expect diagnostic containing ‘required format’.”  
- **REQ-yyy**: **Acceptance test** — “Add TestStep: Config with `description` set, run against Elasticsearch 8.14 (or mock), ExpectError matching ‘Unsupported Feature’.”

Cross-reference existing tests: if `acc_test.go` already has an import step but only for valid id, the opportunity is “Add step: import with invalid id, ExpectError”. If `resource_test.go` has state upgrade tests but none for “unparseable JSON”, add “TestStateUpgradeUnparseableJSON: prior state invalid JSON, expect State Upgrade Error”.
