## Context

Fleet cloud connectors are reusable cloud-credential bundles used by agentless integrations (Cloud Security Posture Management and Cloud Asset Discovery, today) to authenticate against AWS, Azure, and GCP. Kibana exposes them under `/api/fleet/cloud_connectors` with full CRUD plus a list endpoint. The current `generated/kbapi` client in this repo does not yet include these operations, so implementing this change will require adding/regenerating the `kbapi` client support for the cloud connectors endpoints.

The non-trivial parts of this change are not in plumbing — the CRUD pattern matches the existing `internal/fleet/proxy` resource almost exactly — but in three modelling decisions:

1. **`vars` is a polymorphic map.** The API accepts each var value as one of four union arms: a bare `string`, `number`, or `boolean`, or a structured `{type, value, frozen?}` object whose `value` is itself either a string or a `{id, isSecretRef}` saved-secret reference. Terraform Plugin Framework has no native union type, so the schema must encode the union explicitly.
2. **Secrets must be write-only.** The API stores `secret_value` once on Create/Update, then only ever returns a secret reference id; the raw value is never readable again. The resource cannot round-trip the value through state, so attribute-level write-only support is required.
3. **Per-provider ergonomics matter.** Users coming from the CSPM/Asset Discovery docs expect to write `role_arn` and `external_id` (for AWS) or `tenant_id`/`client_id`/`cloud_connector_id` (for Azure) directly, not assemble a generic `vars` map by hand.

This document captures the resulting design. Implementation follows the existing `internal/fleet/proxy` skeleton with `entitycore.KibanaResource[model]`.

## Goals / Non-Goals

**Goals:**
- Full lifecycle (Create, Read, Update, Delete) of Fleet cloud connectors with import support.
- Faithful representation of the API's `vars` union — every arm must be expressible in HCL.
- Per-provider typed sugar (`aws { }`, `azure { }`) for AWS and Azure that compiles to the same wire payload as the generic `vars` map.
- Both the typed block AND the raw `vars` map populated in state after Read, so users can fall back to `vars` when Kibana adds keys the typed block doesn't yet model.
- Drift detection for ALL realistic scenarios — including silent in-config edits to write-only secret values — without requiring user-managed version companions.
- A reusable `internal/utils/writeonlyhash` helper that other secret-bearing resources in the provider can adopt later.
- Data source backed by the list endpoint with server-side `kuery` filtering.

**Non-Goals:**
- Typed GCP block. The API accepts `gcp` as a `cloudProvider`, but as of 9.2/9.3 Elastic has not published a documented agentless GCP flow. GCP users will use the generic `vars` map until a typed block can be defined against a stable documented shape.
- Retrofitting other secret-bearing resources (`kibana_action_connector`, etc.) to use the new write-only hash helper. The helper is built reusable; adoption is a follow-up change.
- Surfacing the `/usage` endpoint (`GET /api/fleet/cloud_connectors/{id}/usage`) as a separate data source. Package policy usage is already visible via `package_policy_count` on the resource.
- Bulk operations or any custom rotation actions. Standard Terraform Create/Update suffices.
- Backwards-compatible support for Kibana versions older than the first version that ships the cloud connectors API.

## Decisions

### Decision 1: Resource layout mirrors `internal/fleet/proxy`

Implement the resource as an `entitycore.KibanaResource[cloudConnectorModel]` with one file per CRUD verb (`create.go`, `read.go`, `update.go`, `delete.go`), `schema.go` for the Plugin Framework schema, `models.go` for the model and API conversions, and `resource.go` for registration. Thin client wrappers live in `internal/clients/fleet/cloud_connector.go`, mirroring the existing `proxy.go` wrappers.

**Why:** The proxy resource is the closest existing analogue (small, space-aware, composite ID, kibana_connection support, version-gated). Reusing the same shape minimises review effort and keeps the Fleet domain internally consistent. Considered: rolling a bespoke skeleton — rejected because `entitycore.KibanaResource` already provides space resolution, connection handling, version checks, and the read-after-write contract.

### Decision 2: Composite ID `<space_id>/<cloud_connector_id>`

Identity: `id` is computed and composite; `cloud_connector_id` is computed only (API-assigned on Create; not user-settable because the POST body has no `id` field); `space_id` defaults to `"default"` and forces replacement on change. Import accepts the composite form.

