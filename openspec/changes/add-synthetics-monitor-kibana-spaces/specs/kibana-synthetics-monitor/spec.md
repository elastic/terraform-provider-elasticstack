## ADDED Requirements

### Requirement: Kibana space visibility (REQ-025)

The resource SHALL support an optional `kibana_spaces` list of Kibana space IDs that controls the spaces in which a Synthetics monitor is visible. On create and update, a configured list SHALL be sent as the monitor API's `spaces` field; a configured empty list SHALL be sent as an empty array, while an omitted list on a new monitor SHALL leave the field absent. Removing a previously configured list SHALL clear additional visibility by sending an empty array. The list SHALL support the `"*"` wildcard. `space_id` SHALL remain the monitor's owning space.

When Kibana adds the owning `space_id` to an API response, the provider SHALL preserve the configured `kibana_spaces` value if that is the only difference. An omitted `kibana_spaces` value SHALL remain omitted when the API response contains only the owning space. Any other API-reported visibility difference SHALL be reflected in state.

#### Scenario: Configured visibility is sent

- **GIVEN** a monitor with `space_id` set to `default` and `kibana_spaces` set to `["observability", "security"]`
- **WHEN** the provider creates or updates the monitor
- **THEN** the monitor API request includes `spaces` set to `["observability", "security"]` and the request path uses `default` as the owning space

#### Scenario: Wildcard visibility is sent

- **GIVEN** a monitor with `kibana_spaces` set to `["*"]`
- **WHEN** the provider creates or updates the monitor
- **THEN** the monitor API request includes `spaces` set to `["*"]`

#### Scenario: Empty visibility is sent

- **GIVEN** a monitor with `kibana_spaces` set to an empty list
- **WHEN** the provider creates or updates the monitor
- **THEN** the monitor API request includes `spaces` as an empty array

#### Scenario: Configured visibility is removed

- **GIVEN** a monitor that previously configured `kibana_spaces`
- **WHEN** the configuration removes `kibana_spaces`
- **THEN** the update request clears additional visibility and Terraform state keeps `kibana_spaces` omitted after Kibana reports only the owning space

#### Scenario: Owning space is automatically reported

- **GIVEN** a monitor with `space_id` set to `default` and `kibana_spaces` set to `["security"]`
- **WHEN** Kibana returns `["security", "default"]` for the monitor's spaces
- **THEN** Terraform state retains `["security"]` for `kibana_spaces`

#### Scenario: Omitted visibility receives only the owning space

- **GIVEN** a monitor with no configured `kibana_spaces` and `space_id` set to `default`
- **WHEN** Kibana returns `["default"]` for the monitor's spaces
- **THEN** Terraform state keeps `kibana_spaces` omitted

#### Scenario: External visibility change is read

- **GIVEN** a monitor with configured `kibana_spaces`
- **WHEN** Kibana returns a spaces list that differs by more than automatic inclusion of the owning space
- **THEN** Terraform state reflects the API-reported spaces list
