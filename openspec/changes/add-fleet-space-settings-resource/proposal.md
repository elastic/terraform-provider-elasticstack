## Why

In multi-team Kibana environments that use one space per team, Fleet's per-space setting `allowed_namespace_prefixes` restricts which data stream namespaces may be written. This is a security control, but the provider cannot manage it today, so it cannot be code-reviewed or drift-detected. The only alternatives are manual per-space configuration or `terraform_data` with `local-exec`, neither of which provides state management or drift detection. Fleet space awareness became the default in Elastic Stack 9.1, so multi-space deployments are increasingly common.

Source: [elastic/terraform-provider-elasticstack#4980](https://github.com/elastic/terraform-provider-elasticstack/issues/4980).

## What Changes

- Add a new Terraform resource `elasticstack_fleet_space_settings` that manages Fleet's per-space settings for one Kibana space, backed by Fleet's space settings API.
- Practitioners can declaratively set `allowed_namespace_prefixes` for a space and observe the read-only `managed_by` value reported by Fleet.
- Destroying the resource resets the space's `allowed_namespace_prefixes` to empty rather than removing anything, because the settings object always exists for a space.
- The API limits writes to 10 prefixes per request (per issue #4980; to be verified in design); the resource rejects larger lists clearly rather than work around the limit. Duplicate prefixes are also rejected, since they carry no meaning in an allow-list. The 10-element limit and the 9.1.0 minimum version were confirmed against the Kibana API documentation.
- Add documentation, an example, and acceptance tests for the resource.

Non-goals:

- No data source; only the managed resource.
- No management of global Fleet settings (`/api/fleet/settings`).
- The resource does not create Kibana spaces; the space must already exist.
- No workaround for the 10-element write limit (e.g. batching); it is validated, not bypassed.

## Capabilities

### New Capabilities
- `fleet-space-settings`: Requirements for the `elasticstack_fleet_space_settings` resource (schema, identity/import, create/update/read/delete behavior, version compatibility, validation).

### Modified Capabilities
<!-- None. Existing Fleet specs are unaffected. -->

## Impact

- New Fleet resource in the provider; the required API operations already exist in the generated Kibana client, so no client regeneration is expected.
- User-facing documentation, an example, and a changelog entry.
- Acceptance tests requiring a Kibana/Fleet version that supports space settings; the minimum version is to be confirmed in design.
