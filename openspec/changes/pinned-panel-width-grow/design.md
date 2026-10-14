## Context

`elasticstack_kibana_dashboard`'s `pinned_panels[]` nested schema currently exposes only `type` and
the four `*_config` blocks (`options_list_control_config`, `range_slider_control_config`,
`esql_control_config`, `time_slider_control_config`) via `pinnedPanelsNestedObject()`
(`internal/kibana/dashboard/schema_control_configs.go:200-217`). The Kibana Dashboards API's
pinned-control-group payload also carries per-panel `width` (enum `small`/`medium`/`large`, default
`medium`) and `grow` (bool, default `false`) on all four `ControlsGroupSchema*Control` structs — for
example `Grow *bool` / `Width *KibanaHTTPAPIsKbnControlsSchemasControlsGroupSchemaOptionsListControlWidth`
on `KibanaHTTPAPIsKbnControlsSchemasControlsGroupSchemaOptionsListControl`
(`generated/kbapi/kibana.gen.go:41637-41649`). The generated client already fully supports both
fields; the gap is entirely in the provider — the Terraform schema, model, and the JSON-remap
conversion path between the Terraform model and the API payload all need to carry `width`/`grow`.

The actual drop happens in `internal/kibana/dashboard/panelkit/control_pinned_handler.go`: `ToAPI`
builds a `DashboardPanelType*` struct from the Terraform model and round-trips it through
`RemapViaJSON` (`internal/kibana/dashboard/panelkit/remap.go`, a plain JSON marshal/unmarshal,
matched by JSON tag) into the API "group" struct; `FromAPI` does the reverse via `AsGroup`. The
intermediate `DashboardPanelType*` structs (e.g. `KibanaHTTPAPIsKbnDashboardPanelTypeOptionsListControl`)
have no `width`/`grow` JSON keys at all, so the tag-based remap can never carry these two fields in
either direction in either `ToAPI` or `FromAPI` — this is the literal mechanism of the bug.

## Goals / Non-Goals

**Goals:**

- Carry `width` and `grow` from Terraform configuration through to the Kibana Dashboards API create
  and update requests for all four pinned control types that share `control_pinned_handler.go`.
- Read `width`/`grow` back from the API on refresh and import, following the package's existing
  null-preservation convention so an omitted attribute does not drift against a server-reported
  default.
- Keep the four `*_config` block schemas and in-grid `panels[]` behavior unchanged; this is
  additive, sibling-attribute scope only.

**Non-Goals:**

- Reworking the generated Kibana API client or the `generated/kbapi` OAS transform pipeline. The
  gap is on the provider side of an already-supported API shape.
- Changing unpinned/regular dashboard panel sizing. The Discover-session `width` Float64Attribute at
  `internal/kibana/dashboard/panel/discoversession/schema.go` is an unrelated, pre-existing field for
  column settings and must not be confused with this pinned-panel `width` enum.
- Changing the four `*_config` block schemas themselves; `width`/`grow` are new sibling attributes on
  the pinned-panel entry, not a rework of existing config fields.

## Decisions

1. **Schema and model placement: `PinnedPanelModel` only, not `PanelModel`**
   Add `width` (`schema.StringAttribute{Optional: true, Validators:
   []validator.String{stringvalidator.OneOf("small", "medium", "large")}}`) and `grow`
   (`schema.BoolAttribute{Optional: true}`) directly to `pinnedPanelsNestedObject()`
   (`internal/kibana/dashboard/schema_control_configs.go`), alongside `type` and the `*_config`
   blocks. Add matching `Width types.String` / `Grow types.Bool` fields to `PinnedPanelModel` only
   (`internal/kibana/dashboard/models/panel.go`) — not to `SyntheticPanel()`'s `PanelModel` return
   value, which has no room for sizing fields and is shared with in-grid panels that don't have
   `width`/`grow` at all. This follows the existing `stringvalidator.OneOf` enum convention used
   elsewhere in the dashboard package (e.g. `internal/kibana/dashboard/panel/optionslist/schema.go`).

2. **Conversion mechanism: typed `AfterRemapToGroup`/read-side callbacks on `ControlPinnedHandler`**
   `control_pinned_handler.go`'s generic type parameter `G` is constrained only by `any`, so group
   fields can't be accessed directly inside the generic handler. Today `AfterRemapToGroup` has
   signature `func(group *G)` with no access to the Terraform model, and it is only invoked inside
   `ToAPI`; only `esqlcontrol/pinned.go` wires it up today (to set `group.Type`) — `optionslist/pinned.go`,
   `rangeslider/pinned.go`, and `timeslider/pinned.go` all leave it nil. Change the signature to
   `func(ppm models.PinnedPanelModel, group *G)` and the `ToAPI` call site to pass `ppm` (the
   existing `ToAPI` parameter) through. Update the existing `esqlcontrol` callback for the new
   signature and add a new callback to each of the other three control types' handler construction;
   each callback sets `group.Width`/`group.Grow` from `ppm.Width`/`ppm.Grow` (converting the
   Terraform string to the generated `*...Width` enum pointer type, and the Terraform bool to
   `*bool`). Add a symmetric typed read callback invoked right after `AsGroup(raw)` inside `FromAPI`,
   wired per control type the same way, that reads `group.Width`/`group.Grow` back onto the
   `PinnedPanelModel` being built.