**Why:** Matches every other space-aware Fleet resource in the provider (`fleet_proxy`, `fleet_output`, `fleet_server_host`, etc.), so users get a consistent import idiom across the Fleet surface.

### Decision 3: `vars` schema exactly mirrors the API union

The `vars` attribute is a `map(object({...}))` whose nested object has six top-level fields, with a `ConfigValidator` per element enforcing the exclusivity rules:

```
vars = map(object({
  # Bare arms (1)-(3): primitive values
  string = optional(string)        # arm (1)
  number = optional(number)        # arm (2), wire is float32
  bool   = optional(bool)          # arm (3)

  # Structured arm (4)
  type   = optional(string)        # passthrough; integration-package types
                                   #   like "text", "password", "yaml"; the API
                                   #   imposes no enum so the provider must not
  frozen = optional(bool)          # only valid alongside `type`

  # Within arm (4), exactly one of:
  value             = optional(string)     # arm (4) Value0 — plain string
  secret_value      = optional(string,     # provider-side write sugar; raw
                              write_only,  #   secret sent to API once
                              sensitive)
  secret_ref        = computed(object({    # arm (4) Value1 — populated on Read
    id            = string
    is_secret_ref = bool                   # plain bool to faithfully echo wire
  }))

  secret_value_wo_version  = computed(number)   # private-state-driven; see D5
}))
```

`ConfigValidator` rules per element:
- Group A (bare): exactly one of `string`, `number`, `bool`; mutually exclusive with Group B.
- Group B (structured): `type` plus exactly one of `value`, `secret_value`, `secret_ref`. `frozen` only valid in Group B.
- `secret_ref` and `secret_value_wo_version` are computed-only; rejected if set in config.

**Why:** This is the smallest schema that captures all four arms faithfully without lossy conversion. A `DynamicAttribute` was considered but rejected because it cannot be sensitivity-marked per-key (whole-value only) and cannot host a write-only sub-attribute, both of which are essential for proper secret handling. A JSON-string fallback was rejected for the same reasons. The cost is a slightly verbose HCL surface for users using `vars` directly — mitigated by the per-provider typed blocks.

### Decision 4: Per-provider typed sugar with dual state population

The resource exposes optional + computed typed blocks for the well-known per-provider authentication flows:

```hcl
aws = optional + computed(object({
  role_arn                = optional(string)
  external_id             = optional(string, write_only, sensitive)
  external_id_secret_ref  = computed(object({ id, is_secret_ref }))
}))

azure = optional + computed(object({
  tenant_id           = optional(string)
  client_id           = optional(string)
  cloud_connector_id  = optional(string)
}))
```

Config-time `ConfigValidator`: exactly one of `aws`, `azure`, `vars` may be set in config. The typed block, if set, must match the resource's `cloud_provider` attribute.

After every Read, state populates `vars` from the raw API response **and** the matching typed block (`aws` or `azure`) — but only when ALL keys the typed block models are present in the response. If Kibana adds an unknown key, the typed block is left null in state and the user can see the new key in `vars` and migrate their config to use `vars` directly if they wish to set it.

Typed blocks compile to the same wire `vars` shape during Create/Update; they are pure sugar.

**Why:** Two observations:
- The user's first design instinct (D) — per-provider blocks — gives the best HCL ergonomics for the documented happy paths.
- The user's second instinct — always populate both — protects users from a fragile coupling between the provider's typed schema and Kibana's evolving var-key set.

Considered: populating only the typed block when matched (and only `vars` otherwise). Rejected because it forces a state-shape change when Kibana adds keys, which is a confusing diff.

The GCP typed block is deliberately omitted (see Non-Goals). GCP users use `vars` directly.

**Plan-time dual representation:** Plugin Framework disallows `Computed` on parent blocks that contain write-only children, so `aws` and `vars` are `Optional` only (unlike `azure`, which has no write-only fields and remains `Optional + Computed + UseStateForUnknown`). After Read populates both the configured representation and its read-populated sibling, a subsequent plan would otherwise mark the unconfigured sibling for removal. `ModifyPlan` preserves the sibling by copying the read-populated attribute from state into plan when the practitioner configures only one representation—the same branching logic `compileVarsForWrite` uses on config. Typed siblings are copied into vars-mode plans only when planned `vars` matches state (so explicit representation or value changes are not overwritten).

### Decision 5: Write-only secret drift detection via bcrypt hash in private state

