# Entity Requirements — Code-Path Checklist and Requirement Categories

Use this checklist when examining a Terraform resource or data source so the requirements document is complete and traceable to code.

---

## 1. Locate the implementation

| Entity type   | Where to look |
|---------------|---------------|
| Resource      | `internal/<domain>/<name>/` (e.g. `internal/elasticsearch/security/role/`) with `resource.go`, `schema.go`, `create.go`, `read.go`, `update.go`, `delete.go`. Some resources use a single `resource.go` or split schema in a `schema/` package. |
| Data source   | `internal/<domain>/<name>/` or `internal/<domain>/<name>_data_source.go` (SDK v2) or a `*_data_source.go` file in the same package as the resource. Framework data sources: `DataSource` with `Metadata`, `Schema`, `Read`. |

**Entity name**: 
* For Plugin Framework based entities, this is defined in `Metadata()`: `resp.TypeName = req.ProviderTypeName + "_..."`. 
* For SDK based entities, this is defined in `internal/provider/provider.go`.
* Use this for the document title and HCL `resource "/ data "` type.

**Implementation path**: The Go package or directory (e.g. `internal/elasticsearch/security/role`) for the “Resource implementation” / “Data source implementation” line.

---

## 2. Schema

| What to capture | Where |
|-----------------|--------|
| All attributes and blocks | `Schema()` or `GetSchema()` in `schema.go` or main resource file. |
| Required vs optional vs computed | Attribute struct: `Required`, `Optional`, `Computed`. |
| Optional+computed | Both `Optional: true` and `Computed: true`. |
| Types | `StringAttribute`, `SetAttribute`, `ListAttribute`, `ObjectAttribute`, `SingleNestedBlock`, `SetNestedBlock`, etc. Element types (e.g. `types.StringType`). |
| JSON attributes | e.g. `jsontypes.Normalized`; note “JSON (normalized) string” in schema section. |
| Plan modifiers | `planmodifier` (e.g. `UseStateForUnknown`, `RequiresReplace`) — drives Lifecycle/Plan/State requirements. |
| Validators | `validator` list (e.g. `SizeAtLeast(1)`) — note in schema or Mapping. |
| Deprecated blocks | e.g. `elasticsearch_connection` with deprecation message. |
| Version-specific behavior | Comments or version checks in create/update/read; document in schema notes and Compatibility requirements. |

Output: HCL-style block with `<required|optional|optional+computed|computed>`, type, and brief notes (version, deprecated).

---

## 3. Identity and import (resources)

| What to capture | Where |
|-----------------|--------|
| `id` format | How `id` is built in Create/Update (e.g. `clients.CompositeID`, `clusterUUID/name`). |
| Computed `id` | Set in Create/Update and stored in state. |
| Import | `ImportState`: `ImportStatePassthroughID` with path, or custom logic. Document accepted `id` format and that it is persisted to state. |
| Import validation | If present: validation of imported `id` format and error diagnostic. |

---

## 4. CRUD and API

| Operation | Where | What to capture |
|-----------|--------|------------------|
| Create    | `Create()` / `create.go` | Which API (e.g. Put role), how request body is built, how `id` is set, read-after-create if any, error handling. |
| Read      | `Read()` / `read.go`     | Which API (e.g. Get role), how `id` is parsed (e.g. `CompositeIDFromStrFw`), “not found” → remove from state, how response is mapped to state. |
| Update    | `Update()` / `update.go` | Same API as create or dedicated update API; plan → API model; read-after-update; error handling. |
| Delete    | `Delete()` / `delete.go` | Which API (e.g. Delete role), how identifier is derived from state, error handling. |

For each API: requirement that the resource uses that API (with doc link if available). One requirement: non-success API responses (except “not found” on read) are surfaced as diagnostics.

---

## 5. Connection

| What to capture | Where |
|-----------------|--------|
| Default client | `Configure()`: use provider client; “use provider’s configured … client by default”. |
| Resource-level override | Block like `elasticsearch_connection`; `MaybeNewAPIClientFromFrameworkResource` or similar in Create/Read/Update/Delete. Requirement: when block is set, use that connection for that resource’s API calls. |

---

## 6. Compatibility (version checks)

