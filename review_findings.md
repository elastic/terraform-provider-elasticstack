# Critical Code Review: Plugin Framework Processor Data Sources

## Summary

All 40 `_pf_data_source.go` files embed `CommonProcessorModel` and call `toCommonProcessorBody`. Default boolean handling for `ignore_failure` and `ignore_missing` is consistent and correct across the board. The overall architecture (generic `processorDataSource` base, shared `marshalAndHash`, per-processor `MarshalBody` methods) is clean and idiomatic.

However, there are **behavioral divergences from the old SDK implementations** that will produce different JSON output (and therefore different computed hashes/IDs) for identical Terraform configurations. There are also schema type mismatches and a few structural issues.

---

## Findings

### 1. Missing SDK default values in multiple processors (CRITICAL)

**Severity: critical**

The old SDK used `schema.Default` for many optional fields and then unconditionally included those values in JSON via `d.Get()`. The PF implementations use `IsKnown()` checks, so when the user does not specify these fields they are omitted from JSON entirely. This produces different `json` and `id` values than the SDK data sources.

| Processor | Field | SDK Default | PF Behavior |
|-----------|-------|-------------|-------------|
| `date` | `target_field` | `"@timestamp"` | omitted |
| `date` | `timezone` | `"UTC"` | omitted |
| `date` | `locale` | `"ENGLISH"` | omitted |
| `date` | `output_format` | `"yyyy-MM-dd'T'HH:mm:ss.SSSXXX"` | omitted |
| `date_index_name` | `timezone` | `"UTC"` | omitted |
| `date_index_name` | `locale` | `"ENGLISH"` | omitted |
| `date_index_name` | `index_name_format` | `"yyyy-MM-dd"` | omitted |
| `csv` | `separator` | `","` | omitted |
| `csv` | `quote` | `"` | omitted |
| `fingerprint` | `target_field` | `"fingerprint"` | omitted |
| `fingerprint` | `method` | `"SHA-1"` | omitted |
| `geoip` | `target_field` | `"geoip"` | omitted |
| `set` | `media_type` | `"application/json"` | omitted |
| `sort` | `order` | `"asc"` | omitted |

**Evidence:**
SDK `date` Read:
```go
processor.Timezone = d.Get("timezone").(string) // always "UTC" even if omitted
```
PF `date` MarshalBody:
```go
if IsKnown(m.Timezone) {
    body.Timezone = m.Timezone.ValueString() // omitted when null
}
```

**Recommended fix:** In each `MarshalBody`, add a null/unknown check that sets both `body.Field` and `m.Field` to the SDK default value, exactly as is done for booleans like `ignore_missing`.

---

### 2. `community_id` seed always serialized in SDK, omitted in PF (CRITICAL)

**Severity: critical**  
**File:** `internal/elasticsearch/ingest/processor_community_id_pf_data_source.go`

The SDK Read function unconditionally sets `processor.Seed = &seed` (where `seed` comes from `d.Get("seed")` which defaults to `0`). The SDK model tags `Seed` as `json:"seed"` (no `omitempty`), so `"seed": 0` is **always** present in JSON.

The PF body tags `Seed` as `json:"seed,omitempty"`, and `MarshalBody` only sets it when `IsKnown(m.Seed)`. When omitted, the field disappears from JSON instead of appearing as `0`.

**Evidence:**
SDK:
```go
seed := d.Get("seed").(int)
processor.Seed = &seed  // always included
```
PF:
```go
if IsKnown(m.Seed) {
    body.Seed = intPtr(m.Seed.ValueInt64()) // omitted when null
}
```

**Recommended fix:** Always include `seed` in the PF body (remove `omitempty` from the body struct tag) or explicitly set `body.Seed = intPtr(0)` and `m.Seed = types.Int64Value(0)` when null to match SDK behavior.

---

### 3. Schema type mismatch: SDK `TypeSet` replaced with PF `ListAttribute` (WARNING)

**Severity: warning**  
**Files:**
- `processor_remove_pf_data_source.go`
- `processor_geoip_pf_data_source.go`
- `processor_user_agent_pf_data_source.go`

