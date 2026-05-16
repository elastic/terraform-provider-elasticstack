## 1. Replace `EnforceMinServerVersion` in the `datastreamoptions` package

- [ ] 1.1 Add `var MinSupportedVersion = version.Must(version.NewVersion("9.1.0"))` to
      `internal/elasticsearch/index/datastreamoptions/version_gating.go`.
- [ ] 1.2 Add the `GetVersionRequirements(tmplObj types.Object) ([]entitycore.VersionRequirement, diag.Diagnostics)`
      function to the same file, using the same `data_stream_options` presence check as the
      existing `EnforceMinServerVersion` function.
- [ ] 1.3 Delete the `EnforceMinServerVersion` function from `datastreamoptions/version_gating.go`
      (it will be replaced entirely — callers are updated in later tasks).

## 2. Implement `WithVersionRequirements` on `componenttemplate.Data`

- [ ] 2.1 Add `GetVersionRequirements() ([]entitycore.VersionRequirement, diag.Diagnostics)` to
      `internal/elasticsearch/index/componenttemplate/models.go` on the `Data` struct:
      - If `d.Template.IsNull() || d.Template.IsUnknown()`, return `nil, nil`.
      - Otherwise delegate to `datastreamoptions.GetVersionRequirements(d.Template)`.
- [ ] 2.2 Update imports in `componenttemplate/models.go` to include
      `github.com/elastic/terraform-provider-elasticstack/internal/entitycore` and
      `github.com/elastic/terraform-provider-elasticstack/internal/elasticsearch/index/datastreamoptions`.

## 3. Remove manual version check from `componenttemplate/create.go`

- [ ] 3.1 Delete the `serverVersion, sdkDiags := client.ServerVersion(ctx)` block and the
      `datastreamoptions.EnforceMinServerVersion(plan.Template, serverVersion)` call from
      `writeComponentTemplate` in `internal/elasticsearch/index/componenttemplate/create.go`.
- [ ] 3.2 Remove the now-unused `diagutil` and `datastreamoptions` imports from `componenttemplate/create.go`
      if they become unused (verify `diagutil` is still referenced by the `PutComponentTemplate` call).

## 4. Implement `WithVersionRequirements` on `template.Model`

- [ ] 4.1 Add `GetVersionRequirements() ([]entitycore.VersionRequirement, diag.Diagnostics)` to
      `internal/elasticsearch/index/template/models.go` on the `Model` struct:
      - Delegate to `datastreamoptions.GetVersionRequirements(m.Template)` for the `data_stream_options`
        requirement.
      - Append an `entitycore.VersionRequirement` for `ignore_missing_component_templates` when
        `m.IgnoreMissingComponentTemplates` is non-null, non-unknown, and has at least one element
        (minimum version `index.MinSupportedIgnoreMissingComponentTemplateVersion`, 8.7.0).
      - Return accumulated requirements and any diagnostics.
- [ ] 4.2 Update imports in `template/models.go` to include `entitycore` and `datastreamoptions`.

## 5. Replace `serverVersion` block in `template/create.go`

- [ ] 5.1 Delete the following block from `internal/elasticsearch/index/template/create.go`:
      ```go
      serverVersion, sdkDiags := client.ServerVersion(ctx)
      resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
      if resp.Diagnostics.HasError() { return }
      resp.Diagnostics.Append(validateIgnoreMissingComponentTemplatesVersion(plan, serverVersion)...)
      resp.Diagnostics.Append(datastreamoptions.EnforceMinServerVersion(plan.Template, serverVersion)...)
      if resp.Diagnostics.HasError() { return }
      ```
- [ ] 5.2 Add the following replacement block immediately after the client is obtained and before
      `plan.toAPIModel`:
      ```go
      vReqs, reqDiags := plan.GetVersionRequirements()
      resp.Diagnostics.Append(reqDiags...)
      if resp.Diagnostics.HasError() { return }
      for _, req := range vReqs {
          ok, sdkDiags := client.EnforceMinVersion(ctx, &req.MinVersion)
          resp.Diagnostics.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
          if resp.Diagnostics.HasError() { return }
          if !ok {
              resp.Diagnostics.AddError("Unsupported server version", req.ErrorMessage)
              return
          }
      }
      ```
