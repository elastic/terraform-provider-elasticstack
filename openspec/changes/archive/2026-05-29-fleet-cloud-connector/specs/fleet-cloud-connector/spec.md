## ADDED Requirements

### Requirement: Resource identity and composite ID

The `elasticstack_fleet_cloud_connector` resource SHALL set its `id` to the composite string `"<space_id>/<cloud_connector_id>"` after every Create and Update. `cloud_connector_id` SHALL be Computed only: the API assigns the ID on Create and the provider SHALL populate it into state from the response. Users SHALL NOT be able to supply a custom `cloud_connector_id` at Create time because the Fleet API does not accept an explicit ID in the POST body. `space_id` SHALL default to `"default"`. Changing `space_id` or `cloud_provider` SHALL force resource replacement.

#### Scenario: Create with API-assigned cloud_connector_id
- **WHEN** the resource is created
- **THEN** the POST request SHALL NOT include a client-supplied connector ID
- **AND** `cloud_connector_id` SHALL be populated from the API-assigned ID in the response
- **AND** `id` SHALL equal `"<space_id>/<cloud_connector_id>"`

#### Scenario: cloud_connector_id is read-only in configuration
- **WHEN** a practitioner inspects the resource schema
- **THEN** `cloud_connector_id` SHALL be Computed only (not Optional)
- **AND** the provider SHALL NOT attempt to influence the assigned ID during Create

#### Scenario: cloud_provider change forces replacement
- **WHEN** `cloud_provider` is changed in config from `"aws"` to `"azure"`
- **THEN** Terraform SHALL destroy and recreate the resource

#### Scenario: space_id change forces replacement
- **WHEN** `space_id` is changed in config
- **THEN** Terraform SHALL destroy and recreate the resource

### Requirement: Cloud provider and account type

The resource SHALL require `cloud_provider` (one of `"aws"`, `"azure"`, `"gcp"`) and SHALL accept an optional `account_type` (one of `"single-account"`, `"organization-account"`). `cloud_provider` SHALL force replacement on change; `account_type` SHALL be updatable in-place because the API PUT body accepts it.

#### Scenario: account_type updated in-place
- **WHEN** `account_type` is changed from `"single-account"` to `"organization-account"`
- **THEN** `PUT /api/fleet/cloud_connectors/{id}` SHALL be called with the new value
- **AND** the resource SHALL NOT be replaced

#### Scenario: Invalid cloud_provider rejected
- **WHEN** `cloud_provider = "digitalocean"` is set in config
- **THEN** Terraform SHALL reject the plan with a validation error

### Requirement: Create

The resource SHALL call `POST /api/fleet/cloud_connectors` (space-aware) with `name`, `cloudProvider`, `vars` (compiled from either the typed block or the user-supplied `vars` map), and optional `accountType`. The response body SHALL be decoded into the generated `kbapi` cloud-connector item type. State SHALL be set from the decoded response, populating both the raw `vars` map AND the matching typed block per Decision 4.

#### Scenario: Successful create with typed AWS block
- **WHEN** a resource with `aws { role_arn = "...", external_id = var.id }` is applied
- **THEN** `POST /api/fleet/cloud_connectors` SHALL be called with a body containing `cloudProvider: "aws"` and `vars` containing typed entries for `role_arn` (`type: "text"`) and `external_id` (`type: "password"`, value sent as raw string)
- **AND** state SHALL contain both `aws.role_arn` and `vars["role_arn"]` populated from the API response
- **AND** state SHALL contain `vars["external_id"].secret_ref.id` populated from the API response
- **AND** state SHALL NOT contain the raw value of `external_id` anywhere

#### Scenario: Successful create with generic vars only
- **WHEN** a resource with only `vars = { ... }` (no typed block) is applied for an `aws` provider
- **THEN** the API SHALL be called with `vars` matching the user-supplied map verbatim
- **AND** state SHALL populate both `vars` (raw) and `aws { }` (if all expected keys are present and match `cloud_provider`)

### Requirement: Read

