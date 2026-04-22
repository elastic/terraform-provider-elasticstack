## Why

The Plugin Framework ILM resource models each phase and each ILM action as **list nested blocks** capped at one element (`listvalidator.SizeBetween(0, 1)`). That matches legacy SDK `MaxItems: 1` behavior but **generated documentation** describes them as unbounded “Block List” instead of a single nested object, unlike the SDK resource.

Separately, with **single nested block** semantics, an empty action block (e.g. `forcemerge {}`) is a present object with unset child attributes. Marking those children `Required: true` produces awkward validation compared to **omitting the block**; the provider should use **optional attributes** plus **object-level `AlsoRequires`** so “block present ⇒ required fields set” is enforced without requiring the block when the action is unused.

## What Changes

- **Schema**: Replace list-nested phases and actions with **`SingleNestedBlock`** in `internal/elasticsearch/index/ilm/schema.go` and `schema_actions.go` (remove list max-one validators where obsolete). Keep **`elasticsearch_connection`** as the shared list nested block.
- **Model / read / expand**: Retype phase and action fields to **`types.Object`** (and object-shaped nested attrs), updating `attr_types.go`, `models.go`, `validate.go`, `policy.go`, `flatten.go`, `value_conv.go`, `model_expand.go` so API expansion and flatten behavior stay equivalent.
- **State**: Bump resource **schema version**; implement **`ResourceWithUpgradeState`** with JSON migration **v0 → v1** that unwraps list-shaped state only for known phase and action keys (not `elasticsearch_connection`). Add a **unit test** for the upgrader.
- **Action validation**: For `forcemerge`, `searchable_snapshot`, `set_priority`, `wait_for_snapshot`, and `downsample`, make previously **required** attributes **optional** and attach **`objectvalidator.AlsoRequires`** on the parent action object (or `SingleNestedBlock` validators) for the former required paths.
- **Tests**: Update **`acc_test.go`** flat state attribute paths (remove `.0` segments for single-nested values); leave **`testdata/**/*.tf`** unchanged.
- **Docs**: Regenerate Terraform resource docs after schema change.
- **OpenSpec**: Delta under `specs/elasticsearch-index-lifecycle/spec.md` adds normative requirements for the above; sync into `openspec/specs/elasticsearch-index-lifecycle/spec.md` when the change is applied.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- **`elasticsearch-index-lifecycle`**: Schema shape (single nested blocks), state upgrade, action validation rules, and spec text aligned with implementation.

## Impact

- **Users**: Saved state upgrades automatically on first apply after upgrade; HCL using `hot { }` / `forcemerge { ... }` remains valid. Flat state keys in `terraform show` and in tests change (no `hot.0.*` list indices for phases/actions).
- **Code**: `internal/elasticsearch/index/ilm/` (schema, model, flatten, expand helpers, resource, new `state_upgrade.go`, tests).
- **Maintenance**: Clearer generated docs (“nested block” vs list); behavior matches legacy SDK cardinality.
