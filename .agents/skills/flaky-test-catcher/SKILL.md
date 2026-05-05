---
name: flaky-test-catcher
description: Detects broken and flaky acceptance tests from recent CI failures on main and opens structured GitHub issues for automated remediation. Follow this skill strictly when analyzing CI failures.
---

# Flaky Test Catcher — Analysis Protocol

## 1. Overview / Purpose

This skill defines the end-to-end protocol for detecting broken and flaky acceptance tests by analyzing recent CI failures on `main`, then opening structured GitHub issues so that automated remediation workflows can address the root causes.

You **must** follow this protocol strictly. Do not improvise or skip steps.

## 2. Inputs from pre-activation context

The workflow pre-activation step has already computed all run-level data. Do **not** re-query GitHub for run lists or issue counts. Use the values injected into your prompt:

| Variable | Meaning |
|---|---|
| `failed_run_ids` | JSON array of run IDs of failed `test.yml` runs on `main` in the last 3 days |
| `total_run_count` | Count of completed runs with meaningful conclusions (success, failure, timed\_out, neutral, action\_required) — cancelled runs excluded |
| `open_issues` | Current count of open `flaky-test` issues |
| `issue_slots_available` | How many new issues you may create (max 3) |

Parse `failed_run_ids` as a JSON array immediately. Example: `["12345678","87654321"]`.

## 3. Fetching job logs

For each run ID in `failed_run_ids`:

### 3.1 List jobs for the run

```
gh api /repos/{owner}/{repo}/actions/runs/{run_id}/jobs?per_page=100
```

Replace `{owner}` and `{repo}` with the values from the repository's GitHub remote (visible in `git remote get-url origin`).

### 3.2 Filter to relevant failing jobs

From the returned `jobs` array, keep only jobs where **both** conditions hold:
- `conclusion == "failure"`
- The job name contains `Matrix Acceptance Test` (this is the job group that runs acceptance tests)

Ignore infrastructure jobs (e.g. lint, build, generate) — they do not produce `--- FAIL:` lines.

### 3.3 Fetch the log for each failing job

The GitHub API log endpoint returns a ZIP archive. Use the `gh` CLI to stream plain-text logs for a job:

```bash
gh run view --job {job_id} --log | grep '^--- FAIL:'
```

**Log size warning**: Job logs can be very large (10 MB+). Do **not** load the full log into context. Instead:
- Stream the log output through `grep` so only matching lines are retained.
- Scan only for lines matching the `--- FAIL:` pattern (see §4).
- Use `grep -B3 -A3` or similar when you also need surrounding context for the "Sample Failure Output" issue section.
- Capture a small surrounding context (3–5 lines before/after each `--- FAIL:` line) for the "Sample Failure Output" issue section.

### 3.4 Pagination

If a job list response has 100 items and there may be more, check the `Link` header for a `next` page URL and repeat the request. In practice, a single run rarely has more than 100 jobs.

## 4. `--- FAIL:` extraction pattern

Scan each log for lines that match this exact pattern:

```
^--- FAIL: TestName (timing)
```

Examples of matching lines:
```
--- FAIL: TestAccResourceAgentConfiguration_alternateEnvironment (12.34s)
--- FAIL: TestAccSomeResource_basic (0.45s)
```

**Rules:**
- The line must start with `--- FAIL:` (three hyphens, a space, `FAIL:`, a space).
- The test name follows Go test naming: `TestAcc...` with optional underscore-separated sub-name.
- Extract **only the test function name** — strip the timing suffix `(12.34s)`.
- **Ignore** bare `FAIL` lines without the `---` prefix; those are package-level failure markers, not individual test failures.

Collect all extracted test names across all runs and all jobs. A test may appear multiple times (once per run where it failed) — track counts.

**Same-run deduplication**: If the same test name appears in multiple failing jobs within a single run (e.g. multiple shards both failing the same test), count it **only once** for that run. Deduplication is by run ID, not job ID — use a set per run when accumulating test names.

## 5. Fail-rate formula and thresholds

For each unique test name:

```
fail_rate = fail_count / total_run_count
```

Where:
- `fail_count` = number of distinct run IDs in which this test appeared as `--- FAIL:`
- `total_run_count` = the value from pre-activation context (already excludes cancelled runs)

| Classification | Condition | Action |
|---|---|---|
| **Broken** | `fail_rate == 1.0` (fails in 100% of runs) | Create issue |
| **Flaky** | `fail_rate >= 0.20` and `< 1.0` | Create issue |
| **Noise** | `fail_rate < 0.20` | Ignore — do not create issues |

## 6. Base-test-name grouping rule

Extract the **base test name** from each test function name using this rule:

> Take the substring from the beginning up to (but not including) the first underscore `_`.

Pattern: `TestAcc[^_]+`

Examples:

| Full test name | Base test name |
|---|---|
| `TestAccResourceAgentConfiguration_alternateEnvironment` | `TestAccResourceAgentConfiguration` |
| `TestAccResourceAgentConfiguration_minimal` | `TestAccResourceAgentConfiguration` |
| `TestAccResourceAgentConfiguration` | `TestAccResourceAgentConfiguration` |
| `TestAccSomeResource_basic` | `TestAccSomeResource` |

**One issue per base test name.** All scenario variants (subtests/suffixes) belonging to the same base name are consolidated into a single issue. List each specific variant inside the issue body.

**Fallback for non-`TestAcc` tests**: All acceptance tests in this project follow the `TestAcc` prefix convention. If a non-`TestAcc` test name appears in the logs, treat everything up to the first `_` (or the full name if no `_`) as the base name.

## 7. Commit analysis steps

