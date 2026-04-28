## 1. Shared Core

- [x] 1.1 Add a new provider-wide `internal/resourcecore` package for Plugin Framework resource wiring.
- [x] 1.2 Define the typed `Component` constants for `elasticsearch`, `kibana`, `fleet`, and `apm`.
- [x] 1.3 Implement the resource core constructor, canonical `Configure`, `Metadata`, and `Client()` accessor.
- [x] 1.4 Document in package comments that `Component` is a type-name namespace and that `resourceName` is a literal suffix segment with no normalization.

## 2. Safety Coverage

- [x] 2.1 Add unit tests for type-name generation across `elasticsearch`, `kibana`, `fleet`, and `apm` components.
- [x] 2.2 Add conformance tests that verify embedded-core resources satisfy `resource.ResourceWithConfigure` without accidentally gaining `resource.ResourceWithImportState`.
- [x] 2.3 Preserve or add explicit concrete-resource interface assertions in pilot resources so promoted methods remain visible in review.

## 3. Trial Rollout

- [x] 3.1 Convert `internal/elasticsearch/ml/jobstate` to embed the shared core while preserving its existing import behavior and type name.
- [x] 3.2 Convert `internal/kibana/agentbuildertool` to embed the shared core while preserving its current `kibana_agentbuilder_tool` type-name suffix.
- [x] 3.3 Convert `internal/fleet/integration` to embed the shared core without adding import support or altering upgrade-state behavior.
- [x] 3.4 Convert `internal/apm/agent_configuration` to embed the shared core using the new `apm` component while preserving its Kibana-backed API logic.

## 4. Verification

- [x] 4.1 Run targeted tests for `internal/resourcecore` and each pilot resource package after conversion.
- [x] 4.2 Confirm the pilot resources retain their existing Terraform type names and import support boundaries.
- [x] 4.3 Reassess readability after the four-resource pilot and decide whether to continue embedded rollout, stop at the pilot, or revert to helper-only usage.
