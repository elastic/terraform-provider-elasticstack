## MODIFIED Requirements

### Requirement: State mapping — output type dispatch (REQ-017)

On read, the resource SHALL dispatch state population based on the output type discriminator. For `OutputElasticsearch`, `OutputLogstash`, `OutputKafka`, and `OutputRemoteElasticsearch` responses, the resource SHALL map all common fields (`id`, `output_id`, `name`, `type`, `hosts`, `ca_sha256`, `ca_trusted_fingerprint`, `default_integrations`, `default_monitoring`, `config_yaml`, `ssl`). For `OutputKafka`, the resource SHALL additionally map all Kafka-specific fields. For `OutputRemoteElasticsearch`, the resource SHALL additionally map remote Elasticsearch-specific fields. If an unrecognized output type is returned, the resource SHALL surface an error diagnostic.

When mapping `config_yaml`, the resource SHALL fold a nil or empty-string value from the Fleet API into a null state value. This normalisation keeps state stable for outputs that were never configured with a `config_yaml`, since Fleet echoes an empty string in update responses for such outputs.

The `config_yaml` attribute SHALL use a YAML-aware custom type that performs semantic equality (whitespace, key ordering, anchor expansion) so that semantically equivalent YAML re-emitted by the Fleet API does not register as a change.

#### Scenario: Unknown output type

- GIVEN the API returns a type not in the known set
- WHEN read runs
- THEN the resource SHALL return an error diagnostic

#### Scenario: Empty config_yaml from Fleet maps to null state

- GIVEN the Fleet API response contains `"config_yaml": ""`
- WHEN read runs
- THEN `config_yaml` SHALL be null in state

#### Scenario: Semantically equivalent config_yaml round-trip

- GIVEN configuration sets `config_yaml = "a: 1\nb: 2\n"` and Fleet re-emits `"b: 2\na: 1\n"`
- WHEN refresh runs
- THEN `config_yaml` SHALL be treated as unchanged and SHALL NOT register a diff

## ADDED Requirements

### Requirement: config_yaml preserves user-removed null across the Fleet API echo (REQ-025)

The Fleet update API treats an omitted `config_yaml` in the PUT body as "no change" and echoes the previously stored value (empty string for outputs that never had one) back in the response. To prevent Terraform's post-apply consistency check from rejecting the resulting null-vs-value mismatch for the sensitive attribute, and to avoid perpetual drift on subsequent refresh, the shared reader (`fromAPICommonFields`) SHALL preserve a null `config_yaml` from the existing model when the API echoes a non-null value back. The preservation applies to both the update path (where the existing model is the plan) and the refresh path (where the existing model is the prior state).

The preservation SHALL NOT apply on import. The reader SHALL detect import by checking whether the existing model's required `name` field is null — on import only the importer-populated identity fields (`output_id`, optionally `space_ids`) are pre-populated and `name` is null, whereas after any successful create / read / update the prior state always carries a non-null `name`. On import the reader SHALL surface the API-returned `config_yaml` (after empty-string normalisation) so the imported state matches Fleet.

The resource documentation SHALL note that removing `config_yaml` from configuration does not clear the value stored by Fleet server-side, and that practitioners must delete and re-create the output to fully clear a previously stored value.

#### Scenario: Removing config_yaml from configuration applies cleanly and stays clean on refresh

- GIVEN an existing Fleet output with `config_yaml` set to a non-empty value
- AND the configuration is updated to omit `config_yaml`
- AND the Fleet update response echoes the previously stored value back
- WHEN `terraform apply` runs the update
- THEN the apply SHALL complete without error
- AND `config_yaml` SHALL be null in state
- AND the subsequent refresh SHALL keep `config_yaml` null in state, even though Fleet still holds the value server-side

#### Scenario: Import populates config_yaml from the Fleet API

- GIVEN an existing Fleet output in Fleet with `config_yaml = "bulk_max_size: 100\n"`
- AND no prior Terraform state for the output
- WHEN `terraform import` runs followed by the framework's import-state-verify read
- THEN `config_yaml` SHALL be `"bulk_max_size: 100\n"` in state

#### Scenario: External drift to a managed config_yaml is surfaced

- GIVEN state has `config_yaml = "a: 1\n"` for a Terraform-managed output
- AND someone changes the value to `"a: 2\n"` outside of Terraform
- WHEN refresh runs
- THEN `config_yaml` SHALL be updated to `"a: 2\n"` in state and the next plan SHALL show the drift