For each write-only secret attribute (`vars[*].secret_value`, `aws.external_id`, future `azure.client_secret` if added), the provider stores a salted bcrypt hash of the most recently applied value in the resource's private state, keyed by a stable per-attribute identifier (e.g. `secret_hash:aws.external_id`, `secret_hash:vars.external_id`).

During `ModifyPlan`, the provider:
1. Reads the current write-only attribute from config (write-only attrs are available in config but not state).
2. Hashes the current value with the stored salt.
3. Compares to the hash already in private state.
4. If mismatched, marks the resource as needing update and emits a plan-time warning diagnostic: `"Detected a change to write-only attribute <name>; the resource will be updated."` This restores plan transparency without leaking the secret.

A reusable utility lives at `internal/utils/writeonlyhash/`:
```
package writeonlyhash

// Hasher manages bcrypt-based private-state hashes for write-only attributes.
type Hasher struct {
    Salt []byte   // per-resource-type, baked into the helper at construction
    Cost int      // bcrypt cost; default 10
}

func New(resourceTypeName string) *Hasher
func (h *Hasher) Compute(value string) ([]byte, error)
func (h *Hasher) Matches(value string, storedHash []byte) bool
func (h *Hasher) PrivateStateKey(attributePath string) string
```

The helper is exported and named for reuse; this change uses it from the cloud connector resource only.

**Why over `_wo_version` companion attributes:**
- Detects scenario B (user silently edits secret_value in config without bumping a version) — invisible with the version-companion pattern.
- Removes a foot-gun: users no longer have to remember to bump a version when changing a secret.
- One-time implementation cost is amortised across future secret-bearing resources via the shared helper.

**Why bcrypt over SHA-256:**
- State files can leak. A fast hash (SHA-256) lets an attacker brute-force low-entropy secrets offline. bcrypt is the conservative choice and matches `random_password.bcrypt_hash`'s existing pattern.
- A per-resource-type salt prevents rainbow-table attacks across state files.
- ~100ms during plan is negligible.

**Plan transparency trade-off:** The version-bump approach makes the cause of an update visible as `~ external_id_wo_version 1 -> 2`. The hash approach hides it. The warning diagnostic during `ModifyPlan` is the mitigation. Considered: not emitting a warning, relying solely on the implicit update. Rejected as too opaque.

**Imported-resource case:** On first refresh after `terraform import`, no hash exists in private state. The provider treats absence-of-hash as "no comparison possible" and produces no drift. On the first apply where a write-only attribute is set in config, the hash is computed and written. This matches `random_password`'s post-import behaviour.

### Decision 6: `cloud_provider` change forces replacement; `account_type` is updatable

The Kibana PUT endpoint does not accept `cloudProvider` in its body, so changing it must trigger replacement. `accountType` IS accepted by PUT and is therefore updatable in-place.

#### Scenario impact

This decision drives `RequiresReplace` plan modifiers on `cloud_provider` and `space_id`, a computed-only `cloud_connector_id`, and a plain Optional+Computed shape on `account_type`.

### Decision 7: `force_delete` as explicit attribute

Expose the API's `?force=true` query parameter as a top-level boolean attribute defaulting to `false`. When `force_delete = false` and the API returns a conflict because `package_policy_count > 0`, the provider surfaces a helpful error mentioning the count and suggesting `force_delete = true` if intentional.

**Why:** Considered: always force-delete (hides accidental destructive operations) — rejected. Considered: relying on Terraform's `lifecycle { prevent_destroy }` meta — rejected because providers cannot observe `lifecycle` blocks and the in-use case still needs server-side override. An explicit attribute is plan-visible and reversible.

### Decision 8: Read returns 404 → remove from state; Delete 404 → success

Standard provider convention. Read on a missing connector removes the resource from state without error; Delete tolerates 404 as already-deleted.

### Decision 9: Version-gate via `entitycore.VersionRequirement`

Minimum supported Kibana version is the first version that ships the `/api/fleet/cloud_connectors` endpoints. The cloud-connector naming feature is preview in 9.2 and GA in 9.3. The provider gates the resource via `entitycore.VersionRequirement` returning a helpful error against older stacks. The exact minimum version is verified against a real Kibana during implementation and pinned in the model's `GetVersionRequirements`.

### Decision 10: Data source returns the full list with server-side `kuery`

