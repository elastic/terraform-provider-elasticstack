# `elasticstack_elasticsearch_ml_trained_model_alias` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/ml/trainedmodelalias`

## Purpose

Define schema and behavior for the Elasticsearch ML trained model alias resource: API usage, identity and import, connection, lifecycle (force-new on `model_alias`), create/read/update/delete flows, and mapping between Terraform state and the Elasticsearch Machine Learning trained model alias API — including in-place reassignment via the `reassign` flag and drift handling when the alias is modified out of band.

## Schema

```hcl
resource "elasticstack_elasticsearch_ml_trained_model_alias" "example" {
  id          = <computed, string>  # internal identifier: <cluster_uuid>/<model_alias>; UseStateForUnknown
  model_alias = <required, string>  # force new; unique logical alias name; cannot end in digits per ES validation
  model_id    = <required, string>  # the trained model this alias refers to; mutable (not force new)
  reassign    = <optional, bool>    # default: true; when false, PUT fails if alias already refers to a different model

  # Resource-level Elasticsearch connection override (injected by entitycore)
  elasticsearch_connection {
    endpoints                = <optional, list(string)>
    username                 = <optional, string>
    password                 = <optional, string>
    api_key                  = <optional, string>
    bearer_token             = <optional, string>
    es_client_authentication = <optional, string>
    insecure                 = <optional, bool>
    ca_file                  = <optional, string>
    ca_data                  = <optional, string>
    cert_file                = <optional, string>
    cert_data                = <optional, string>
    key_file                 = <optional, string>
    key_data                 = <optional, string>
    headers                  = <optional, map(string)>
  }
}
```

## ADDED Requirements

### Requirement: Resource identity (REQ-001)

The resource SHALL use `model_alias` as the Terraform resource identity (mapped to `GetResourceID()`). The composite state `id` SHALL be `<cluster_uuid>/<model_alias>` and SHALL be set during Create using `client.ID(ctx, modelAlias).String()`.

The `model_alias` attribute SHALL be marked ForceNew (RequiresReplace plan modifier). Changing `model_alias` destroys the old alias and creates a new one.

The `model_id` attribute SHALL NOT be marked ForceNew. Changing `model_id` triggers an in-place update via the PUT API (see REQ-004).

#### Scenario: Composite id is set on create
- GIVEN a plan with `model_alias = "my-alias"` and `model_id = "model-1"`
- WHEN Create is called
- THEN `id` SHALL be set to `<cluster_uuid>/my-alias`

#### Scenario: Changing model_alias forces replacement
- GIVEN a resource exists with `model_alias = "alias-a"`
- WHEN the configuration is updated to `model_alias = "alias-b"`
- THEN Terraform SHALL plan a destroy of the existing resource and create a new resource

#### Scenario: Changing model_id triggers an update, not replacement
- GIVEN a resource exists with `model_id = "model-1"`
- WHEN the configuration is updated to `model_id = "model-2"` with `reassign = true`
- THEN Terraform SHALL plan an in-place update
- AND no replacement SHALL be planned

### Requirement: API — Create (REQ-002)

The resource SHALL call `PUT _ml/trained_models/{model_id}/model_aliases/{model_alias}` with the `reassign` query parameter set to the plan value (default true) to create the alias.

When `reassign = true`, the PUT succeeds even if the alias already points to a different model.

When `reassign = false` and an alias with the same name already exists pointing to a different model, Elasticsearch returns an error; the resource SHALL surface that error and leave no state.

#### Scenario: Create new alias
- GIVEN a plan with `model_alias = "my-alias"`, `model_id = "model-1"`
- AND no alias named `my-alias` exists
- WHEN Create is called
- THEN `PUT _ml/trained_models/model-1/model_aliases/my-alias?reassign=true` is called
- AND the composite `id` is set in state as `<cluster_uuid>/my-alias`

#### Scenario: Create succeeds with reassignment when alias already exists
- GIVEN an alias named `my-alias` already exists pointing to `model-2`
- AND the plan has `model_id = "model-1"`
- WHEN Create is called
- THEN `PUT _ml/trained_models/model-1/model_aliases/my-alias?reassign=true` is called
- AND the alias is reassigned to `model-1`
- AND the composite `id` is set in state

#### Scenario: Create fails when alias already exists with reassign disabled
- GIVEN an alias named `my-alias` already exists pointing to `model-2`
- AND the plan has `reassign = false`
- WHEN Create is called
- THEN Elasticsearch returns an error
- AND the resource SHALL surface the error with no state persisted

### Requirement: API — Read (REQ-003)

The resource SHALL call `GET _ml/trained_models/{model_alias}` (using the alias name as the `model_id` parameter to `GetTrainedModels`) to resolve the current model the alias points to.

When the response has HTTP status 404, or when the returned model list is empty, the resource SHALL signal not-found (returning `found = false`), causing the framework to remove the resource from state.

When the alias is found, the resource SHALL map the resolved `model_id` from the API response into state.

