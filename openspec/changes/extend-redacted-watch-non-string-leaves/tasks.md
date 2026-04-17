## 1. Broaden the merge helper and unit-test coverage

- [x] 1.1 In `internal/elasticsearch/watcher/watch/actions_merge.go`, change the redacted-string branch in `mergePreserveRedactedLeaves` so the prior value at the redacted path is substituted whenever it is non-nil and is not itself the sentinel, regardless of its JSON type.
- [x] 1.2 Extract the precondition into a helper `isRedactedOrAbsent(priorVal any) bool` that returns `true` only when `priorVal` is `nil` or equals `elasticsearchWatcherRedactedSecret`.
- [x] 1.3 In `internal/elasticsearch/watcher/watch/actions_merge_test.go`, replace `TestMergeActionsPreservingRedactedLeaves_priorTypeMismatch` with `TestMergeActionsPreservingRedactedLeaves_priorObjectReplacesRedacted`, asserting that a `map[string]any{"not": "a string"}` prior is substituted in.
- [x] 1.4 Add `TestMergeActionsPreservingRedactedLeaves_scriptReferenceHeader` covering `headers.Authorization = {"id": "service-now-key"}` at a redacted path; assert the prior object is preserved and sibling fields stay from the API.
- [x] 1.5 Add `TestMergeActionsPreservingRedactedLeaves_inlineScriptHeader` covering an inline-script object prior (`{"source": "...", "lang": "painless"}`).
- [x] 1.6 Add `TestMergeActionsPreservingRedactedLeaves_arrayPriorReplacesRedacted` covering a `[]any` prior at a redacted leaf.
- [x] 1.7 Add `TestMergeActionsPreservingRedactedLeaves_roundTripJSON_userPayload` that round-trips the user's example watcher payload (`actions.to_echo.webhook.headers.Authorization = {"id": "$SCRIPT_ID"}`) through JSON to assert the merge produces the prior object at the redacted path with all sibling fields intact.
- [x] 1.8 Run `go test -run 'TestMerge' ./internal/elasticsearch/watcher/watch/...` and confirm all merge tests pass.

## 2. Add Plugin Framework acceptance coverage

- [x] 2.1 Add testdata at `internal/elasticsearch/watcher/watch/testdata/TestAccResourceWatch_redactedScriptHeaderPreserved/create/main.tf` defining a webhook action whose `headers.Authorization` is an inline-script object (`{ source = "return 'Bearer x'", lang = "painless" }`).
- [x] 2.2 Add testdata at `internal/elasticsearch/watcher/watch/testdata/TestAccResourceWatch_redactedScriptHeaderPreserved/update_throttle/main.tf` that keeps the same actions and only changes `throttle_period_in_millis`.
- [x] 2.3 Add `TestAccResourceWatch_redactedScriptHeaderPreserved` in `internal/elasticsearch/watcher/watch/acc_test.go` modeled on `TestAccResourceWatch_redactedWebhookAuthPreserved`: step 1 creates and asserts the script attributes survive in `actions`; step 2 expects a non-empty plan for the throttle update only and reasserts the script attributes are preserved (no `::es_redacted::` in state).
- [x] 2.4 Run the targeted acceptance test with `TF_ACC=1 go test -v -run 'TestAccResourceWatch_redactedScriptHeaderPreserved' ./internal/elasticsearch/watcher/watch/` against the running stack and confirm both steps pass.

## 3. Sync the `elasticsearch-watch` delta into canonical specs

- [x] 3.1 Update `openspec/specs/elasticsearch-watch/spec.md` so the **Read (REQ-014–REQ-016)** and **JSON field mapping — read/state (REQ-023–REQ-027)** requirements match the broadened wording from `openspec/changes/extend-redacted-watch-non-string-leaves/specs/elasticsearch-watch/spec.md`, including the new non-string scenarios.
- [x] 3.2 Run `OPENSPEC_TELEMETRY=0 npx openspec validate --all` (or `make check-openspec`) and confirm no validation errors.
