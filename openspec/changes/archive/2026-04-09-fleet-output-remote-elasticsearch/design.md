## Context

The provider already implements Fleet output CRUD for `elasticsearch` and `kafka`, and the data source already recognizes `OutputRemoteElasticsearch` in list responses. However, the resource schema and request/response mapping do not expose `remote_elasticsearch`, which prevents Terraform-managed creation and lifecycle of this Fleet output type.

Fleet remote Elasticsearch outputs require a service token to let Fleet Server generate API keys against the target cluster. The provider must model this as a sensitive input and handle API responses that may redact or omit stored secret values after creation.

## Goals / Non-Goals

**Goals:**
- Add first-class support for `type = "remote_elasticsearch"` in `elasticstack_fleet_output`.
- Define resource schema and validation for remote output fields, including required authentication material and optional TLS/mTLS options.
- Ensure CRUD and state mapping are deterministic for remote outputs, including behavior when secret fields are write-only or redacted in API responses.
- Extend tests and documentation to cover the new output type and expected constraints.

**Non-Goals:**
- Changing Fleet API semantics or introducing provider behavior not supported by Fleet.
- Refactoring unrelated output types (`elasticsearch`, `kafka`, `logstash`) beyond changes required to share common flow.
- Implementing automatic cross-cluster integration synchronization workflows beyond exposing fields Fleet already supports.

## Decisions

- Add `remote_elasticsearch` to the allowed `type` set for the resource.
  - Rationale: keeps a single Fleet output resource with type-dispatched behavior, consistent with existing provider UX.
  - Alternative considered: introducing a separate resource for remote outputs; rejected because it duplicates lifecycle logic and diverges from Fleet’s unified outputs model.

- Model remote auth token and related key material as sensitive Terraform attributes.
  - Rationale: avoids plaintext exposure in state and plans where possible, and aligns with Fleet secret semantics.
  - Alternative considered: forcing all auth material into `config_yaml`; rejected because it weakens schema validation and user ergonomics.

- Preserve configured secret values in state when Fleet read responses omit or redact them.
  - Rationale: avoids perpetual diffs and failed drift reconciliation for write-only secret fields.
  - Alternative considered: setting missing secret fields to null on read; rejected because this causes churn and unintended updates.

- Keep remote output support as an incremental extension of current type-dispatch and model conversion logic.
  - Rationale: minimizes risk and keeps behavior for existing output types unchanged.
  - Alternative considered: replacing all output mappings with a new generalized translation layer; rejected due to larger migration risk for this scoped change.

## Risks / Trade-offs

- [Secret field read-back differences across Fleet versions] -> Preserve prior state when API omits redacted fields and add coverage for create/read/update cycles.
- [Validation mismatch with Fleet server-side constraints] -> Mirror known Fleet constraints in schema validation and rely on API diagnostics for remaining server-side validation.
- [Acceptance environment complexity for remote connectivity] -> Prefer focused acceptance tests that validate schema/API wiring; gate network-dependent scenarios behind explicit env setup.

## Migration Plan

- No state migration is expected when adding support for a new `type` value and additional optional attributes.
- Existing resources remain unchanged; behavior differences apply only when users configure `type = "remote_elasticsearch"`.
- Rollback strategy: remove the new type/fields and re-apply previous provider version. Existing remote outputs created by the interim version may need import or manual management if rolled back.

## Open Questions

- Which remote output fields are consistently returned versus redacted by all supported Fleet versions, and do any require explicit `UseStateForUnknown` plan modifiers?
- Should the provider expose automatic integration synchronization and wired-stream related options immediately, or defer to a follow-up once Fleet API compatibility is confirmed for targeted versions?
