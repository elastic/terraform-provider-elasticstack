## Context

The Kibana connector API (`PUT /api/actions/connector/{id}`) accepts a `secrets` field that is typed as an optional pointer (`omitempty`) in the generated `kbapi` client. The GET API **never** returns connector secrets — `ConnectorResponseToModel` does not set `SecretsJSON` — so write-only attributes fit naturally: the framework already expects a null read-back for write-only fields.

The `terraform-plugin-framework` version in `go.mod` is v1.19.0, which includes write-only support (available since v1.13.0). The `terraform-plugin-framework-validators` version is v0.19.0, which includes `stringvalidator.PreferWriteOnlyAttribute`.

## Decisions

| Topic | Decision |
|-------|----------|
| Approach | Additive: new `secrets_wo` + `secrets_wo_version` alongside unchanged `secrets`. Mirrors `password_wo`/`password_wo_version` on `elasticsearch_security_user`. |
| `secrets_wo` type | `schema.StringAttribute{Optional: true, Sensitive: true, WriteOnly: true}` |
| `secrets_wo_version` type | `schema.StringAttribute{Optional: true}` (string, consistent with `password_wo_version`) |
| Validators on `secrets_wo` | `stringvalidator.ConflictsWith(path.MatchRoot("secrets"))` |
| Validators on `secrets_wo_version` | `stringvalidator.AlsoRequires(path.MatchRoot("secrets_wo"))` |
| Validators on `secrets` | Add `stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("secrets_wo"))` and `stringvalidator.ConflictsWith(path.MatchRoot("secrets_wo"))` |
| Model fields | Add `SecretsWo jsontypes.Normalized` (mirror `Secrets`) and `SecretsWoVersion types.String` to `tfModel` |
| `toAPIModel()` | If `SecretsWo` is known, use it as `SecretsJSON`; otherwise fall back to `Secrets` (unchanged logic) |
| Create handler | Read `request.Config` (not just `request.Plan`) to retrieve `secrets_wo` — write-only values are available in config but not plan |
| Update handler | Always re-send `secrets_wo` from `request.Config` if set, regardless of version change. This is necessary because the Kibana API behavior when `secrets` is omitted is unknown (see Open questions). Version comparison can be used as an optimization once the omission behavior is confirmed. |
| State migration | None required — additive change only |
| Deprecation of `secrets` | Out of scope for this change; tracked as a follow-up |

## Implementation notes

### Reading write-only values

In Plugin Framework, write-only values are available in `request.Config` but are zeroed in `request.Plan`. Both `Create` and `Update` must therefore read:

```go
var cfg tfModel
response.Diagnostics.Append(request.Config.Get(ctx, &cfg)...)
// cfg.SecretsWo is the live ephemeral value
```

The existing `plan.Secrets` fall-through remains for users who still use the `secrets` attribute.

### `toAPIModel()` adjustment

```go
// Prefer write-only secrets over persisted secrets
if typeutils.IsKnown(model.SecretsWo) {
    apiModel.SecretsJSON = model.SecretsWo.ValueString()
} else if typeutils.IsKnown(model.Secrets) {
    apiModel.SecretsJSON = model.Secrets.ValueString()
}
```

Because `SecretsWo` is always null in state, this function must be called with the **config model** (not state/plan) when write-only secrets are needed.

### `populateFromAPI()` — no change needed

`secrets_wo` stays null in state by framework contract. `secrets_wo_version` persists normally (it is not write-only).

### `is_missing_secrets` drift signal

The existing `is_missing_secrets` attribute continues to serve as a passive drift signal for out-of-band secret deletion — no change needed.

## Open Questions

1. **Kibana update behavior with omitted secrets**: When the Kibana `PUT /api/actions/connector/{id}` is called *without* a `secrets` key in the body, does it preserve existing secrets or clear them? The generated type marks the field `omitempty` (pointer), so it can be omitted. If Kibana preserves secrets on omission, the provider could skip re-sending `secrets_wo` when `secrets_wo_version` is unchanged—making the ephemeral source optional for non-rotation updates. If Kibana clears secrets on omission, the provider must always re-send and the ephemeral source must be available on every apply. Until this is confirmed during acceptance testing, the implementation MUST always re-send `secrets_wo` from config on update.
2. **Version attribute type**: `password_wo_version` is a string. The issue example uses `secrets_wo_version = 1` (numeric literal). String is recommended for consistency; users can quote the value.
3. **`secrets_wo_version` required vs. optional**: Following `password_wo_version`, `secrets_wo_version` SHOULD be optional (not required) when `secrets_wo` is set. The `AlsoRequires` validator ensures `secrets_wo_version` cannot be set without `secrets_wo`, but the reverse is not enforced. Practitioners who choose not to track rotation simply leave `secrets_wo_version` unset.

## Out of Scope

- Deprecating or removing the `secrets` attribute (follow-up deprecation cycle).
- Changes to the connector data source (`data_source.go`) — it is read-only and does not expose secrets.
- Any changes to the Kibana API client beyond accepting an optional secrets payload on updates.
- Approach B (in-place migration of `secrets` to write-only) — deferred as a follow-up after write-only adoption is established.
