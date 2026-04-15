## Context

`elasticstack_elasticsearch_watch` currently has a narrow acceptance suite that covers a basic create/update flow, but it does not exercise import, omitted defaults, or clearing a previously configured `transform`. That gap matters now because the current SDK implementation preserves a stale `transform` value when Elasticsearch no longer returns one, and the Plugin Framework migration should start from behavior that is already covered and correct.

## Goals / Non-Goals

**Goals:**
- Add focused acceptance coverage for import, defaulted `active`, defaulted `throttle_period_in_millis`, and removing `transform`.
- Fix watch refresh behavior so state stops retaining a stale `transform` after it has been removed remotely or via Terraform.
- Express the corrected watch behavior in OpenSpec before the migration change builds on it.

**Non-Goals:**
- Migrating the resource to the Terraform Plugin Framework.
- Redesigning the watch schema or changing the resource type, import format, or ID format.
- Expanding test coverage beyond the watch-specific regressions needed to de-risk the upcoming migration.

## Decisions

Clear `transform` from state when the API response omits it.
The current SDK read path keeps the old value, which makes state diverge from the remote watch after `transform` is removed. The fix should make refresh authoritative so subsequent plans can see and reconcile the drift.

Alternative considered: preserve the old state value when Elasticsearch omits `transform`.
Rejected because it leaves stale state in place and prevents acceptance tests from proving that a removed transform stays removed.

Add the new acceptance coverage before the Plugin Framework migration.
The new tests should protect the current behavior boundary first, then travel with the resource during migration.

Alternative considered: fold coverage additions into the migration change.
Rejected because it would blur whether a regression came from the behavior fix or the framework port.

Keep the change scoped to the existing watch acceptance layout.
The current `create` and `update` config directories already establish the main watch flow. The new cases should extend that suite with the smallest set of extra configs and checks needed to cover defaults, import, and removal.

## Risks / Trade-offs

- [Risk] Clearing `transform` on refresh could expose previously hidden drift for users whose state already contains a stale value. -> Mitigation: codify the new behavior in the delta spec and cover it with an acceptance test that removes `transform`.
- [Risk] Adding more acceptance steps can increase runtime and fixture complexity. -> Mitigation: keep the new cases narrowly focused on import, defaults, and transform removal instead of duplicating the existing broad update scenario.
- [Risk] The migration change could still need to revisit edge cases outside these tests. -> Mitigation: target the cases most likely to break during the SDK-to-PF port and treat this hardening change as a prerequisite for the migration.

## Migration Plan

1. Extend the current watch acceptance suite with import and omitted-default scenarios.
2. Add a regression test step that removes a previously configured `transform`.
3. Update watch read behavior so refresh clears `transform` when Elasticsearch no longer returns it.
4. Validate the watch-specific acceptance suite and use it as the baseline for the Plugin Framework migration.

## Open Questions

- None.