Search for version checks (e.g. `version.Compare`, `MinSupported...Version`) in create/update/read or in validators.

| What to capture | Requirement pattern |
|-----------------|---------------------|
| Feature gated by server version | When attribute/block X is configured, verify server version ≥ Y; if lower, fail with “Unsupported Feature” (or equivalent) diagnostic. |

---

## 7. Lifecycle (resources)

| What to capture | Where |
|-----------------|--------|
| Requires replace | Schema: `RequiresReplace()` on attributes (e.g. `name`). Requirement: when attribute X changes, resource shall require replacement. |

---

## 8. Mapping (config ↔ API ↔ state)

| What to capture | Where |
|-----------------|--------|
| JSON parsing | Where config JSON (e.g. `global`, `metadata`, `query`) is parsed. Requirement: if parsing fails, return “Invalid JSON” (or similar) and do not call the API. |
| Empty vs null | Handling of empty lists/sets from API (e.g. preserve null vs `[]` to avoid drift). |
| Unknown during plan | Plan modifiers or logic that preserve prior state for unknown values; document as Plan/State requirements. |
| API → state | How response is mapped to state (e.g. empty `applications` → null in state; `global`/`metadata` normalized JSON strings). |

---

## 9. State upgrade (resources)

| What to capture | Where |
|-----------------|--------|
| Upgraders | `UpgradeState()`: map of version → `StateUpgrader`. |
| Per-version logic | Each upgrader function: what it changes (e.g. remove empty `global`/`metadata`, convert `field_security` list to object), and how. |
| Errors | If unmarshal/marshal fails: “State Upgrade Error” (or similar) diagnostic and no upgraded state. |

Requirements: one per supported upgrade path (e.g. v0→v1) and one for upgrade failure behavior.

---

## 10. Data sources only

| What to capture | Where |
|-----------------|--------|
| Read only | No Create/Update/Delete; only Read and which API is used. |
| Required vs optional arguments | Often one or more required args (e.g. `name`) to identify the remote object. |
| Computed attributes | All read-only attributes documented as computed. |

---

## Requirement categories and phrasing

Use these categories and “shall” phrasing so the doc stays consistent:

| Category | Use for |
|----------|---------|
| **API** | Which API is used for create/update/read/delete; surfacing API errors in diagnostics. |
| **Identity** | Format and computation of `id`; composite identifiers. |
| **Import** | Import support, `id` format for import, validation and errors. |
| **Lifecycle** | Requires replace when an attribute changes. |
| **Connection** | Default client; resource-level connection override. |
| **Compatibility** | Server/version checks and “Unsupported Feature” (or equivalent) errors. |
| **Create/Update** | Create and update flow: API call, read-back, error when resource not found after update. |
| **Read** | Refresh: parse `id`, call API, remove from state if not found, set state (including stable `name` from `id` where applicable). |
| **Delete** | Delete flow: derive identifier from state, call delete API. |
| **Mapping** | JSON parsing, validation, sending optional fields (omit when unset). |
| **Plan/State** | Preserving unknown values, handling empty vs null to avoid drift. |
| **State** | How API response is stored in state (null vs empty, normalized JSON). |
| **StateUpgrade** | Version-to-version upgrade behavior and upgrade failure. |

Example phrases:

- **API**: “The resource shall use the \<API name\> API to create and update \<objects\> ([docs](link)).”
- **Import**: “The resource shall support import by accepting an `id` in the format \<format\> and persisting it to state.”
- **Identity**: “The resource shall expose a computed `id` representing … in the format \<format\>.”
- **Compatibility**: “When \<attribute\> is configured, the resource shall verify the \<server\> version is at least \<ver\>, and if it is lower the resource shall fail with an ‘Unsupported Feature’ error.”
- **StateUpgrade**: “During v0→v1 upgrade, if \<field\> is null or empty string, the resource shall remove the attribute from the upgraded state.”

---

## File layout for requirements docs

- Path: `dev-docs/requirements/<domain>/<name>.md`.
- Match existing layout: e.g. `dev-docs/requirements/elasticsearch/security/role.md`, `dev-docs/requirements/kibana/slo/slo.md`.
- One document per Terraform resource or data source (not per Go package if one package serves multiple entities).
