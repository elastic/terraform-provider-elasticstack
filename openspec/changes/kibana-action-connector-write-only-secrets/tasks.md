## 1. Schema changes

- [x] 1.1 Add `secrets_wo` to `internal/kibana/connectors/schema.go` using the same JSON-aware type/validation as the existing `secrets` attribute (for example, `schema.StringAttribute` with `CustomType: jsontypes.NormalizedType{}`), while keeping it `Optional: true`, `Sensitive: true`, and `WriteOnly: true`
- [x] 1.2 Add `secrets_wo_version` (`schema.StringAttribute{Optional: true}`) with `AlsoRequires(secrets_wo)` validator
- [x] 1.3 Add `ConflictsWith(secrets_wo)` and `PreferWriteOnlyAttribute(secrets_wo)` validators to the existing `secrets` attribute
- [x] 1.4 Add `ConflictsWith(secrets)` validator to `secrets_wo` and ensure `secrets_wo` preserves the same JSON normalization/validation behavior as `secrets`

## 2. Model changes

- [x] 2.1 Add `SecretsWo types.String` and `SecretsWoVersion types.String` fields to `tfModel` in `internal/kibana/connectors/models.go`
- [x] 2.2 Update `toAPIModel()` to prefer `SecretsWo` over `Secrets` when `SecretsWo` is known

## 3. CRUD handler changes

- [x] 3.1 Update `Create` in `internal/kibana/connectors/create.go` to read `request.Config` into a separate `cfgModel` and pass `SecretsWo` to `toAPIModel()` (write-only values are zero in plan, non-zero in config)
- [x] 3.2 Update `Update` in `internal/kibana/connectors/update.go` with the same config-reading logic; always re-send `secrets_wo` from config if set (pending resolution of open question on Kibana omit-secrets behavior)

## 4. Tests

- [x] 4.1 Add a unit test for the updated `toAPIModel()` that covers both `secrets_wo` and `secrets` paths
- [x] 4.2 Add or extend an acceptance test in `internal/kibana/connectors/` that exercises `secrets_wo` + `secrets_wo_version` with a connector that accepts a `secrets` payload (e.g. `.pagerduty` or `.webhook`)
- [x] 4.3 Verify that the acceptance test confirms `secrets_wo` is not present in state after apply (read state and check attribute is null)
- [x] 4.4 Add an acceptance test step that bumps `secrets_wo_version` to confirm the update path re-sends the secret

## 5. Documentation and validation

- [x] 5.1 Update the resource documentation / description strings (add note that `secrets_wo` is preferred for ephemeral secret sources and never persisted to state)
- [x] 5.2 Run `make build`
- [x] 5.3 Run targeted acceptance tests for `internal/kibana/connectors/...`
- [x] 5.4 Run `make check-openspec`
