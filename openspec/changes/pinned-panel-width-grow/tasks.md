## 1. `pinned_panels[]` entries accept and apply `width`/`grow`

- [x] 1.1 Add optional `width` (`stringvalidator.OneOf("small", "medium", "large")`) and `grow`
      (bool) attributes to `pinnedPanelsNestedObject()`
      (`internal/kibana/dashboard/schema_control_configs.go`), alongside `type` and the existing
      `*_config` blocks.
- [x] 1.2 Add matching `Width types.String` / `Grow types.Bool` fields to `PinnedPanelModel`
      (`internal/kibana/dashboard/models/panel.go`). Do not add these fields to `PanelModel` or
      `SyntheticPanel()`.
- [x] 1.3 Change `ControlPinnedHandler.AfterRemapToGroup`'s signature from `func(group *G)` to
      `func(ppm models.PinnedPanelModel, group *G)` and update the `ToAPI` call site
      (`internal/kibana/dashboard/panelkit/control_pinned_handler.go`) to pass `ppm` through. Update
      the existing `esqlcontrol/pinned.go` callback for the new signature.
- [x] 1.4 Add an `AfterRemapToGroup` callback to `optionslist/pinned.go`, `rangeslider/pinned.go`, and
      `timeslider/pinned.go` that sets `group.Width`/`group.Grow` from `ppm.Width`/`ppm.Grow`.
- [x] 1.5 Add a symmetric typed read callback to `ControlPinnedHandler`, invoked in `FromAPI` right
      after `AsGroup(raw)`, and wire it in all four `panel/*/pinned.go` files to read
      `group.Width`/`group.Grow` back onto the `PinnedPanelModel` being built.

## 2. `width`/`grow` round-trip correctly on refresh and import

- [x] 2.1 In the new read callback, when `prior` is nil (import or a missing index), set
      `width`/`grow` directly from the API's `Width`/`Grow` pointers; when any prior exists at that
      index (even with a different control type), use `PreserveString`/`PreserveBool` against
      `prior.Width`/`prior.Grow`.
- [x] 2.2 Unit-test `pinned_panels_mapping_test.go` (or the per-control-type model tests) for all four
      control types covering: omitted `width`/`grow` on refresh with a non-null API default (stays
      null, no diff), a configured `small`/`false` or `large`/`true` value round-tripping unchanged,
      drift when the API value changes while a same-type prior exists, a different control type at
      the same list position preserving a null prior and reading API values for a configured prior, and the no-prior/import branch.
- [x] 2.3 Add or extend acceptance test coverage in
      `internal/kibana/dashboard/acc_pinned_panels_test.go` (using
      `testdata/TestAccResourceKibanaDashboardPinnedPanels/...`) for at least the options-list control
      type: a create step with `width = "small"` / `grow = false`, an update step changing to `width =
      "large"` / `grow = true`, a step omitting both attributes that produces no diff on a subsequent
      plan, and an `ImportStateVerify` pass asserting both configured values import correctly.
- [x] 2.4 Confirm all four pinned control types (options list, range slider, ES|QL, time slider)
      expose the same `width`/`grow` behavior, gated at the existing `MinDashboardAPISupport` version
      check; no additional version gate is required unless the open question in `design.md` resolves
      otherwise.

## 3. Documentation

- [x] 3.1 Document the new `width` and `grow` attributes on `pinned_panels[]` in the resource
      documentation (generated provider docs / schema `MarkdownDescription`), noting the enum values
      and defaults (`medium` / `false`) applied server-side when omitted.
