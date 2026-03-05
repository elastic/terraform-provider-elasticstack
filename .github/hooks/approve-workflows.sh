#!/usr/bin/env bash
set -euo pipefail

# Best-effort: hooks should never block the agent.
if ! command -v gh >/dev/null 2>&1; then
  exit 0
fi

# gh authenticates via GH_TOKEN; it also supports GITHUB_TOKEN.
if [[ -z "${GH_TOKEN:-}" && -z "${GITHUB_TOKEN:-}" ]]; then
  exit 0
fi

REPO="${GITHUB_REPOSITORY:-}"
if [[ -z "$REPO" ]]; then
  exit 0
fi

BRANCH="${GITHUB_HEAD_REF:-}"
if [[ -z "$BRANCH" ]]; then
  BRANCH="$(git rev-parse --abbrev-ref HEAD 2>/dev/null || true)"
fi

if [[ -z "$BRANCH" || "$BRANCH" == "HEAD" ]]; then
  exit 0
fi

approve_runs_for_status() {
  local status="$1"

  local run_ids=""
  run_ids="$(gh api "repos/${REPO}/actions/runs?branch=${BRANCH}&status=${status}&per_page=100" \
    --jq '.workflow_runs[].id' 2>/dev/null || true)"

  if [[ -z "$run_ids" ]]; then
    return 0
  fi

  for run_id in $run_ids; do
    gh api --method POST "repos/${REPO}/actions/runs/${run_id}/approve" >/dev/null 2>&1 || true
  done
}

# The UI "Approve and run workflows" state typically maps to action_required.
approve_runs_for_status "action_required" || true

# Some runs may surface as waiting in the API.
approve_runs_for_status "waiting" || true

exit 0
