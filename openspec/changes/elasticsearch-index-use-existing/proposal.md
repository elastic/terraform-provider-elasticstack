## Why

`elasticstack_elasticsearch_index` always treats create as "create a brand-new index". When an index with the same name already exists at create time, the Create Index API returns `resource_already_exists_exception` and the resource fails. This affects two real scenarios:

- **Replacement race (issue [#966](https://github.com/elastic/terraform-provider-elasticstack/issues/966))** — Changing a static setting like `mapping_coerce` forces replacement. Between the destroy and the subsequent create, the index can be re-created out-of-band by indexing load or by an `auto_create_index` template, so the create then fails. Today there is no provider-side mitigation.
- **Adopting an existing index** — Operators sometimes want to manage an index that was created out-of-band (bootstrap scripts, a different team, or an index template that pre-creates the concrete index) without first running `terraform import`.

Both cases need the same primitive: an opt-in fallback during create that detects an existing index and brings it under management instead of failing.

## What Changes

- Add a new optional boolean attribute `use_existing` to `elasticstack_elasticsearch_index` (defaults to `false`).
- When `use_existing = true` and the configured `name` is a static index name, the create flow SHALL look up the index by name and, if it exists, adopt it: build a synthetic prior state from the Get Index API response and run the resource's existing update logic to reconcile aliases, dynamic settings, and mappings against the plan.
- Adoption SHALL be strict on static settings: if the existing index's static settings differ from any static setting explicitly set in config, the resource SHALL fail with an error diagnostic listing every mismatch and SHALL NOT mutate the cluster.
- Adoption SHALL surface a warning diagnostic whenever it occurs, so practitioners can see in apply output that an existing index was adopted rather than created.
- When `use_existing = true` but the configured `name` is a date math expression, the resource SHALL skip the existence check, emit a warning explaining `use_existing` is ignored for date math names, and fall through to the normal create path.
- When `use_existing = false` (the default), the resource SHALL behave exactly as it does today.
- After adoption the resource is fully managed by Terraform: subsequent reads, updates, and deletes follow the existing flows (including `deletion_protection` semantics).

Out of scope:

- Tolerating `resource_already_exists_exception` when `use_existing = false` (no implicit fallback).
- Race-window handling between the existence check and the subsequent create call (the existing 409 surfaces unchanged).
- A separate "release on destroy" / non-destructive removal mode — adopt implies full ownership.
- Adopting indices via wildcards, aliases, or date math expressions.

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `elasticsearch-index`: extend the create flow to optionally adopt an existing index by name, gated by the new `use_existing` attribute.

## Impact

- `internal/elasticsearch/index/index/schema.go` for the new `use_existing` attribute.
- `internal/elasticsearch/index/index/models.go` to track `use_existing` on the Terraform model.
- `internal/elasticsearch/index/index/create.go` for the new pre-create existence check and adopt branch, including static-setting strict validation, mapping/alias/dynamic-setting reconciliation via the existing update helpers, and warning diagnostics.
- `internal/elasticsearch/index/index/update.go` to expose its dynamic-settings, mappings, and aliases reconciliation helpers (or share them) so the create-time adopt path can reuse them.
- `internal/elasticsearch/index/index/acc_test.go` and `testdata/` for acceptance coverage of adopt success, static-setting mismatch error, and the date-math warning fallback.
- `internal/elasticsearch/index/index/models_test.go` (or new unit tests) for static-setting comparison behavior.
- `openspec/specs/elasticsearch-index/spec.md` via the delta spec for the new requirement and the modified create-flow requirement.
- `docs/resources/elasticsearch_index.md` regenerated to document the new attribute.
- `CHANGELOG.md` entry referencing issue #966.
