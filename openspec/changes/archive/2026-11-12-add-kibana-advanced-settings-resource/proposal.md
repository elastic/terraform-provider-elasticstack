# Proposal

## Why

Teams that manage Kibana declaratively with Terraform have no native way to manage Kibana Advanced
Settings (for example `dateFormat:tz`, `defaultRoute`, and other UI/behavior toggles). The only
workaround today is `elasticstack_kibana_import_saved_objects` with an exported saved-object blob.
That approach is all-or-nothing, unreadable in `terraform plan`, makes targeted changes and
validation hard, defeats drift detection on individual settings, and is error-prone during upgrades
and environment promotions. A customer has asked for first-class, reviewable, per-key management of
these settings as part of their infrastructure-as-code workflow.

## What Changes

- Add a new managed resource, `elasticstack_kibana_advanced_settings`, that manages Kibana Advanced
  Settings for a single Kibana space, identified by `space_id` (defaulting to the `default` space).
- The resource takes setting values in two maps keyed by setting name: one for plain-string values
  (the common case, such as a timezone or default route) and one for JSON-encoded values of other
  types (booleans, numbers, arrays, objects). Keys are passed through to Kibana without a
  provider-side allow list, so settings introduced by future Kibana versions, or by specific
  solutions, work without a provider release.
- The resource manages **only the keys declared in configuration**. Settings that exist in the space
  but are not in configuration (for example, ones an operator set through the Kibana UI) are left
  untouched and do not produce drift.
- Create and update apply the declared keys. Removing a key from configuration resets that setting to
  its Kibana default. Destroying the resource resets the keys it manages to their Kibana defaults
  and does not touch any other setting.
- Read refreshes the managed keys from Kibana so that out-of-band changes to a managed key show up
  as drift in the plan.
- Validate the declared keys and values at plan time where Kibana's limits are known: keys must be
  1-1024 characters, a key may appear in only one of the two maps, and JSON values must be valid
  and non-null. An empty configuration is accepted and manages nothing. Kibana's settings routes
  silently accept a space that does not exist, so for spaces other than `default` the provider
  checks that the space exists through the existing spaces API: if the space no longer exists, the
  resource is dropped from state on refresh and destroy succeeds, and create and update fail with a
  clear "the space must exist" error and write nothing.
- Kibana's own validation is the source of truth for known keys. Invalid values for registered
  settings, and attempts to change settings pinned by the Kibana server configuration, are surfaced
  to the user as clear errors rather than being re-implemented in the provider.
- Add provider-side client support for the Kibana advanced-settings HTTP endpoints, which exist in
  Kibana but are not described in the OpenAPI export that the provider's generated Kibana client is
  built from.
- Support the standard resource-level `kibana_connection` override, consistent with other Kibana
  resources.
- Keep the resource's identity and schema compatible with later adding support for Kibana's global
  (cross-space) advanced settings, so that follow-up work is additive and does not require breaking
  changes or state migration for users of this resource.
- Add acceptance coverage, generated documentation, and an example configuration.

Background and sources: the [Kibana advanced settings reference](https://www.elastic.co/docs/reference/kibana/advanced-settings)
and Kibana's UI settings service. The endpoints are defined in Kibana's
`src/core/packages/ui-settings/server-internal/src/routes/` (`get.ts`, `set.ts`, `set_many.ts`,
`delete.ts`).

### Non-goals

- Dedicated typed attributes for individual known settings (such as `date_format_tz` or
  `default_route`) and any provider-maintained registry of known keys. Kibana has hundreds of
  settings that vary by version and solution, so a hard-coded list would go stale; Kibana already
  validates registered keys on write.
- Managing global (cross-space) advanced settings in this change. Only the extensibility described
  above is in scope.
- A data source for reading Advanced Settings.
- `terraform import` support. Because the resource manages only the keys declared in configuration,
  the set of keys to adopt on import is ambiguous; this is deferred.
- Managing settings that are set through Kibana server configuration (`kibana.yml`
  `uiSettings.overrides`); the provider reports Kibana's refusal rather than working around it.

## Capabilities

### New Capabilities

- `kibana-advanced-settings`: Behavior of the `elasticstack_kibana_advanced_settings` resource:
  per-space scoping, managing only declared keys, passthrough of unknown keys, reset-to-default on
  key removal and on destroy, drift detection on managed keys, error surfacing for Kibana
  validation and server-pinned settings, plan-time validation of keys and values, treating an
  empty configuration as a no-op, handling a missing space (checked through the spaces API), reporting a missing settings route as an error, and the `kibana_connection` override.
  (Forward compatibility with global advanced settings is a design constraint on schema and
  identity, captured in `design.md`, not a runtime requirement in this spec.)

### Modified Capabilities

<!-- None. No existing capability's requirements change. -->

## Impact

- **New code**: a new resource package under `internal/kibana/`, registered in the provider; a
  Kibana client wrapper under `internal/clients/kibanaoapi/` for the advanced-settings endpoints.
- **Generated client**: the advanced-settings endpoints are absent from the Kibana OpenAPI export
  used to generate `generated/kbapi`, so this change must decide how to provide typed client
  coverage for them (resolved in design).
- **Docs and examples**: new generated resource documentation and an example under `examples/`;
  changelog entry.
- **Tests**: new acceptance tests (including a non-default space) and unit tests; no change to
  existing tests expected.
- **Compatibility**: purely additive. No existing resource, schema, or state changes. The
  advanced-settings endpoints exist in every Kibana release checked but are internal routes
  (no public API contract) that Kibana 9.x only serves when a request carries an internal-origin
  header, so Elastic could change them. No product version floor has been found, and behavior has been
  verified only on current stateful Kibana, so the acceptance tests are gated to stack versions
  9.5.0 and later (see `design.md`).
- **Permissions**: the Terraform identity needs Kibana's `manage_advanced_settings` privilege to
  write settings; this is called out in the resource documentation. The space-existence check also
  reads the space through Kibana's spaces API, which needs read access to the target space
  (verified live: credentials holding only `advancedSettings: all` in one space get 200 for that
  space and 403 for any other, and the provider surfaces that 403 as an error).
- **Specs**: adds one new canonical capability, `kibana-advanced-settings`, on archive.
