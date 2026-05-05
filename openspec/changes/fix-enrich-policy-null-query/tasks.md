## 1. Client-Layer Fix

- [ ] 1.1 In `internal/clients/elasticsearch/enrich.go`, inside `GetEnrichPolicy`, after
  `queryBytes, err := json.Marshal(policy.Query)`, add a guard:
  ```go
  if string(queryBytes) != "null" {
      queryStr = string(queryBytes)
  }
  ```
  This ensures `queryStr` is empty (not `"null"`) when the marshal result is JSON null.

## 2. Test Helper Hardening

- [ ] 2.1 In `internal/elasticsearch/enrich/acc_test.go`, update `checkEnrichPolicyQueryNull`
  to reject the string value `"null"` as a valid null. Change the existing condition:
  ```go
  // before
  if !ok || value == "" || value == "null" {
  ```
  to:
  ```go
  // after
  if !ok || value == "" {
  ```
  and add a corresponding error case when `value == "null"`:
  ```go
  if value == "null" {
      return fmt.Errorf("Expected query to be null (TF null), got the string %q — the null-as-string bug is present", value)
  }
  ```

## 3. Idempotency Acceptance Test

- [ ] 3.1 In `TestAccResourceEnrichPolicyQueryOmitted`
  (`internal/elasticsearch/enrich/acc_test.go`), add a second `resource.TestStep`
  after the existing create step:
  ```go
  {
      ProtoV6ProviderFactories: acctest.Providers,
      ConfigDirectory:          acctest.NamedTestCaseDirectory("create"),
      ConfigVariables:          config.Variables{"name": config.StringVariable(name)},
      PlanOnly:                 true,
      ExpectNonEmptyPlan:       false,
  },
  ```
  This step re-plans with the same configuration and fails if Terraform schedules any
  changes (i.e., if `query = "null"` reappears in the refreshed state).

## 4. Requirements Spec Update

- [ ] 4.1 In `openspec/specs/elasticsearch-enrich-policy/spec.md`, update the body of
  REQ-013 to add the marshaled-null scenario. See the delta spec in
  `openspec/changes/fix-enrich-policy-null-query/specs/elasticsearch-enrich-policy/spec.md`.

## 5. Validation

- [ ] 5.1 Run `make build` to verify the client change compiles.
- [ ] 5.2 Run `make check-lint` to ensure lint passes.
- [ ] 5.3 If an Elasticsearch stack is available, run the targeted acceptance test:
  ```
  TF_ACC=1 go test -v -run TestAccResourceEnrichPolicyQueryOmitted ./internal/elasticsearch/enrich/... -timeout 20m
  ```
