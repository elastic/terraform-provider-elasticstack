## Context

The `elasticstack_elasticsearch_enrich_policy` resource uses the Terraform Plugin Framework and stores its identity as `<cluster_uuid>/<policy_name>`. All other resource attributes (`name`, `policy_type`, `indices`, `match_field`, `enrich_fields`, `query`, `execute`) require replacement, so the resource is effectively immutable once created.

The `execute` attribute is Terraform-only (not returned by the Elasticsearch Get enrich policy API). On normal read, it retains the value persisted from the previous create — the framework reads existing state then calls `Read`, which does not overwrite `Execute`. On import, however, the initial state has only `id` set; after `Read`, `execute` would remain null. Because `execute` has `RequiresReplace`, a subsequent `terraform plan` would see null-in-state vs. default-true-in-plan and mark the resource for replacement — defeating the purpose of the import.

## Goals / Non-Goals

**Goals:**
- Allow `terraform import elasticstack_elasticsearch_enrich_policy.<name> <cluster_uuid>/<policy_name>` to succeed.
- Ensure that a `terraform plan` immediately after import shows no diff (no spurious replacement).
- Add acceptance test coverage for the import step.

**Non-Goals:**
- Changing the ID format.
- Adding import support to the enrich policy data source (data sources do not support import).
- Detecting the actual historical value of `execute` from Elasticsearch (not available from the API).

## Decisions

### Use a custom `ImportState` method rather than `ImportStatePassthroughID` alone

`ImportStatePassthroughID` copies the raw import ID into the `id` attribute and leaves all other attributes null. For most resources this is fine because `Read` repopulates everything from the API. For the enrich policy, `execute` is not in the API response. Leaving it null in state after import means the next plan will compare null (state) with the default `true` (plan), triggering a spurious replacement.

**Decision**: Implement a custom `ImportState` that:
1. Validates and copies the import ID into the `id` attribute (`resource.ImportStatePassthroughID`).
2. Also sets `execute = true` in the response state — matching the computed default.

**Alternative considered**: Set `execute` to null and document that a plan diff is expected after import. Rejected: users would not understand why import forces a replacement of a perfectly healthy policy.

**Alternative considered**: Use `ImportStatePassthroughID` only, and teach `Read` to default `execute` to `true` when it finds a null value. Rejected: would silently mutate state during normal refresh, which is a broader change with more surface area. The custom `ImportState` keeps the fix isolated.

## Risks / Trade-offs

- [Risk]: A user imports a policy that was originally created with `execute = false`. The import will set `execute = true` in state. If they later specify `execute = false` in config, a replacement will be planned. → **Mitigation**: Document in the import guidance that `execute` defaults to `true` on import. Because `execute` is `RequiresReplace` anyway, the user must explicitly configure it; the imported default of `true` is correct for almost every real policy (one that has been executed and has an enrich index).

## Migration Plan

Additive change only. No existing state is affected; the new `ImportState` method is only invoked during `terraform import`.
