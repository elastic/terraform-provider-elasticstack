## MODIFIED Requirements

### Requirement: Create, update, and read (REQ-013–REQ-016)

On create/update, the resource SHALL construct a `models.IndexTemplate` request body from the Terraform plan and submit it with the Put index template API. After a successful Put request, the resource SHALL set `id` and perform a read to refresh state, using the decoded Terraform configuration (not the plan) as the prior-state seed for read-after-write to avoid false drift from Unknown placeholders in computed set elements. On read, the resource SHALL parse `id`, fetch the index template by name, and remove the resource from state when the template is not found. If the Get index template API returns a result count other than exactly one template, the read path SHALL return an error diagnostic.

The resource SHALL use the `WriteFunc[T]` callback contract via the entitycore envelope for both Create and Update. The resource SHALL NOT override the envelope's `Create` or `Update` method receivers.

On update, when prior state had `data_stream.allow_custom_routing=true` and the configuration does not explicitly set `allow_custom_routing=true`, the resource SHALL include `allow_custom_routing=false` in the API request body (8.x workaround). The write callback SHALL determine this by comparing `req.Prior` (prior state) against `req.Config` (decoded configuration).

#### Scenario: Template not found on refresh

- **WHEN** read runs after the template was deleted in Elasticsearch
- **THEN** the resource SHALL be removed from state

#### Scenario: Read-after-write uses config as seed model

- **WHEN** create or update runs and the Put index template API succeeds
- **THEN** the resource SHALL seed the read-after-write call with the decoded Terraform configuration model (not the plan)
- **AND** the refreshed state SHALL reflect the server-returned template without spurious Unknown-placeholder drift

#### Scenario: allow_custom_routing 8.x workaround applied on update

- **WHEN** update runs and prior state has `data_stream.allow_custom_routing=true`
- **AND** the configuration does not explicitly set `allow_custom_routing=true`
- **THEN** the API request SHALL include `allow_custom_routing=false` in the `data_stream` block

#### Scenario: allow_custom_routing 8.x workaround not applied when config sets true

- **WHEN** update runs and the configuration explicitly sets `data_stream.allow_custom_routing=true`
- **THEN** the API request SHALL include `allow_custom_routing=true` and the workaround SHALL NOT override it
