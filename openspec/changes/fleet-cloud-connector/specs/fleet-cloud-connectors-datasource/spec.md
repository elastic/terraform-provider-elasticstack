## ADDED Requirements

### Requirement: Data source identity and inputs

The `elasticstack_fleet_cloud_connectors` data source SHALL accept the following input attributes: `space_id` (optional, default `"default"`), `kuery` (optional, server-side KQL filter), `page` (optional), and `per_page` (optional). It SHALL also accept a `kibana_connection` block for resource-level connection override, consistent with other Fleet data sources.

#### Scenario: Default inputs
- **WHEN** the data source is read with no inputs other than `kibana_connection`
- **THEN** `space_id` SHALL default to `"default"`
- **AND** no `kuery`, `page`, or `per_page` query parameters SHALL be sent

#### Scenario: Server-side kuery passed through
- **WHEN** `kuery = "cloud_connectors.attributes.cloudProvider:aws"` is set on the data source
- **THEN** the request to `GET /api/fleet/cloud_connectors` SHALL include `kuery=cloud_connectors.attributes.cloudProvider:aws` in the query string

### Requirement: List endpoint backing

The data source SHALL call `GET /api/fleet/cloud_connectors` (space-aware) and return the full list of items in the response. The data source SHALL NOT paginate internally on the user's behalf in v1; if the response is truncated by the API, the user is responsible for setting `per_page` and/or refining `kuery`.

#### Scenario: Successful list
- **WHEN** the data source is read against a Kibana with three cloud connectors
- **THEN** `cloud_connectors` SHALL contain three entries
- **AND** each entry SHALL include `cloud_connector_id`, `name`, `cloud_provider`, `account_type`, `namespace`, `package_policy_count`, `verification_status`, `verification_started_at`, `verification_failed_at`, `created_at`, `updated_at`

### Requirement: `vars` omitted from data source output

The data source SHALL NOT expose the `vars` field of each cloud connector in its output. Users who need the full `vars` (including secret references) SHALL `terraform import` the specific resource.

#### Scenario: No vars in output
- **WHEN** the data source is read against a Kibana with cloud connectors that have `vars`
- **THEN** the output entries SHALL NOT contain a `vars` attribute

### Requirement: Empty list handling

The data source SHALL return an empty `cloud_connectors` list (not null and not an error) when the API returns no items.

#### Scenario: No connectors
- **WHEN** the data source is read against a Kibana with zero cloud connectors
- **THEN** the output `cloud_connectors` SHALL be an empty list
- **AND** no error SHALL be raised

### Requirement: Version gating

The data source SHALL declare a `GetVersionRequirements` entry identical to the resource's, failing against Kibana versions older than the first version that ships the cloud connectors API.

#### Scenario: Pre-cloud-connectors Kibana version
- **WHEN** the data source is read against a Kibana older than the configured minimum
- **THEN** Terraform SHALL fail with an error message stating the minimum required version
