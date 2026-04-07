# Proposal: Require `searchable_snapshot` in the ILM `frozen` phase

## Why

Issue [#482](https://github.com/elastic/terraform-provider-elasticstack/issues/482) reports that the provider's generated documentation and schema shape currently imply that `frozen.searchable_snapshot` is optional, but Elasticsearch requires the `frozen` phase to mount a searchable snapshot.

Today, practitioners can write a `frozen {}` block that passes provider-side validation and only fails later when Elasticsearch rejects the policy. That creates a poor Terraform experience:

- the generated docs are misleading for the `frozen` phase
- invalid configuration survives plan-time validation
- the real requirement is only discovered from an API error during apply

This should be expressed directly in the Terraform schema and requirements so the provider rejects invalid `frozen` phase configuration before any API call and the generated docs describe the block correctly.

## What Changes

- **Require `frozen.searchable_snapshot` at the schema level** whenever the `frozen` phase block is present.
- **Extend ILM config validation** so a `frozen` phase without `searchable_snapshot` is rejected before calling the Elasticsearch ILM API.
- **Add a new `elasticsearch-index-lifecycle` requirement** describing the frozen-phase constraint and its plan-time validation behavior.
- **Update generated documentation and acceptance coverage** so docs and tests both reflect that `searchable_snapshot` is mandatory for the `frozen` phase.

## Capabilities

### New Capabilities

- Practitioners receive a provider validation error when they declare `frozen {}` without a `searchable_snapshot` block.
- Generated resource documentation describes `frozen.searchable_snapshot` as required within the `frozen` phase instead of implying it is optional.

### Modified Capabilities

- `elasticsearch-index-lifecycle`: frozen phase validation becomes stricter and aligned with Elasticsearch API requirements.

## Impact

- **Scope:** `internal/elasticsearch/index/ilm/`, generated docs for `elasticstack_elasticsearch_index_lifecycle`, and `openspec/specs/elasticsearch-index-lifecycle/spec.md` via delta spec.
- **User impact:** invalid `frozen` configurations fail earlier, during Terraform validation/planning rather than at apply time.
- **Compatibility:** additive validation only for configurations that are already invalid against Elasticsearch; valid configurations are unchanged.
