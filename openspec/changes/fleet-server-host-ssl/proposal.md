## Why

`elasticstack_fleet_server_host` never sends or reads back the Kibana `ssl` field on `POST /api/fleet/fleet_server_hosts` and `PUT /api/fleet/fleet_server_hosts/{itemId}`. That field carries the TLS configuration Fleet Server presents to Elastic Agents (certificate, key, trusted CAs, and client-auth mode). Without it, practitioners cannot manage TLS for a Fleet Server host through Terraform at all — a significant gap for production use, since enabling mutual TLS between agents and Fleet Server is a common hardening requirement. This closes the `kibana-api-gap` triage on [`#5366`](https://github.com/elastic/terraform-provider-elasticstack/issues/5366).

## What Changes

- Add an optional `ssl` single-nested-object attribute to `elasticstack_fleet_server_host` with sub-attributes `certificate` (string), `key` (string, sensitive), `certificate_authorities` (list of string, max 10 entries), and `client_auth` (string enum: `none`, `optional`, `required`). This mirrors the primary (non-`agent_*`/`es_*`) fields of the Kibana `ssl` object, using `internal/fleet/agentdownloadsource`'s `ssl` block as the direct structural template and the shared `internal/fleet/sslcore` helpers for certificate/key/CA conversion.
- Map `ssl` to and from the `ssl` object on the Fleet create, update, and read response bodies.
- Reject, at plan time, an `ssl` object that is present but all-null (extending the `agentdownloadsource` all-nil-subattribute validator pattern to include `client_auth`).
- Gate `ssl` with an attribute-scoped minimum Kibana version requirement (9.1.0, the version that introduced the field per Kibana release notes), applied only when `ssl` is configured — existing configurations that never set `ssl` are unaffected on any stack version. This requires adding a `GetVersionRequirements` method to `serverHostModel`, which does not currently implement one.
- Determine and apply the correct update-removal semantics (in-place clear vs. forced replacement) for `ssl` based on what the Fleet API actually does when a previously-set `ssl` is omitted, sent as an explicit empty object, or sent as explicit JSON `null` on `PUT` — each behavior will be probed independently during implementation, per `agentdownloadsource`'s precedent for the same open question on its own endpoint.
- Determine and apply the correct hydration behavior for `ssl.key` on read-after-write and import, handling the case where the API redacts the key the same way `agentdownloadsource` does for its secret-backed key, if that is what testing shows.

## Capabilities

### New Capabilities

<!-- None — this change extends an existing capability. -->

### Modified Capabilities

- `fleet-server-host`: Adds the `ssl` nested attribute (certificate, key, certificate_authorities, client_auth), its plan-time validation, its attribute-scoped version gate, and its create/update/read mapping to the Kibana `ssl` object.

## Impact

- **Changed code**: `internal/fleet/serverhost/schema.go` (new `ssl` attribute), `internal/fleet/serverhost/models.go` (new `GetVersionRequirements`, `ssl` mapping in `populateFromAPI`/`toAPICreateModel`/`toAPIUpdateModel`), new `internal/fleet/serverhost/models_ssl.go` and `internal/fleet/serverhost/ssl_validator.go` (or equivalent, following `agentdownloadsource`'s file layout) for the `ssl` conversion and validation logic.
- **No generated-client changes**: `kbapi.PostFleetFleetServerHostsJSONBody.Ssl` / `PutFleetFleetServerHostsItemidJSONBody.Ssl` already exist in `generated/kbapi/kibana.gen.go`; no client regeneration is required.
- **Docs/examples**: `docs/resources/fleet_server_host.md` and `examples/resources/elasticstack_fleet_server_host/resource.tf` gain `ssl` usage; a CHANGELOG entry is added.
- **Acceptance tests**: new coverage in `internal/fleet/serverhost/acc_test.go` for `ssl` create/update/removal/validation and version gating, modeled on `internal/fleet/agentdownloadsource/acc_test.go`.
- **Out of scope (explicitly deferred)**: the `agent_certificate`/`agent_certificate_authorities`/`agent_key` and `es_certificate`/`es_certificate_authorities`/`es_key` sub-fields of `ssl`, and the `secrets.ssl` Fleet-managed-secret storage mechanism for any of these keys. These are lower-risk follow-ups once this primary block ships.
- **Backward compatibility and migration**: existing resources with no server-side `ssl` remain unaffected when configuration omits `ssl`. Because read hydration will expose an already-configured server-side `ssl`, users managing such hosts must add the existing SSL settings to configuration as part of the provider upgrade; otherwise Terraform will plan to remove `ssl` (or replace the host if the API cannot clear it in place). This is a documented configuration migration, not a no-diff guarantee for remotely configured SSL. No state upgrader is needed.
