## Context

The `elasticstack_kibana_synthetics_private_location` resource exposes `space_id` for space-scoped Synthetics Private Location API calls. That integration is only available from Elastic Stack **9.4.0-SNAPSHOT** onward (aligned with other 9.4-gated features in this provider, e.g. `elasticstack_kibana_stream`). Fleet agent policies gained `space_ids` at **9.1.0**, so acceptance tests that only check `agentpolicy.MinVersionSpaceIDs` are insufficient: they run on 9.1–9.3 where private-location space scoping is still unsupported.

## Goals / Non-Goals

**Goals:**

- Fail fast with a clear diagnostic when a practitioner uses a non-default space for this resource on a stack below **9.4.0-SNAPSHOT**.
- Reuse the existing `APIClient.EnforceMinVersion` pattern used by Fleet and Streams.
- Skip `TestSyntheticPrivateLocationResource_nonDefaultSpace` unless the stack is at least **9.4.0-SNAPSHOT**.
- Document the minimum version on the `space_id` attribute (embedded markdown), consistent with other version-gated resources.

**Non-Goals:**

- Changing default-space behavior for stacks before 9.4 (empty `space_id` remains valid).
- Bumping or relaxing Fleet’s own `MinVersionSpaceIDs` constant.

## Decisions

1. **Minimum version string: `9.4.0-SNAPSHOT`**  
   Matches `internal/kibana/streams` (`minVersionStreams`) and the user’s “9.4-SNAPSHOT” requirement. `go-version` compares prereleases consistently with existing tests.

2. **When to enforce**  
   Call `EnforceMinVersion` on create, read, and delete when the **effective** Kibana space for API calls is non-default: `effectiveSpaceID(...)` is non-empty after resolving `space_id` and optional composite `id` (import) segments. Empty string means default space and does not trigger the gate. If composite import ever yields the literal segment `default`, treat it as default space (normalize so we do not require 9.4 for that edge case).

3. **Acceptance test skip**  
   Use `versionutils.CheckIfVersionIsUnsupported` with a package-level `version.Must(version.NewVersion("9.4.0-SNAPSHOT"))` (exported constant in `privatelocation` or a local test var mirroring the implementation constant) so the test tracks the same floor as runtime checks. This supersedes relying solely on `agentpolicy.MinVersionSpaceIDs` (9.1.0) for this test.

4. **Alternatives considered**  
   - *Schema-only documentation*: insufficient; users still hit opaque API errors.  
   - *Validators without API client*: Terraform validate cannot know cluster version; runtime enforcement is required.

## Risks / Trade-offs

- **[Risk]** Practitioners on 9.1–9.3 with a genuine need for Fleet `space_ids` but not yet on 9.4 for Synthetics private locations may be blocked from using `space_id` on this resource — **Mitigation**: diagnostic explains the minimum version; default space continues to work.

- **[Risk]** Duplicate version constants if not exported from one place — **Mitigation**: single exported `MinVersionSpaceID` (or similarly named) in `privatelocation`, referenced from tests.

## Migration Plan

- No state migration: existing configurations targeting default space are unchanged.
- After upgrade, stacks below 9.4 that incorrectly used non-default `space_id` will fail with an explicit error instead of a raw API error; upgrade the stack or remove `space_id` to use the default space.

## Open Questions

- None for this change; the 9.4.0-SNAPSHOT floor is fixed by product constraint.
