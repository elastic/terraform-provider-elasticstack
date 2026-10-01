## MODIFIED Requirements

### Requirement: Configuration validation (REQ-009)

The resource SHALL validate that the index named in `write_index.name` does not also appear in any entry of `read_indices`, either by literal name equality or, once expressions are resolved during plan and apply, as a concrete target of any `read_indices` expression. If the same index appears in both roles, the resource SHALL return an "Invalid Configuration" error diagnostic and SHALL NOT proceed to create or update. Literal-name validation is applied both at plan time (`ValidateConfig`) and at apply time before API calls.

The resource SHALL reject multi-target selector syntax in `write_index.name`: a name containing `*`, `?` or `,`, a name beginning with `-`, or the name `_all`. This validation SHALL be syntactic, SHALL be skipped for unknown values, and SHALL NOT reject single-target names that were previously accepted.

#### Scenario: Write index in read indices

- GIVEN `write_index.name` equals one of the `read_indices[*].name` values
- WHEN config is validated
- THEN the provider SHALL return an "Invalid Configuration" error diagnostic

#### Scenario: Write index matched by a read index expression

- GIVEN `write_index.name` is `logs-1` and a `read_indices` expression `logs-*` resolves to include `logs-1`
- WHEN plan or apply resolves the expressions
- THEN the provider SHALL return an "Invalid Configuration" error diagnostic

#### Scenario: Wildcard write index rejected

- GIVEN `write_index.name` is `logs-*`, `a,b`, `-a` or `_all`
- WHEN config is validated
- THEN the provider SHALL return an error diagnostic before any API call

#### Scenario: Concrete write index accepted

- GIVEN `write_index.name` is `my.index_1-v2`
- WHEN config is validated
- THEN no selector validation error SHALL be returned

### Requirement: Create (REQ-010–REQ-012)

On create, the resource SHALL convert `write_index` (if set) and all entries in `read_indices` (if set) into `add` alias actions and submit them in a single atomic Update Aliases API call. Each `read_indices[*].name` expression SHALL be resolved at apply time (not only at plan time) to its current concrete targets, and one `add` action SHALL be produced per resolved target. The `write_index` entry SHALL be submitted with `is_write_index: true`; `read_indices` entries SHALL be submitted without `is_write_index`. After a successful API call, the resource SHALL perform a read to refresh state before storing the final state.

#### Scenario: Atomic create

- GIVEN both `write_index` and `read_indices` are configured
- WHEN create runs
- THEN a single Update Aliases call SHALL include one `add` action per configured index

#### Scenario: Wildcard create attaches every match

- GIVEN `read_indices` contains `name = "traces-apm*"` and `traces-apm-default` and `traces-apm.rum-default` exist
- WHEN create runs
- THEN a single Update Aliases call SHALL include an `add` action for each of the two concrete indices and none for the pattern string
- AND the apply SHALL complete without an inconsistent-result error

#### Scenario: Index created between plan and apply

- GIVEN a matching index is created after plan but before apply
- WHEN create or update runs
- THEN the alias SHALL also be attached to that index

### Requirement: Update (REQ-013–REQ-015)

On update, the resource SHALL resolve each `read_indices[*].name` expression to concrete targets and compute a diff between the live alias associations (from the Get Alias API, keyed by concrete name) and the desired set of targets. Current members absent from the desired set SHALL be submitted as `remove` actions. Desired targets that are not current members, or are current members with changed settings, SHALL be submitted as `add` actions. Targets in both with identical settings SHALL be skipped. Selector strings SHALL NOT be used as diff keys or sent in `remove` actions. All resulting actions SHALL be submitted in a single atomic Update Aliases API call. If no actions are required, the Update Aliases API SHALL NOT be called. After a successful update (or when no actions were needed), the resource SHALL perform a read to refresh state.

If two expressions resolve to the same target with identical settings, the resource SHALL produce a single action for that target. If they resolve to the same target with different settings, the resource SHALL return a configuration error diagnostic and SHALL NOT call the Update Aliases API.

