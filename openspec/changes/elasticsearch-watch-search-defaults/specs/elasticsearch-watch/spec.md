## ADDED Requirements

### Requirement: Semantic equality for Watcher search-request defaults on `input` and `transform` (REQ-031)

The `input` and `transform` attributes SHALL use a custom string type that performs apply-time semantic equality by populating known Elasticsearch-injected Watcher search-request defaults on both the prior and new value before comparing them, so that Elasticsearch's Get Watch response injecting these defaults into a search request the practitioner did not set them on SHALL NOT be treated as a change.

Wherever a JSON object within a parsed `input` or `transform` value has a `search` key whose value is itself an object containing a `request` object, the resource SHALL populate the following defaults on that `request` object when the corresponding key is absent:

- `rest_total_hits_as_int`: default `true`
- `search_type`: default `"query_then_fetch"`

This default-population SHALL be applied recursively at any nesting depth, so it covers `request` objects nested inside `chain` input sub-inputs (each chained `search` sub-input SHALL be defaulted independently) as well as the top-level `input.search.request` / `transform.search.request` case.

Default-population SHALL only fill a key that is **absent** from the `request` object; a key already present (including an explicit `false` for `rest_total_hits_as_int` or a non-default `search_type`) SHALL be left unchanged, and MUST NOT be overwritten by the default value.

This default-population SHALL be used exclusively for the Plugin Framework's semantic-equality comparison (`StringSemanticEquals`). It SHALL NOT alter the JSON string value written to Terraform state or read back from configuration — the practitioner's authored JSON (or the API's redaction-merged JSON produced by existing `input`/`actions` redaction-preservation behavior) is stored unchanged; only the transient values used for the equality comparison have defaults populated.

#### Scenario: Omitted rest_total_hits_as_int does not cause apply inconsistency

- **GIVEN** a watch `input` configured as a `search` request whose `request` object omits `rest_total_hits_as_int`
- **WHEN** the resource is created and the Get Watch response used to refresh state includes `rest_total_hits_as_int: true` injected by Elasticsearch
- **THEN** the apply SHALL succeed without a "Provider produced inconsistent result after apply" error
- **AND** a subsequent `terraform plan` SHALL show no changes

#### Scenario: Omitted search_type does not cause apply inconsistency

- **GIVEN** a watch `input` configured as a `search` request whose `request` object omits `search_type`
- **WHEN** the resource is created and the Get Watch response used to refresh state includes `search_type: "query_then_fetch"` injected by Elasticsearch
- **THEN** the apply SHALL succeed without a "Provider produced inconsistent result after apply" error
- **AND** a subsequent `terraform plan` SHALL show no changes

#### Scenario: transform search request defaults

- **GIVEN** a watch `transform` configured as a `search` request whose `request` body omits `rest_total_hits_as_int` and `search_type`
- **WHEN** the resource is created and the Get Watch response used to refresh state includes both keys injected by Elasticsearch
- **THEN** the apply SHALL succeed without a "Provider produced inconsistent result after apply" error
- **AND** a subsequent `terraform plan` SHALL show no changes

#### Scenario: Chained search sub-inputs each default independently

- **GIVEN** a watch `input` configured as a `chain` input containing two or more `search` sub-inputs, each omitting `rest_total_hits_as_int` and/or `search_type`
- **WHEN** the resource is created and Elasticsearch injects those defaults independently into each chained search request's `request` object
- **THEN** the apply SHALL succeed without a "Provider produced inconsistent result after apply" error
- **AND** a subsequent `terraform plan` SHALL show no changes

#### Scenario: Explicit non-default values are preserved and still detected as real changes

- **GIVEN** a watch `input` search request that explicitly sets `rest_total_hits_as_int = false` and `search_type = "dfs_query_then_fetch"`
- **WHEN** the resource is created and read back
- **THEN** the state SHALL preserve the practitioner's explicit values unchanged
- **AND** a subsequent `terraform plan` SHALL show no changes
- **AND** if the practitioner subsequently changes either value in configuration, the next plan SHALL show that change (defaulting only fills absent keys, so explicit values continue to participate in equality comparison)

#### Scenario: Non-search inputs and transforms are unaffected

- **GIVEN** a watch `input` or `transform` that does not contain a `search` key (for example an `http` or `simple` input, or a `script` transform)
- **WHEN** the resource is created, updated, or read
- **THEN** no `rest_total_hits_as_int` or `search_type` defaults SHALL be injected for comparison purposes
- **AND** existing semantic-equality behavior (key order/whitespace-insensitive comparison) SHALL be unchanged for that input or transform