The resource SHALL call `GET /api/fleet/cloud_connectors/{id}` (space-aware). On HTTP 404 the resource SHALL be removed from state without error. On success, both the raw `vars` map AND the matching typed block (when all of its expected keys are present in the response) SHALL be populated in state.

#### Scenario: Resource deleted out of band
- **WHEN** the API returns HTTP 404 on Read
- **THEN** the resource SHALL be removed from state without error

#### Scenario: Read populates both representations
- **WHEN** Read succeeds against an AWS cloud connector with `role_arn` and `external_id` vars
- **THEN** state SHALL contain `vars["role_arn"]`, `vars["external_id"]`, AND `aws.role_arn`
- **AND** the typed block's secret-bearing field SHALL contain the secret reference id under `aws.external_id_secret_ref`, never the raw secret value

#### Scenario: Read with unknown var keys
- **WHEN** the API response contains a var key the typed block does not model (e.g. a future `region` key for AWS)
- **THEN** state SHALL contain the new key in `vars`
- **AND** the typed `aws` block SHALL be null in state for that resource
- **AND** no error SHALL be raised

### Requirement: Update

The resource SHALL call `PUT /api/fleet/cloud_connectors/{id}` (space-aware) with `name`, `accountType` (when set), and the recompiled `vars`. `cloudProvider` SHALL NOT appear in the PUT body because the API does not accept it. After Update, state SHALL be repopulated from the PUT response in the same dual-representation form as Read.

#### Scenario: Update name
- **WHEN** `name` is changed in config
- **THEN** `PUT /api/fleet/cloud_connectors/{id}` SHALL be called with the new name
- **AND** state SHALL reflect the new name

#### Scenario: Update vars without changing secret_value
- **WHEN** a non-secret var (e.g. `role_arn`) is changed in config and secret values are unchanged
- **THEN** the PUT body SHALL contain the new value for the changed var
- **AND** secret-bearing vars SHALL be sent with their existing `secret_ref` (not re-sent as raw strings)

### Requirement: Delete

The resource SHALL call `DELETE /api/fleet/cloud_connectors/{id}` (space-aware), passing `?force=true` when the `force_delete` attribute is `true`. HTTP 404 SHALL be treated as success. When `force_delete` is `false` and the API returns a conflict because the connector is referenced by package policies, the provider SHALL surface an error that names the `package_policy_count` and suggests setting `force_delete = true` if the deletion is intentional.

#### Scenario: Successful delete
- **WHEN** the resource is destroyed and `force_delete = false`
- **THEN** `DELETE /api/fleet/cloud_connectors/{id}` SHALL be called without the `force` query parameter
- **AND** no error SHALL be returned on success

#### Scenario: Force delete passes force=true
- **WHEN** the resource is destroyed and `force_delete = true`
- **THEN** `DELETE /api/fleet/cloud_connectors/{id}?force=true` SHALL be called

#### Scenario: Already-deleted resource
- **WHEN** the API returns HTTP 404 on Delete
- **THEN** the resource SHALL be removed from state without error

#### Scenario: In-use conflict without force
- **WHEN** the API returns a conflict on Delete because `package_policy_count > 0` and `force_delete = false`
- **THEN** the provider SHALL return an error that includes the current `package_policy_count`
- **AND** the error message SHALL suggest setting `force_delete = true` if intentional

### Requirement: Import

The resource SHALL support import via the composite ID `"<space_id>/<cloud_connector_id>"`. On import, Read SHALL parse the composite ID to derive `space_id` and `cloud_connector_id` for the API call. On first refresh after import, no hash SHALL exist in private state, so no write-only drift SHALL be detected; the hash SHALL be populated on the first apply that includes a write-only secret in config.

#### Scenario: Import by composite ID
- **WHEN** `terraform import elasticstack_fleet_cloud_connector.x "default/my-connector"` is run
- **THEN** `cloud_connector_id` SHALL be set to `"my-connector"`
- **AND** `space_id` SHALL be set to `"default"`
- **AND** all other attributes SHALL be populated from the API response per the Read contract

