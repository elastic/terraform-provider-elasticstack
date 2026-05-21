## 1. Parameterise `StringIsJSONObject` validator

- [x] 1.1 Add `NonEmpty bool` field to `StringIsJSONObject` struct in `internal/elasticsearch/index/validation.go`
- [x] 1.2 Update `ValidateString` to check `len(m) == 0` when `NonEmpty` is true, with a descriptive error diagnostic
- [x] 1.3 Update `Description` and `MarkdownDescription` to mention non-empty when applicable

## 2. Wire validator into `indexmappings` schema

- [x] 2.1 Add `index.StringIsJSONObject{NonEmpty: true}` to `Validators` slice for the `mappings` attribute in `internal/elasticsearch/index/indexmappings/schema.go`

## 3. Add unit tests for validator

- [x] 3.1 Create table-driven tests in `internal/elasticsearch/index/validation_test.go` covering:
  - `{}` with zero-value `StringIsJSONObject{}` → passes
  - `{}` with `NonEmpty: true` → fails with error
  - non-object values (`[]`, `"hello"`, `123`) → fails with error
  - valid object with keys and zero-value → passes
  - valid object with keys and `NonEmpty: true` → passes

## 4. Add acceptance test for empty mappings rejection

- [x] 4.1 Create test data directory `testdata/TestAccResourceIndexMappings_emptyMappings` with a main.tf file defining `mappings = jsonencode({})`
- [x] 4.2 Add `TestAccResourceIndexMappings_emptyMappings` to `acc_test.go` with a single step using `ExpectError` matching the validation error

## 5. Validate and review

- [x] 5.1 Run `make build` to confirm the project compiles
- [x] 5.2 Run `go test ./internal/elasticsearch/index/...` to confirm unit tests pass
- [x] 5.3 Review all modified call sites of `StringIsJSONObject{}` to confirm zero-value behavior is preserved
