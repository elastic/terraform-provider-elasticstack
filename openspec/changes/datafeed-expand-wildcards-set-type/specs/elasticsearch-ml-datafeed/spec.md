# Delta spec — `elasticsearch-ml-datafeed`: expand_wildcards set type

Capability: `elasticsearch-ml-datafeed`
Change: `datafeed-expand-wildcards-set-type`
Amends: `openspec/specs/elasticsearch-ml-datafeed/spec.md`

The attribute `indices_options.expand_wildcards` changes from `list(string)` to `set(string)` backed
by a custom type with semantic equality. Elasticsearch normalizes the shorthand token `"all"` to
`["open", "closed", "hidden"]` in GET responses. Without semantic equality, this causes a perpetual
plan diff whenever a user writes `expand_wildcards = ["all"]`.

The schema change in the HCL block is:

```
expand_wildcards = <optional+computed, set(string)>  # values: all, open, closed, hidden, none
```

Valid element values remain unchanged: `all`, `open`, `closed`, `hidden`, `none`.

---

## ADDED Requirements

### Requirement: expand_wildcards semantic equality (REQ-035)

The `indices_options.expand_wildcards` attribute SHALL use a custom set type (`ExpandWildcardsType` / `ExpandWildcardsValue`) that implements `basetypes.SetValuableWithSemanticEquals`. During plan, if the configured value and the prior state value are semantically equal under the rules below, the provider SHALL NOT produce a diff for this attribute.

Semantic equality rules:

1. If both values are null, they are semantically equal.
2. If both values are unknown, they are semantically equal.
3. If one is null or unknown and the other is not, they are not semantically equal.
4. Otherwise, normalize each value independently: expand the token `"all"` to `{"open", "closed", "hidden"}`; leave all other tokens (including `"none"`) as literals. Compare the two normalized token sets for equality without regard to order.

#### Scenario: `all` token does not produce a perpetual diff

- GIVEN `indices_options.expand_wildcards = ["all"]` in configuration
- WHEN the Elasticsearch API returns `expand_wildcards: ["open", "closed", "hidden"]` on read
- THEN Terraform SHALL NOT show a diff for `expand_wildcards` and SHALL NOT trigger an update

#### Scenario: Order-insensitive comparison

- GIVEN `indices_options.expand_wildcards = ["closed", "open"]` in configuration
- WHEN the Elasticsearch API returns `expand_wildcards: ["open", "closed"]` on read
- THEN Terraform SHALL NOT show a diff for `expand_wildcards`

#### Scenario: Partial expansion is not equal to `all`

- GIVEN `indices_options.expand_wildcards = ["all"]` in configuration
- WHEN the Elasticsearch API returns `expand_wildcards: ["open", "closed"]` on read
- THEN Terraform SHALL show a diff for `expand_wildcards` because `all` expands to `{open, closed, hidden}` which differs from `{open, closed}`

#### Scenario: `none` token compared literally

- GIVEN `indices_options.expand_wildcards = ["none"]` in configuration
- WHEN the Elasticsearch API returns `expand_wildcards: ["none"]` on read
- THEN Terraform SHALL NOT show a diff for `expand_wildcards`

#### Scenario: `none` is not equal to other tokens

- GIVEN `indices_options.expand_wildcards = ["none"]` in configuration
- WHEN the Elasticsearch API returns `expand_wildcards: ["open"]` on read
- THEN Terraform SHALL show a diff for `expand_wildcards`

---

## MODIFIED Requirements

### Requirement: expand_wildcards schema type (REQ-034, amended)

The `indices_options.expand_wildcards` attribute SHALL be declared as `schema.SetAttribute` with `CustomType: ExpandWildcardsType` (backed by `basetypes.SetType{ElemType: types.StringType}`). Element ordering SHALL NOT affect plan equality. The attribute SHALL continue to use `UseStateForUnknown` plan behavior, implemented via `setplanmodifier.UseStateForUnknown()`. Valid element values remain: `all`, `open`, `closed`, `hidden`, `none`.

#### Scenario: Configuration with unordered elements produces no diff

- GIVEN `indices_options.expand_wildcards = ["hidden", "open", "closed"]` in configuration
- WHEN the Elasticsearch API returns `expand_wildcards: ["open", "closed", "hidden"]` on read
- THEN Terraform SHALL NOT show a diff for `expand_wildcards`

#### Scenario: Unknown value preserved across plan

- GIVEN an existing datafeed with a known `expand_wildcards` set in state
- WHEN a plan is generated without changing `indices_options`
- THEN `expand_wildcards` SHALL remain known (not unknown) in the plan due to `UseStateForUnknown`

### Requirement: Mapping — API response to state for expand_wildcards (REQ-028, amended)

On read, when `indices_options.expand_wildcards` is non-empty in the API response, the resource SHALL store the values as an `ExpandWildcardsValue` set. When the API response contains an empty or nil `expand_wildcards`, the resource SHALL store a null `ExpandWildcardsValue`. The `IndicesOptions` model struct field `ExpandWildcards` SHALL be typed as `ExpandWildcardsValue` (not `types.List`). The `GetIndicesOptionsAttrTypes()` helper and all `map[string]attr.Type` usages for `"expand_wildcards"` SHALL reference `ExpandWildcardsType` instead of `types.ListType`.

#### Scenario: Non-empty expand_wildcards stored as set value

- GIVEN the Elasticsearch API returns `expand_wildcards: ["open", "closed"]` for a datafeed
- WHEN read runs
- THEN `expand_wildcards` in state SHALL contain the elements `"open"` and `"closed"` as a set

#### Scenario: Empty expand_wildcards stored as null

- GIVEN the Elasticsearch API returns `expand_wildcards: []` or omits the field
- WHEN read runs
- THEN `expand_wildcards` in state SHALL be null
