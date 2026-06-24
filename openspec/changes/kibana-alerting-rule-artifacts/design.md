## Context

The Kibana alerting rule API accepts an `artifacts` object in the create (`POST /api/alerting/rule/{id}`) and update (`PUT /api/alerting/rule/{id}`) request bodies. The generated `kbapi` client already models this in `AlertingRuleAPIBodyGeneric.Artifacts` (dashboards list + investigation_guide.blob) and in `GetAlertingRuleIdResponse.JSON200.Artifacts`. Neither `internal/models.AlertingRule` nor the Terraform resource schema currently surface these fields.

**Version history (from Kibana source):**

| Capability | Minimum version | Kibana reference |
|------------|-----------------|------------------|
| Write (`artifacts` on POST/PUT) — dashboards | **8.19.0** / **9.1.0** | [PR #216292](https://github.com/elastic/kibana/pull/216292), backported to 8.19 as [PR #218920](https://github.com/elastic/kibana/pull/218920) |
| Write — `investigation_guide` | **8.19.0** / **9.1.0** | [PR #216377](https://github.com/elastic/kibana/pull/216377), backported to 8.19 as [PR #219943](https://github.com/elastic/kibana/pull/219943) |
| Read (`artifacts` on public GET / `_find`) | Documented from 8.19; **bugfix in 9.5.0** | Public GET/find omitted `artifacts` until [PR #247279](https://github.com/elastic/kibana/pull/247279) ([kibana#242792](https://github.com/elastic/kibana/issues/242792)); backports to older lines were in progress at research time |

On stacks between **8.19.0** and the GET fix, create/update with `artifacts` works but authoritative re-read after apply may not populate `artifacts` from the API. The provider MUST preserve configured values in state per REQ-048 when the API omits the field.

## Goals

- Expose `artifacts` on `elasticstack_kibana_alerting_rule` so practitioners can link Kibana dashboards and attach investigation guides to rules as code.
- Support **inline content** (`content`) and **file-path-based content** (`content_path`) for the investigation guide; the latter uses a SHA-256 checksum (computed at plan time) to detect when the source file changes outside Terraform.
- Map `artifacts` correctly on create, update, and read; when `artifacts` is absent from the Terraform config, omit it from the PUT body so Kibana does not wipe existing rule artifacts.

## Non-Goals

- Changing Kibana alerting rule behavior beyond what the API's `artifacts` object controls.
- Adding new Kibana API behavior (this tracks existing API semantics).
- Supporting dashboard creation from within this resource.

## Decisions

| Topic | Decision |
|-------|-----------|
| Schema shape | `artifacts` as a `schema.SingleNestedBlock` (needs nested blocks for the `dashboards` list). `dashboards` as `schema.ListNestedBlock`. `investigation_guide` as `schema.SingleNestedBlock`. |
| Dashboards | Each `dashboards {}` entry holds a single required `id` string attribute (Kibana dashboard saved-object id). Maps 1-to-1 to `artifacts.dashboards[].id` in the API. |
| Investigation guide — inline | `content` (optional string): passed directly to API as `blob`. On read, API-returned `blob` is stored back as `content` in state. |
| Investigation guide — file-based | `content_path` (optional string): provider reads the file at plan time, computes SHA-256, and at apply time sends file content as `blob`. `checksum` (computed string) is written to state after create/update. On the next plan the provider re-reads the file and compares the new checksum to state; if different, computed fields are marked unknown (same pattern as `elasticstack_fleet_custom_integration`). |
| Mutual exclusion | When `investigation_guide` is present, exactly one of `content` or `content_path` MUST be set. Enforced via `objectvalidator.ExactlyOneOf` or equivalent plan-time validation. |
| `checksum` in state | Only meaningful when `content_path` is used. Written to state at apply time (create or update). On read, `checksum` is not overwritten from the API (the API returns `blob`, not a checksum); the plan modifier manages drift. |
| Update when absent | When `artifacts` is not in the Terraform config, the provider SHALL omit `artifacts` from the PUT body so Kibana leaves existing rule artifacts unchanged. |
| Read path — content vs content_path | If prior state has `content` set (and `content_path` null): populate `content` from API `blob` after read. If prior state has `content_path` set (and `content` null): do not update `content` or `content_path` from the API; the plan modifier handles drift detection via `checksum`. |
| Plan modifier | A `ModifyPlan` hook on the resource (mirroring `elasticstack_fleet_custom_integration`) reads `content_path` during plan, computes SHA-256, and if it differs from the stored `checksum`, marks `checksum` (and potentially `id`) as unknown so Terraform shows a non-empty plan. |
| Version gate (write) | When `artifacts` is configured with known values, create/update SHALL fail on stack **&lt; 8.19.0** (8.x line) or **&lt; 9.1.0** (9.x line), mirroring the `flapping` / `flapping.enabled` split-gate pattern in `models.go`. Diagnostic text SHALL name both minimums. **9.0.x does not support `artifacts`.** |
| Public GET limitation | When the API omits `artifacts` on GET/find (known on 8.19–9.4.x before [PR #247279](https://github.com/elastic/kibana/pull/247279) backports land), the read path SHALL preserve prior known state rather than clearing `artifacts`. Document this in resource descriptions. |
| Acceptance-test gates | Write tests: skip below **8.19.0** or **9.1.0** (use version constraints like `flapping` tests). Read round-trip assertions (dashboard IDs / inline `content` from API): prefer gating at **≥ 9.5.0** unless CI runs a stack with the GET fix backported to an earlier minor. |

## Non-Goals (implementation)

- Do not add `content_path` support to the `investigation_guide.content` attribute; keep them as separate, mutually exclusive attributes.

## Risks / Trade-offs

- **Removing `artifacts` from Terraform config** still results in an update that omits `artifacts`; Kibana keeps any previously stored artifacts. After refresh, state may again show `artifacts` from the API while configuration omits the block, producing a **plan diff** until the practitioner aligns config with the API or clears rule artifacts outside Terraform. Document in resource docs.
- **External file changes**: When `content_path` is used, an external change to the file (without running `terraform plan`) will not be detected until the next `terraform plan` (at which point the plan modifier detects the checksum mismatch). This is consistent with the custom integration behavior and acceptable.
- **Content drift from Kibana**: Kibana may normalise the blob (e.g. trim whitespace). If so, `content`-based state may show perpetual drift. If this is observed during testing, consider trimming both sides before comparison or treating blob as opaque (read-only after first write). This should be investigated at implementation time.
- **GET does not return `artifacts` on some stacks**: Between 8.19.0 and the [PR #247279](https://github.com/elastic/kibana/pull/247279) fix (9.5.0+), public GET/find may omit `artifacts` even after a successful write. Refresh can show drift when config includes `artifacts` but the API response does not. Preserving prior state (REQ-048) avoids wiping practitioner config; document that `terraform refresh` may not reflect server-side artifacts until the stack includes the GET fix.

## Open Questions

1. **Blob normalisation**: Does Kibana modify the investigation guide blob (e.g. trim whitespace, normalise line endings) before storing? If so, read-path comparison for `content` may produce spurious drift. Investigate during acceptance testing.
2. **GET-fix backport matrix**: [PR #247279](https://github.com/elastic/kibana/pull/247279) was labeled `v9.5.0` with `backport:skip` on the PR itself, but maintainers noted backports were planned. Confirm which 8.x/9.x minors receive the fix before tightening acceptance-test skip thresholds below 9.5.0.

## Migration / State

- `artifacts` is a new optional block; no state upgrade is required. Existing resources without `artifacts` configured will have it absent from state after upgrade; if Kibana returns artifacts for those rules, a drift plan will appear (expected behaviour — practitioner must configure or ignore).
