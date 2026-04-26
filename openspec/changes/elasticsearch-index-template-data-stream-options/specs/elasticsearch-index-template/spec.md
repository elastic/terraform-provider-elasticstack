# Delta Spec: `elasticstack_elasticsearch_index_template` — `data_stream_options` support

Base spec: `openspec/specs/elasticsearch-index-template/spec.md`
Last requirement in base spec: REQ-031
This delta introduces: REQ-032

---

This delta defines the target behavior introduced by change `elasticsearch-index-template-data-stream-options`. It describes the requirements the implementation will satisfy once the change is applied; it does not claim that the current implementation already meets these behaviors.

## ADDED Requirements

### Requirement: Schema — `template.data_stream_options` block (REQ-032)

The `template` block SHALL support an optional `data_stream_options` block (at most one) containing an optional `failure_store` block (at most one) with an optional boolean `enabled` attribute. When omitted, the field SHALL NOT be included in the Elasticsearch API request or response mapping.

```hcl
template {
  data_stream_options {
    failure_store {
      enabled = <optional, bool>
    }
  }
}
```

On write (create or update), when `template.data_stream_options` is configured, the resource SHALL include the `data_stream_options` object in the `template` payload sent to the Elasticsearch Put index template API. When `data_stream_options.failure_store.enabled` is set, it SHALL be sent as the value of `template.data_stream_options.failure_store.enabled` in the API request body. When `data_stream_options` is not configured, the field SHALL be omitted from the API request body.

On read, when the Elasticsearch Get index template API response includes `template.data_stream_options`, the resource SHALL deserialize the `failure_store.enabled` value and store it in Terraform state. When `template.data_stream_options` is absent from the API response, the `data_stream_options` block SHALL remain unset in state.

The data source for `elasticstack_elasticsearch_index_template` SHALL expose the same `data_stream_options` block schema and populate it from API read responses using the same deserialization behavior.

#### Scenario: Failure store enabled on create

- **WHEN** configuration sets `template.data_stream_options.failure_store.enabled = true`
- **THEN** the Put index template API request body SHALL include `"data_stream_options": {"failure_store": {"enabled": true}}` inside the `template` object

#### Scenario: Failure store disabled on create

- **WHEN** configuration sets `template.data_stream_options.failure_store.enabled = false`
- **THEN** the Put index template API request body SHALL include `"data_stream_options": {"failure_store": {"enabled": false}}` inside the `template` object

#### Scenario: `data_stream_options` omitted when not configured

- **WHEN** the `template` block is configured without a `data_stream_options` block
- **THEN** the Put index template API request body SHALL NOT include a `data_stream_options` key inside the `template` object

#### Scenario: Read populates state from API response

- **WHEN** the Get index template API returns `"data_stream_options": {"failure_store": {"enabled": true}}` in the template
- **THEN** `template.data_stream_options.failure_store.enabled` SHALL be `true` in Terraform state

#### Scenario: Read clears state when field absent from API response

- **WHEN** the Get index template API returns a `template` object without `data_stream_options`
- **THEN** `template.data_stream_options` SHALL be unset (null or empty) in Terraform state

#### Scenario: Update removes `data_stream_options`

- **WHEN** a previously applied configuration had `data_stream_options` set and the updated configuration omits `data_stream_options`
- **THEN** the Put index template API request body SHALL NOT include `data_stream_options` in the `template` object

## MODIFIED Requirements

None. All base spec requirements (REQ-001 through REQ-031) remain unchanged.