- [ ] 5.3 Remove the now-unused `datastreamoptions` import from `template/create.go`.

## 6. Replace `serverVersion` block in `template/update.go`

- [ ] 6.1 Apply the same replacement (task 5.1 → 5.2) to the equivalent block in
      `internal/elasticsearch/index/template/update.go` (lines 57–67 in the current file).
- [ ] 6.2 Remove the now-unused `datastreamoptions` import from `template/update.go`.

## 7. Delete `template/version_gating.go`

- [ ] 7.1 Verify that `validateIgnoreMissingComponentTemplatesVersion` has no remaining callers
      (it should have none after tasks 5 and 6).
- [ ] 7.2 Delete `internal/elasticsearch/index/template/version_gating.go`.

## 8. Add unit tests — `componenttemplate/version_requirements_test.go`

- [ ] 8.1 Create `internal/elasticsearch/index/componenttemplate/version_requirements_test.go` with
      `TestData_GetVersionRequirements` covering:
      - `Data` with a null `Template` → returns `nil, nil`.
      - `Data` with a `Template` object where `data_stream_options` is null → returns `nil, nil`.
      - `Data` with a `Template` object where `data_stream_options` is a configured (non-null) object
        → returns exactly one `VersionRequirement` (min ES 9.1.0).

## 9. Add unit tests — `template/version_requirements_test.go`

- [ ] 9.1 Create `internal/elasticsearch/index/template/version_requirements_test.go` with
      `TestModel_GetVersionRequirements` covering:
      - Neither `data_stream_options` nor `ignore_missing_component_templates` configured → returns `nil`.
      - Only `data_stream_options` configured → returns exactly one requirement (ES 9.1.0).
      - Only `ignore_missing_component_templates` configured (non-empty list) → returns exactly one
        requirement (ES 8.7.0).
      - Both attributes configured → returns exactly two requirements.
      - `ignore_missing_component_templates` is an empty list → returns no requirement for it.

## 10. Refactor `template/expand_flatten_test.go`

- [ ] 10.1 Refactor the test at lines 211–227 in
      `internal/elasticsearch/index/template/expand_flatten_test.go` (currently calling
      `datastreamoptions.EnforceMinServerVersion`) to instead build a `Model` value with the
      `data_stream_options`-carrying template object and assert on
      `model.GetVersionRequirements()` — confirming one requirement is returned when
      `data_stream_options` is set and none when it is absent.
- [ ] 10.2 Remove the `datastreamoptions` import from `expand_flatten_test.go` if it becomes unused.

## 11. Requirements update

- [ ] 11.1 Update REQ-027 in `openspec/specs/elasticsearch-index-component-template/spec.md` to reflect
      that version gating for `data_stream_options` is now implemented via `WithVersionRequirements`
      and routes through `client.EnforceMinVersion` (correctly handling Serverless clusters).
- [ ] 11.2 Update REQ-012 in `openspec/specs/elasticsearch-index-template/spec.md` to specify that
      `ignore_missing_component_templates` version gating uses `Model.GetVersionRequirements()` and
      `client.EnforceMinVersion`.
- [ ] 11.3 Update REQ-033 in `openspec/specs/elasticsearch-index-template/spec.md` to specify that
      `data_stream_options` version gating uses `Model.GetVersionRequirements()` and
      `client.EnforceMinVersion`, and applies to both Create/Update (explicit loop) and Read
      (envelope enforcement).

## 12. Validation

- [ ] 12.1 Run `make build` to confirm the provider compiles without errors.
- [ ] 12.2 Run `go test ./internal/elasticsearch/index/componenttemplate/... ./internal/elasticsearch/index/template/...`
      to confirm all unit tests pass.
- [ ] 12.3 Run `make check-lint` to confirm lint passes.
