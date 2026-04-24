## Why

The provider now has a shared Plugin Framework `resourcecore`, but most remaining Plugin Framework resources still hand-roll the same `Configure` and `Metadata` boilerplate. That duplication keeps issue [#2454](https://github.com/elastic/terraform-provider-elasticstack/issues/2454) alive, makes provider-wide wiring fixes harder to roll out consistently, and creates avoidable review noise across dozens of resources.

## What Changes

- Migrate the remaining compatible Plugin Framework resources that still duplicate canonical provider-data conversion and Terraform type-name construction to embed `internal/resourcecore.Core`.
- Preserve each resource's existing Terraform type name, import behavior, CRUD logic, schema, and state-upgrade behavior while removing only the repeated `client` field plus `Configure` and `Metadata` wiring.
- Add a provider-package unit test that iterates the resources registered in `provider/plugin_framework.go` and verifies the registered Plugin Framework resources embed `resourcecore.Core`.
- Add or extend verification so the broader rollout proves `resourcecore` works across the remaining compatible resource shapes without changing user-facing behavior.
- Document the broader rollout in the `provider-framework-resource-core` capability so the shared-core contract covers provider-wide adoption instead of only the original pilot resources.

## Capabilities

### New Capabilities
<!-- None -->

### Modified Capabilities
- `provider-framework-resource-core`: expand the shared-core capability from a pilot embedding to the remaining compatible Plugin Framework resources that currently duplicate canonical `Configure` and `Metadata` wiring

## Impact

- Affected code primarily under `internal/fleet/` and `internal/kibana/`, plus any additional Plugin Framework resource packages found during implementation to already match the canonical `resourcecore` semantics.
- A new provider-package registry test, plus any additional targeted tests and/or compile-time conformance checks for the widened rollout.
- No Terraform schema, type-name, import-identifier, or API contract changes are intended for end users.