#### Scenario: Index removed from plan

- GIVEN an existing alias with two read indices
- WHEN one read index is removed from configuration
- THEN the update SHALL issue a `remove` action for the removed index and an `add` action (if changed) or no action for the unchanged index

#### Scenario: Unchanged wildcard produces no churn

- GIVEN an alias attached to every index matching `traces-apm*` with the configured settings
- WHEN update runs for an unrelated change
- THEN no `remove` or `add` action SHALL be issued for those indices

#### Scenario: Newly matching target is attached

- GIVEN a new index now matches a configured `read_indices` expression but is not attached to the alias
- WHEN update runs
- THEN an `add` action SHALL be issued for that index only

#### Scenario: Overlapping expressions with identical settings

- GIVEN two `read_indices` expressions resolve to the same index with identical settings
- WHEN update runs
- THEN exactly one `add` action SHALL be issued for that index

#### Scenario: Overlapping expressions with conflicting settings

- GIVEN two `read_indices` expressions resolve to the same index with different settings
- WHEN create or update runs
- THEN the resource SHALL return an error diagnostic and SHALL NOT call the Update Aliases API

### Requirement: Read (REQ-016–REQ-018)

On read, the resource SHALL call the Get Alias API with the alias name from state. If the API returns an empty result (no indices) or the alias name is not present in any returned index, the resource SHALL remove itself from state without an error. When alias data is returned, the resource SHALL classify each index as write or read based on the `is_write_index` flag from the API response and populate `write_index` and `read_indices` accordingly.

For each `read_indices` element in prior state or plan, the resource SHALL resolve its `name` expression and set the element's `concrete_indices` to the intersection of the resolved targets and the alias's actual members. Alias members that are not covered by any configured `read_indices` expression and are not the write index SHALL be returned as additional `read_indices` entries whose `name` and sole `concrete_indices` value are the concrete index name, so that the next plan removes them. When no prior configuration exists (for example immediately after import), every read member SHALL be returned in this singleton form.

#### Scenario: Alias not found on read

- GIVEN the Get Alias API returns an empty map or the alias name is missing
- WHEN read runs
- THEN the resource SHALL be removed from state

#### Scenario: State classification

- GIVEN the API returns two indices, one with `is_write_index: true`
- WHEN read runs
- THEN `write_index` SHALL hold the write index and `read_indices` SHALL hold the remaining index

#### Scenario: Wildcard element reconciles after read

- GIVEN `read_indices` contains `name = "traces-apm*"` and the alias is attached to both matching indices
- WHEN read runs
- THEN state SHALL contain a single `read_indices` element with `name = "traces-apm*"` and `concrete_indices` containing both indices

#### Scenario: Missing attachment is not hidden

- GIVEN `traces-apm*` resolves to two indices but the alias is attached to only one
- WHEN read runs
- THEN the element's `concrete_indices` SHALL contain only the attached index
- AND the next plan SHALL show an in-place update to attach the other

#### Scenario: Unconfigured member is drift

- GIVEN the alias is attached to an index not covered by any configured `read_indices` expression
- WHEN read runs
- THEN state SHALL include a singleton `read_indices` entry for that index
- AND the next plan SHALL remove it

### Requirement: Delete (REQ-019)

On delete, the resource SHALL read the alias's live associations with the Get Alias API and submit `remove` alias actions, in a single atomic Update Aliases API call, for every index currently associated with the alias. It SHALL NOT submit selector expressions or rely on state element names. If the alias has no current associations, the Update Aliases API SHALL NOT be called.

#### Scenario: Delete removes all indices

- GIVEN an alias with a write index and one read index
- WHEN delete runs
- THEN a single Update Aliases call SHALL include `remove` actions for both indices

#### Scenario: Delete with wildcard configuration

- GIVEN state holds `read_indices` with `name = "traces-apm*"` and the alias is attached to two concrete indices
- WHEN delete runs
- THEN the Update Aliases call SHALL include `remove` actions for the two concrete indices and none for `traces-apm*`

