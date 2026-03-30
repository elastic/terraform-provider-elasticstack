## 1. Update workflow runtime requirements

- [ ] 1.1 Update the `ci-aw-openspec-verification` canonical spec and the verify-label workflow source so the review environment is documented around explicit `runtimes.go.version: "1.26.1"` and `runtimes.node.version: "24"` only.
- [ ] 1.2 Remove any review-bootstrap use of `actions/setup-go` from the workflow source and related generated workflow artifacts while preserving Terraform CLI setup behavior.

## 2. Extend repository runtime validation

- [ ] 2.1 Update the make-based runtime validation so the workflow Go runtime is still checked against `go.mod`.
- [ ] 2.2 Add Node runtime validation so the workflow `runtimes.node.version` is verified against the `package.json` `engines.node` range in the same check path used for Go runtime drift detection.

## 3. Regenerate and verify workflow artifacts

- [ ] 3.1 Recompile the verify-label workflow outputs after the source changes so the committed generated artifacts match the source.
- [ ] 3.2 Run the relevant repository checks for workflow/runtime alignment and OpenSpec validation, then resolve any drift or validation failures.
