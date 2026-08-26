## Context

See proposal.md — Why. Design-relevant constraints in the existing implementation:

- `internal/fleet/integration_policy/models.go` builds the request body in a single `toAPIModel(ctx, feat)` function that **both** Create and Update call. It receives only the plan model and an `integrationPolicyFeatures` struct — it has no access to prior state, so any "was set, now removed" logic cannot live there without new plumbing.
- `integrationPolicyFeatures` (`capabilities.go`) already resolves per-server capability bits via `client.EnforceMinVersion`, and `toAPIModel` already uses one of them (`SupportsPolicyIDs`) to decide whether to send an empty `policy_ids` array for clearing.
- `GetVersionRequirements` on the model declares attribute-level version floors that `entitycore.EnforceVersionRequirements` enforces before any API call, and it only reports a requirement when the attribute is known.
- `populateFromAPI(ctx, pkg, data)` is shared by Create, Update, and Read. For non-`Computed` attributes, whatever it writes must match the plan exactly on Create/Update, or Terraform raises "Provider produced inconsistent result after apply".
- The generated client already carries the field: `PackagePolicyRequestMappedInputs.AdditionalDatastreamsPermissions` on the request and `PackagePolicy.AdditionalDatastreamsPermissions` on the response, both `*[]string`.
- `internal/fleet/managedintegration` is a working precedent for the same field on a different resource, but it uses a different write path (`toCreateBody` with an explicit `sendExplicitEmptyScalars` option and prior-state comparison in `update.go`).

## Goals / Non-Goals

**Goals:**

- Add the attribute without introducing a prior-state parameter into `toAPIModel`, keeping Create and Update on one code path.
- Reuse the resource's existing capability-bit and version-requirement mechanisms rather than inventing a new gate.
- Guarantee the attribute round-trips so no plan diff appears after apply, on supporting and non-supporting servers alike.

**Non-Goals:**

- Reproducing `managedintegration`'s `sendExplicitEmptyScalars` machinery here. That resource has a broader in-place-update allowlist; this change needs one field.
- Local validation of permission values against space namespace prefixes. Kibana owns that rule and its error message is more accurate than anything the provider could reproduce.
- Refactoring `toAPIModel` to take prior state. That is a larger change affecting every attribute and is not needed for the chosen approach.

## Decisions

### Clear by always sending an empty array on supporting servers

`toAPIModel` will send the configured values when the attribute is known and non-null, and send an explicit empty array when the attribute is null **and** the resolved capability bit says the server supports the field. On servers below 9.1.0 the field is omitted entirely.

This is exactly the shape `policy_ids` already uses a few lines above in the same function, so it needs no new state plumbing: "unset" and "was set, now removed" both arrive at `toAPIModel` as a null plan value, and sending `[]` is correct for both. The cost is a redundant `[]` on every create against a 9.1+ server, which is a no-op server-side.

Alternatives considered:

- **Compare prior state in `Update` and only send `[]` when the attribute was previously set.** More precise, but requires threading prior state into `toAPIModel` (or a second body-mutation step after it), which complicates the one path both Create and Update share. Precision buys nothing here because `[]` is idempotent.
- **Always send the field, empty array when unset.** Rejected: on servers below 9.1.0 an unconfigured attribute would start sending an unsupported field, breaking existing 8.x configurations that never asked for this feature.

### Reject an empty list at plan time instead of supporting it as a value

A schema-level `listvalidator.SizeAtLeast(1)` rejects `additional_datastreams_permissions = []`, mirroring `agent_policy_ids` in the same schema.

Without this, `[]` in configuration is a trap: the API returns an empty array or omits the field, `populateFromAPI` would map that to null, and the apply fails with "Provider produced inconsistent result after apply" on a non-`Computed` attribute. The alternatives were to carve out an exception in `populateFromAPI` that preserves a known empty list, or to make the attribute `Optional+Computed`. Both add subtle state-handling rules to support a value that means the same thing as omitting the attribute. Rejecting it at plan time keeps `populateFromAPI` to a single rule — non-empty API list becomes state, anything else becomes null — and gives the user a clear message pointing at the supported way to revoke permissions.

### Keep the attribute `Optional` and not `Computed`

The user either grants extra permissions or does not; there is no meaningful server-computed default to adopt. Leaving it non-`Computed` means an unconfigured attribute is null in state and stays null, which also makes the null-versus-empty question above decidable. `agent_policy_ids` and `output_id` on this resource are `Optional`-only for the same reason.

### Gate at Kibana 9.1.0 via both existing mechanisms

Add `MinVersionAdditionalDatastreamsPermissions = 9.1.0` in `resource.go`, report it from `GetVersionRequirements` when the attribute is known (so a user on 8.x gets an attribute-level error rather than an opaque Fleet 400), and resolve a matching `SupportsAdditionalDatastreamsPermissions` bit in `resolveIntegrationPolicyFeatures` (so the clearing behaviour above knows whether it may send `[]`).

Both are needed and they answer different questions: the version requirement guards *configured* use, the capability bit guards *unconfigured* clearing. Kibana PR elastic/kibana#210452 carries `backport:skip` and `v9.1.0`, so 9.1.0 is a hard floor with no 8.x variant to accommodate.

### No schema version bump or state upgrader

Adding an optional attribute to schema V3 does not invalidate existing V3 state: Terraform reads the missing attribute as null, which is the correct value for a policy that never had permissions configured. A bump would force a fourth upgrader that does nothing but copy fields.

## Risks / Trade-offs

- **A policy whose permissions were granted outside Terraform is silently cleared on the next apply.** → This follows from sending `[]` when the attribute is unset, and it is the standard Terraform contract: configuration is authoritative. Mitigated by documenting the behaviour on the attribute and by making import populate the attribute, so an imported policy carries its existing permissions into state and the user can keep them.
- **Acceptance tests need a Kibana 9.1.0+ stack and will skip below it.** → Follow the existing skip-helper pattern used for other version-gated attributes in this package so the suite still passes against older stacks rather than failing.
- **The redundant empty array on create could mask a future Kibana change that treats `[]` differently from omission.** → Acceptance coverage asserts the API-side state after create with the attribute unset, so a behaviour change surfaces as a test failure rather than as silent permission loss.
- **`SizeAtLeast(1)` is a hard error for anyone who writes `[]` expecting it to mean "none".** → The validator message names the supported alternative (remove the attribute), and the docs say the same.
