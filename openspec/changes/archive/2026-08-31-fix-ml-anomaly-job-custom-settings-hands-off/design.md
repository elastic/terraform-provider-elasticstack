## Context

Canonical requirements for this resource live in
[`openspec/specs/elasticsearch-ml-anomaly-detection-job/spec.md`](../../specs/elasticsearch-ml-anomaly-detection-job/spec.md).
Implementation lives in
[`internal/elasticsearch/ml/anomalydetectionjob/`](../../../internal/elasticsearch/ml/anomalydetectionjob/).

This proposal's design was worked out in the issue thread itself (#4729), in a research
comment from `@tobio` posted 2026-08-31 followed immediately by an "Agreed path" comment
from the same author recording the decision. Both are human comments on the issue (not
an automated research-comment artifact), but they are the direct investigation and
agreed resolution for this exact bug, authored by the person who both reported it and
triggered this proposal run. This design adopts that agreed path (Path 1, "atomic
hands-off") as its spine; the issue title and body remain authoritative for scope (fix
the failed-apply loop on `custom_settings`), and nothing here contradicts that scope —
it only settles which of several ways to fix it. Notably, the agreed path explicitly
rules out the fix the issue body itself proposed (`Computed: true` +
`UseStateForUnknown()`); see "Rejected alternative" below for why.

### What actually breaks (today)

```
0.14.5 — perpetual plan diff, apply "succeeds"
  config omit -> plan: state{...} -> null
  update      -> BuildFromPlan sees plan.CustomSettings.IsNull(), no API call, returns early
  state       -> still {...} (no re-read)
  next plan   -> same diff, forever (harmless)

0.15.0+ (current) — same plan, apply dies
  config omit -> plan: state{...} -> null
  update      -> no API call (same guard as above)
  envelope    -> read-after-write always runs -> fromAPIModel copies API value -> state = {...}
  core        -> planned null != new state -> "Provider produced inconsistent result after apply"
```

Read-after-write is required by this resource's existing contract (REQ-016-018 in the
canonical spec: "the envelope SHALL still perform read-after-write to refresh state from
the API" even when no Update Job call was made) and is not being changed here — the fix
has to make the *read result* consistent with the plan, not skip the refresh.

### Facts that dominate the design (from the issue-thread research)

- **Elasticsearch does not invent this field; Kibana writes it.** `created_by` is a
  wizard/module provenance label; `custom_urls` is user-facing drill-down config that
  Kibana still stores even when empty (`[]`). The field is free-form JSON — future Kibana
  versions or operators can add keys this provider has never seen.
- **Update Job replaces the whole object; it does not merge.** The existing comprehensive
  acceptance test (`TestAccResourceAnomalyDetectionJobComprehensive`, see
  `testdata/.../update/anomaly_detection.tf`) already proves this: updating
  `{"custom_key":...}` to a different map drops `custom_key`. Any design that reads a
  subset and later writes only that subset would delete every key the user did not
  mention, including Kibana's.
- **A Go `map[string]any{}` with a `json:",omitempty"` tag is dropped from the marshaled
  body.** `internal/elasticsearch/ml/anomalydetectionjob/models_api.go:163` currently
  declares `UpdateAPIModel.CustomSettings map[string]any \`json:"custom_settings,omitempty"\``.
  Any path that uses `custom_settings = "{}"` as a wipe signal needs the wire encoding
  changed so `{}` is actually sent — `omitempty` on a plain map cannot distinguish "empty
  object" from "not set."
- **Whether `{"custom_settings":{}}` actually clears existing keys server-side is
  asserted by the agreed path but not yet proven against a live cluster** — flagged
  explicitly by the research as wanting a spike or acceptance-test confirmation before
  this ships (see Open Questions).

## Goals / Non-Goals

**Goals:**

- Stop the failed-apply loop: omitted/null `custom_settings` config must never diff
  against a Kibana/operator-populated server value, on create, update, import, or plain
  read.
- Give users a way to explicitly clear an existing `custom_settings` bag, since state no
  longer mirrors the server value unconditionally.
- Keep `custom_settings` atomic: when the attribute is set to a real object, Terraform
  owns the whole bag and extras (Kibana or otherwise) are visible as drift, matching
  today's Update Job replace semantics.
