## Why

Plugin Framework resources in this provider have no uniform `timeouts` support. Four resources (`fleet/customintegration`, `elasticsearch/ml/jobstate`, `elasticsearch/ml/datafeed_state`, `elasticsearch/ml/anomalydetectionjob`) each implement their own `timeouts` field, schema entry, and per-callback `context.WithTimeout` wrapping; the other 62 entitycore resources have no way for practitioners to bound long-running Create/Read/Update/Delete operations. This closes the SDKv2 → Plugin Framework migration gap tracked by [#780](https://github.com/elastic/terraform-provider-elasticstack/issues/780) for timeouts (retries are explicitly out of scope — the Plugin Framework no longer offers a generic retry mechanism).

The action envelope already solves the same problem for actions via `ActionTimeoutsField`, `WithActionTimeouts`, auto-injected schema, and ctx-wrapping inside the generic `Invoke` prelude. This change applies the same blueprint to the Elasticsearch and Kibana resource envelopes, using the attribute-style `timeouts.AttributesAll` schema (the modern Plugin Framework convention) instead of a block.

## What Changes

### Envelope (Elasticsearch and Kibana resource envelopes)

- **`internal/entitycore/resource_timeouts.go`** (new file): Introduce shared timeouts plumbing usable by both envelopes:
  - `ResourceTimeoutsField` — embeddable struct with `Timeouts timeouts.Value `tfsdk:"timeouts"`` (using `github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts`) and a `GetTimeouts()` value-receiver method
  - `WithResourceTimeouts` interface — `GetTimeouts() timeouts.Value`
  - `ResourceTimeouts` struct with fields `Create, Read, Update, Delete time.Duration` for per-op default overrides; any zero field falls back to the framework default
  - Package-level constants `DefaultResourceCreateTimeout = 20*time.Minute`, `DefaultResourceReadTimeout = 5*time.Minute`, `DefaultResourceUpdateTimeout = 20*time.Minute`, `DefaultResourceDeleteTimeout = 20*time.Minute`
- **`internal/entitycore/resource_envelope.go`**: Tighten `ElasticsearchResourceModel` to embed `WithResourceTimeouts`. Add `Timeouts ResourceTimeouts` to `ElasticsearchResourceOptions[T]`. In `Schema`, inject `"timeouts": timeouts.AttributesAll(ctx)` into `schema.Attributes` (mirroring how the existing `elasticsearch_connection` block is injected into `schema.Blocks`); any pre-existing `"timeouts"` attribute in the factory output is silently overwritten — same contract as action-envelope block injection. In `Create`, `Read`, `Update`, and `Delete`, decode the model and apply `model.GetTimeouts().Create/Read/Update/Delete(ctx, opts.Timeouts.<Op> or framework default)` before invoking the user callback; defer the cancel.
- **`internal/entitycore/kibana_resource_envelope.go`**: Same treatment — `KibanaResourceModel` embeds `WithResourceTimeouts`, `KibanaResourceOptions[T]` gains `Timeouts ResourceTimeouts`, `Schema` injects the attribute, `Create/Read/Update/Delete` wrap ctx.
- **`internal/entitycore/resource_envelope_test.go`** and **`internal/entitycore/kibana_resource_envelope_test.go`**: Update every test model to embed `ResourceTimeoutsField`; add scenarios covering schema injection, ctx-wrap per op, per-op default override via Options, and silent overwrite of any factory-supplied `timeouts` attribute.

### Mechanical migration: 62 entitycore resources gain `timeouts` (additive)

All 62 entitycore-envelope resources that do not currently expose `timeouts` gain it through model embedding. The work per resource is a one-line model change:

```go
type myModel struct {
    entitycore.<Es|Kb>ConnectionField
    entitycore.ResourceTimeoutsField  // ← added
    // existing fields...
}
```

The schema is auto-injected by the envelope; no schema, callback, or test code changes are required for these 62 resources. Provider documentation regenerates to include the `timeouts` attribute.

### Migration of the 3 existing attribute-style timeouts resources (additive)

- **`internal/fleet/customintegration`**:
  - Replace the bespoke `Timeouts timeouts.Value` field in `models.go` with embedded `entitycore.ResourceTimeoutsField`
  - Delete the bespoke `"timeouts": timeouts.Attributes(...)` entry from `schema.go`
  - Delete the manual `plan.Timeouts.Create(ctx, 20*time.Minute) → context.WithTimeout` block from `create.go`
  - Delete the equivalent block from `update.go`
  - Pass per-op defaults via `entitycore.NewKibanaResource` options: `Timeouts: entitycore.ResourceTimeouts{Create: 20*time.Minute, Update: 20*time.Minute}`
  - Schema effect: gains `read` and `delete` (additive, non-breaking)
- **`internal/elasticsearch/ml/jobstate`**: Same pattern; per-op default `Create: 5*time.Minute, Update: 5*time.Minute`
- **`internal/elasticsearch/ml/datafeed_state`**: Same pattern; per-op default `Create: 5*time.Minute, Update: 5*time.Minute`

### Breaking migration of the 1 block-style timeouts resource

- **`internal/elasticsearch/ml/anomalydetectionjob`**: Migrate from `timeouts.Block(ctx, timeouts.Opts{Delete: true})` (HCL block syntax `timeouts {}`) to envelope-injected attribute syntax `timeouts = {}`. This is a **breaking change to existing Terraform configuration** for any practitioner using a `timeouts {}` block on this resource. Mitigations:
  - CHANGELOG entry under `BREAKING CHANGES` with before/after HCL
  - Migration note in resource documentation

## Capabilities

### Modified Capabilities

- **`entitycore-resource-envelope`** — gains the `WithResourceTimeouts` model constraint, auto-injected `timeouts` attribute, and per-op ctx-wrap behavior in Create/Read/Update/Delete
- **`entitycore-kibana-resource-envelope`** — same additions on the Kibana side

### New Capabilities

None — practitioners gain a new `timeouts` attribute on resources, but the resources themselves already exist and are specified in their own capability specs (no entity behavior changes beyond the timeouts surface).

## Impact

- `internal/entitycore/resource_timeouts.go` (new ~60 LOC)
- `internal/entitycore/resource_envelope.go` (interface tighten, Schema injection, four ctx-wrap blocks)
- `internal/entitycore/kibana_resource_envelope.go` (same)
- `internal/entitycore/resource_envelope_test.go` and `kibana_resource_envelope_test.go` (test models updated + new scenarios)
- 62 resource model files: one-line embed addition each (`internal/<comp>/.../models*.go` / `models_tf.go`)
- 4 resource migrations (3 additive, 1 breaking) — see What Changes above
- `CHANGELOG.md` entry for breaking change to `elasticsearch_ml_anomaly_detection_job`
- Resource documentation regenerates (`make docs-generate`) for all 66 entitycore-envelope resources to reflect the unified `timeouts` attribute (62 gain it for the first time; 4 migrate from their bespoke shape)

## Out of Scope

- **Retries**: Plugin Framework no longer offers a generic retry primitive equivalent to SDKv2's `resource.RetryError`/`resource.StateChangeConf` lifecycle. Per-callback polling stays a per-resource concern.
- **Action envelope changes**: Actions already have timeouts (block-style, single-op). Out of scope to migrate them to attribute style here.
- **Data source / ephemeral resource timeouts**: These envelopes do not gain timeouts in this change. Data source reads are typically fast; ephemeral resources have their own lifecycle.