The SDK used `schema.TypeSet` for `remove.field`, `geoip.properties`, and `user_agent.properties`. Sets are unordered and deduplicated in Terraform. PF uses `ListAttribute` (ordered, duplicates allowed). This changes config semantics and could affect any tooling or modules that relied on set hash stability.

**Evidence:**
SDK remove:
```go
"field": { Type: schema.TypeSet, Required: true, MinItems: 1 }
```
PF remove:
```go
"field": schema.ListAttribute{ Required: true, ElementType: types.StringType, Validators: []validator.List{listvalidator.SizeAtLeast(1)} }
```

**Recommended fix:** If strict SDK parity is required, change these to `schema.SetAttribute` with `types.SetType`. If list semantics are acceptable, document the change.

---

### 4. Stricter list validators added in PF (WARNING)

**Severity: warning**  
**Files:**
- `processor_geoip_pf_data_source.go` (`properties`)
- `processor_fingerprint_pf_data_source.go` (`fields`)
- `processor_set_security_user_pf_data_source.go` (`properties`)

The SDK did **not** enforce `MinItems: 1` / `SizeAtLeast(1)` on these fields. The PF versions do. Configs that explicitly passed an empty list (e.g., `properties = []`) would have been accepted by the SDK but will fail validation in PF.

**Evidence:**
SDK geoip properties had no `MinItems`:
```go
"properties": { Type: schema.TypeSet, Optional: true, ... }
```
PF geoip properties:
```go
"properties": schema.ListAttribute{ Optional: true, Validators: []validator.List{listvalidator.SizeAtLeast(1)} }
```

**Recommended fix:** Decide whether the stricter validation is desired. If SDK parity is the goal, remove `SizeAtLeast(1)` from these optional lists.

---

### 5. `dissect` `append_separator` omitted when empty in PF (WARNING)

**Severity: warning**  
**File:** `internal/elasticsearch/ingest/processor_dissect_pf_data_source.go`

The SDK model serializes `append_separator` unconditionally (`json:"append_separator"` with no `omitempty`). The PF body uses `json:"append_separator,omitempty"`, so when the user does not specify it, the field is omitted from JSON rather than serialized as `""`.

**Evidence:**
SDK model:
```go
AppendSeparator string `json:"append_separator"`
```
PF body:
```go
AppendSeparator string `json:"append_separator,omitempty"`
```

**Recommended fix:** Remove `omitempty` from `processorDissectBody.AppendSeparator` to match SDK JSON output.

---

### 6. `circle` field description is copy-pasted from `trim` (NIT)

**Severity: nit**  
**File:** `internal/elasticsearch/ingest/processor_circle_pf_data_source.go:98`

The description for the `field` attribute says "The string-valued field to trim whitespace from." This was inherited from the SDK `trim` processor schema and is incorrect for a `circle` processor.

**Recommended fix:** Use a circle-appropriate description (e.g., "The field containing the WKT circle string to convert.").

---

### 7. Inconsistent comments for `ignore_failure` default handling (NIT)

**Severity: nit**  
**Files:** Multiple

Some `MarshalBody` functions include the comment `// Ensure ignore_failure default is reflected in state.` before the `ignore_failure` null-check block (e.g., `processor_append_pf_data_source.go`, `processor_drop_pf_data_source.go`), while most others omit the comment entirely.

**Recommended fix:** Either add the comment consistently everywhere or remove it everywhere for uniformity.

---

## Positive Observations

1. **Common processor embedding is 100% correct.** Every file embeds `CommonProcessorModel` and calls `toCommonProcessorBody`.
2. **Default boolean handling is consistent.** `ignore_failure`, `ignore_missing`, `override`, `trace_match`, `ignore_empty_value`, etc. all correctly set defaults in both body and model state.
3. **GeoIP and user_agent common fields are correctly added.** Both now support `description`, `if`, `ignore_failure`, `on_failure`, and `tag`, which they lacked in the SDK. The PF implementations correctly include these when set.
4. **No nil pointer dereference risks.** All type assertions on `types.String` / `types.Bool` list elements include `ok` checks and append diagnostics rather than panicking.
5. **Generic data source base is well-designed.** `processorDataSource[T]` and `marshalAndHash` eliminate significant duplication.
