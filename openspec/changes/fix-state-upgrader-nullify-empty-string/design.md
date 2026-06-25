# Design: Fix state upgrader — nullify empty-string JSON fields

## Overview

Both `migrateComponentTemplateStateV0ToV1` and `migrateIndexTemplateStateV0ToV1` need
`stateutil.NullifyEmptyString` calls for the JSON-string fields that SDK state can persist as empty
strings: `template.mappings`, `template.settings`, and top-level `metadata`. The transform resource
upgrader (`internal/elasticsearch/transform/state_upgrade.go:55`) already uses this pattern:

```go
stateutil.NullifyEmptyString(stateMap, "metadata", "pivot", "latest")
```

## Change details

### `componenttemplate/state_upgrade.go`

In `migrateComponentTemplateStateV0ToV1`, after `stateutil.EnsureMapKeys(tmpl, ...)`:

```go
if tmpl, ok := stateMap[attrTemplate].(map[string]any); ok {
    stateutil.EnsureMapKeys(tmpl, attrAlias, attrMappings, attrSettings, attrDataStreamOptions)
    stateutil.NullifyEmptyString(tmpl, attrMappings, attrSettings)  // ADD
    aliasutil.NormalizeTemplateAliasesInV1State(tmpl)
}
```

Additionally, nullify top-level `metadata` when it arrives as `""`:

```go
stateutil.NullifyEmptyString(stateMap, "metadata")  // ADD — after the template block
```

### `template/state_upgrade.go`

Same addition in `migrateIndexTemplateStateV0ToV1`:

```go
if tmpl, ok := stateMap[attrTemplate].(map[string]any); ok {
    stateutil.EnsureMapKeys(tmpl, attrAlias, attrMappings, attrSettings, attrLifecycle, attrDataStreamOptions)
    stateutil.NullifyEmptyString(tmpl, attrMappings, attrSettings)  // ADD
    aliasutil.NormalizeTemplateAliasesInV1State(tmpl)
}
stateutil.NullifyEmptyString(stateMap, "metadata")  // ADD
```

## Testing strategy

### Unit tests

Add three test cases each to `state_upgrade_test.go` for both resources:

1. **`settings_only_empty_string_mappings`** — state with `"mappings": ""` and a valid `settings`
   JSON. Assert the upgraded state has `mappings = null`, `settings` preserved, and that the result
   decodes against the v1 schema without error.
2. **`mappings_only_empty_string_settings`** — state with a valid `mappings` JSON and
   `"settings": ""`. Assert the upgraded state has `settings = null`, `mappings` preserved, and that
   the result decodes against the v1 schema without error.
3. **`metadata_empty_string`** — state with `"metadata": ""` at top level. Assert the upgraded
   state has `metadata = null` and decodes cleanly.

### Acceptance tests

Add a settings-only variant to each resource's SDK upgrade acceptance test:

- **component template**: A new acceptance test `TestAccResourceComponentTemplateFromSDKSettingsOnly`
  (or an additional step in `TestAccResourceComponentTemplateFromSDK`) that creates the resource with
  provider `0.14.5` using only a `settings` block (no `mappings`), then upgrades to the Plugin
  Framework provider and asserts a no-op plan.
- **index template**: Same shape for `TestAccResourceIndexTemplateFromSDKSettingsOnly`.

The acceptance tests serve as the end-to-end regression gate for the reported issue.

## Validation questions

- Can the `index_template` empty-string bug be reproduced with a targeted acceptance test to confirm
  the symmetric fix works?

This question is non-blocking for implementation; the symmetric fix is correct regardless. The
settings-only index template SDK upgrade acceptance test should be used as the end-to-end validation
gate.

## Invariants

- `stateutil.NullifyEmptyString` is idempotent: calling it on a key that is already `nil` or absent
  is a no-op.
- The fix runs once, during the first `terraform plan` after the provider upgrade. Subsequent plans
  read state written by the Plugin Framework, which always emits `null` for unset JSON strings.
- No change is needed to `stateutil.NullifyEmptyString` itself; it already handles the empty-string
  case.
