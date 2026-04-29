# Task 2 Coverage/Test Analysis Report

## Scope
Task 2 changes in `remove-7x-support` touch ONLY the `Makefile` — specifically the Fleet image fallback logic.

## 1. Tests for Makefile Fleet Image Logic

**Result: NO tests exist.**

- There are no unit tests, shell tests, or Makefile-level tests that exercise the `FLEET_IMAGE` fallback condition.
- The relevant Makefile logic is:
  ```makefile
  ifneq (,$(filter 8.0.% 8.1.%,$(STACK_VERSION)))
  FLEET_IMAGE := elastic/elastic-agent
  endif
  ```
- The only tests found that mention "fleet" are Go **acceptance tests** (`*_test.go` in `internal/fleet/...`) that exercise the Terraform provider's Fleet resources against a running stack. These do not test, validate, or assert the Makefile image selection behavior.
- No test file sets `STACK_VERSION=8.0.x` or `8.1.x` to verify the fallback path.

## 2. Remaining `7.17` References in Workflows / Makefile

**Result: NO inappropriate `7.17` references remain.**

| Location | `7.17` Found? | Notes |
|----------|---------------|-------|
| `Makefile` | No | Fleet fallback only references `8.0.%` and `8.1.%`. No `7.x` logic. |
| `.github/workflows/test.yml` | No | Test matrix starts at `8.0.1`. No `7.x` versions. |
| `.github/workflows-src/test/workflow.yml.tmpl` | No | Source template matrix also starts at `8.0.1`. |
| `.github/workflows-src/` (other templates) | No | Incidental `7.` matches are step numbers in markdown instructions, not version refs. |
| `.github/workflows/code-factory-issue.lock.yml` | No | Fleet setup steps present, but no `7.17` refs. |
| `.github/workflows/copilot-setup-steps.yml` | No | Fleet setup steps present, but no `7.17` refs. |

## 3. `7.17` References Elsewhere in Codebase (Unrelated to Task 2)

The grep sweep did find `7.17` in a few files, but **none are related to the Makefile fleet image logic or CI orchestration**:

- `internal/clients/elasticsearch_scoped_client_test.go` — test data asserting `7.17.0` does not satisfy min version `8.0.0`.
- `internal/clients/kibanaoapi/status_test.go` — test data parsing a `7.17.0` status response.
- `internal/clients/elasticsearch/transform.go` — functional code using `7.17.0` as the minimum supported version for the `timeout` parameter in the Transform API.
- `internal/fleet/customintegration/acc_test.go` — YAML test fixture using `kibana.version: "^7.17.0 || ^8.0.0 || ^9.0.0"` for a custom integration package manifest.

## 4. Untested Paths

- The Makefile `FLEET_IMAGE` fallback for `STACK_VERSION=8.0.x` or `8.1.x` is **untested**.
- This is standard for build/orchestration Makefiles in this repo; there is no existing test harness for Makefile variable evaluation.
- The logic is trivial (a `filter` conditional) and is mirrored by the CI workflow template, which explicitly sets `fleetImage: elastic/elastic-agent` for `8.0.1` and `8.1.3` matrix entries.

## Conclusion

- **No tests exist** for the Makefile Fleet image fallback logic.
- **No remaining `7.17` references** in workflows or Makefile that should be removed.
- The change is a build/infra-only Makefile edit with no Terraform entity impact.
- **Nothing further is needed** for Task 2 coverage verification.