#### Scenario: First apply after import baselines write-only hash
- **WHEN** the imported resource is applied for the first time with `aws { external_id = var.id }` in config
- **THEN** the bcrypt hash of `external_id` SHALL be written to private state
- **AND** subsequent plans SHALL detect drift in `external_id` per the write-only drift detection contract

### Requirement: `vars` schema covers all four API union arms

The `vars` attribute SHALL be modelled as `map(object({...}))`. Each map element SHALL faithfully represent one of the four API union arms via the following nested attributes: `string` (arm 1), `number` (arm 2), `bool` (arm 3), and `type`+`frozen`+exactly-one-of(`value`, `secret_value`, `secret_ref`) (arm 4). `secret_ref` SHALL be Computed-only and rejected if set in config. A `ConfigValidator` SHALL enforce that within each element exactly one arm is configured.

#### Scenario: Arm (1) bare string
- **WHEN** `vars = { "k" = { string = "hello" } }` is set
- **THEN** the API SHALL be sent `{"k": "hello"}` in the `vars` body

#### Scenario: Arm (2) bare number
- **WHEN** `vars = { "k" = { number = 3.14 } }` is set
- **THEN** the API SHALL be sent `{"k": 3.14}` in the `vars` body

#### Scenario: Arm (3) bare boolean
- **WHEN** `vars = { "k" = { bool = true } }` is set
- **THEN** the API SHALL be sent `{"k": true}` in the `vars` body

#### Scenario: Arm (4) structured with plain string value
- **WHEN** `vars = { "k" = { type = "text", value = "abc", frozen = false } }` is set
- **THEN** the API SHALL be sent `{"k": {"type": "text", "value": "abc", "frozen": false}}` in the `vars` body

#### Scenario: Arm (4) structured with write-only secret value
- **WHEN** `vars = { "k" = { type = "password", secret_value = "shhh" } }` is applied for the first time
- **THEN** the API SHALL be sent `{"k": {"type": "password", "value": "shhh"}}` in the `vars` body
- **AND** state SHALL NOT contain the raw value of `secret_value`
- **AND** the next Read SHALL populate `vars["k"].secret_ref` with the `{id, isSecretRef}` ref returned by the API

#### Scenario: Multiple arms set on the same element is rejected
- **WHEN** `vars = { "k" = { string = "a", number = 1 } }` is set in config
- **THEN** Terraform SHALL reject the plan with a validation error naming the conflicting attributes

#### Scenario: secret_ref in config is rejected
- **WHEN** `vars = { "k" = { type = "password", secret_ref = { id = "x", is_secret_ref = true } } }` is set in config
- **THEN** Terraform SHALL reject the plan with a validation error indicating `secret_ref` is computed-only

### Requirement: Typed `aws` and `azure` blocks compile to `vars`

The resource SHALL expose typed blocks `aws { role_arn, external_id, external_id_secret_ref }` and `azure { tenant_id, client_id, cloud_connector_id, tenant_id_secret_ref, client_id_secret_ref }`. Parent blocks `aws` and `vars` SHALL be Optional only (not Computed) because Plugin Framework disallows Computed on blocks containing write-only children; inner fields such as `external_id_secret_ref` remain Computed where populated from Read. A `ConfigValidator` SHALL enforce that exactly one of `aws`, `azure`, or `vars` is configured, and that any configured typed block matches the resource's `cloud_provider`. Typed blocks SHALL compile to the same wire `vars` payload during Create and Update; on Read, the typed block SHALL be populated only when ALL of its modelled keys appear in the API response under the expected provider.

#### Scenario: Typed block input compiles to vars wire payload
- **WHEN** `aws { role_arn = "arn:..." external_id = "secret" }` is applied
- **THEN** the API SHALL receive `vars: { "role_arn": {"type":"text","value":"arn:..."}, "external_id": {"type":"password","value":"secret"} }`

#### Scenario: Mismatched typed block and cloud_provider rejected
- **WHEN** `cloud_provider = "aws"` and `azure { ... }` is configured
- **THEN** Terraform SHALL reject the plan with a validation error

#### Scenario: Multiple representations rejected
- **WHEN** both `aws { ... }` and `vars = { ... }` are configured
- **THEN** Terraform SHALL reject the plan with a validation error

