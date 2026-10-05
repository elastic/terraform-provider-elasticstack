#!/usr/bin/env bash
# Compute the targeted acceptance test packages for a provider.yml test-matrix
# shard (the `targeted` / compute-packages step). Extracted verbatim from the
# inline workflow step so the logic is exercised by tests
# (.github/scripts/workflows/lib/targeted-testacc-compute.test.mjs).
#
# Inputs, passed via the step's env: block:
#   EVENT_NAME    - github.event_name; only "pull_request" runs the classifier
#   PR_BASE_SHA   - github.event.pull_request.base.sha (PR events only)
#   SHARD         - matrix.shard
#   GITHUB_OUTPUT - GitHub Actions step-output file (tests point this at a
#                   temp file)
#
# set -u note: the provider.yml env: block always defines EVENT_NAME,
# PR_BASE_SHA, and SHARD, so they are bound in production; if a required input
# is ever missing the script aborts (fail-closed) rather than silently
# skipping tests, so -u is kept.
set -euo pipefail

if [[ "${EVENT_NAME}" != "pull_request" ]]; then
  echo "has_packages=true" >>"$GITHUB_OUTPUT"
  echo "targeted_pkgs=" >>"$GITHUB_OUTPUT"
  exit 0
fi
# The actions/checkout step gives us a shallow merge commit only, so
# merge-base / HEAD~1 resolution fails and the tool falls back to the
# full suite. Fetch the PR base commit and pass it explicitly so the
# tool can compute the real diff for this PR.
compute_targeted() {
  local out
  if ! out=$(go run ./scripts/targeted-testacc/... "$@"); then
    echo "::error::targeted-testacc failed; refusing to skip tests"
    exit 1
  fi
  targeted=$(printf '%s' "$out" | tr '\n' ' ')
}
base_ref="refs/remotes/origin/pr-base"
if git fetch origin +"${PR_BASE_SHA}":"${base_ref}" --depth=1; then
  compute_targeted --base="${base_ref}" --total-shards=2 --shard-index="${SHARD}"
else
  # Base-commit fetch failed (e.g. shallow history / pruned ref);
  # fall back to re-invoking the tool without --base, which resolves
  # the baseline from merge-base origin/main HEAD, or HEAD~1 when no
  # origin/main is available.
  compute_targeted --total-shards=2 --shard-index="${SHARD}"
fi
targeted=$(echo "$targeted" | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//')
if [[ -n "$targeted" ]]; then
  echo "has_packages=true" >>"$GITHUB_OUTPUT"
  echo "targeted_pkgs=$targeted" >>"$GITHUB_OUTPUT"
else
  echo "has_packages=false" >>"$GITHUB_OUTPUT"
fi
