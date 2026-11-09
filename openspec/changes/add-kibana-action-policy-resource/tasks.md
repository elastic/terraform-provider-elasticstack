# Tasks

## 1. Practitioner can create, read, update and delete a workflow-routed action policy in the default space (experimental opt-in)

- [ ] 1.1 Add the action policy client wrapper in `internal/clients/kibanaoapi/action_policy.go` covering create (POST), get, full replace (PUT) and delete, with space-aware routing. Get reports not-found without an error, delete treats 204 and 404 as success, and other statuses surface the HTTP status and Kibana's message. Verify with wrapper unit tests against an `httptest` server.
- [ ] 1.2 Add `elasticstack_kibana_action_policy` in `internal/kibana/actionpolicy/` on the entitycore Kibana envelope with `space_id`, computed `policy_id` and `id`, `name`, `description` and the `destinations` set. Create uses POST and records the server-generated identifier; update uses a full PUT; final state comes from read-after-write, ignoring snooze and audit fields. Verify with an acceptance test covering create (asserting `policy_id`/`id` hold the server identifier), in-place rename, reordered destinations giving an empty plan, an out-of-band snooze giving an empty plan, and destroy.
- [ ] 1.3 Validate the slice-1 attributes at plan time: destinations need at least one entry with type `workflow` only, and `name` and `description` must be non-empty and have no surrounding whitespace. Verify with validator unit tests and `ExpectError` acceptance steps for an unsupported destination type, no destinations, a name with a trailing space, and an empty description.
- [ ] 1.4 Make refresh remove a policy deleted outside Terraform, and make destroy succeed when the policy is already gone. Verify with acceptance steps that delete the policy out of band.
- [ ] 1.5 Gate the resource at Kibana 9.6.0, so older stacks fail with `Unsupported server version` before any API call. Surface 400/403/409/503 and transport failures as diagnostics with the status and Kibana's message. Verify with unit tests for the version gate and error mapping, and with acceptance tests skipped below 9.6.0.
- [ ] 1.6 Honor a resource-level `kibana_connection` override for all API calls. Verify with an acceptance step following the existing Kibana `kibana_connection` acceptance pattern.
- [ ] 1.7 Register the resource only in the provider's experimental resource set, so it is unknown without `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL` and available with it. Verify with a new provider registration unit test asserting absence without the variable and presence with it.
- [ ] 1.8 Write `resource-description.md` and the generated docs with two notices, and add a CHANGELOG "Added" entry marked experimental. Verify that docs generation and the docs check pass and the generated page contains each notice.
  - Availability notice: technical preview; Kibana 9.6.0+; Enterprise license; Alerting v2 enabled; unavailable in serverless Elasticsearch and Security; routing via rule routing tags.
  - Credentials notice: last-writer binding re-bound on every update; the Elasticsearch API key caveat; Cloud API keys rejected; required Kibana privileges.

## 2. Practitioner can control which alerts a policy routes, and how they are grouped and throttled

- [ ] 2.1 Map `matcher` (the `tags` set and the KQL `expression`), `grouping` (`per_alert`/`all`/`per_field` with `fields`) and `throttle` (four strategies with `interval`) to and from the API unions. Absent blocks read back as null, and no Kibana defaults enter state. Verify with mapping round-trip unit tests for every mode, strategy and matcher combination, and with an acceptance test that uses tags plus an expression, per-field grouping and a time-interval throttle, then expects an empty follow-up plan.
- [ ] 2.2 Clear `description`, `matcher`, `grouping` and `throttle` on the policy when they are removed from configuration; removing `matcher` makes the policy catch-all. Verify with acceptance steps that remove each block and assert the stored policy and an empty follow-up plan.
- [ ] 2.3 Validate matcher, grouping and throttle shape at plan time. Verify with validator unit tests and `ExpectError` acceptance steps. The rules:
  - enum values for `mode` and `strategy`
  - `fields` required and non-empty only for `per_field`, with entries that are non-empty and have no surrounding whitespace
  - `interval` required only for interval strategies, in the duration format
  - a `matcher` with at least one of `tags`/`expression`
  - non-empty `tags`
  - an `expression` without surrounding whitespace
- [ ] 2.4 Leave grouping/throttle compatibility and size limits to Kibana. Verify with an acceptance step where an incompatible grouping/throttle pair plans successfully and fails at apply with Kibana's message.
- [ ] 2.5 Document routing configuration. Verify that docs generation includes the example and the limitation.
  - Add an example under `examples/resources/elasticstack_kibana_action_policy/` showing tag routing, grouping and throttle.
  - Note in the resource description that Kibana does not validate the KQL `expression` when the policy is written, so an invalid expression matches nothing.

## 3. Practitioner can choose the policy identifier, place policies in named spaces, and import existing policies without overwriting them

- [ ] 3.1 Support an optional `policy_id`: 1–150 characters of `[a-zA-Z0-9_-]`, validated at plan time. When it is set, create checks for an existing policy in the space and fails with an import hint without writing; otherwise create uses that identifier. Verify with `httptest` unit tests for the guard (existing → no write; missing → create with that id) and acceptance steps for a configured id and for the import-hint error.
- [ ] 3.2 Force replacement when `space_id` or `policy_id` changes, and create policies in a named space. Verify with acceptance steps that create a policy in a non-default space and assert replacement on a space change.
- [ ] 3.3 Import policies by `<space_id>/<policy_id>`, populating every configurable attribute from Kibana, and reject import ids without a space segment. Add `examples/resources/elasticstack_kibana_action_policy/import.sh` so the generated docs show the import format. Verify with an acceptance import step with state verification, an `ExpectError` step for a malformed id, and the import section on the generated page.

## 4. Practitioner can enable or disable a policy, and a failed enable/disable never orphans the policy

- [ ] 4.1 Add the opt-in persist-on-error option to the entitycore write result. When a write callback sets it and returns errors, the envelope skips read-after-write, saves the returned model (including timeouts) and returns the errors. The default leaves every existing resource unchanged. Verify with envelope unit tests for the opt-in and default paths across the Kibana and Elasticsearch adapters.
- [ ] 4.2 Manage `enabled` as optional and computed through new enable and disable client wrapper functions.
  - When configured, create and update bring the policy to that state, calling the dedicated endpoints only on a mismatch.
  - When not configured, the server value is recorded and never enforced, including when the plan runs without refresh.

  Verify with:
  - wrapper and `httptest` unit tests for the reconcile decision (config null, matching, mismatching)
  - acceptance steps for creating a disabled policy, re-enabling it, and an out-of-band disable with `enabled` unset
  - the import step's state verification now covering `enabled`
- [ ] 4.3 Keep the policy in state when enable/disable fails after a successful write: create saves it (Terraform marks it tainted), and update saves the written fields with the server's `enabled`. Verify with `httptest` unit tests that simulate a failing disable after create and after update, asserting the persisted state and the returned error.