- Preserve the resource's existing read-after-write contract (REQ-016-018) unchanged.

**Non-goals** (explicitly ruled out by the agreed path):

- Preserving Kibana/operator extras across a user write (that is Path 2, "subset/merge").
- An allow-list of known server-injected keys (Path 3) — cheap but reopens the bug the
  moment Kibana adds a new key.
- `Computed: true` + `UseStateForUnknown()` on `custom_settings` (the issue body's own
  suggestion — see "Rejected alternative" below).
- A custom `StringSemanticEquals` type or private-state key tracking for this attribute.
- Changing the Terraform schema declaration for `custom_settings` in any way (type,
  optionality, computability, or plan modifiers).

## Decisions

### Contract

```
omitted / null   ->  do not write; never copy API custom_settings into state
                     (import included — stays null; plain read/refresh also stays null
                     when prior state is null)
"{}"             ->  the only wipe; send an empty object; persist "{}" even if
                     Get Jobs omits the field afterward
any other object ->  replace the whole bag with exactly that object;
                     extras seen on a later refresh -> plan diff -> apply replaces
```

### Read: `fromAPIModel` (`models_tf.go`)

`fromAPIModel` is called as `state.fromAPIModel(ctx, apiModel)` from
`readAnomalyDetectionJob` (`read.go`), where the receiver already holds, in
`CustomSettings`, the value the caller wants preserved as the "owned" baseline:

- On the write path (create/update), the envelope calls
  `r.read(ctx, client, readResourceID, written.Model)` — `written.Model` is the plan the
  write callback returned, so the incoming `CustomSettings` is the plan's configured
  value (null, `"{}"`, or an object).
- On the plain read/refresh path, `baseResourceEnvelope.Read` decodes `req.State` into
  the model passed to `read` — so the incoming `CustomSettings` is whatever was
  previously persisted in state.

In both cases the incoming value is exactly the signal needed to decide what to do,
without changing any function signature or threading extra context through the envelope.
Capture it before it is overwritten:

```go
priorCustomSettings := plan.CustomSettings // "plan" is the fromAPIModel receiver

switch {
case priorCustomSettings.IsNull():
    plan.CustomSettings = jsontypes.NewNormalizedNull()
case isEmptyJSONObject(priorCustomSettings):
    plan.CustomSettings = jsontypes.NewNormalizedValue("{}")
case apiModel.CustomSettings != nil:
    customSettingsJSON, err := json.Marshal(apiModel.CustomSettings)
    if err != nil {
        diags.AddError("Failed to marshal custom_settings", err.Error())
        return diags
    }
    plan.CustomSettings = jsontypes.NewNormalizedValue(string(customSettingsJSON))
default:
    // Owned (non-null, non-"{}") in prior state/plan, but the API returned nothing
    // (e.g. Elasticsearch dropped the object entirely) — treat as cleared to "{}"
    // rather than resurrecting the prior owned value, so state matches the server.
    plan.CustomSettings = jsontypes.NewNormalizedValue("{}")
}
```

`isEmptyJSONObject` decodes the prior value into `map[string]any` and checks `len(...)
== 0`, rather than comparing the raw string, since `jsontypes.Normalized` values may
differ in formatting (`"{}"` vs `"{ }"`) while being semantically identical.

