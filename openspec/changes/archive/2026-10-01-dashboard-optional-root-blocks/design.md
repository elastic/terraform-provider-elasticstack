## Context

The Dashboards API treats `time_range`, `refresh_interval`, and `query` as optional. The resource currently requires them. Create (`dashboardToAPICreateRequest`) already uses pointer fields and is nil-safe. Update (`dashboardToAPIUpdateRequest`) is not, because the PUT request body type is generated from a local overlay in `generated/kbapi/dashboard-paths.json` whose `required` list includes these blocks and `options`.

Verified API behavior (Kibana 9.4.2 and 9.6.0-SNAPSHOT):

| Probe | Result |
|---|---|
| POST with `title` only | blocks absent in response |
| GET of that dashboard | blocks absent |
| PUT with `title` only | 200, blocks absent |
| PUT with all three set, then PUT with `title` only | blocks reset to absent (PUT is a full replace) |
| PUT with zero-value blocks (current behavior for nil blocks) | 400 on `query.language` |
| `time_range.mode` sent on PUT | not returned on GET (existing REQ-009 handling applies) |

The only version difference is that 9.6 returns `description: ""` when omitted, already handled. Kibana 9.5 was not tested.

## Goals / Non-Goals

**Goals**
- A title-only `elasticstack_kibana_dashboard` configuration is valid, applies, re-plans empty, and imports without drift.
- Null blocks are omitted from POST and PUT payloads.

**Non-Goals**
- Provider-side defaults or default detection for the three blocks.
- Overlay/path drift beyond the PUT `required` list; `query` wire-shape changes; list/search endpoints.

## Decisions

1. **Plain `Optional`, no `Computed`, no read-side null-preservation.** Kibana applies no defaults and returns the blocks as absent, so the existing nil mapping in `models.go` (`TimeRange`, `RefreshInterval`, `Query` set to nil when the API omits them) already yields a stable round-trip. `time_range.mode` preservation (REQ-009) is retained when `time_range` is present.
2. **Removal clears, it does not reset.** PUT is a full replace and Kibana has no defaults for these blocks, so removing a block from configuration makes it absent. Docs say so explicitly.
3. **Overlay fix lands in the same change.** Set the PUT `required` list in `generated/kbapi/dashboard-paths.json` to `["title"]`, regenerate `kibana.gen.go` per `dev-docs/high-level/generated-clients.md`, and put the generated diff in its own commit for reviewability. `dashboardToAPIUpdateRequest` is then adjusted for pointer fields on `TimeRange`, `RefreshInterval`, `Query`, and `Options` (set only when the corresponding model block is non-nil).
4. **Schema.** Flip the three attributes to `Optional: true`; call `timeRangeSingleNestedAttribute(..., false)` for `time_range`. Nested attributes stay `Required`. No schema version bump, no state upgrader (required to optional is non-breaking).
5. **Examples.** Existing examples stay explicit. `panel_no_time_range.tf` drops its now-unneeded root blocks, and a new title-only example is added.
6. **Panel-level interactions.** The root blocks are only read in `models.go`. Panel-level `use_time_range` lives in the per-panel packages and does not touch them. It was not probed against Kibana, so the implementation SHOULD cover a dashboard panel with `use_time_range` on a title-only dashboard if cheap, otherwise note it as unverified.

## Risks / Trade-offs

- **A future Kibana may begin returning defaults for these blocks.** With plain `Optional`, that would cause "inconsistent result after apply" for configs that omit them. Mitigation: the title-only acceptance test is the tripwire; a follow-up would add null-preservation modelled on `isDashboardOptionsDefaultSet`.
- **Kibana 9.5 was not probed.** The acceptance test runs across the supported version matrix and will surface any difference.
- **Overlay edit requires client regeneration.** Mitigated by keeping the generated diff in its own commit and limiting the overlay edit to the PUT `required` list.
- **Existing PUT callers.** Any other caller constructing `PutDashboardsIdJSONBody` must be updated for the pointer fields; the compiler will flag them.

## Migration Plan

No migration. Existing configurations (which set all three blocks) and state remain valid. No state version change.

## Deviation from the research comment

The implementation-research comment recommended Approach A *with* read-side default detection (null-preserving Kibana defaults, mirroring `isDashboardOptionsDefaultSet`). The follow-up comment on the issue from the maintainer, after live verification, narrowed this to Approach A *without* the default-detection helper because Kibana returns no defaults for these blocks. This design follows the maintainer's agreed path. The research comment's remaining open questions were answered in that follow-up (see Decisions above).

## Open questions

None remaining. Both were answered by CI on the PR: the new acceptance tests, including panel-level `use_time_range` on a title-only dashboard (`_titleOnlyPanelUseTimeRange`), passed on Kibana 9.5.4 and 9.6.0-SNAPSHOT.
