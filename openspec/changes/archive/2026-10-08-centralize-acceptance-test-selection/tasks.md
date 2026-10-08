## 1. Machine-readable targeted selection

- [x] 1.1 Make targeted acceptance selection produce and validate the JSON shard plan, including rationale, empty-plan, threshold, and shard-coverage behavior; verify selector unit tests cover normal, dry-run, invalid-count, zero, small, and large selections.
- [x] 1.2 Update local targeted acceptance execution to consume the selected JSON plan shard or an explicitly supplied package list; verify existing targeted test invocation behavior remains available without plaintext selector output.

## 2. Dynamic pull-request acceptance execution

- [x] 2.1 Extend acceptance-matrix preparation to compute one pull-request plan, publish version, shard, package, and package-presence outputs, and fail closed on invalid planning; verify preparation-script tests cover base resolution, fallback, and failures.
- [x] 2.2 Build the pull-request acceptance matrix from populated prepared shards and run assigned packages without reselection; verify workflow-structure tests cover zero-, small-, and large-selection matrices.
- [x] 2.3 Preserve successful no-op acceptance jobs for zero-package pull requests while skipping costly setup, stack lifecycle, diagnostics, and teardown; verify the dedicated unit-test job remains independent of the prepared plan and provider-gate inputs remain successful.
- [x] 2.4 Preserve pre-pull retry behavior and diagnostics and teardown for populated pull-request jobs; verify no-op gating does not alter the populated-job lifecycle.

## 3. Stable authoritative full-suite execution

- [x] 3.1 Preserve fixed two-shard full-suite execution and failure handling for push, merge-group, and workflow-dispatch events; verify workflow tests assert the non-pull-request path does not use targeted planning.
- [x] 3.2 Validate the completed change with focused Go and workflow tests, OpenSpec validation, and the repository build.
