## 1. Create and read

- [x] 1.1 Add the templates API wrapper and resource schema/model under `internal/clients/kibanaoapi/` and `internal/kibana/casetemplate/`, including the 9.5.2 version requirement, composite ID, YAML syntax validation, `Get`/not-found behavior, and provider registration. Verify invalid YAML fails before an API call (`TestAccResourceKibanaCaseTemplateInvalidYAML`) and sub-9.5.2 operations fail before a templates API call.
- [x] 1.2 Verify at resource level that an unscoped resource uses the provider-level Kibana client and a resource with `kibana_connection` uses its scoped client (`TestAccResourceKibanaCaseTemplateKibanaConnection`). Include mocked API coverage for the create/read routes in the requested space and verify read 404 removes the resource from state.
- [x] 1.3 Add acceptance tests for minimal create (`owner`, `name`, `definition_yaml` only) and create with all optional fields (`description`, `tags`, `is_enabled = false`). Assert non-empty `template_id`, `template_version == 1`, composite `id`, `is_enabled == true` when omitted, and that `description`/`tags` are null when omitted; verify successful create is followed by a read that supplies final state.
- [x] 1.4 Live-verify (or, if no cluster is available during implementation, re-confirm from kibana#280144/kibana#283476 PR and release-tag metadata) the 9.5.2 minimum version and the `xpack.cases.templates.enabled` feature-flag prerequisite; update design.md's Implementation findings and the schema description with whatever is confirmed, including whether a disabled flag is indistinguishable from a too-old server at the HTTP level.
- [x] 1.5 Add the create/read resource example and generated documentation, including the Technical Preview note and the `xpack.cases.templates.enabled` prerequisite.

## 2. Update, diff suppression, and null-vs-empty behavior

- [x] 2.1 Implement update using the required `owner`, `name`, and `definition_yaml` fields in the PUT body, and perform a read after successful update to derive final state. Verify model-to-request and read-response-to-model mappings, including `description`/`tags`/`field_count`/`usage_count`/`is_default` present versus absent.
- [x] 2.2 Implement the `definition_yaml` plan-time diff-suppression modifier (parse both state and config as YAML, compare for structural equality, override the planned value to the prior state value when equal) and add a unit test covering semantically-equal (whitespace/comment/key-order-only) versus genuinely-different YAML.
- [x] 2.3 Add an acceptance test step editing `definition_yaml` only in formatting (reordered keys, added comments, changed indentation) using `ConfigPlanChecks.PreApply` + `plancheck.ExpectEmptyPlan()`, and verify `template_version` is unchanged after apply. Add a paired step with a genuine structural edit using `plancheck.ExpectNonEmptyPlan()`, verifying `template_version` increments.
- [x] 2.4 Add acceptance tests for omitting a previously set `description`/`tags` on update (asserting the live-verified clear-on-omit behavior and resulting null state) and for explicitly setting `description = ""` / `tags = []` (asserting the resulting state preserves the empty values distinctly from null).
- [x] 2.5 Add acceptance tests for updating `tags` with a duplicate entry and a specific order, verifying the API preserves both.
- [x] 2.6 Add acceptance test steps for an in-place rename and an `is_enabled` toggle; assert state reflects each change and `template_version` increments after each update.
- [x] 2.7 Update generated documentation and the resource example for update behavior, the diff-suppression guarantee, and the clear-on-omit semantics for `description`/`tags`.

## 3. Import, destroy, and errors

- [ ] 3.1 Implement import through `resource.ImportStatePassthroughID` on `id`; add acceptance tests importing `<space_id>/<template_id>` in the default and a named space and verify refresh populates all attributes.
- [ ] 3.2 Implement delete status handling so HTTP 204 is accepted as normal success and 404 is treated as already deleted. Add mocked API tests for these statuses and acceptance tests for normal destroy (HTTP 204) and external deletion followed by refresh/plan.
- [ ] 3.3 Add acceptance tests for the negative cases: duplicate `name` within the same `owner` (409), malformed template schema such as an invalid field `control` value (400), and invalid `owner` rejected at plan-time validation.
- [ ] 3.4 Verify named-space path routing against a live 9.5.2+ cluster if available; record confirmed behavior in design.md's Implementation findings.
- [ ] 3.5 Add a 9.6+-gated acceptance test for a template whose `definition_yaml` references a field via `$ref`, verifying successful create.
- [ ] 3.6 Update generated documentation and examples for composite-ID import and destroy behavior (including the all-versions soft-delete note); no `CHANGELOG.md` edit is needed because the changelog is generated from merged PRs (`dev-docs/high-level/contributing.md`).