For each base test name that will receive an issue, investigate whether any recent commit may already address the failure:

### 7.1 Find the oldest failing run timestamp

From the `failed_run_ids` list, identify the oldest run's `created_at` timestamp. You can get metadata for a single run:

```
gh api /repos/{owner}/{repo}/actions/runs/{run_id}
```

Take the minimum `created_at` across all failed runs.

### 7.2 Fetch commits on `main` since that timestamp

```
gh api "/repos/{owner}/{repo}/commits?sha=main&since={timestamp}&per_page=50"
```

Replace `{timestamp}` with the ISO 8601 value from step 7.1 (e.g. `2024-01-15T12:00:00Z`).

### 7.3 For each commit, check relevance

For each commit returned:

**a. Commit message relevance** — does it reference any of:
  - The base test name (e.g. `TestAccResourceAgentConfiguration`)
  - The resource name (derive from the test name, e.g. `agent_configuration` → `AgentConfiguration`)
  - Keywords: `fix`, `flaky`, `test`, `revert`

**b. Changed file relevance** — fetch the full commit detail to get changed file paths:

```
gh api /repos/{owner}/{repo}/commits/{sha}
```

Check if any file in `files[].filename` matches patterns like:
  - `*_test.go` files whose name contains a token from the resource name
  - Files in the same Go package directory as the test

### 7.4 Include findings in issue body

- If a relevant commit is found:
  ```
  ⚠️ may already be addressed in `{short_sha}` — {one-line message summary}
  ```
- If no relevant commits found:
  ```
  No recent commits appear to address this failure.
  ```

Frame the analysis as "has this been fixed yet?" — not as blame attribution.

**Do not suppress issue creation**: Even if a fix commit is found, always proceed with creating the issue and include the fix-detection note in the Commit Analysis section. The issue serves as the remediation trigger regardless.

## 8. Issue deduplication

Before creating an issue for a base test name, check whether one already exists:

```
gh api "/repos/{owner}/{repo}/issues?labels=flaky-test&state=open&per_page=100"
```

- Inspect each returned issue's `title`.
- The **full rendered title** for flaky-test issues is: `[flaky-test] {BaseTestName}` (the `[flaky-test] ` prefix is applied automatically by the `create-issue` safe output). When checking for duplicates, compare against this full title as it appears in GitHub.
- If an **exact title match** exists, skip creating a new issue for that base test name.
- Do **not** re-query or recalculate `issue_slots_available`; use only the value from pre-activation.
- If the response contains exactly 100 results, check for a `next` link in the `Link` response header and repeat the request for subsequent pages until all open issues are fetched.

## 9. Required issue body sections

**Issue title**: Pass only `{BaseTestName}` to the `create-issue` safe output — the `[flaky-test] ` prefix is added automatically.

Every issue you create must contain exactly these 5 sections in this order:

```markdown
## Broken Tests

List each test function name (including scenario suffix) that failed in 100% of runs:
- ❌ `TestAccResourceFoo_basic` — failed in 5/5 runs

## Flaky Tests

List each test function name that failed in ≥ 20% but < 100% of runs, with the observed rate:
- ⚠️ `TestAccResourceFoo_update` — failed in 3/5 runs (60%)
- ⚠️ `TestAccResourceFoo_import` — failed in 1/5 runs (20%)

## Commit Analysis

{Output from §7. Either a ⚠️ note about a possible fix commit, or the "No recent commits" message. Include commit SHA, message, and affected file paths when relevant.}

## Sample Failure Output

{Short excerpt (5–15 lines) of the actual log output surrounding a `--- FAIL:` line. Include any immediately preceding error messages for context.}

## Affected Stack Versions

{List the Elastic Stack versions / matrix dimension values (e.g. Elasticsearch version, Kibana version) from the failing job names or log metadata. If not determinable, write "Unknown — not present in log output".}
```

**Formatting rules:**
- Use `❌` for broken tests (100% fail rate).
- Use `⚠️` for flaky tests (20%–99% fail rate), and always include the fraction and percentage.
- Keep "Sample Failure Output" to the most informative excerpt; do not paste hundreds of lines.

**Cap enforcement**: Before creating each issue, verify that the number of issues created so far in this run has not reached `issue_slots_available`. Stop creating issues once the cap is reached, even if additional base test names remain.

## 10. Noop conditions

Call `noop` with a descriptive explanation (do not create any issues) when **any** of these conditions holds:

1. **All failures already have open issues** — every qualifying base test name matched an existing open `flaky-test` issue during deduplication; nothing new to open.

2. **All failures are below the 20% threshold** — every observed `--- FAIL:` test has `fail_rate < 0.20`; there are no broken or flaky tests to report.

3. **No `--- FAIL:` patterns found** — none of the logs for the provided `failed_run_ids` contained a `--- FAIL:` line; the CI failures were likely infrastructure failures (network timeouts, setup errors, etc.) rather than test logic failures.

When calling `noop`, state which condition applied and include basic counts (e.g. "3 failures observed, all below 20% threshold").

## Summary of execution order

1. Parse `failed_run_ids` from pre-activation context.
2. For each run ID: list jobs → filter to failing `Matrix Acceptance Test` jobs → fetch and scan logs for `--- FAIL:` lines.
3. Aggregate per-test fail counts across all runs.
4. Apply fail-rate thresholds; discard noise (`< 20%`).
5. Group surviving tests by base test name.
6. Deduplicate against existing open `flaky-test` issues.
7. For each remaining base test name (up to `issue_slots_available`): run commit analysis, then create an issue with all 5 required sections.
8. If no issues were created, call `noop`.
