## 1. Add `IDValidatorWithoutLength()` to the ML validator package

- [x] 1.1 In `internal/elasticsearch/ml/idvalidator.go`, add a new exported function
  `IDValidatorWithoutLength()` that returns `stringvalidator.All(stringvalidator.LengthAtLeast(1),
  stringvalidator.RegexMatches(pathIDRegexp, idAllowedCharsMessage))` — identical to
  `IDValidator()` but without `LengthBetween(1, 64)`.
- [x] 1.2 Update the doc comment on `IDValidator()` to note that `IDValidatorWithoutLength()`
  exists for resources where Elasticsearch imposes no upper-bound length.

## 2. Swap validators in the affected schemas

- [x] 2.1 In `internal/elasticsearch/ml/datafeed/schema.go:65`, replace `ml.IDValidator()` with
  `ml.IDValidatorWithoutLength()` on the `datafeed_id` attribute.
- [x] 2.2 In `internal/elasticsearch/ml/datafeed_state/schema.go:57`, replace `ml.IDValidator()`
  with `ml.IDValidatorWithoutLength()` on the `datafeed_id` attribute.
- [x] 2.3 In `internal/elasticsearch/ml/filter/schema.go:52`, replace `ml.IDValidator()` with
  `ml.IDValidatorWithoutLength()` on the `filter_id` attribute.

## 3. Unit tests

- [x] 3.1 In `internal/elasticsearch/ml/idvalidator_test.go`, add `TestIDValidatorWithoutLength`
  covering:
  - A 68-character valid ID (e.g. the example from the issue: `datafeed-opserv-riskviewxml-customer-transaction-volume-decline-stop`) — must pass.
  - A 65-character valid ID — must pass (previously rejected by `IDValidator`).
  - A single-character valid ID — must pass.
  - An empty string — must fail (non-empty still enforced).
  - An uppercase character — must fail (character class still enforced).
  - A leading underscore — must fail (anchor regex still enforced).

## 4. Spec sync

- [x] 4.1 Run `make check-openspec` and resolve any reported issues.
- [x] 4.2 Confirm `make build` succeeds.
- [x] 4.3 Run `go test ./internal/elasticsearch/ml/...` to confirm all unit tests pass.
