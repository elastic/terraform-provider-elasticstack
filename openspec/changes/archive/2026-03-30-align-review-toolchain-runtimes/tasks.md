## 1. Update workflow runtime requirements

- [x] 1.1 Update the `ci-aw-openspec-verification` canonical spec, delta spec, and verify-label workflow source so Go comes from `go.mod`, Node comes from `package.json`, and `runtimes.go` is not used.
- [x] 1.2 Add the `Capture GOROOT for AWF chroot mode` step immediately after Go setup while preserving Terraform CLI setup behavior.

## 2. Remove legacy runtime maintenance

- [x] 2.1 Remove the legacy verify-label runtime maintenance Makefile targets and any `check-lint` dependency on them now that the workflow reads repository version files directly.
- [x] 2.2 Remove the supporting `makefile-workflows` requirement text for that legacy runtime maintenance path.

## 3. Regenerate and verify workflow artifacts

- [x] 3.1 Recompile the verify-label workflow outputs after the source changes so the committed generated artifacts match the source.
- [x] 3.2 Run the relevant repository checks for workflow generation and OpenSpec validation, then resolve any drift or validation failures.
