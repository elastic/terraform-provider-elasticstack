#!/usr/bin/env bash
# PR planning is capped at two shards to match provider.yml's packages_0/1 outputs.
#
# Inputs, passed via the step's env: block:
#   EVENT_NAME    - github.event_name; only "pull_request" runs the selector
#   PR_BASE_SHA   - github.event.pull_request.base.sha (PR events only)
#   GITHUB_OUTPUT - GitHub Actions step-output file (tests point this at a
#                   temp file)
#
# Outputs (consumed by the load-matrix job and the matrix test job):
#   has_packages - whether the plan selects any package
#   shards       - JSON array of numeric shard indices for the matrix axis
#   packages_0   - space-joined packages assigned to shard 0
#   packages_1   - space-joined packages assigned to shard 1
#
set -euo pipefail

if [[ "${EVENT_NAME}" != "pull_request" ]]; then
  {
    echo "has_packages=true"
    echo "shards=[0,1]"
    echo "packages_0="
    echo "packages_1="
  } >>"$GITHUB_OUTPUT"
  exit 0
fi
# The actions/checkout step gives us a shallow merge commit only, so
# merge-base / HEAD~1 resolution fails and the tool falls back to the
# full suite. Fetch the PR base commit and pass it explicitly so the
# tool can compute the real diff for this PR.
selector_output() {
  local out
  if ! out=$(go run ./scripts/targeted-testacc/... "$@"); then
    echo "::error::targeted-testacc failed; refusing to skip tests" >&2
    exit 1
  fi
  printf '%s' "$out"
}

base_ref="refs/remotes/origin/pr-base"
if git fetch origin +"${PR_BASE_SHA}":"${base_ref}" --depth=1; then
  plan=$(selector_output --base="${base_ref}" --total-shards=2)
else
  echo "::warning::PR base commit fetch failed; running the selector without --base, which may fall back to the full suite" >&2
  plan=$(selector_output --total-shards=2)
fi

if ! jq -e '
  (.has_packages == ((.selected_packages | length) > 0)) and
  (.shards | length >= 1) and
  (.shards | length <= 2) and
  all(.shards[]; type == "array") and
  (if (.selected_packages | length) == 0
   then (.shards | length == 1) and (.shards[0] | length == 0)
   else all(.shards[]; length > 0)
   end) and
  ((.selected_packages | sort) == ([.shards[]] | flatten | sort))
' <<<"$plan" >/dev/null; then
  echo "::error::invalid targeted-testacc shard plan; refusing to skip tests" >&2
  exit 1
fi

outputs=$(jq -r '
  "has_packages=\(.has_packages)",
  "shards=\([range(0; (.shards | length))] | tojson)",
  "packages_0=\(.shards[0] | join(" "))",
  "packages_1=\((.shards[1] // []) | join(" "))"
' <<<"$plan")
printf '%s\n' "$outputs" >>"$GITHUB_OUTPUT"