## ADDED Requirements

### Requirement: Read indices selector expressions

`read_indices[*].name` SHALL accept any Elasticsearch multi-target expression, including `*` and `?` wildcards, comma-separated lists, `-` exclusions and `_all`, and SHALL retain the user's original expression in state. The resource SHALL resolve expressions during plan (`ModifyPlan`), create, read and update using an explicit `expand_wildcards=all` policy (open, closed and hidden targets) and `allow_no_indices=true`. An expression with no current matches SHALL be valid and SHALL resolve to an empty set; it SHALL be re-resolved on later plans. Only resolved concrete indices and data streams SHALL be accepted as targets; if an expression resolves to an alias or to a remote-cluster target the resource SHALL return an error diagnostic identifying the target.

#### Scenario: No-match expression

- GIVEN a `read_indices` expression matches no indices
- WHEN create runs
- THEN the apply SHALL succeed and the element's `concrete_indices` SHALL be an empty set

#### Scenario: Later-created match is attached on next apply

- GIVEN an expression had no matches at create time and a matching index is created later
- WHEN the next plan runs
- THEN the plan SHALL show an in-place update, and applying it SHALL attach the alias to the new index

#### Scenario: Hidden and closed targets included

- GIVEN a closed index and a hidden index match an expression
- WHEN the expression is resolved
- THEN both SHALL be included as targets

#### Scenario: Comma, exclusion and `_all` expressions

- GIVEN expressions such as `a-*,-a-skip`, `a,b` and `_all`
- WHEN they are resolved
- THEN the resolved targets SHALL follow Elasticsearch multi-target semantics

#### Scenario: Alias or remote target rejected

- GIVEN an expression resolves to an alias or to a remote-cluster target
- WHEN plan, create or update runs
- THEN the resource SHALL return an error diagnostic naming that target

### Requirement: Computed concrete_indices

Each `read_indices` element SHALL expose a computed `concrete_indices` attribute of type `set(string)` containing the concrete indices and data streams currently attached to the alias for that element's `name` expression. During plan the resource SHALL populate it with the full set of currently resolved targets for the expression, so that a resolved but unattached target produces an in-place update. During read it SHALL contain only the intersection of the resolved targets and the alias's actual members. When `name` is unknown at plan time, `concrete_indices` SHALL be unknown.

State created before this attribute existed SHALL be handled without error: the initial refresh and plan SHALL succeed and converge to the same result as for newly created state.

#### Scenario: Plan populates concrete indices

- GIVEN `read_indices` contains `name = "traces-apm*"` and two indices match
- WHEN plan runs on create
- THEN the planned `concrete_indices` SHALL contain both indices

#### Scenario: Unattached resolved target plans an update

- GIVEN state records one attached index for an expression and a second matching index exists
- WHEN plan runs
- THEN the plan SHALL show an in-place update for that `read_indices` element

#### Scenario: Existing state without concrete_indices

- GIVEN state saved by a provider version without `concrete_indices`
- WHEN the new provider version refreshes and plans
- THEN no error SHALL occur and the plan SHALL be empty when membership matches the configuration

### Requirement: Out-of-band setting drift scope

Drift detection for `read_indices` SHALL be membership-based. An out-of-band change to `filter`, routing or hidden settings on an already attached concrete target SHALL NOT, by itself, produce a plan when membership is unchanged. Whenever an update is otherwise required, the resource SHALL compare the live alias associations and apply the configured settings to every resolved target whose live settings differ.

#### Scenario: Settings drift without membership change

- GIVEN an attached target's filter is changed out-of-band and membership is unchanged
- WHEN plan runs
- THEN no change SHALL be planned for that reason alone

#### Scenario: Settings repaired during another update

- GIVEN an attached target's filter was changed out-of-band
- WHEN an update is applied for another reason
- THEN an `add` action restoring the configured settings SHALL be issued for that target
