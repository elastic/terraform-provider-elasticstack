## Why

Fleet cloud connectors are reusable cloud-credential bundles that authenticate agentless integrations (Cloud Security Posture Management and Cloud Asset Discovery, today) against AWS, Azure, and GCP. Kibana exposes a full CRUD API for them under `/api/fleet/cloud_connectors`, but the Terraform provider has no resource or data source covering this surface. Teams adopting agentless CSPM/Asset Discovery in cloud environments must therefore click through Kibana to create the connector, then reference its ID by hand in their Terraform package policies — breaking the infrastructure-as-code workflow and making per-environment provisioning (dev/staging/prod) error-prone.

This change adds a first-class resource and data source for Fleet cloud connectors, closing the "Fleet cloud connectors" gap called out in [`elastic/kibana#260388`](https://github.com/elastic/kibana/issues/260388) and resolving [`#2122`](https://github.com/elastic/terraform-provider-elasticstack/issues/2122).

## What Changes

- Add `elasticstack_fleet_cloud_connector` resource with full CRUD lifecycle against `/api/fleet/cloud_connectors` (POST/GET/PUT/DELETE), import via composite `<space_id>/<cloud_connector_id>`, and `kibana_connection` override consistent with other Fleet resources.
- Model the cloud-connector `vars` shape as a typed Terraform map covering all four API union arms (bare `string`/`number`/`bool` plus the structured `{type, value | secret_ref, frozen}` arm), with computed-only `secret_ref` and write-only `secret_value` for the sugar that converts a raw credential into a saved-secret reference.
- Add typed provider-specific blocks (`aws { role_arn, external_id }` and `azure { tenant_id, client_id, cloud_connector_id }`) as ergonomic sugar over `vars`. The typed blocks compile to the same wire payload; on Read, both the typed block (when keys match) AND the raw `vars` map are populated in state for forward-compatibility with future Kibana additions.
- Add `elasticstack_fleet_cloud_connectors` data source backed by `GET /api/fleet/cloud_connectors`, exposing the `kuery` parameter for server-side filtering plus pagination.
- Introduce a reusable `internal/utils/writeonlyhash` helper that hashes write-only attribute values into per-resource private state (using bcrypt + per-resource-type salt) so that silent in-config edits to write-only secrets are detected at plan time. This helper is structured for adoption by other secret-bearing resources (`kibana_action_connector`, etc.) in follow-up work, without refactoring them in this change.
- Add `force_delete` attribute on the resource to surface the API's `?force=true` query parameter; default `false` so a delete fails (with a helpful error mentioning `package_policy_count`) when the connector is still in use.
- Gate the resource behind the minimum Kibana version that exposes the endpoints; cloud connector naming is preview in 9.2 and generally available in 9.3+.

## Capabilities

### New Capabilities

- `fleet-cloud-connector`: Defines the schema and runtime behavior of the `elasticstack_fleet_cloud_connector` resource, including the typed-blocks-plus-vars dual representation, write-only secret handling with hash-based drift detection, and the force-delete passthrough.
- `fleet-cloud-connectors-datasource`: Defines the `elasticstack_fleet_cloud_connectors` data source backed by the list endpoint, including server-side `kuery` filtering and pagination semantics.
- `writeonly-secret-hashing`: A reusable provider utility for storing bcrypt hashes of write-only secret values in per-resource private state so that plan-time drift detection works for write-only attributes without requiring user-managed version companions.

### Modified Capabilities

<!-- None — this change adds new capabilities only. -->

## Impact

- **New code**: `internal/fleet/cloudconnector/` (resource), `internal/fleet/cloudconnector/datasource/` (data source), `internal/clients/fleet/cloud_connector.go` (thin client wrappers), `internal/utils/writeonlyhash/` (reusable helper).
- **New docs/examples**: `docs/resources/fleet_cloud_connector.md`, `docs/data-sources/fleet_cloud_connectors.md`, `examples/resources/elasticstack_fleet_cloud_connector/`, `examples/data-sources/elasticstack_fleet_cloud_connectors/`.
- **Provider registration**: register the new resource and data source in the Plugin Framework provider entrypoint.
- **Dependencies**: adds `golang.org/x/crypto/bcrypt` (or equivalent) for the write-only hash helper if not already present.
- **Generated clients**: no `kbapi` regeneration needed — `PostFleetCloudConnectors`, `GetFleetCloudConnectorsCloudconnectorid`, `PutFleetCloudConnectorsCloudconnectorid`, `DeleteFleetCloudConnectorsCloudconnectorid`, and `GetFleetCloudConnectors` are already generated.
- **Acceptance test infra**: requires Kibana ≥ 9.2 (cloud connectors are preview-in-9.2 and GA-in-9.3); skip-gating uses the existing `entitycore.VersionRequirement` pattern.
- **Backward compatibility**: additive only — no breaking changes to existing resources or data sources.