This mirrors the existing `detector_description` pattern in the same resource (canonical
spec REQ-027-031: "When the prior detector configuration omitted `detector_description`
and Elasticsearch returns an auto-generated description, the resource SHALL keep
`detector_description` null in state") — read-time preservation keyed off the
already-known prior value, no new plumbing required.

### Write: `BuildFromPlan` / wire encoding (`models_api.go`)

Current guard (`models_api.go:234`):

```go
if !plan.CustomSettings.Equal(state.CustomSettings) && !plan.CustomSettings.IsNull() {
    ...
    u.CustomSettings = customSettings
    hasChanges = true
}
```

This already does the right thing for "omitted/null → don't send" (the `!IsNull()`
guard). It does the wrong thing for `"{}"`: `json.Unmarshal("{}", &customSettings)`
produces a non-nil `map[string]any{}` of length 0, assigned to
`u.CustomSettings map[string]any \`json:"custom_settings,omitempty"\`` — but `omitempty`
treats a zero-length map as empty and drops the field from the marshaled JSON entirely,
so the wipe never reaches the API.

Fix the wire encoding, not the guard logic: change `UpdateAPIModel.CustomSettings` (and
the analogous `APIModel.CustomSettings` field used by `toPutJobRequest`, if create-time
`"{}"` is also in scope — see REQ list) from `map[string]any` with `omitempty` to a
representation that distinguishes "absent" from "present-but-empty," for example a
`json.RawMessage` populated by marshaling the parsed map unconditionally when
`BuildFromPlan`'s guard decides to send it at all:

```go
if !plan.CustomSettings.Equal(state.CustomSettings) && !plan.CustomSettings.IsNull() {
    var customSettings map[string]any
    if err := json.Unmarshal([]byte(plan.CustomSettings.ValueString()), &customSettings); err != nil {
        diags.AddError("Failed to parse custom_settings", err.Error())
        return false, diags
    }
    raw, err := json.Marshal(customSettings) // marshals {} for an empty map, never omitted
    if err != nil {
        diags.AddError("Failed to encode custom_settings", err.Error())
        return false, diags
    }
    u.CustomSettings = json.RawMessage(raw)
    hasChanges = true
}
```

with `UpdateAPIModel.CustomSettings json.RawMessage \`json:"custom_settings,omitempty"\``
— `omitempty` on `json.RawMessage` only drops a `nil`/zero-length slice, and a marshaled
`{}` is 2 bytes, so it survives. (`APIModel.CustomSettings` already flows through
`toPutJobRequest` via an explicit `json.Marshal` into `req.CustomSettings
json.RawMessage`, per `models_api.go:297-300` — create-time already uses the right wire
type; only `UpdateAPIModel` needs the type change.)

### Import

`ImportState` (`resource.go`) sets only `id` and `job_id`; `custom_settings` is left at
its Go zero value, which for `jsontypes.Normalized` is null. The subsequent
framework-triggered plain `Read` then goes through the same `fromAPIModel` path above
with `priorCustomSettings.IsNull() == true`, so imported state stays null regardless of
what the server holds — matching decision #3 from the agreed path ("Import stays null").
This differs intentionally from `detector_description`, which does store the
Elasticsearch-generated value when there is no prior — `custom_settings` does not, per
the agreed contract.

## Rejected alternative: `Computed: true` + `UseStateForUnknown()`

The issue body itself suggests this (matching the pattern already used four lines below
by `daily_model_snapshot_retention_after_days`). It would stop the crash and would adopt
whatever Elasticsearch/Kibana last wrote into state when config is omitted. It is
rejected here because:

- It cannot express "clear this field" — once adopted, a Kibana-authored bag stays in
  state forever with no config-driven way to remove it (short of `terraform state rm` /
  reimport games).
- It silently takes ownership of Kibana metadata the user never asked to manage, which is
  a bigger behavior change than "don't crash" — a later Terraform-driven write of a
  partial `custom_settings` object would then unexpectedly diff away from an
  auto-adopted baseline the user never set.
- `UseStateForUnknown` only fires when the planned value is *unknown*, not when it's
  merely *null*-and-differs-from-state; it would not by itself resolve the null-vs-populated
  mismatch this bug is about without further changes to how the plan value is computed at
  all.

## Open Questions

- **Does `POST .../_update` with `{"custom_settings":{}}` actually clear every existing
  key server-side, or does Elasticsearch treat an empty object as a no-op?** The agreed
  path calls this likely (Update Job replaces, not merges) but flags it as needing
  confirmation via a spike or a live-cluster acceptance test before shipping. The
  acceptance test added for this change (tasks.md §4) should be the confirmation
  mechanism — if the wipe does not clear existing keys against a real cluster, the wire-
  encoding fix in "Write" above still stands (it is required regardless), but the "{}"
  semantics documented in the delta spec need to be revisited before merge.
- **None of the other paths considered (subset/Path 2, allow-list/Path 3) are open** —
  they were evaluated and explicitly rejected in the agreed-path comment; this design
  does not revisit them.
