## 1. Schema (`internal/kibana/dashboard/panel/links/schema.go`)

- [x] 1.1 Create the `links` package under `internal/kibana/dashboard/panel/links/`.
- [x] 1.2 Define `SchemaAttribute()` returning `panelkit.PanelConfigBlock(...)` with `panelType = "links"` and `Required: true` (REQ-LINKS-001).
- [x] 1.3 Implement the inner attributes map:
  - `by_value`: `SingleNestedAttribute` (optional) with sub-attributes:
    - `layout`: required string, `stringvalidator.OneOf("horizontal", "vertical")` (REQ-LINKS-001).
    - `title`, `description`: optional strings.
    - `hide_title`, `hide_border`: optional bools.
    - `links`: required `ListNestedAttribute`, `listvalidator.SizeAtLeast(1)` (REQ-LINKS-001), with link item attributes per 1.4.
  - `by_reference`: `SingleNestedAttribute` (optional) with sub-attributes:
    - `ref_id`: required string, `stringvalidator.LengthAtLeast(1)` (REQ-LINKS-001).
    - `title`, `description`: optional strings.
    - `hide_title`, `hide_border`: optional bools.
  - Both blocks carry `objectvalidator.ConflictsWith(...)` pointing at the peer block (REQ-LINKS-001).
- [x] 1.4 Define link item nested attributes (flat, all at same level):
  - `type`: required string, `stringvalidator.OneOf("dashboard", "external")` (REQ-LINKS-001).
  - `destination`: required string.
  - `label`: optional string.
  - `open_in_new_tab`: optional bool.
  - `use_filters`: optional bool (dashboard only — no schema guard here; enforced by validator).
  - `use_time_range`: optional bool (dashboard only).
  - `encode_url`: optional bool (external only).
- [x] 1.5 Implement `linksConfigModeValidator` (by_value / by_reference mutual exclusion) and attach to `links_config` block as an `ExtraValidators` entry (REQ-LINKS-001).
- [x] 1.6 Implement `linksItemTypeValidator` on each link item object enforcing type-specific field isolation (REQ-LINKS-001):
  - `type = "dashboard"` → `encode_url` must be null/unknown.
  - `type = "external"` → `use_filters` and `use_time_range` must be null/unknown.
  - Skip when `type` is null or unknown.

## 2. Model (`internal/kibana/dashboard/panel/links/model.go`)

- [x] 2.1 Define `linksModel` (top-level `links_config` attributes: `ByValue`, `ByReference`).
- [x] 2.2 Define `linksByValueModel` (`Layout`, `Title`, `Description`, `HideTitle`, `HideBorder`, `Links`).
- [x] 2.3 Define `linksByReferenceModel` (`RefID`, `Title`, `Description`, `HideTitle`, `HideBorder`).
- [x] 2.4 Define `linkItemModel` (`Type`, `Destination`, `Label`, `OpenInNewTab`, `UseFilters`, `UseTimeRange`, `EncodeURL`).
- [x] 2.5 Define `AttrTypes()` helpers for each model struct (required for `types.ObjectType` construction and `types.Object.As()`).

## 3. API mapping (`internal/kibana/dashboard/panel/links/api.go` and `populate.go`)

- [x] 3.1 Implement `Handler.ToAPI` (REQ-LINKS-001):
  - Reject `config_json` via `panelkit.RejectConfigJSON`.
  - Branch on `links_config.by_value` vs `links_config.by_reference`.
  - For `by_value`: build `KibanaHTTPAPIsKbnDashboardPanelTypeLinksConfig0`; for each link item translate `type` value (`"dashboard"` → `"dashboardLink"`, `"external"` → `"externalLink"`) and populate the appropriate union variant.
  - For `by_reference`: build `KibanaHTTPAPIsKbnDashboardPanelTypeLinksConfig1` with `ref_id`.
  - Set optional display fields only when non-null.
- [x] 3.2 Implement `Handler.FromAPI` using `panelkit.SimpleFromAPI` and a `populateLinksPanelFromAPI` helper (REQ-LINKS-001):
  - Detect branch by inspecting whether `RefId` is set on the config.
  - Map API discriminator values back to Terraform enum (`"dashboardLink"` → `"dashboard"`, `"externalLink"` → `"external"`).
  - Apply REQ-009 null-preservation: carry null from prior state for display fields when user omitted them (REQ-LINKS-001).
  - Map optional link item fields; leave null in state when absent from API response (REQ-LINKS-001).
- [x] 3.3 Implement all remaining `iface.Handler` methods:
  - `PanelType()` returns `"links"`.
  - `SchemaAttribute()` delegates to `schema.go`.
  - `ClassifyJSON` returns `false`.
  - `PopulateJSONDefaults` returns config unchanged.
  - `PinnedHandler` returns `nil`.
  - `AlignStateFromPlan` is a no-op.
  - `ValidatePanelConfig` returns an error when `links_config` is absent for a `links` panel.

## 4. Registry (`internal/kibana/dashboard/registry.go`)

- [x] 4.1 Add `links.Handler{}` to the `panelHandlers` slice (REQ-LINKS-001).
- [x] 4.2 Add the `links` import to the file's import block.

## 5. Tests

- [x] 5.1 Unit test for `linksConfigModeValidator` — both branches set, neither set, only `by_value`, only `by_reference`.
- [x] 5.2 Unit test for `linksItemTypeValidator` — `type = "dashboard"` with `encode_url`, `type = "external"` with `use_filters`/`use_time_range`, valid cases.
- [x] 5.3 Unit tests for `ToAPI` — `by_value` round-trip with both link types; `by_reference` round-trip; null display fields preserved.
- [x] 5.4 Unit tests for `FromAPI` — `by_value` branch; `by_reference` branch; REQ-009 null-preservation for display fields; optional link item fields null in state when absent.
- [x] 5.5 Acceptance test `TestAccKibanaDashboard_LinksPanel_ByValue` (REQ-LINKS-001):
  - Creates a dashboard with a `links` panel, `by_value`, one `dashboard` link and one `external` link.
  - Verifies plan idempotency, apply, import.
- [x] 5.6 Acceptance test `TestAccKibanaDashboard_LinksPanel_ByReference` (REQ-LINKS-001):
  - Creates a dashboard with a `links` panel, `by_reference`.
  - Verifies plan idempotency, apply, import.
- [x] 5.7 Run `go test ./internal/kibana/dashboard/...` (unit), `make build`, and `go vet ./...`.

## 6. Spec sync

- [x] 6.1 Add the `links_config` schema entry to `openspec/specs/kibana-dashboard/spec.md` (a new numbered REQ referencing REQ-LINKS-001 from this delta spec).
- [x] 6.2 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate kibana-dashboard-links-panel --type change` and fix any issues.
- [x] 6.3 Run `make check-openspec`.
