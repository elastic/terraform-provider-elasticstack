## Why

The `elasticstack_kibana_slo` resource still depends on a separately generated SLO OpenAPI client under `generated/slo`, while the rest of the Kibana surface is consolidating on the shared `generated/kbapi` client and `internal/clients/kibanaoapi` helpers. Maintaining two generators and client stacks for the same Kibana SLO HTTP API increases drift risk, complicates authentication and transport wiring, and blocks removal of the legacy `generated/slo` tree.

## What Changes

- Introduce `kibanaoapi` helper functions for Kibana SLO operations: find (list/search), get (single SLO with summary), create, update, and delete, built on `generated/kbapi` request/response types and consistent with existing `kibanaoapi` patterns (typed errors, diagnostics, `kbn-xsrf` where required).
- Replace imports of `github.com/elastic/terraform-provider-elasticstack/generated/slo` in `internal/kibana/slo` and `internal/models/slo.go` with the equivalent SLO models from `generated/kbapi` (including discriminated unions for indicators and `group_by` wire encoding).
- Migrate `internal/clients/kibana/slo.go` to call the new helpers using the shared `kibanaoapi.Client` (backed by `kbapi.ClientWithResponses`) instead of `slo.SloAPI` from `generated/slo`.
- Remove or stop building the legacy `generated/slo` client for SLO once all references are gone (follow-up cleanup as part of the same change where feasible).
- Preserve existing Terraform-visible behavior and **all** documented version gates for `prevent_initial_backfill`, `data_view_id`, `group_by`, and multi-value `group_by` (no relaxation of minimum stack versions or schema semantics).

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- `kibana-slo`: Add normative implementation requirements that SLO HTTP traffic uses the shared `kbapi` client via `kibanaoapi` helpers, without altering user-facing resource behavior or version-gated attributes.

## Impact

- **New / updated Go**: `internal/clients/kibanaoapi/slo.go` (or similarly named) for find/get/create/update/delete helpers; refactors in `internal/clients/kibana/slo.go`, `internal/models/slo.go`, and `internal/kibana/slo/*` model wiring and tests.
- **Client factory / scoped client**: `internal/clients/kibana_scoped_client.go`, `internal/clients/provider_client_factory.go`, and `internal/clients/api_client.go` — remove or replace `GetSloClient` / `buildSloClient` paths that bind `generated/slo`; align SLO auth with whatever mechanism `kbapi` + `kibanaoapi` use for other resources (may subsume `SetSloAuthContext` or reimplement equivalent credential injection).
- **Generated code**: Prefer `generated/kbapi` types for SLO; deprecate `generated/slo` for provider use.
- **Specs**: Delta under `openspec/changes/migrate-kibana-slo-to-kbapi/specs/kibana-slo/spec.md` documenting the implementation contract; canonical `openspec/specs/kibana-slo/spec.md` unchanged until archive/sync.
- **Tests**: `internal/clients/kibana/slo_test.go` and `internal/kibana/slo/*_test.go` updated for new types and client entry points.
