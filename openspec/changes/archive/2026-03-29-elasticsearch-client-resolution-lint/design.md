## Context

Elasticsearch resources and data sources under `internal/elasticsearch/**` can reach Elasticsearch through two kinds of sink calls: package-level helper functions in `internal/clients/elasticsearch` that accept `*clients.APIClient`, and direct method calls on `*clients.APIClient`. The provider already has approved helper flows for constructing these clients, but the current protection is social rather than enforced, which makes it easy for new code to bypass the helper path with provider-meta casts or direct construction.

The analyzer target is `analysis/esclienthelper`, wired through repository lint execution. Because client values can be delegated through small wrapper functions, the design needs to balance strictness with false-positive control by combining a narrow sink definition, type-based analysis, and interprocedural provenance facts.

## Goals / Non-Goals

**Goals:**

- Enforce approved `*clients.APIClient` origins specifically at concrete sink call sites in Elasticsearch entity code.
- Accept helper-derived wrappers either through an explicit allowlist or through analyzer-exported facts that prove helper-derived returns.
- Prefer conservative failure when provenance cannot be proven, while keeping diagnostics actionable for contributors.
- Make the rule part of normal lint and CI execution, with regression tests covering compliant and non-compliant cases.

**Non-Goals:**

- Rewriting entity connection handling or changing the approved helper APIs themselves.
- Performing broad style linting outside the defined sink set.
- Proving arbitrary dataflow across patterns the analyzer cannot model precisely.

## Decisions

- **Sink-based enforcement**: Report only when an in-scope sink consumes a non-derived client, rather than flagging every suspicious assignment. This keeps the analyzer focused on behavior that can actually reach Elasticsearch and reduces noisy intermediate findings.
- **Type-driven sink detection**: Identify sinks using type information for `*clients.APIClient`, package/function identity, and method receivers rather than variable names. This directly addresses false-positive risk from naming-only heuristics.
- **Dual wrapper policy**: Support both an explicit allowlist keyed by fully qualified function name and analyzer facts for inferred wrappers/factories. The allowlist covers intentionally blessed helpers, while facts cover straightforward delegated helper wrappers without growing static configuration for every local function.
- **Conservative provenance model**: If a client value cannot be proven helper-derived, treat it as non-compliant at the sink. This prevents bypass patterns from slipping through because of analysis uncertainty.
- **Actionable diagnostics**: Diagnostics should explain that the sink uses a non-helper-derived client and point developers to approved helpers such as `clients.NewAPIClientFromSDKResource(...)` and `clients.MaybeNewAPIClientFromFrameworkResource(...)`.

## Risks / Trade-offs

- **[Risk] Conservative analysis flags wrappers the analyzer cannot yet prove** -> Mitigation: support both explicit allowlisting and exported provenance facts, and keep diagnostics clear about how to remediate.
- **[Risk] Sink matching drifts as helper packages evolve** -> Mitigation: define sinks in terms of package/type identity and maintain analyzer tests for both helper-function and receiver-call cases.
- **[Risk] Interprocedural facts miss complex flows** -> Mitigation: scope fact inference to relevant return behavior, prefer false negatives over unsound approval, and add targeted tests for supported wrapper patterns.

## Migration Plan

1. Add the new `elasticsearch-client-resolution-lint` delta spec that captures sink scope, approved sources, diagnostics, CI expectations, and regression coverage.
2. Implement or update `analysis/esclienthelper` to enforce helper-derived provenance at the defined sinks, including wrapper allowlisting and fact export/import.
3. Wire the analyzer into repository lint execution so `make check-lint` fails on violations.
4. Add analyzer tests covering compliant direct-helper usage, violating bypass paths, wrapper allowlist behavior, and fact-proven wrappers.

## Open Questions

- None.
