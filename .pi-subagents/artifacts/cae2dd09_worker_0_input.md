# Task for worker

Implement top-level task 4 of OpenSpec change `selective-acceptance-tests`: update `.github/workflows/provider.yml`.

**Scope:**
- Add `merge_group:` trigger to the `on:` section (the OpenSpec context was updated to support the merge queue as a non-PR full-suite event).
- Add a `compute-packages` step in the `test` job after `make vendor` and before `Pre-pull fleet image`. The step id should be `targeted`. It must set two outputs:
  - `has_packages` (`true` or `false`)
  - `targeted_pkgs` (space-separated list, or empty string)
  Implementation:
  - For non-PR events (`github.event_name != 'pull_request'`): set `has_packages=true` and `targeted_pkgs=` unconditionally.
  - For PR events: run `git fetch origin main --depth=1`, then run `go run ./scripts/targeted-testacc/... --total-shards=2 --shard-index=${{ matrix.shard }}` capturing stdout. If it emits packages, set `has_packages=true` and `targeted_pkgs=<packages>`; else `has_packages=false`.
- Gate the following expensive steps with `if: steps.targeted.outputs.has_packages == 'true'`:
  - `Pre-pull fleet image` — condition should be `if: matrix.fleetImage && steps.targeted.outputs.has_packages == 'true'`
  - `Start stack with docker compose`
  - `Wait for stack readiness`
  - `Get ES API key`
  - `Setup Fleet` — AND with existing version condition
  - `Force install synthetics` — AND with existing version condition
  - `TF acceptance tests` — add the condition and route as follows:
    - if `steps.targeted.outputs.targeted_pkgs != ''`: run `make targeted-testacc TARGETED_PKGS=<packages>`
    - else: run `make testacc`
    - In either case pass `ACCTEST_TOTAL_SHARDS=2 ACCTEST_SHARD_INDEX=${{ matrix.shard }}`.
- Verify/ensure `Tear down docker compose stack` step still has `if: always()` and runs `make docker-clean`.

**Requirements:**
- Follow the spec in `openspec/changes/selective-acceptance-tests/specs/ci-provider-acceptance-tests/spec.md`.
- Do not alter the matrix definition (versions, shards, include entries).
- Do not push.
- Create one focused git commit.
- Validate the YAML is syntactically sane with a lightweight action (e.g., `actionlint` if available, or `python3 -c 'import yaml; yaml.safe_load(open(".github/workflows/provider.yml"))'` using pyyaml, or at least visually verify indentation).

**Context files to read first:**
- `/Users/tobio/Projects/terraform-provider-elasticstack/selective-acceptance-tests/openspec/changes/selective-acceptance-tests/specs/ci-provider-acceptance-tests/spec.md`
- `/Users/tobio/Projects/terraform-provider-elasticstack/selective-acceptance-tests/openspec/changes/selective-acceptance-tests/tasks.md`
- `/Users/tobio/Projects/terraform-provider-elasticstack/selective-acceptance-tests/.github/workflows/provider.yml`

Report back:
- the exact diff summary for `provider.yml`
- any validation command run and its result
- commits created
- blockers

## Acceptance Contract
Acceptance level: checked
Completion is not accepted from prose alone. End with a structured acceptance report.

Criteria:
- criterion-1: Implement the requested change without widening scope

Required evidence: changed-files, tests-added, commands-run, residual-risks, no-staged-files

Finish with a fenced JSON block tagged `acceptance-report` in this shape:
Use empty arrays when no items apply; array fields contain strings unless object entries are shown.
```acceptance-report
{
  "criteriaSatisfied": [
    {
      "id": "criterion-1",
      "status": "satisfied",
      "evidence": "specific proof"
    }
  ],
  "changedFiles": [
    "src/file.ts"
  ],
  "testsAddedOrUpdated": [
    "test/file.test.ts"
  ],
  "commandsRun": [
    {
      "command": "command",
      "result": "passed",
      "summary": "short result"
    }
  ],
  "validationOutput": [
    "validation output or concise summary"
  ],
  "residualRisks": [
    "none"
  ],
  "noStagedFiles": true,
  "diffSummary": "short description of the diff",
  "reviewFindings": [
    "blocker: file.ts:12 - issue found, or no blockers"
  ],
  "manualNotes": "anything else the parent should know"
}
```