3. **Read-side default handling: explicit no-prior branch, then preserve**
   The package's convention is explicit null-preservation, not unconditional copy of the API value:
   `PreserveString`/`PreserveBool` (`internal/kibana/dashboard/panelkit/nullpreserve.go`) return the
   existing Terraform value unchanged when it is null/unknown, and only convert the API pointer to a
   `ControlPinnedHandler.FromAPI` receives the prior panel at the same list index, which is not
   necessarily the same control type as the API response. The read callback must therefore use the
   same type-match criterion as `SeedPinnedPanelForRead`: when `prior` is nil, its type is unknown,
   or its type differs from the current discriminator, set `width`/`grow` directly from the API's
   `Width`/`Grow` pointers (nil pointer → null Terraform value). When a same-type prior exists, use
   `PreserveString`/`PreserveBool` against `prior.Width`/`prior.Grow` so an omitted config value that
   reads back as the server default (`medium`/`false`) stays null instead of being overwritten. This
   mirrors the existing import-vs-refresh and type-change branching used by the pinned-panel path.

4. **Validator and type conversion for `width`**
   The Terraform attribute is a plain `types.String` validated with `stringvalidator.OneOf("small",
   "medium", "large")`; the API field is a typed string enum (`*...ControlWidth`) distinct per
   control type. The `AfterRemapToGroup`/read callbacks convert between the two by string value
   (Terraform string ↔ enum's underlying string), not by importing a shared enum type, since each of
   the four control types has its own generated enum type with an identical set of string values.

5. **Alternatives considered**
   - *Extend the generated-client intermediate `DashboardPanelType*` structs to carry `Width`/`Grow`
     so the existing tag-based `RemapViaJSON` carries them through unchanged*: rejected. This is
     incomplete on its own — `ToAPI`'s write path starts from `ppm.SyntheticPanel()`, which returns a
     `PanelModel` with no `Width`/`Grow` field, so the remap would still have nothing to carry without
     also giving `PanelModel` its own sizing fields (duplicating state already tracked on
     `PinnedPanelModel`) or adding a bespoke pre-remap copy step — at which point it has reintroduced
     a form of the per-control-type callback above, just at a different boundary. It also requires
     hand-editing a generated client (`generated/kbapi/kibana.gen.go`) or extending the OAS
     transform pipeline (`generated/kbapi/transform_schema.go`), which would be overwritten on
     regeneration unless the generator itself is changed — a much larger, pipeline-level change for a
     two-field, provider-side gap.

## Risks / Trade-offs

- **[Risk]** Changing `AfterRemapToGroup`'s signature from `func(group *G)` to `func(ppm
  models.PinnedPanelModel, group *G)` is a breaking change to an existing callback, not purely
  additive. **Mitigation**: `AfterRemapToGroup` is private to the `panelkit`/`panel/*` packages and
  has exactly one existing caller (`esqlcontrol/pinned.go`) to update; the blast radius is contained
  to the four `pinned.go` files touched by this change.
- **[Risk]** Practitioners who already rely on the server defaulting `width`/`grow` to
  `medium`/`false` after this change ships could see unexpected diffs if the no-prior/preserve branch
  is implemented incorrectly (e.g. always copying the API value instead of preserving a null on
  refresh). **Mitigation**: unit tests cover both the no-prior/import branch and the
  preserve-on-refresh branch explicitly, matching the acceptance criteria below.
- **[Trade-off]** Each of the four control types' generated `Width` enum is a distinct Go type even
  though they share the same three string values, so the conversion helper cannot be fully shared
  across control types without either generics or a small per-type wrapper. This is accepted as a
  minor duplication cost consistent with how the rest of `panel/*` already duplicates thin
  per-control-type glue around the shared `ControlPinnedHandler`.

## Migration Plan

- No state migration. `width` and `grow` are new optional attributes added to an existing nested
  object (`pinned_panels[]`); existing state with `pinned_panels` entries that predate this change
  has no value for either attribute and will read back as null via the no-prior/preserve branch
  (treated as "no prior value to preserve from" only on the first read after upgrade if the entry's
  stored width/grow are absent — otherwise handled like any other newly-added optional attribute by
  the Plugin Framework, which does not require a schema version bump for additive optional
  attributes).
- After this change, a previously-applied `pinned_panels` entry with `width`/`grow` set in
  configuration starts actually taking effect on the next apply instead of being silently dropped.

## Open questions

- Does `state_upgrade.go` need a new state-upgrade step for existing `pinned_panels` state written
  before this change, or is adding two new optional attributes to an existing nested object
  non-breaking for Terraform's schema versioning? (non-blocking, but affects task list)
- The Dashboards API itself is version-gated (`MinDashboardAPISupport = "9.4.0-SNAPSHOT"` in
  `dashboardacctest/dashboardacctest.go`); the OAS schema shows no separate version marker for
  `width`/`grow` within the already-gated pinned-panel schema, so no additional version gate beyond
  the existing `MinDashboardAPISupport` check appears needed, but this should be confirmed against
  the Elastic Stack changelog for the release that introduced `width`/`grow` on control-group panels
  specifically (not just the Dashboards API as a whole). (non-blocking: affects acceptance-test skip
  conditions, not the implementation)
