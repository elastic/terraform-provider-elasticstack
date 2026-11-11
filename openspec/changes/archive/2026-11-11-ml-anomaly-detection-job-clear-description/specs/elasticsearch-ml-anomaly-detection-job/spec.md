## MODIFIED Requirements

### Requirement: Update — partial update with only mutable fields (REQ-016–REQ-018)

On update, the envelope SHALL invoke the resource update callback with `WriteRequest[TFModel]` (where `req.Prior` is a non-nil pointer to the prior-state model), supplying planned and prior models plus raw Terraform config. The callback SHALL only send fields that have changed and that are permitted by the Elasticsearch Update Job API. The following fields MAY be updated in place: `description`, `groups`, `model_plot_config`, `analysis_limits` (specifically `model_memory_limit`), `allow_lazy_open`, `background_persist_interval`, `custom_settings`, `daily_model_snapshot_retention_after_days`, `model_snapshot_retention_days`, `renormalization_window_days`, and `results_retention_days`. When no updateable fields have changed, the callback SHALL log a warning and SHALL NOT call the Update Job API. When `description` is set in prior state and null in the plan (removed from configuration), the callback SHALL treat it as changed and SHALL send `description` as an empty string so Elasticsearch clears it; the read maps the empty string back to null.

After the update callback returns successfully — including the case where no Update Job API call was made because no updatable fields changed — the envelope SHALL refresh the resource by reading the job back from the API via the shared read callback and SHALL persist state from that read result.

#### Scenario: No updateable fields changed

- GIVEN an update where only non-updateable fields (e.g. analysis_config) have changed according to plan
- WHEN update runs
- THEN the resource SHALL not call the Update Job API and SHALL log a warning
- AND the envelope SHALL still perform read-after-write to refresh state from the API

#### Scenario: Updateable fields changed

- GIVEN an update where description has changed
- WHEN update runs
- THEN the resource SHALL call the Update Job API with only the changed fields
- AND the envelope SHALL refresh state via read-after-write afterward

#### Scenario: Description removed from configuration

- GIVEN a job whose prior state has a non-empty `description`
- AND the configuration no longer sets `description`
- WHEN update runs
- THEN the resource SHALL call the Update Job API with `description` set to an empty string
- AND after read-after-write `description` SHALL be null in state