#### Scenario: Read existing alias
- GIVEN an alias `my-alias` exists and points to `model-1`
- WHEN Read is called
- THEN `GET _ml/trained_models/my-alias` returns the TrainedModelConfig for model-1
- AND `model_id` in state is set to `model-1`

#### Scenario: Read missing alias returns not-found
- GIVEN the alias does not exist (Elasticsearch returns 404 or empty list)
- WHEN Read is called
- THEN the resource SHALL be removed from state with no error

#### Scenario: Drift — alias reassigned externally
- GIVEN the alias `my-alias` was reassigned out-of-band from `model-1` to `model-3`
- WHEN Read is called during refresh
- THEN `model_id` in state is updated to `model-3`
- AND the next plan shows a diff on `model_id`

### Requirement: API — Update (REQ-004)

When `model_id` or `reassign` changes, the resource SHALL call `PUT _ml/trained_models/{model_id}/model_aliases/{model_alias}` using the planned `model_id` and `reassign` values.

To change `model_id` in-place, `reassign` MUST NOT be `false`; otherwise Elasticsearch returns an error because the alias already exists pointing to a different model.

#### Scenario: Update model_id
- GIVEN a resource with `model_alias = "my-alias"`, `model_id = "model-1"`, `reassign = true`
- AND the plan has `model_id = "model-2"`
- WHEN Update is called
- THEN `PUT _ml/trained_models/model-2/model_aliases/my-alias?reassign=true` is called
- AND Read is called afterward; state reflects `model_id = "model-2"`

#### Scenario: Update model_id with reassign disabled fails
- GIVEN a resource with `model_alias = "my-alias"`, `model_id = "model-1"`, `reassign = false`
- AND the plan has `model_id = "model-2"`, `reassign = false`
- WHEN Update is called
- THEN Elasticsearch returns an error because `reassign = false`
- AND the resource SHALL surface the error

#### Scenario: Update reassign flag only
- GIVEN a resource with `reassign = true` and `model_id = "model-1"`
- AND the plan changes only `reassign = false`
- WHEN Update is called
- THEN `PUT _ml/trained_models/model-1/model_aliases/{alias}?reassign=false` is called successfully

### Requirement: API — Delete (REQ-005)

The resource SHALL first call `GET _ml/trained_models/{model_alias}` to resolve the current model the alias points to. If the alias does not exist (404 or empty result), Delete SHALL treat this as already-gone and remove the resource from state without error.

If the alias exists, the resource SHALL call `DELETE _ml/trained_models/{resolved_model_id}/model_aliases/{model_alias}` using the resolved `model_id` from the GET response.

A 404 response during DELETE SHALL be treated as idempotent success.

Any other API error SHALL be surfaced as a "Failed to delete ML trained model alias" error.

#### Scenario: Delete existing alias
- GIVEN a resource with `model_alias = "my-alias"` and `model_id = "model-1"` in state
- WHEN Delete is called
- THEN `GET _ml/trained_models/my-alias` resolves the alias to `model-1`
- AND `DELETE _ml/trained_models/model-1/model_aliases/my-alias` is called
- AND the resource is removed from state

#### Scenario: Delete already-removed alias is idempotent
- GIVEN the alias was deleted out-of-band before Terraform runs destroy
- WHEN Delete is called
- THEN `GET _ml/trained_models/my-alias` returns 404 / empty result
- AND the resource SHALL be removed from state without error

### Requirement: Import (REQ-006)

The resource SHALL support import by composite `<cluster_uuid>/<model_alias>` ID.

On import, the resource SHALL set `id` to the full composite ID and call Read using `model_alias` as the resource identity. `model_id` is populated from the API response. `reassign` defaults to true on import (the API does not persist or return this flag).

#### Scenario: Import by composite id
- GIVEN an alias `my-alias` exists in Elasticsearch pointing to `model-1`
- WHEN `terraform import elasticstack_elasticsearch_ml_trained_model_alias.example <cluster_uuid>/my-alias` is run
- THEN `model_alias` is set to `my-alias`
- AND `model_id` is set to `model-1`
- AND `reassign` defaults to true

### Requirement: Drift handling (REQ-007)

When the alias is deleted out of band, the next plan SHALL detect the missing resource (Read returns not-found) and plan a re-create.

When the alias is reassigned out of band to a different model, the next plan SHALL show `model_id` as changed and plan an in-place update.

#### Scenario: Alias deleted out of band triggers re-create
- GIVEN the resource exists in Terraform state
- AND the alias is deleted via a direct API call
- WHEN Read is called during refresh
- THEN the resource is removed from state
- AND the next plan shows it as to-be-created

#### Scenario: Alias reassigned out of band triggers update
- GIVEN the resource has `model_id = "model-1"` in state
- AND the alias is reassigned to `model-3` via a direct API call
- WHEN Read is called during refresh
- THEN state is updated with `model_id = "model-3"`
- AND the next plan shows `model_id` as changed from `"model-3"` back to the desired `"model-1"`
