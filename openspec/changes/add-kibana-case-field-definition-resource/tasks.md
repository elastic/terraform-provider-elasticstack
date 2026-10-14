## 1. Create and read

- [x] 1.1 Add the field-definition API wrapper and resource schema/model under `internal/clients/kibanaoapi/` and `internal/kibana/casefielddefinition/`, including the 9.6.0 version requirement, composite ID, YAML validation, `Get`/not-found behavior, and provider registration. Use the schema in "Schema and validation" and verify invalid YAML fails before an API call and sub-9.5.0 operations fail before a field-definition API call.
- [x] 1.2 Verify at resource level that an unscoped resource uses the provider-level Kibana client and a resource with `kibana_connection` uses its scoped client. Include mocked API coverage for the create/read routes in the requested space and verify read 404 removes the resource from state.
- [x] 1.3 Add acceptance tests for minimal create (only `owner` and `definition`, YAML `name` present) and create with all optional fields (top-level `name` matching YAML, `description`, `is_global = true`). Assert computed `name`, non-empty `field_definition_id`, and composite `id`; verify successful create is followed by a read that supplies final state.
- [x] 1.4 Add an acceptance test creating with `description`/`is_global` omitted to probe the server response, and resolve the "Server-side omission semantics for description and is_global" open question. If the probe indicates behavior inconsistent with the canonical Optional+Computed requirement, update the spec before implementation and then implement the revised requirement; otherwise retain Optional+Computed. Verify nil response pointers are handled safely.
- [x] 1.5 Add the create/read resource example and generated documentation, including the Technical Preview note, and verify the resource is described accurately.

## 2. Update and identity behavior

- [x] 2.1 Implement update using the required `owner` and `definition` fields in the PUT body, and perform a read after successful update to derive final state. Verify model-to-request and read-response-to-model mappings, including `description`/`is_global` present versus absent.
- [x] 2.2 Add acceptance tests for omitting a previously set `description`/`is_global` (asserting the live-verified preserve-on-omit state), clearing `description` with `""`, and updating both explicitly; changing non-identity YAML fields in place with an unchanged `field_definition_id`; and a semantically identical YAML reformat (record the existing schema-description limitation if semantic equality does not suppress the diff).
- [x] 2.3 Add acceptance tests for changing top-level `name` with a matching YAML name (replacement), changing only YAML-internal `name` (API 400 diagnostic), and changing only YAML-internal `type` (API `field_identity_immutable` 409 diagnostic).
- [x] 2.4 Update generated documentation and the resource example for update behavior and document the distinction between the API's name-mismatch and type-identity errors.

## 3. Import and destroy

- [x] 3.1 Implement import through `resource.ImportStatePassthroughID` on `id`; add acceptance tests importing `<space_id>/<field_definition_id>` in the default and a named space and verify refresh populates all attributes.
- [x] 3.2 Implement delete status handling so HTTP 204 is accepted as normal success and 404 is treated as already deleted, while 409 remains an error. Add mocked API tests for these statuses and acceptance tests for normal destroy (HTTP 204), external deletion followed by refresh/plan, and destroy blocked by an active reference (409).
- [x] 3.3 Verify named-space path routing against a live 9.6+ cluster and resolve whether the 409 demotion error for a linked global field needs a clearer provider message; record confirmed behavior or an explicit follow-up decision in design.md.
- [x] 3.4 Update generated documentation and examples for composite-ID import and destroy behavior; no CHANGELOG.md edit is needed because the changelog is generated from merged PRs (`dev-docs/high-level/contributing.md`).
- [x] 3.5 Run `openspec validate add-kibana-case-field-definition-resource --strict`, `make build`, the new package's unit tests, and `make docs-generate` / `make check-docs`; resolve any failures.