#### Scenario: Typed block populated in state when input was vars
- **WHEN** only `vars = { "role_arn" = { type = "text", value = "..." }, "external_id" = { type = "password", secret_value = var.s } }` is configured for an AWS cloud connector
- **THEN** after Read, state SHALL contain both `vars` (raw) AND `aws { role_arn = "...", external_id_secret_ref = { ... } }`

### Requirement: GCP provider falls back to generic vars

The resource SHALL NOT expose a typed `gcp` block in the initial version. Users with `cloud_provider = "gcp"` SHALL configure credentials via the generic `vars` map.

#### Scenario: GCP without typed block
- **WHEN** a resource with `cloud_provider = "gcp"` and `vars = { ... }` is applied
- **THEN** the API SHALL be called with the user-supplied vars
- **AND** no typed `gcp` block SHALL be populated in state

#### Scenario: GCP with no typed block defined
- **WHEN** a resource with `cloud_provider = "gcp"` and no `vars`, `aws`, or `azure` is configured
- **THEN** Terraform SHALL reject the plan with a validation error indicating one of `vars` is required

### Requirement: Write-only secret drift detection via private-state hash

Every write-only secret attribute on the resource (`vars[*].secret_value`, `aws.external_id`, and any future write-only secrets) SHALL have its most-recently-applied value hashed with bcrypt and stored in resource private state, keyed by a stable per-attribute identifier. During `ModifyPlan`, the provider SHALL recompute the hash from the current config value, compare against the stored hash, and mark the resource as needing update when they differ. When such drift is detected, the provider SHALL emit a plan-time warning diagnostic naming the changed attribute (without revealing the value). On first refresh after import, no hash SHALL exist and no drift SHALL be reported.

#### Scenario: Silent in-config secret edit detected at plan time
- **WHEN** `aws.external_id` is changed in config from `"old"` to `"new"` between two `terraform apply` runs
- **THEN** the second plan SHALL show the resource as needing update
- **AND** a warning diagnostic SHALL be emitted naming `aws.external_id` as the changed attribute
- **AND** applying the plan SHALL re-send the secret to the API and update the stored hash

#### Scenario: No change to secret produces no plan drift
- **WHEN** a `terraform plan` is run with no changes to `aws.external_id` since the last apply
- **THEN** the plan SHALL show no diff related to the secret

#### Scenario: First apply after import does not over-report drift
- **WHEN** an imported resource is planned with `aws.external_id` set in config for the first time
- **THEN** the plan SHALL still show the resource as needing update (to baseline the hash)
- **BUT** the diagnostic SHALL clearly indicate this is the import-baseline case

### Requirement: Computed read-only fields

The resource SHALL expose the following computed attributes, populated from the API response: `namespace`, `package_policy_count`, `verification_status`, `verification_started_at`, `verification_failed_at`, `created_at`, `updated_at`. None of these SHALL block Create from succeeding; verification fields MAY remain null on first Read because verification is asynchronous in Kibana.

#### Scenario: Computed fields populated after Create
- **WHEN** a resource is successfully created
- **THEN** `created_at`, `updated_at`, `namespace`, and `package_policy_count` SHALL be populated in state from the API response
- **AND** `verification_status` and related verification fields MAY be null or partially populated without raising an error

### Requirement: Connection override and version gating

The resource SHALL obtain its Fleet client via the resource-level `kibana_connection` block when provided, otherwise via the provider-level Kibana configuration. Space-aware requests SHALL use `space_id` via the existing `spaceAwarePathRequestEditor` helper. The resource SHALL declare a `GetVersionRequirements` entry that fails with a helpful error against Kibana versions older than the first version that ships `/api/fleet/cloud_connectors`.

#### Scenario: Resource-level kibana_connection override
- **WHEN** `kibana_connection` is configured on the resource
- **THEN** all API calls SHALL use that connection instead of the provider-level Kibana connection

#### Scenario: Pre-cloud-connectors Kibana version
- **WHEN** the resource is planned against a Kibana version older than the configured minimum
- **THEN** Terraform SHALL fail with an error message stating the minimum required version
