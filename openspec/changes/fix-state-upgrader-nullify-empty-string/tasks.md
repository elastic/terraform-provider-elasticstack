# Tasks

## TASK-1: Fix `componenttemplate` state upgrader

File: `internal/elasticsearch/index/componenttemplate/state_upgrade.go`

In `migrateComponentTemplateStateV0ToV1`, after the existing `stateutil.EnsureMapKeys` call inside
the `if tmpl, ok := stateMap[attrTemplate].(map[string]any); ok` block, add:

```go
stateutil.NullifyEmptyString(tmpl, attrMappings, attrSettings)
```

After the `template` block (and outside the `if` guard), add:

```go
stateutil.NullifyEmptyString(stateMap, "metadata")
```

The resulting block should look like:

```go
if tmpl, ok := stateMap[attrTemplate].(map[string]any); ok {
    stateutil.EnsureMapKeys(tmpl, attrAlias, attrMappings, attrSettings, attrDataStreamOptions)
    stateutil.NullifyEmptyString(tmpl, attrMappings, attrSettings)
    aliasutil.NormalizeTemplateAliasesInV1State(tmpl)
}

stateutil.NullifyEmptyString(stateMap, "metadata")
```

Place the `NullifyEmptyString(stateMap, "metadata")` call before the `version` cleanup block that
follows.

## TASK-2: Fix `template` (index template) state upgrader

File: `internal/elasticsearch/index/template/state_upgrade.go`

Same change as TASK-1, applied to `migrateIndexTemplateStateV0ToV1`. In the second
`if tmpl, ok := stateMap[attrTemplate].(map[string]any); ok` block (after all the
`CollapseListPath` calls), after `stateutil.EnsureMapKeys`, add:

```go
stateutil.NullifyEmptyString(tmpl, attrMappings, attrSettings)
```

After the `template` block, add:

```go
stateutil.NullifyEmptyString(stateMap, "metadata")
```

## TASK-3: Add unit tests to `componenttemplate/state_upgrade_test.go`

Add two test cases to the `TestComponentTemplateUpgradeState_template_path` table (or a new table
function):

1. **`settings_only_empty_string_mappings`**: Input state has `template` as a single-element list
   containing `{"mappings": "", "settings": "{\"index\":{\"number_of_replicas\":\"1\"}}"}`.
   Assert: upgraded state has `template.mappings == nil` (null), `template.settings` is preserved,
   and `requireUpgradedStateDecodes` passes.

2. **`metadata_empty_string`**: Input state has `"metadata": ""` at top level with no `template`
   block. Assert: upgraded state has `metadata == nil` (null) and `requireUpgradedStateDecodes`
   passes.

## TASK-4: Add unit tests to `template/state_upgrade_test.go`

Mirror of TASK-3 for `migrateIndexTemplateStateV0ToV1`. Add:

1. **`settings_only_empty_string_mappings`**: Same as TASK-3 case 1 but for index template state
   shape (include `composed_of`, `index_patterns` in base state as the resource requires them).
2. **`metadata_empty_string`**: Same as TASK-3 case 2, for index template state.

Reference the existing `runUpgrade` / `requireUpgradedStateDecodes` helpers (or equivalents) in
that package.

## TASK-5: Add acceptance test — settings-only component template SDK upgrade

File: `internal/elasticsearch/index/componenttemplate/acc_from_sdk_test.go`

Add a new acceptance test `TestAccResourceComponentTemplateFromSDKSettingsOnly`:

- **Step 1** (SDK 0.14.5): Create a component template with only a `settings` block (no `mappings`,
  no `alias`). Verify `template.0.settings` is set.
- **Step 2** (Plugin Framework): Re-apply. Assert no error, `template.settings` is set, and
  `template.mappings` is empty/null.
- **Step 3** (no-op plan): `PlanOnly: true, ExpectNonEmptyPlan: false`.

Create the required Terraform configuration in
`internal/elasticsearch/index/componenttemplate/testdata/TestAccResourceComponentTemplateFromSDKSettingsOnly/config/`
mirroring the layout of the existing `TestAccResourceComponentTemplateFromSDK` test data.

## TASK-6: Add acceptance test — settings-only index template SDK upgrade

File: `internal/elasticsearch/index/template/acc_from_sdk_test.go`

Add a new acceptance test `TestAccResourceIndexTemplateFromSDKSettingsOnly` mirroring TASK-5 for
index templates:

- **Step 1** (SDK 0.14.5): Create an index template with `index_patterns` and only a `settings`
  block (no `mappings`, no `data_stream`).
- **Step 2** (Plugin Framework): Re-apply and assert no error.
- **Step 3** (no-op plan).

Create the required Terraform configuration in
`internal/elasticsearch/index/template/testdata/TestAccResourceIndexTemplateFromSDKSettingsOnly/`.
