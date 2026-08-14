## ADDED Requirements

### Requirement: Semantic equality for Watcher-injected JSON defaults (REQ-031)

The `input`, `transform`, `condition`, and `actions` attributes SHALL use a custom string type that performs semantic equality by populating known Elasticsearch-injected Watcher defaults on both the prior and new value before comparing them. Elasticsearch's Get Watch response injecting those defaults into JSON the practitioner omitted SHALL NOT be treated as a change.

The custom type SHALL be `customtypes.JSONWithDefaultsType[map[string]any]` with a single shared populate function for all four attributes. `trigger` and `metadata` SHALL remain `jsontypes.Normalized`.

Default-population SHALL recurse into JSON objects **and** JSON arrays so Watcher `chain.inputs` (an array of named wrapper objects) and any deeper nesting are visited. The populate function SHALL copy maps (and slices when descending into arrays) before inserting keys; it SHALL NOT mutate the tree it was given.

Wherever a JSON object has a `search` key whose value is an object containing a `request` object, the resource SHALL populate the following defaults on that `request` object when the corresponding key is absent:

- `rest_total_hits_as_int`: default `true`
- `search_type`: default `"query_then_fetch"`
- `indices`: default `[]`

These three keys SHALL be injected only on a `request` object reached via a `search` key. They SHALL NOT be injected into an `http` input's `request` object or any other `request` object.

Wherever a JSON object has a `script` key whose value is an object, the resource SHALL populate `lang: "painless"` on that object when `lang` is absent. The resource SHALL NOT treat arbitrary objects that merely contain `source` or `id` as scripts.

Wherever a JSON object has a `logging` key whose value is an object, the resource SHALL populate `level: "info"` on that object when `level` is absent. The resource SHALL NOT treat arbitrary objects that merely contain `level` as logging actions.

Default-population SHALL only fill a key that is **absent**; a key already present (including explicit `rest_total_hits_as_int: false`, a non-default `search_type`, a non-empty `indices` array, a non-painless `lang`, or a non-info logging `level`) SHALL be left unchanged.

Default-population SHALL be used exclusively for the Plugin Framework's semantic-equality comparison (`StringSemanticEquals`), which runs after create, update, and read, and during plan. It SHALL NOT be applied as a mutation inside `fromAPIModel`. `fromAPIModel` SHALL continue to marshal the Get Watch (redaction-merged) payload into the resource model as required by REQ-023–REQ-027. When `StringSemanticEquals` reports equality, the Framework SHALL keep the prior value in Terraform state (the plan on apply, prior state on refresh), so practitioner-authored JSON is preserved. Semantic equality SHALL NOT treat the redacted sentinel `::es_redacted::` as equal to a concrete secret; redaction preservation SHALL remain the responsibility of `fromAPIModel` and SHALL run before the custom type is constructed.

On import, when Terraform has no prior plan value, the resource MAY store the API JSON including injected defaults. A subsequent plan against configuration that omits only those defaults SHALL be empty.

#### Scenario: Omitted rest_total_hits_as_int does not cause apply inconsistency

- **GIVEN** a watch `input` configured as a `search` request whose `request` object omits `rest_total_hits_as_int`
- **WHEN** the resource is created and the Get Watch response used to refresh state includes `rest_total_hits_as_int: true` injected by Elasticsearch
- **THEN** the apply SHALL succeed without a "Provider produced inconsistent result after apply" error
- **AND** a subsequent `terraform plan` SHALL show no changes
- **AND** Terraform state SHALL keep the practitioner's `input` JSON (without the injected key)

#### Scenario: Omitted search_type does not cause apply inconsistency

- **GIVEN** a watch `input` configured as a `search` request whose `request` object omits `search_type`
- **WHEN** the resource is created and the Get Watch response used to refresh state includes `search_type: "query_then_fetch"` injected by Elasticsearch
- **THEN** the apply SHALL succeed without a "Provider produced inconsistent result after apply" error
- **AND** a subsequent `terraform plan` SHALL show no changes

