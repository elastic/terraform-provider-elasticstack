## Context

`elasticstack_elasticsearch_cluster_settings` gained `terraform import` support during the Terraform Plugin Framework migration (commit `4f6f0914`, released in v0.11.x). The `ImportState` handler is `resource.ImportStatePassthroughID` on the `id` attribute, and an acceptance test step (`TestAccResourceClusterSettings/import`) validates the full import path. The implementation is complete and correct.

The only remaining gap is that the Terraform Registry documentation page has no **## Import** section, because `tfplugindocs` generates this section exclusively from an `import.sh` example file, and no such file exists under `examples/resources/elasticstack_elasticsearch_cluster_settings/`.

This resource is a **singleton**: exactly one instance exists per cluster, identified by `<cluster_uuid>/cluster-settings`. The `Read` callback only surfaces settings already tracked in Terraform state; it does not auto-populate all live cluster settings. This means that after a successful `terraform import`, the state will contain only the composite `id` — no `persistent` or `transient` blocks. The `import.sh` comment MUST explain this so users know to add the desired settings to their configuration before running `terraform plan`.

## Goals / Non-Goals

**Goals:**
- Add `examples/resources/elasticstack_elasticsearch_cluster_settings/import.sh`.
- Regenerate `docs/resources/elasticsearch_cluster_settings.md` via `make docs-generate` to include the `## Import` section.
- Sync REQ-020 from the delta spec into the main spec.

**Non-Goals:**
- Modifying the `Read` callback to auto-populate all live cluster settings on import (Approach B from research — deferred as a separate enhancement).
- Adding `ImportStateVerify` to the acceptance test (the current comment in `acc_test.go` notes it is intentionally absent because the read only returns settings already in state).
- Any change to the `persistent`/`transient` schema, validation, or Go implementation.

## Decisions

- **Spec sync.** The existing REQ-006 in the main spec covers import at a functional level, but it currently references the SDKv2 helper `schema.ImportStatePassthroughContext`; the implementation uses the Plugin Framework `resource.ImportStatePassthroughID`. REQ-020 is added to the delta spec to formally require the `import.sh` documentation example. This surfaces the documentation gap in the spec history.
## Risks / Trade-offs

- None material. The change is additive (one new file) and the regenerated doc is deterministic from the existing Go source and example files. No Go code changes means no risk of regressions.

## Open questions

1. Should the `import.sh` comment explain how to discover `<cluster_uuid>` (e.g., via `elasticstack_elasticsearch_info` data source or `GET /`)? Chosen answer: yes — include both options as a brief comment, consistent with the pattern used in `elasticstack_elasticsearch_script/import.sh`.
2. Is the absence of `ImportStateVerify` in the acceptance test intentional and permanent, or is Approach B desired for a future release? The answer does not change this change's scope. A follow-up issue can be opened if Approach B is desired.