`data "elasticstack_fleet_cloud_connectors"` is backed by `GET /api/fleet/cloud_connectors`. It exposes:
- `space_id` (optional, default `"default"`)
- `kuery` (optional, server-side filter passed verbatim)
- `page` (optional)
- `per_page` (optional, default the API default)
- `cloud_connectors` (computed list of objects: id, name, cloud_provider, account_type, namespace, package_policy_count, verification_status, verification_started_at, verification_failed_at, created_at, updated_at)

`vars` is **omitted** from the data source output. The list endpoint returns vars verbatim including secret references, but exposing them here would invert the secret-protection guarantees of the resource (vars can include secrets that are visible only via their refs; a user wanting full vars should `terraform import` the specific resource). This decision is reversible if real users ask for it; defaulting to safer-first.

**Why server-side `kuery`:** the user explicitly requested it. Cheaper than client-side filtering for large connector inventories and surfaces Kibana's existing query language without reimplementation.

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| **Cloud connector var keys evolve in Kibana** (e.g., new `region` field on AWS connectors) | Both `vars` and the typed block are populated in state. New keys appear in `vars` immediately; typed block gracefully drops out when unknown keys are present rather than corrupting state. |
| **Plan output is opaque when only a write-only secret changed** | `ModifyPlan` emits a warning diagnostic naming the changed attribute (without revealing the value), restoring "what changed" visibility. |
| **State-file leak exposes hashes** | bcrypt + per-resource-type salt makes offline brute-force impractical for any reasonably-entropic secret. Documented in the resource help text so users know the hash is intentional. |
| **First refresh after `terraform import` produces no drift signal for secret-changed-since-import** | Documented in the resource help text; users should `terraform apply` after import to baseline the hash. Same behaviour as `random_password`. |
| **Cloud connectors are PREVIEW in 9.2** | Mark resource as preview/experimental in docs; version-gate via `entitycore.VersionRequirement` to fail loud against older stacks. |
| **Typed AWS block and integrations PR #16985 disagree on `aws.role_arn` vs `role_arn`** | The acceptance test exercises a real Kibana create-then-read cycle for each typed block, so the wire format the provider sends is validated end-to-end. If Kibana evolves the key, the test catches it and the typed-block compiler updates accordingly. |
| **Typed block + ConfigValidator interaction is subtle** | The validator runs against config only; state can have both populated after Read. This is the standard PF pattern, but the acceptance test explicitly covers (a) typed-only input → both populated in state; (b) vars-only input → typed populated when matchable; (c) mixed input → validation error. |
| **`writeonlyhash` helper is used in only one place initially** | Helper is intentionally minimal and well-tested in isolation. Follow-up adoption is mechanical. Avoid the temptation to overfit to cloud connectors. |

## Open Questions

1. **Exact minimum Kibana version**: ANSWERED — pinned at `9.2.0` (preview release of cloud connectors). Naming is GA in 9.3; the preview API surface is present in 9.2 and is what `GetVersionRequirements` gates against. Acceptance tests run against the current default stack (9.4+), which exceeds the minimum.
2. **`bcrypt` dependency**: ANSWERED — `golang.org/x/crypto v0.52.0` is already a direct module dependency and includes `golang.org/x/crypto/bcrypt`. No `go get` required.
3. **Whether to surface `verification_*` fields with documentation about their async nature**: They will be `Computed`, but the docs should call out that values may not stabilise immediately on first Create. Confirmed yes in the spec.
4. **Whether the data source should also be space-aware**: Current decision is yes (mirrors the resource). Confirm during implementation that the list endpoint accepts the space-aware path.
5. **Plugin Framework `WriteOnly` support**: ANSWERED — provider is on `terraform-plugin-framework v1.19.0` (≥ 1.11), which supports `WriteOnly` on string attributes. No bump required.
6. **`entitycore.KibanaResource` private-state hook**: ANSWERED — hash persistence runs in the `OnWritten` callback on `KibanaResourceOptions[T]`, invoked by the envelope after a successful Create or Update (after read-after-write refresh sets state and any optional `PostRead` runs). `OnWritten` receives the final model, the original config (so write-only values are available), and `privateState any` (the framework response `Private` field). `ModifyPlan` is implemented directly on the concrete `Resource` type via `resource.ResourceWithModifyPlan`, which has full access to `req.Private`/`resp.Private`. The envelope extension for `OnWritten` was added for this change; no hash writes occur in `PostRead`.
