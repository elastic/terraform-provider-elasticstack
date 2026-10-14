## Why

`elasticstack_fleet_agent_download_source` manages Fleet's agent binary download source settings, but its schema never maps the Kibana `auth` request field on either `POST /api/fleet/agent_download_sources` or `PUT /api/fleet/agent_download_sources/{sourceId}`. `auth` is how Elastic Agents authenticate to a private/self-hosted artifact registry when downloading binaries — HTTP Basic (`username`+`password`), API key (`api_key`), and/or custom HTTP headers sent on every request. There is no opaque/passthrough attribute on this resource today, so a practitioner configuring a private download host cannot set credentials through Terraform at all; they must configure `auth` out-of-band through the Kibana UI and the Terraform resource silently ignores it.

This closes the gap flagged in [`#5241`](https://github.com/elastic/terraform-provider-elasticstack/issues/5241): the Kibana API gap scanner found `auth` present on both the create and update request bodies' schema but absent from the Terraform schema and model.

## What Changes

- Add an optional `auth` `SingleNestedAttribute` to `elasticstack_fleet_agent_download_source`, mirroring the existing `output` resource's `kafka` block shape: `username` (plain string), `password` (string, sensitive), `api_key` (string, sensitive), `headers` (list of `{key, value}`, max 100 entries).
- Enforce the same credential-mode exclusivity Kibana's server validates: `username` requires `password` (`AlsoRequires`) and `password`/`api_key` are mutually exclusive (`ConflictsWith`), reusing the relative-path validator pattern from `internal/schema/es_connection_shared.go`. `headers` stays independent and valid on its own (headers-only `auth` block).
- Map `auth` into `toAPICreateModel`/`toAPIUpdateModel`. On update, when the `auth` block is removed from configuration, send `auth: {}` explicitly (verified live against Kibana 9.4.0: omitting the `auth` key preserves the existing credentials via merge semantics, while `auth: {}` clears them) rather than omitting the key.
- Extend read hydration (`readAndHydrateState`) to decode `auth.username` and `auth.headers` verbatim from the API response (both come back unredacted), and apply the `kafka.password`-style nil-pointer preservation pattern to `auth.password`/`auth.api_key`: Kibana omits both fields from every read response (moving the real secret into its internal secret store and returning only an opaque, drift-unusable reference id), confirmed by live probing of POST/GET/PUT/list responses against Kibana 9.4.0 — there is no masked placeholder to detect.
- Add a per-field minimum-version requirement (9.4.0) on `auth`, following the `agentpolicy/version_requirements.go` `addMinimum` pattern, leaving the resource's existing 8.13.0 baseline untouched when `auth` is absent from configuration.
- Scope `ImportStateVerify` for `auth.password`/`auth.api_key` with a targeted `ImportStateVerifyIgnore` plus a custom `ImportStateCheck` asserting both are null after import (following the `apm/source_map` / `security/user` precedent), since there is no prior Terraform state to fall back to on first import and the API never echoes those two fields.
- Update the existing OpenSpec requirement "Auth and secrets not in v1 schema" to reflect that `auth` (credentials and headers) is now in scope; the `secrets.auth` secret-store-reference mechanism (referencing an existing Fleet secret by ID instead of supplying a literal value) remains explicitly out of scope for this change.

## Capabilities

### Modified Capabilities

- `fleet-agent-download-source`: Adds the `auth` nested attribute (username/password/api_key/headers) to the resource schema, including credential-mode validators, version gating, update-time clearing semantics, and read-time secret preservation/redaction handling.

## Impact

- **Changed code**: `internal/fleet/agentdownloadsource/schema.go` (new `auth` attribute + validators), `internal/fleet/agentdownloadsource/models.go` (create/update mapping, version requirement), `internal/fleet/agentdownloadsource/read.go` (`readAndHydrateState` gains an `auth` fallback parameter; `auth` decode), `internal/fleet/agentdownloadsource/write.go` (`finalizeWrite` passes the planned `auth` as read-back fallback).
- **New/updated acceptance tests**: `internal/fleet/agentdownloadsource/acc_test.go` and `testdata/TestAccResourceFleetAgentDownloadSource*/` — basic auth, API key, headers-only, password/API-key rotation, mode switching, removal (`auth_remove`), empty object (`auth = {}`), header count boundaries (100/101), validator `ExpectError` cases, import with server-side-only credentials, and a negative version-gate case.
- **Docs**: `docs/resources/fleet_agent_download_source.md` and `examples/resources/elasticstack_fleet_agent_download_source/` will need an `auth` example once implemented (not produced in this proposal-only change).
- **No `generated/kbapi` changes**: the generated client already exposes `Auth` on both request bodies and the response item; no regeneration is required.
- **Backward compatibility**: additive only. Existing configurations that omit `auth` are unaffected: create omits `Auth`, and update omits it when it was also absent from prior state. If a previously configured `auth` block is removed, update sends an explicit empty `Auth` object to clear the server-side credentials.
