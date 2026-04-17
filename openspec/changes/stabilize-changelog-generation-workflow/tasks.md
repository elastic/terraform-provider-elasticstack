## 1. Trigger and contract alignment

- [x] 1.1 Update the changelog workflow source and generated artifacts so release-preparation runs use `pull_request_target` with the existing `prep-release-*` head-branch guard.
- [x] 1.2 Align workflow fixture tests and any helper/unit tests with the `pull_request_target` trigger contract and supported `github-script` env wiring.
- [x] 1.3 Verify the OpenSpec delta for `ci-changelog-generation` matches the authored workflow behavior now present on this branch.

## 2. Artifact-backed evidence handoff

- [x] 2.1 Change `pre_activation` to materialize the gathered manifest as an `evidence.json` file and upload it as a workflow artifact while retaining scalar release-context outputs.
- [x] 2.2 Change the `agent` job to download the evidence artifact directly to `/tmp/gh-aw/agent/evidence.json` before agent reasoning starts.
- [x] 2.3 Remove the full-manifest cross-job `evidence_json` output contract and retire or refactor the redundant manifest bridge step/helper accordingly.

## 3. Validation and regression coverage

- [x] 3.1 Update changelog workflow tests to assert artifact upload/download behavior and the absence of cross-job manifest transport through `evidence_json`.
- [x] 3.2 Regenerate `.github/workflows/changelog-generation.md` and `.github/workflows/changelog-generation.lock.yml` from the updated workflow source.
- [x] 3.3 Run the relevant workflow/unit test suite and `make check-openspec` (or equivalent OpenSpec validation) and resolve any resulting issues.