#### Scenario: Omitted indices does not cause apply inconsistency

- **GIVEN** a watch `input` configured as a `search` request whose `request` object omits `indices`
- **WHEN** the resource is created and the Get Watch response used to refresh state includes `indices: []` injected by Elasticsearch
- **THEN** the apply SHALL succeed without a "Provider produced inconsistent result after apply" error
- **AND** a subsequent `terraform plan` SHALL show no changes

#### Scenario: transform search request defaults

- **GIVEN** a watch `transform` configured as a `search` request whose `request` object omits `rest_total_hits_as_int`, `search_type`, and `indices`
- **WHEN** the resource is created and the Get Watch response used to refresh state includes those keys injected by Elasticsearch
- **THEN** the apply SHALL succeed without a "Provider produced inconsistent result after apply" error
- **AND** a subsequent `terraform plan` SHALL show no changes

#### Scenario: Chained search sub-inputs each default independently

- **GIVEN** a watch `input` configured as a `chain` input whose `inputs` array contains two or more named sub-inputs, each a `search` request omitting `rest_total_hits_as_int`, `search_type`, and/or `indices`
- **WHEN** the resource is created and Elasticsearch injects those defaults independently into each chained search request's `request` object
- **THEN** the apply SHALL succeed without a "Provider produced inconsistent result after apply" error
- **AND** a subsequent `terraform plan` SHALL show no changes

#### Scenario: Action-level search transform defaults

- **GIVEN** a watch `actions` JSON containing a search transform whose `request` object omits `rest_total_hits_as_int`, `search_type`, and/or `indices`
- **WHEN** the resource is created and Elasticsearch injects those defaults into that nested `request` object
- **THEN** the apply SHALL succeed without a "Provider produced inconsistent result after apply" error
- **AND** a subsequent `terraform plan` SHALL show no changes

#### Scenario: Omitted script lang does not cause apply inconsistency

- **GIVEN** a watch `condition` or `transform` configured as a `script` whose script object omits `lang`
- **WHEN** the resource is created and the Get Watch response includes `lang: "painless"` injected by Elasticsearch
- **THEN** the apply SHALL succeed without a "Provider produced inconsistent result after apply" error
- **AND** a subsequent `terraform plan` SHALL show no changes

#### Scenario: Omitted logging action level does not cause apply inconsistency

- **GIVEN** a watch `actions` JSON containing a logging action whose `logging` object omits `level`
- **WHEN** the resource is created and the Get Watch response includes `level: "info"` injected by Elasticsearch
- **THEN** the apply SHALL succeed without a "Provider produced inconsistent result after apply" error
- **AND** a subsequent `terraform plan` SHALL show no changes

#### Scenario: Explicit non-default values are preserved and still detected as real changes

- **GIVEN** a watch `input` search request that explicitly sets `rest_total_hits_as_int = false`, `search_type = "dfs_query_then_fetch"`, and a non-empty `indices` array
- **WHEN** the resource is created and read back
- **THEN** the state SHALL preserve the practitioner's explicit values unchanged
- **AND** a subsequent `terraform plan` SHALL show no changes
- **AND** if the practitioner subsequently changes any of those values in configuration, the next plan SHALL show that change

#### Scenario: HTTP input request is not given search-request defaults

- **GIVEN** a watch `input` configured as an `http` input (which has a `request` object that is not a Watcher search request)
- **WHEN** semantic equality comparison runs
- **THEN** no `rest_total_hits_as_int`, `search_type`, or `indices` defaults SHALL be injected onto that `http.request` object
- **AND** existing key-order/whitespace-insensitive comparison SHALL be unchanged for that input

#### Scenario: Import then plan with omitted defaults is empty

- **GIVEN** a watch is imported and first state stores Get Watch JSON including injected search-request defaults
- **AND** configuration omits those defaults
- **WHEN** `terraform plan` runs
- **THEN** the plan SHALL show no changes on `input` / `transform` / `condition` / `actions` for those omitted defaults
