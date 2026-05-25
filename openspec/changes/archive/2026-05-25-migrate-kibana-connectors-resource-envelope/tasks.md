## 1. Resource audit

- [x] 1.1 Confirm full envelope migration is feasible for `action_connector` with real write callbacks
- [x] 1.2 Record Terraform-visible invariants for import, composite IDs, read-after-write refresh, and state upgrade behavior
- [x] 1.3 Confirm how version-gated preconfigured connector ID validation should live after migration (model requirement vs callback-local check)

**Audit notes**

- **1.1** Feasible: create/update/delete map cleanly to `KibanaWriteFunc` / `kibanaDeleteFunc`; read uses envelope-supplied `resourceID`/`spaceID`.
- **1.2 Invariants preserved**
  - Import: verbatim `id` copy (no parsing).
  - Composite ID: state `id` is `<spaceID>/<connectorID>`; envelope resolves identity via composite parse.
  - Read-after-write: envelope refreshes via `readConnector` after create/update.
  - State upgrade: wrapper-level `UpgradeState` v0→v1 unchanged.
- **1.3 Version gating**: removed `WithVersionRequirements` from `tfModel` (envelope would enforce on every Read/Update once state carries API-assigned `connector_id`). The 8.8.0 gate now runs only in `createConnector` via `enforceUserSuppliedConnectorIDVersion` when the create plan includes a user-supplied `connector_id`.

## 2. Envelope migration

- [x] 2.1 Convert `internal/kibana/connectors/` resource wiring to embed `entitycore.NewKibanaResource`
- [x] 2.2 Adapt the model to satisfy the Kibana envelope contract without changing the schema
- [x] 2.3 Move create/update/read/delete orchestration into envelope callbacks while preserving current connector behavior
- [x] 2.4 Preserve wrapper-level `ImportState` and `UpgradeState` behavior unchanged

## 3. Validation and regression coverage

- [x] 3.1 Update or add unit tests for any new model/helper or version-gating behavior
- [x] 3.2 Keep or extend acceptance coverage for CRUD, import, and upgrade-state-relevant behavior
- [x] 3.3 Run targeted tests for `internal/kibana/connectors/...`
- [x] 3.4 Run `make build`
- [x] 3.5 Run `make check-openspec`
