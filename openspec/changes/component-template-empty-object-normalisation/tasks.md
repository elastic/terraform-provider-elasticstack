## 1. Harden flattenTemplateBlock

- [ ] 1.1 In `internal/elasticsearch/index/componenttemplate/flatten.go`, function `flattenTemplateBlock`:
  - Change the `t.Mappings != nil` guard to `len(t.Mappings) > 0` so a non-nil empty map produces
    `MappingsNull()` instead of `MappingsValue("{}")`.
  - Change the `t.Settings != nil` guard to `len(t.Settings) > 0` so a non-nil empty map produces
    `IndexSettingsNull()` instead of `IndexSettingsValue("{}")`.
  - No other changes to the function signature or surrounding logic.

## 2. Add test config for issue-609 scenario

- [ ] 2.1 Create
  `internal/elasticsearch/index/componenttemplate/testdata/TestAccResourceComponentTemplateIssue609NoDrift/apply/main.tf`
  with the exact reporter configuration:
  - Resource `elasticstack_elasticsearch_component_template.test`
  - `template.alias { name = "my_template_test" }`
  - `template.settings = jsonencode({ number_of_shards = "3" })` (short-form, NOT nested under `index = {}`)
  - No `template.mappings` block

## 3. Add regression acceptance tests

- [ ] 3.1 In `internal/elasticsearch/index/componenttemplate/acc_test.go`, add a second step to
  `TestAccResourceComponentTemplate` immediately after the existing create step:
  ```go
  {
      ProtoV6ProviderFactories: acctest.Providers,
      ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
      ConfigVariables:          config.Variables{"name": config.StringVariable(templateName)},
      PlanOnly:                 true,
      ExpectNonEmptyPlan:       false,
  },
  ```
- [ ] 3.2 Add a new test function `TestAccResourceComponentTemplateIssue609NoDrift`:
  ```go
  func TestAccResourceComponentTemplateIssue609NoDrift(t *testing.T) {
      templateName := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlphaNum)
      resource.Test(t, resource.TestCase{
          PreCheck:     func() { acctest.PreCheck(t) },
          CheckDestroy: checkResourceComponentTemplateDestroy,
          Steps: []resource.TestStep{
              {
                  ProtoV6ProviderFactories: acctest.Providers,
                  ConfigDirectory:          acctest.NamedTestCaseDirectory("issue-609"),
                  ConfigVariables:          config.Variables{"name": config.StringVariable(templateName)},
                  Check: resource.ComposeTestCheckFunc(
                      resource.TestCheckResourceAttr("elasticstack_elasticsearch_component_template.test", "name", templateName),
                  ),
              },
              {
                  ProtoV6ProviderFactories: acctest.Providers,
                  ConfigDirectory:          acctest.NamedTestCaseDirectory("issue-609"),
                  ConfigVariables:          config.Variables{"name": config.StringVariable(templateName)},
                  PlanOnly:                 true,
                  ExpectNonEmptyPlan:       false,
              },
          },
      })
  }
  ```

## 4. Build and validate

- [ ] 4.1 Run `make build` to confirm the change compiles.
- [ ] 4.2 Run `go vet ./internal/elasticsearch/index/componenttemplate/...` to confirm no vet errors.
- [ ] 4.3 Run acceptance tests (requires `TF_ACC=1` and a running Elasticsearch):
  `go test ./internal/elasticsearch/index/componenttemplate/... -run TestAccResourceComponentTemplate -v`
  `go test ./internal/elasticsearch/index/componenttemplate/... -run TestAccResourceComponentTemplateIssue609NoDrift -v`

## 5. Spec sync

- [ ] 5.1 Run `make check-openspec`.
