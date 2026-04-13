## Context

`elasticstack_elasticsearch_index` currently treats `name` as both the configured input and the durable Elasticsearch identity. That works for static names, but it breaks down for date math expressions because Elasticsearch resolves the path to a concrete index name during creation.

The current implementation has three related problems:

- schema validation assumes every valid index name fits the static lowercase-name regex
- create/update identity is computed from the configured `name`, not the concrete index Elasticsearch created
- read repopulates `name` from the looked-up index name, which would overwrite the user's configured date math expression and create perpetual drift

This change needs to preserve existing behavior for static names while adding a separate, intentionally narrow path for plain date math names.

## Goals / Non-Goals

**Goals:**
- Accept plain date math index names without weakening validation for ordinary static names.
- Encode accepted date math names inside the provider before calling the Create Index API.
- Preserve `name` as the configured user intent in state.
- Persist the concrete server-side index identity in a computed `concrete_name` attribute and in `id`.
- Ensure read, update, and delete operations target the concrete index name after creation.
- Keep imported and legacy state readable by deriving `concrete_name` from `id` when needed.

**Non-Goals:**
- Requiring practitioners to URI-encode date math expressions in Terraform configuration.
- Generalizing the resource to support aliases or wildcards as the primary identity.
- Redesigning the resource to manage rollover behavior beyond the initial concrete index created from a date math expression.
- Replacing the existing static name validator with a single broad regex that tries to cover all cases.

## Decisions

Use split validation with explicit regex branches.
The schema should keep the existing non-regex checks in place and replace the current single allowed-characters regex with an explicit `stringvalidator.Any(...)` branch:

```go
stringvalidator.Any(
  stringvalidator.RegexMatches(
    regexp.MustCompile(`^[a-z0-9!$%&'()+.;=@[\]^{}~_-]+$`),
    indexNameAllowedCharsMessage,
  ),
  stringvalidator.RegexMatches(
    regexp.MustCompile(`^<[^<>]*\{[^<>]+\}[^<>]*>$`),
    dateMathIndexNameMessage,
  ),
)
```

The first regex preserves the current static-name character rules. The second regex intentionally stays shape-based rather than trying to fully parse date math syntax: it requires angle brackets around the whole value and at least one `{...}` section inside, which is enough to distinguish date math expressions from normal static names before provider-side URI encoding.

Alternative considered: broaden the current static-name regex to also match date math expressions.

Rejected because it would either become unreadably permissive or encode too much date math syntax in one regex. Keeping separate validators preserves the simple static rules and makes the date math path explicit.

Encode date math names at the API boundary.
The provider should accept plain date math syntax in `name`, validate it as such, and URI-encode it immediately before constructing the Create Index API path. The encoded transport detail should not leak into Terraform configuration or state.

Alternative considered: require users to submit already-encoded date math names.

Rejected because it makes the resource harder to author and review, and it turns an HTTP transport detail into part of the user-facing schema contract.

Track configured and concrete names separately.
The resource should keep `name` as the configured value and add computed `concrete_name` for the concrete index returned or targeted by Elasticsearch. `id` should use the concrete index name, not the configured expression.

Alternative considered: overwrite `name` with the concrete index name after create.

Rejected because it loses the user's declared configuration and creates a permanent mismatch between config and state.

Capture the concrete name during create, not by guessing from a later read.
The create flow should parse the Create Index API response and store the returned `index` field as the concrete name. That concrete name becomes the basis for `id` and later CRUD calls.

Alternative considered: create with the date math expression and then discover the concrete name by calling Get Index with the original expression.

Rejected because Get Index response keys are keyed by the concrete index name, so a helper that looks up the response by the requested expression can miss the created index entirely.

Use concrete identity for all post-create operations.
Update, read, and delete should target the persisted concrete index name from `id` / `concrete_name`, never recomputing identity from `plan.name`.

Alternative considered: continue to use `plan.name` for update operations and only switch read/delete to concrete identity.

Rejected because alias, settings, and mappings updates would still fail or drift when `plan.name` is a date math expression.

Backfill for import and legacy state.
When imported or pre-change state lacks `concrete_name`, read should derive it from `id.ResourceID`. If `name` is absent in state, read may backfill it from that same concrete name so import remains usable without inventing a date math expression.

Alternative considered: require a state upgrader or special import format before enabling the feature.

Rejected because the resource already has a durable composite id that contains the needed concrete name.

## Risks / Trade-offs

- [Risk] The date math validator could accept malformed plain expressions if it is too loose -> Mitigation: keep it purpose-built, test representative valid and invalid date math inputs, and avoid replacing the existing static-name validation rules.
- [Risk] Provider-side encoding could accidentally change static names or double-encode date math names -> Mitigation: confine encoding to the validated date math path and add focused tests for the exact request path sent to Elasticsearch.
- [Risk] Existing helper code assumes requested index name and returned map key are the same -> Mitigation: capture the concrete name from create responses and add targeted tests around read-after-create behavior.
- [Risk] Imported resources do not preserve an original date math expression -> Mitigation: document that import restores the concrete index identity; only resources created through Terraform retain the original configured expression in `name`.
- [Risk] Users may assume the resource tracks future rollover generations -> Mitigation: keep the contract explicit that `concrete_name` represents the concrete index managed by this resource instance.

## Migration Plan

1. Add the new schema attribute and split validator while keeping static-name validation behavior unchanged for existing configs.
2. Update create to persist the concrete index name returned by Elasticsearch and compute `id` from that value.
3. Update read, update, and delete to operate on the persisted concrete identity.
4. Add targeted unit and acceptance coverage for validation, stable state after apply, and post-create updates.

## Open Questions

- None. The main remaining work is implementation and regression coverage.
