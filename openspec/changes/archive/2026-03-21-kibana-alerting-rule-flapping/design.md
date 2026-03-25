## Context

Rule-level `flapping` is a JSON object on Kibana alerting rule create/update bodies: `look_back_window`, `status_change_threshold`, and optionally `enabled` (see `generated/kbapi/kibana.gen.go`). The internal `models.AlertingRule` type and `kibanaoapi` request builders currently ignore this field; read-path `ConvertResponseToModel` does not unmarshal `flapping`.

## Goals

- Expose flapping settings on `elasticstack_kibana_alerting_rule` with **integer** `look_back_window` and `status_change_threshold`.
- Enforce **both integers required** whenever the practitioner configures a `flapping` block; **`enabled` is optional** and never sufficient without the two integers.
- **Version gates**:
  - If a non-absent `flapping` block is known at create/update **without** relying on `enabled` (or with `enabled` unset), require stack **≥ 8.16.0** (aligned with [kibana#190019](https://github.com/elastic/kibana/pull/190019)).
  - If **`flapping.enabled`** is set to a known value, require stack **≥ 9.3.0**; below that version the provider returns an explicit diagnostic (do not send the field and do not pretend success).
- **Update semantics**: if the practitioner does **not** configure `flapping`, the provider **omits** `flapping` from the update payload so Kibana leaves existing server state for that property unchanged.

## Non-Goals

- Changing global Kibana flapping defaults or space-level settings via Terraform.
- Defining new API behavior in Kibana (this change tracks existing HTTP API semantics).

## Decisions

| Topic | Decision |
|--------|-----------|
| Terraform shape | Single nested block `flapping` at rule level (consistent with nested objects like `actions.frequency`). |
| Numeric types | **Int64** in schema; serialize to API as numbers (Kibana/OpenAPI may use floats; whole numbers are sent). |
| Block validation | If `flapping` is present, **`look_back_window` and `status_change_threshold` are required** (Framework `AlsoRequires` or equivalent). |
| `enabled` | Optional; may be set together with the two required integers. **Only supported from stack 9.3.0+**; configuring it on lower versions is a provider error. |
| Update when block absent | **Omit** `flapping` key from PUT body entirely. |
| Read | Unmarshal `flapping` from GET responses into state when present. |

## Risks / Trade-offs

- **Removing `flapping` from Terraform config** still results in an update that **omits** `flapping`; Kibana **keeps** any previously stored rule-specific flapping. After refresh, state may again show `flapping` from the API while configuration omits the block, producing a **plan diff** until the practitioner aligns config with the API or clears rule settings outside Terraform. This is inherent to “omit field = unchanged server” semantics; document in resource docs when implementing.

## Migration / State

- If schema version bump is needed for the new block, follow existing resource `Schema.Version` and state upgrade conventions; otherwise add attributes/blocks in a backward-compatible way (optional block).

## Open Questions

- None for this proposal (resolved: required integers, omit-on-update, integer schema types, **`enabled` minimum stack 9.3.0** with provider-side enforcement and acceptance tests split by version).
