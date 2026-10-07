## MODIFIED Requirements

### Requirement: Settings mapping (REQ-025–REQ-027)

On create, the resource SHALL map each Terraform attribute to its corresponding Elasticsearch settings key (e.g. `mapping_total_fields_limit` → `mapping.total_fields.limit`) using dot-notation key conversion. Analysis settings (`analysis_analyzer`, `analysis_tokenizer`, `analysis_char_filter`, `analysis_filter`, `analysis_normalizer`) SHALL be parsed from JSON and nested under the `analysis` key in the create settings payload; these settings are applied only at index creation time and are not sent on update. When the deprecated `settings` block is also present, its `name`/`value` pairs SHALL be merged into the settings map; if any key conflicts with a dedicated attribute, the resource SHALL return an error diagnostic and SHALL not call the API.

The dynamic-setting attributes among these (every entry in `internal/elasticsearch/index.DynamicSettingsKeys`) SHALL be defined by the shared `GetDynamicSettingAttributes() map[string]schema.Attribute` function in `internal/elasticsearch/index/settings_keys.go`, merged into this resource's schema map, rather than hand-declared individually in `internal/elasticsearch/index/index/schema.go`. The resource's `tfModel` SHALL obtain the corresponding typed fields by anonymously embedding `internal/elasticsearch/index/dynamicsettings.Model`. This is an internal refactor: attribute names, types, descriptions, `Optional`/`Computed` flags, and plan-modifier/validator behavior for every dynamic-setting attribute SHALL remain unchanged from before the refactor (see REQ-UNCHANGED below). Static (creation-time-only) setting attributes are unaffected by this change and remain hand-declared in `schema.go`.

#### Scenario: Duplicate setting detected

- GIVEN a setting is defined both via a dedicated attribute and in the deprecated `settings` block
- WHEN the provider builds the settings map
- THEN it SHALL return a "duplicate setting definition" error diagnostic and SHALL not call the Create or Put Settings API

#### Scenario: Dynamic setting attribute behavior is unchanged after the schema refactor (REQ-UNCHANGED)

- GIVEN the dynamic-setting attributes are now sourced from `GetDynamicSettingAttributes()` instead of being hand-declared
- WHEN a configuration sets any `DynamicSettingsKeys`-derived attribute (e.g. `mapping_total_fields_limit`, `number_of_replicas`, `refresh_interval`)
- THEN the attribute's type, optionality, description, and validation behavior SHALL be identical to its pre-refactor behavior
- AND existing `elasticstack_elasticsearch_index` configurations SHALL plan with no unexpected diff solely due to this refactor
- AND existing unit and acceptance test coverage for these attributes SHALL pass unchanged

## ADDED Requirements

### Requirement: Shared reflection helpers support anonymous embedded struct fields

`internal/elasticsearch/index/reflectutil.go`'s `GetFieldValueByTagValue` and `SetFieldValueByTagValue` SHALL recurse into anonymous (embedded) struct fields when searching for a `tfsdk`-tagged field, in addition to the immediate struct's own fields, so that a `tfsdk`-tagged field defined on an anonymously-embedded struct (such as `dynamicsettings.Model`) is found and set identically to a field declared directly on the outer struct.

#### Scenario: Field on an anonymously embedded struct is found

- GIVEN a struct `Outer` that anonymously embeds a struct `Inner` with a field tagged `tfsdk:"example_key"`
- WHEN `GetFieldValueByTagValue` is called on an `Outer` value with `tagName = "example_key"`
- THEN it SHALL return the value of `Inner`'s `example_key` field and `true`

#### Scenario: Field on an anonymously embedded struct can be set

- GIVEN a struct `Outer` that anonymously embeds a struct `Inner` with a field tagged `tfsdk:"example_key"`
- WHEN `SetFieldValueByTagValue` is called on a pointer to `Outer` with `tagName = "example_key"` and a new `attr.Value`
- THEN `Inner`'s `example_key` field SHALL be updated to the new value and the function SHALL return `true`

#### Scenario: Existing flat (non-embedded) field lookup is unaffected

- GIVEN a struct with a directly-declared field tagged `tfsdk:"flat_key"` and no anonymous embedding
- WHEN `GetFieldValueByTagValue` or `SetFieldValueByTagValue` is called with `tagName = "flat_key"`
- THEN behavior SHALL be identical to before this change (found and get/set the direct field)
