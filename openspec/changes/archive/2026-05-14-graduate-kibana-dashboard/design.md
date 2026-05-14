## Context

The Plugin Framework provider has two registration paths in `provider/plugin_framework.go`:

- `Provider.resources(...)` returns the always-registered Plugin Framework resources.
- `Provider.experimentalResources(...)` returns resources that are only appended when `p.version == AccTestVersion` or when `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true` is set in the environment.

Today `dashboard.NewResource` is returned from `experimentalResources()`, which means practitioners building against a released provider must opt in via the environment variable. The dashboard resource has matured: its capability is documented in `openspec/specs/kibana-dashboard/spec.md`, it has a large acceptance test suite under `internal/kibana/dashboard/acc_*_test.go`, and previous changes (e.g. `archive/2026-04-24-migrate-remaining-pf-resources-to-resourcecore`) already promoted it through internal refactors. The capability is ready to graduate.

The provider-side mechanism for graduation is a single-line registration move; the rest of the work is ensuring the documentation surface, capability spec, and changelog reflect that the resource is part of the standard surface.

## Goals / Non-Goals

**Goals:**

- Move `dashboard.NewResource` out of `experimentalResources()` into `resources()` in `provider/plugin_framework.go`.
- Update the `kibana-dashboard` capability spec to record that the resource is part of the standard Plugin Framework resource set.
- Regenerate `docs/resources/kibana_dashboard.md` from the resource schema so practitioners can read the documentation without enabling experimental resources locally.
- Add a CHANGELOG entry under the unreleased section.

**Non-Goals:**

- Graduating `elasticstack_kibana_stream`. It remains experimental.
- Any schema, behavioral, validation, or state changes for the dashboard resource.
- Removing `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL` plumbing or otherwise changing the experimental resource opt-in mechanism.
- Rewriting or restructuring the existing `kibana-dashboard` capability requirements beyond the registration section.

## Decisions

### D1. Graduate by registration move, not by introducing a separate flag

The provider's existing graduation pattern (see prior Kibana resources that have graduated) is to move the resource constructor from `experimentalResources()` into `resources()`. We follow the same pattern instead of introducing a per-resource flag.

Alternatives considered:

- *Per-resource opt-in flags*: rejected because it would diverge from the existing graduation pattern and the resource is already proven via the acceptance test suite.
- *Leaving the resource experimental but defaulting `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL` to true for dashboards*: rejected; it muddles the "experimental" contract for the still-experimental `streams` resource.

### D2. Update the `kibana-dashboard` spec with a registration requirement

The `kibana-stream` spec explicitly says it is "registered through the provider's experimental Plugin Framework resource set." The `kibana-dashboard` spec has no equivalent statement today. We add a symmetric requirement (standard, non-experimental registration) so the capability documents its registration status and so the spec stays in sync with `provider/plugin_framework.go`. This is the only spec modification.

Alternatives considered:

- *No spec change at all*: rejected because then the spec gives no signal about which provider surface the resource lives on, leaving the registration as an undocumented implementation detail.

### D3. Generate `docs/resources/kibana_dashboard.md` via existing `make docs-generate`

Documentation is generated from the resource schema using the existing `tfplugindocs`-based workflow described in `dev-docs/high-level/documentation.md`. We do not author the docs page by hand. We also do not add a custom template under `templates/resources/` for this resource; the default generator output is sufficient and matches the pattern used by most other Kibana resources.

### D4. CHANGELOG entry under the unreleased "Improvements" section

The CHANGELOG follows the project's existing format. The graduation is a practitioner-visible improvement and not a breaking change (configurations that previously worked with `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true` will continue to work without it).

## Risks / Trade-offs

- **[Risk]** Practitioners already using the resource via `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true` will not need the flag for dashboards anymore, but the flag is still required for `elasticstack_kibana_stream`. → **Mitigation:** CHANGELOG entry clearly says only the dashboard resource graduated; the `kibana-stream` spec already documents its experimental status.
- **[Risk]** Generated documentation may diverge from existing descriptions used elsewhere (the resource embeds descriptions via `descriptions/*.md`). → **Mitigation:** `make docs-generate` is deterministic from the schema; we run it as part of the change and commit the generated file.
- **[Risk]** The provider registry test (`provider/plugin_framework_entitycore_test.go`) iterates `p.Resources(ctx)` using `AccTestVersion`, which already exercises experimental resources, so the test continues to pass without modification. → **Mitigation:** No test change required; verify by running `go test ./provider/...` after the move.
- **[Trade-off]** We add a tiny "Provider registration" requirement to the dashboard spec. This couples the spec to a registration detail, but it mirrors the existing pattern in `kibana-stream` and prevents the spec from silently drifting from the provider surface.
