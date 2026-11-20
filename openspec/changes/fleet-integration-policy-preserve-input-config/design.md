## Context

- `Update` (`internal/fleet/integration_policy/update.go`) builds a `PackagePolicyRequestMappedInputs` body with `toAPIModel` and sends it with `fleet.UpdatePackagePolicy` (`PUT /api/fleet/package_policies/{id}?format=simplified`).
- In the simplified format Kibana regenerates each input from the package template (`simplifiedPackagePolicytoNewPackagePolicy`), so fields that are not part of the mapped-input shape are dropped. The generated `PackagePolicyRequestMappedInput` type has no `config` field; `PackagePolicyRequestTypedInput` and `PackagePolicyRequestTypedInputStream` do.
- Verified against Kibana 9.4.2: after creating an APM package policy, Kibana's APM plugin writes `inputs[0].config["apm-server"].value.{rum.source_mapping,agent.config}.elasticsearch.api_key` (revision 2). A simplified PUT (revision 3) leaves `config["apm-server"].value` in place but without both `elasticsearch` objects. A subsequent typed PUT carrying only `version`, `policy_id`, `policy_ids` and `inputs` (with the keys restored) keeps every other top-level field (name, namespace, description, policy ids, package) and the values the simplified PUT set.
- `fleet.GetDefendPackagePolicy` / `fleet.UpdateDefendPackagePolicy` already perform typed-format GET/PUT for the Defend resource.

## Goals / Non-Goals

**Goals:**

- Never lose input- or stream-level `config` on update, regardless of package.
- Keep the simplified API as the primary write path so package upgrades, var defaults and stream generation keep their current behaviour.
- Add no writes when nothing needs restoring.

**Non-Goals:**

- Exposing `config` as an attribute or detecting drift in it.
- Changing create or read. Kibana callbacks populate `config` after create; read does not model it.

## Decisions

### Restore after the simplified update instead of replacing it with a typed update

A single typed update built from the current policy plus the plan would avoid a second revision, but it would have to re-implement what Kibana does for simplified requests: mapping `<policy_template>-<type>` input ids and dataset-keyed stream ids to typed inputs, filling package defaults for new vars, adding streams introduced by a package upgrade, and handling secret references. Keeping the simplified update and repairing only `config` reuses Kibana's own logic and limits the new code to a well-defined merge.

Cost: when `config` was lost there are two policy revisions in quick succession. Agents converge on the second one within seconds. For packages without `config` (the large majority) there is no second revision.

### Match inputs by type and policy template, streams by dataset

Input ids change when a policy moves from the typed to the simplified representation (for example `<package-policy-id>` becomes `apm-apmserver-<package-policy-id>`), so ids are not stable. An input is identified by `type` plus `policy_template`; a stream within it by `data_stream.dataset`. These are the same identities the simplified format derives its keys from. Identities are counted over all inputs and streams, with or without `config`, both before and after the update; only identities that are unique on both sides are restored, so `config` can never land on the wrong element or on more than one.

### Deep fill, current values win

For each matched input/stream, every key of the prior `config` that is missing in the current `config` is copied over, including keys whose prior value is null; a key that is null now is replaced by a non-null prior value; when both sides hold objects the merge recurses; when both sides hold a value the current one is kept. The typed request type omits a null `value`, `type` or `frozen` on a config entry, so a restored null at that level cannot be sent; when the restored inputs encode to the same request as the current inputs, no restoring write is sent. This restores data that the update dropped without undoing anything the update (or a Kibana callback that ran during it, such as the APM plugin rewriting `rum.source_mapping.metadata`) wrote.

### Send only `version`, the agent policy assignment and `inputs` in the restoring request

Kibana merges a typed update body over the stored policy, so omitting other fields leaves them untouched. The exception is the agent policy assignment: Kibana 8.x clears `policy_id` on a typed update that omits `policy_ids`, so `policy_id` and `policy_ids` are sent back exactly as read. Sending the `version` token from the read immediately preceding the write makes a concurrent modification fail with a conflict rather than be overwritten. Compiled fields (`compiled_input`, `compiled_stream`) are stripped from the request.

### Detect concurrent writes by revision instead of a conditional simplified update

The snapshot is read before the simplified update, so another client could change `config` in between, for example Osquery deleting a pack. Restoring from the snapshot would then bring the deleted pack back. The simplified update cannot be made conditional on the snapshot: the generated `PackagePolicyRequestMappedInputs` type has no `version` field, and Kibana's `simplifiedPackagePolicytoNewPackagePolicy` does not carry `version` through to the saved-object update. Kibana does increment `revision` by exactly one on every package policy update (callbacks such as the APM plugin's mutate the policy inside that same write). So when a restoring write is needed and the restoring read's `revision` is not the snapshot's plus one, another write happened after the snapshot, either before or after the simplified update, and the resource returns an error without restoring. The restoring write's own `version` token covers the remaining window between the restoring read and the restoring write.

### Snapshot failures are errors

If the typed read before the update fails or does not find the policy, the update is not attempted: proceeding would risk exactly the silent data loss this change prevents. A failure after the simplified update has been applied (the restoring read, a policy that is no longer found, the stale-snapshot check, or the restoring write) is reported as an error. At that point any `config` the simplified update removed is gone from the server, and a later apply cannot bring it back because its snapshot is taken from the stripped policy. The resource therefore adds an error that says so, so the user knows to check and recreate the `config` (for example re-assign Osquery packs). Terraform keeps the prior state on an update error; the next refresh reads the policy as the simplified update left it.

## Risks / Trade-offs

- An extra GET on every update. Fleet package policy reads are cheap compared with the write and package-info lookups `Update` already performs.
- A Kibana callback that wrote the package policy a second time during an update would bump the revision by two. The resource would then report a concurrent modification instead of restoring. No callback for the packages this resource manages does that today, and the error is the safe outcome: restoring from a possibly stale snapshot could resurrect deleted data.
- If Kibana intentionally removes a `config` key during an update, this change re-adds it. No such behaviour exists today for simplified updates (the simplified path never computes `config`), and Kibana callbacks that do manage `config` keep their freshly written values because current values win.
