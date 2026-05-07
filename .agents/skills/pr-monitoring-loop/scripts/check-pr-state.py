#!/usr/bin/env python3
"""Collect deterministic GitHub PR state for watcher subagents.

The script is intentionally read-only against GitHub. It fetches PR metadata,
checks (pinned to the current head SHA), reviews, comments, review threads, and
issue events, then computes a `summary` block that watcher subagents use to
decide whether the PR is actionable.

Key concepts:

- `--state-file` persists across watcher restarts. It records seen comment/
  review IDs and timestamps so "new comment since last poll" survives the
  short-lived nature of fresh subagents.
- New-vs-old discrimination is purely ID/timestamp based. There is NO author
  filtering: bots (verify-openspec, macroscope, github-actions) emit
  first-class signal, and the PR author is frequently the human reviewer
  because the agent commits on their behalf.
- `summary.reviews.verifyOpenspec.runState` derives the verify-openspec
  workflow state. The workflow REMOVES its own label as soon as it picks the
  PR up, so label absence on `pr.labels` is not a signal to re-request.
"""

from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys
import time
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Callable, Iterable, Optional


PR_VIEW_FIELDS = [
    "author",
    "baseRefName",
    "baseRefOid",
    "headRefName",
    "headRefOid",
    "isDraft",
    "labels",
    "mergeStateStatus",
    "mergeable",
    "number",
    "reviewDecision",
    "state",
    "statusCheckRollup",
    "title",
    "updatedAt",
    "url",
]

CHECK_FIELDS = [
    "bucket",
    "completedAt",
    "conclusion",
    "detailsUrl",
    "link",
    "name",
    "startedAt",
    "state",
    "workflow",
]

VERIFY_OPENSPEC_LABEL = "verify-openspec"
# The verify-openspec workflow runs under GitHub Actions and posts as
# `github-actions[bot]`. Author identity alone is not enough (that bot posts
# many other things), so we ALSO require the body to contain the workflow's
# distinctive "OpenSpec verify" / "Verification Report" signature.
VERIFY_OPENSPEC_REVIEW_AUTHORS = {"github-actions[bot]", "github-actions"}
VERIFY_OPENSPEC_REVIEW_BODY_MARKERS = (
    "OpenSpec verify",
    "Verification Report",
)

EXIT_OK = 0
EXIT_TRANSIENT = 2
EXIT_TIMEOUT = 124

DEFAULT_INTERVAL_SECONDS = 60
DEFAULT_MAX_DURATION_SECONDS = 1800
SEEN_ID_CAP = 1000


# ---------------------------------------------------------------------------
# Subprocess wrappers (mockable seam for tests)
# ---------------------------------------------------------------------------


class TransientGhError(RuntimeError):
    """Raised when a `gh` invocation fails after retries."""


def run_gh(
    args: list[str],
    *,
    allow_failure: bool = False,
    retries: int = 1,
) -> str:
    """Run a `gh` command, optionally retrying once on transient failures."""

    attempt = 0
    last_stderr = ""
    while True:
        result = subprocess.run(
            ["gh", *args],
            check=False,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
        )
        if result.returncode == 0:
            return result.stdout
        last_stderr = result.stderr.strip()
        if attempt >= retries:
            break
        attempt += 1
        time.sleep(1.0 * attempt)

    if allow_failure:
        return ""

    raise TransientGhError(
        json.dumps(
            {
                "error": "gh command failed",
                "command": ["gh", *args],
                "stderr": last_stderr,
                "returncode": result.returncode,
            }
        )
    )


def run_git(
    args: list[str], *, allow_failure: bool = False
) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        ["git", *args],
        check=not allow_failure,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
    )


def gh_json(
    args: list[str],
    *,
    default: Any,
    allow_failure: bool = False,
    retries: int = 1,
) -> Any:
    output = run_gh(args, allow_failure=allow_failure, retries=retries)
    if not output.strip():
        return default
    try:
        return json.loads(output)
    except json.JSONDecodeError:
        if allow_failure:
            return default
        raise TransientGhError(
            json.dumps(
                {
                    "error": "gh command returned invalid JSON",
                    "command": ["gh", *args],
                    "stdout": output[:2000],
                }
            )
        )


# ---------------------------------------------------------------------------
# Data fetchers
# ---------------------------------------------------------------------------


def repo_info() -> dict[str, str]:
    data = gh_json(["repo", "view", "--json", "owner,name"], default={})
    owner = data.get("owner", {})
    return {"owner": owner.get("login", ""), "name": data.get("name", "")}


def pr_view(pr: str) -> dict[str, Any]:
    return gh_json(
        ["pr", "view", pr, "--json", ",".join(PR_VIEW_FIELDS)],
        default={},
    )


def pr_checks(pr: str) -> list[dict[str, Any]]:
    return gh_json(
        ["pr", "checks", pr, "--json", ",".join(CHECK_FIELDS)],
        default=[],
        allow_failure=True,
    )


def commit_check_runs(owner: str, repo: str, sha: str) -> list[dict[str, Any]]:
    """Check-runs pinned to a specific commit SHA (canonical for the head)."""

    if not (owner and repo and sha):
        return []
    data = gh_json(
        [
            "api",
            f"repos/{owner}/{repo}/commits/{sha}/check-runs",
            "--paginate",
        ],
        default={},
        allow_failure=True,
    )
    if isinstance(data, list):
        runs: list[dict[str, Any]] = []
        for page in data:
            if isinstance(page, dict):
                runs.extend(page.get("check_runs", []))
        return runs
    if isinstance(data, dict):
        return data.get("check_runs", [])
    return []


def commit_combined_status(
    owner: str, repo: str, sha: str
) -> dict[str, Any]:
    if not (owner and repo and sha):
        return {}
    return gh_json(
        ["api", f"repos/{owner}/{repo}/commits/{sha}/status"],
        default={},
        allow_failure=True,
    )


def issue_comments(owner: str, repo: str, number: int) -> list[dict[str, Any]]:
    return gh_json(
        ["api", f"repos/{owner}/{repo}/issues/{number}/comments", "--paginate"],
        default=[],
        allow_failure=True,
    )


def review_comments(owner: str, repo: str, number: int) -> list[dict[str, Any]]:
    return gh_json(
        ["api", f"repos/{owner}/{repo}/pulls/{number}/comments", "--paginate"],
        default=[],
        allow_failure=True,
    )


def reviews(owner: str, repo: str, number: int) -> list[dict[str, Any]]:
    return gh_json(
        ["api", f"repos/{owner}/{repo}/pulls/{number}/reviews", "--paginate"],
        default=[],
        allow_failure=True,
    )


def issue_events(owner: str, repo: str, number: int) -> list[dict[str, Any]]:
    return gh_json(
        ["api", f"repos/{owner}/{repo}/issues/{number}/events", "--paginate"],
        default=[],
        allow_failure=True,
    )


def review_threads(owner: str, repo: str, number: int) -> list[dict[str, Any]]:
    query = """
query($owner: String!, $repo: String!, $number: Int!, $cursor: String) {
  repository(owner: $owner, name: $repo) {
    pullRequest(number: $number) {
      reviewThreads(first: 100, after: $cursor) {
        pageInfo { hasNextPage endCursor }
        nodes {
          id
          isResolved
          isOutdated
          path
          line
          originalLine
          comments(first: 50) {
            nodes {
              id
              databaseId
              author { login }
              body
              createdAt
              updatedAt
              diffHunk
              path
              line
              originalLine
              outdated
              url
            }
          }
        }
      }
    }
  }
}
"""
    nodes: list[dict[str, Any]] = []
    cursor = ""
    while True:
        api_args = [
            "api", "graphql",
            "-f", f"owner={owner}",
            "-f", f"repo={repo}",
            "-F", f"number={number}",
            "-f", f"query={query}",
        ]
        if cursor:
            api_args.extend(["-f", f"cursor={cursor}"])
        data = gh_json(api_args, default={}, allow_failure=True)
        threads = (
            data.get("data", {})
            .get("repository", {})
            .get("pullRequest", {})
            .get("reviewThreads", {})
        )
        nodes.extend(threads.get("nodes", []))
        page_info = threads.get("pageInfo", {})
        if not page_info.get("hasNextPage"):
            return nodes
        cursor = page_info.get("endCursor", "")
        if not cursor:
            return nodes


def fetch_pr_merge_refs(pr: dict[str, Any]) -> dict[str, str]:
    number = pr["number"]
    base_ref = pr["baseRefName"]
    local_base_ref = f"refs/remotes/origin/{base_ref}"
    local_pr_ref = f"refs/remotes/origin/pr-{number}-head"
    run_git(
        [
            "fetch", "--quiet", "origin",
            f"+refs/heads/{base_ref}:{local_base_ref}",
            f"+refs/pull/{number}/head:{local_pr_ref}",
        ],
        allow_failure=True,
    )
    base_sha = run_git(["rev-parse", local_base_ref], allow_failure=True).stdout.strip()
    head_sha = run_git(["rev-parse", local_pr_ref], allow_failure=True).stdout.strip()
    return {"base": base_sha, "head": head_sha}


def parse_merge_tree_conflicts(output: str) -> list[str]:
    files: set[str] = set()
    for line in output.splitlines():
        if "\t" in line:
            metadata, path = line.split("\t", 1)
            parts = metadata.split()
            if len(parts) == 4 and parts[2] in {"1", "2", "3"}:
                files.add(path)
                continue
        marker = " Merge conflict in "
        if marker in line:
            files.add(line.split(marker, 1)[1])
    return sorted(files)


def merge_conflicts(pr: dict[str, Any]) -> dict[str, Any]:
    fallback_conflict = pr.get("mergeable") == "CONFLICTING" or pr.get(
        "mergeStateStatus"
    ) == "DIRTY"
    try:
        refs = fetch_pr_merge_refs(pr)
        if not refs.get("base") or not refs.get("head"):
            raise RuntimeError("could not resolve PR merge refs")
        result = run_git(
            ["merge-tree", "--write-tree", refs["base"], refs["head"]],
            allow_failure=True,
        )
    except (KeyError, subprocess.CalledProcessError, FileNotFoundError, RuntimeError) as exc:
        return {
            "analysisAvailable": False,
            "hasConflicts": fallback_conflict,
            "files": [],
            "source": "github-mergeability-fallback",
            "error": str(exc),
        }
    output = "\n".join(part for part in [result.stdout, result.stderr] if part)
    files = parse_merge_tree_conflicts(output)
    return {
        "analysisAvailable": result.returncode in {0, 1},
        "hasConflicts": bool(files) or fallback_conflict,
        "files": files,
        "source": "git-merge-tree",
        "base": refs["base"],
        "head": refs["head"],
        "mergeTreeExitCode": result.returncode,
        "details": output.strip(),
    }


# ---------------------------------------------------------------------------
# State file
# ---------------------------------------------------------------------------


def default_state_file_path(pr_number: int) -> str:
    """Default to a state directory alongside this script.

    We deliberately do NOT use ``$GIT_DIR``: this repo (and many like it)
    relies on git worktrees, where ``.git`` is a file pointing at a worktree-
    specific git dir. Storing state under the worktree's git dir would
    fragment "seen IDs" across worktrees that share the same PR. Putting the
    file alongside the script keeps state global to the checkout, which
    matches how `pr-monitoring-loop` is meant to be used.
    """

    state_dir = Path(__file__).resolve().parent / "state"
    return str(state_dir / f".pr-monitor-{pr_number}.json")


def load_state(path: str | None) -> dict[str, Any]:
    if not path:
        return {}
    try:
        with open(path, "r", encoding="utf-8") as fh:
            data = json.load(fh)
            if isinstance(data, dict):
                return data
    except (FileNotFoundError, json.JSONDecodeError, OSError):
        pass
    return {}


def save_state(path: str | None, state: dict[str, Any]) -> None:
    if not path:
        return
    try:
        target = Path(path)
        target.parent.mkdir(parents=True, exist_ok=True)
        with open(target, "w", encoding="utf-8") as fh:
            json.dump(state, fh, indent=2, sort_keys=True)
    except OSError:
        # State persistence is best-effort; we don't fail the poll over it.
        pass


def cap_seen_ids(ids: Iterable[Any], cap: int = SEEN_ID_CAP) -> list[Any]:
    deduped: list[Any] = []
    seen: set[Any] = set()
    for item in ids:
        if item in seen:
            continue
        seen.add(item)
        deduped.append(item)
    if len(deduped) > cap:
        deduped = deduped[-cap:]
    return deduped


# ---------------------------------------------------------------------------
# Pure helpers (used by compute_payload and unit tests)
# ---------------------------------------------------------------------------


def parse_iso(ts: str | None) -> Optional[datetime]:
    if not ts:
        return None
    try:
        if ts.endswith("Z"):
            ts = ts[:-1] + "+00:00"
        return datetime.fromisoformat(ts)
    except (TypeError, ValueError):
        return None


def now_iso() -> str:
    return datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")


def normalize_check(check: dict[str, Any]) -> dict[str, Any]:
    state = (check.get("state") or check.get("status") or "").upper()
    conclusion = (check.get("conclusion") or "").upper()
    bucket = (check.get("bucket") or "").lower()

    failed = bucket == "fail" or conclusion in {
        "ACTION_REQUIRED",
        "CANCELLED",
        "FAILURE",
        "SKIPPED",
        "STALE",
        "TIMED_OUT",
    }
    pending = bucket in {"pending", "running", "unknown"} or state in {
        "EXPECTED",
        "IN_PROGRESS",
        "PENDING",
        "QUEUED",
        "REQUESTED",
        "WAITING",
    }
    passed = bucket == "pass" or conclusion == "SUCCESS"

    return {
        **check,
        "derived": {
            "failed": failed,
            "pending": pending and not failed,
            "passed": passed and not failed,
        },
    }


def normalize_check_run(run: dict[str, Any]) -> dict[str, Any]:
    """Normalize a REST `check-runs` entry into the same `derived` shape."""

    status = (run.get("status") or "").upper()
    conclusion = (run.get("conclusion") or "").upper()
    failed = conclusion in {
        "ACTION_REQUIRED",
        "CANCELLED",
        "FAILURE",
        "TIMED_OUT",
        "STALE",
    }
    pending = status in {"QUEUED", "IN_PROGRESS", "PENDING", "WAITING"}
    passed = conclusion == "SUCCESS"
    return {
        **run,
        "derived": {
            "failed": failed,
            "pending": pending and not failed,
            "passed": passed and not failed,
        },
    }


def normalize_combined_status_entry(entry: dict[str, Any]) -> dict[str, Any]:
    state = (entry.get("state") or "").upper()
    failed = state in {"FAILURE", "ERROR"}
    pending = state == "PENDING"
    passed = state == "SUCCESS"
    return {
        **entry,
        "derived": {
            "failed": failed,
            "pending": pending and not failed,
            "passed": passed and not failed,
        },
    }


def derive_verify_openspec(
    review_data: list[dict[str, Any]],
    event_data: list[dict[str, Any]],
    head_sha: str,
) -> dict[str, Any]:
    """Compute verify-openspec workflow state for the current head SHA.

    The verify-openspec workflow removes its own label as soon as it starts
    processing, so label absence is NOT a signal to re-request. We derive
    runState from the labeled/unlabeled timeline plus the bot's reviews.
    """

    def is_verify_review(review: dict[str, Any]) -> bool:
        login = ((review.get("user") or {}).get("login") or "").lower()
        if login not in {a.lower() for a in VERIFY_OPENSPEC_REVIEW_AUTHORS}:
            return False
        body = review.get("body") or ""
        return any(marker in body for marker in VERIFY_OPENSPEC_REVIEW_BODY_MARKERS)

    label_applied_at: Optional[str] = None
    label_removed_at: Optional[str] = None
    for event in event_data:
        ev_type = event.get("event")
        label = (event.get("label") or {}).get("name")
        if label != VERIFY_OPENSPEC_LABEL:
            continue
        created = event.get("created_at")
        if ev_type == "labeled":
            if not label_applied_at or (parse_iso(created) or datetime.min.replace(tzinfo=timezone.utc)) > (parse_iso(label_applied_at) or datetime.min.replace(tzinfo=timezone.utc)):
                label_applied_at = created
        elif ev_type == "unlabeled":
            if not label_removed_at or (parse_iso(created) or datetime.min.replace(tzinfo=timezone.utc)) > (parse_iso(label_removed_at) or datetime.min.replace(tzinfo=timezone.utc)):
                label_removed_at = created

    verify_reviews = sorted(
        [r for r in review_data if is_verify_review(r)],
        key=lambda r: parse_iso(r.get("submitted_at")) or datetime.min.replace(tzinfo=timezone.utc),
    )
    reviews_for_head = [r for r in verify_reviews if r.get("commit_id") == head_sha]

    last_approval = None
    for r in reversed(verify_reviews):
        if (r.get("state") or "").upper() == "APPROVED":
            last_approval = r
            break

    last_approval_at = (last_approval or {}).get("submitted_at")
    last_approval_head = (last_approval or {}).get("commit_id")

    label_applied_dt = parse_iso(label_applied_at)
    label_removed_dt = parse_iso(label_removed_at)
    label_active = (
        label_applied_dt is not None
        and (label_removed_dt is None or label_removed_dt < label_applied_dt)
    )
    label_consumed = (
        label_applied_dt is not None
        and label_removed_dt is not None
        and label_removed_dt >= label_applied_dt
    )

    run_state = "none"

    if reviews_for_head:
        last_for_head = reviews_for_head[-1]
        last_for_head_at = parse_iso(last_for_head.get("submitted_at"))
        # Was a label applied AFTER this review? Then the workflow has been
        # re-requested for the same head SHA and the prior review is stale.
        if (
            label_applied_dt is not None
            and last_for_head_at is not None
            and label_applied_dt > last_for_head_at
        ):
            run_state = "pending-pickup" if label_active else "in-progress"
        else:
            state_upper = (last_for_head.get("state") or "").upper()
            if state_upper == "APPROVED":
                run_state = "approved-current"
            elif state_upper == "CHANGES_REQUESTED":
                run_state = "changes-requested"
            else:
                run_state = "none"
    elif label_active:
        run_state = "pending-pickup"
    elif label_consumed:
        run_state = "in-progress"
    elif verify_reviews:
        last_old = verify_reviews[-1]
        if (last_old.get("state") or "").upper() == "APPROVED":
            run_state = "approved-stale"
        else:
            run_state = "none"

    return {
        "lastLabelAppliedAt": label_applied_at,
        "lastLabelRemovedAt": label_removed_at,
        "lastApprovalAt": last_approval_at,
        "lastApprovalHeadSha": last_approval_head,
        "approvalIsCurrent": run_state == "approved-current",
        "runState": run_state,
    }


def derive_latest_by_reviewer(
    review_data: list[dict[str, Any]],
) -> dict[str, dict[str, Any]]:
    latest: dict[str, dict[str, Any]] = {}
    for review in review_data:
        login = ((review.get("user") or {}).get("login") or "").lower()
        if not login:
            continue
        state = (review.get("state") or "").upper()
        # Ignore COMMENTED-only reviews for "decision" purposes.
        if state not in {"APPROVED", "CHANGES_REQUESTED", "DISMISSED"}:
            continue
        existing = latest.get(login)
        existing_dt = parse_iso((existing or {}).get("submittedAt"))
        candidate_dt = parse_iso(review.get("submitted_at"))
        if existing is None or (
            candidate_dt is not None
            and (existing_dt is None or candidate_dt > existing_dt)
        ):
            latest[login] = {
                "state": state,
                "submittedAt": review.get("submitted_at"),
                "id": review.get("id"),
                "commitId": review.get("commit_id"),
            }
    return latest


def derive_effective_decision(latest_by_reviewer: dict[str, dict[str, Any]]) -> str:
    """Return APPROVED, CHANGES_REQUESTED, or REVIEW_REQUIRED."""

    if not latest_by_reviewer:
        return "REVIEW_REQUIRED"
    states = {entry["state"] for entry in latest_by_reviewer.values()}
    if "CHANGES_REQUESTED" in states:
        return "CHANGES_REQUESTED"
    if "APPROVED" in states:
        return "APPROVED"
    return "REVIEW_REQUIRED"


def select_new(
    items: list[dict[str, Any]],
    *,
    id_key: str,
    seen_ids: set[Any],
    timestamp_keys: list[str],
    since_dt: Optional[datetime],
) -> list[dict[str, Any]]:
    """Return items not in seen_ids and (when applicable) newer than `since`."""

    out: list[dict[str, Any]] = []
    for item in items:
        item_id = item.get(id_key)
        if item_id is not None and item_id in seen_ids:
            continue
        if since_dt is not None:
            ts: Optional[datetime] = None
            for key in timestamp_keys:
                ts = parse_iso(item.get(key))
                if ts is not None:
                    break
            if ts is not None and ts <= since_dt:
                # Older than `since` AND not previously seen — treat as new on
                # first run (no state file). Honour `since` only when a since
                # was explicitly provided.
                continue
        out.append(item)
    return out


def thread_root_comment_id(thread: dict[str, Any]) -> Optional[Any]:
    nodes = (thread.get("comments") or {}).get("nodes") or []
    if not nodes:
        return None
    return nodes[0].get("id")


def thread_last_updated(thread: dict[str, Any]) -> Optional[datetime]:
    latest: Optional[datetime] = None
    for comment in (thread.get("comments") or {}).get("nodes") or []:
        for key in ("updatedAt", "createdAt"):
            ts = parse_iso(comment.get(key))
            if ts is not None and (latest is None or ts > latest):
                latest = ts
                break
    return latest


# ---------------------------------------------------------------------------
# Payload computation (pure: no I/O, easy to test)
# ---------------------------------------------------------------------------


def compute_payload(
    *,
    repo: dict[str, str],
    pr: dict[str, Any],
    pr_check_data: list[dict[str, Any]],
    commit_check_run_data: list[dict[str, Any]],
    commit_status_data: dict[str, Any],
    review_data: list[dict[str, Any]],
    issue_comment_data: list[dict[str, Any]],
    review_comment_data: list[dict[str, Any]],
    thread_data: list[dict[str, Any]],
    event_data: list[dict[str, Any]],
    merge_conflict_data: dict[str, Any],
    state: dict[str, Any],
    head_sha_override: Optional[str] = None,
    since_override: Optional[str] = None,
) -> tuple[dict[str, Any], dict[str, Any]]:
    """Build the JSON payload and the next-state-file content."""

    head_sha = head_sha_override or pr.get("headRefOid") or ""

    # --- normalize checks -------------------------------------------------
    raw_pr_checks = [normalize_check(c) for c in pr_check_data]
    pinned_check_runs = [normalize_check_run(r) for r in commit_check_run_data]
    pinned_statuses = [
        normalize_combined_status_entry(s)
        for s in (commit_status_data.get("statuses") or [])
    ]

    # Decide the canonical set of checks for actionable detection. Prefer
    # commit-pinned data when available; fall back to gh pr checks otherwise.
    pinned_combined = pinned_check_runs + pinned_statuses
    canonical_checks = pinned_combined if pinned_combined else raw_pr_checks
    failed_checks = [c for c in canonical_checks if c["derived"]["failed"]]
    pending_checks = [c for c in canonical_checks if c["derived"]["pending"]]
    passed_checks = [c for c in canonical_checks if c["derived"]["passed"]]

    # --- new-since discrimination ----------------------------------------
    seen_issue_ids: set[Any] = set(state.get("seenIssueCommentIds") or [])
    seen_review_comment_ids: set[Any] = set(state.get("seenReviewCommentIds") or [])
    seen_review_ids: set[Any] = set(state.get("seenReviewIds") or [])
    seen_thread_ids: set[Any] = set(state.get("seenReviewThreadIds") or [])

    since_str = since_override or state.get("lastPolledAt")
    since_dt = parse_iso(since_str)
    last_head_sha_in_state = state.get("lastHeadSha")
    head_pushed_recently = (
        last_head_sha_in_state is not None and last_head_sha_in_state != head_sha
    )

    new_issue_comments = select_new(
        issue_comment_data,
        id_key="id",
        seen_ids=seen_issue_ids,
        timestamp_keys=["created_at", "updated_at"],
        since_dt=since_dt,
    )
    new_review_comments = select_new(
        review_comment_data,
        id_key="id",
        seen_ids=seen_review_comment_ids,
        timestamp_keys=["created_at", "updated_at"],
        since_dt=since_dt,
    )
    new_reviews = select_new(
        review_data,
        id_key="id",
        seen_ids=seen_review_ids,
        timestamp_keys=["submitted_at"],
        since_dt=since_dt,
    )

    # --- threads ----------------------------------------------------------
    unresolved_threads = [
        t
        for t in thread_data
        if not t.get("isResolved") and not t.get("isOutdated")
    ]
    unresolved_new = []
    unresolved_updated_since_head = []
    for thread in unresolved_threads:
        thread_id = thread.get("id")
        if thread_id and thread_id not in seen_thread_ids:
            unresolved_new.append(thread)
        last_updated = thread_last_updated(thread)
        if last_updated is not None and since_dt is not None and last_updated > since_dt:
            if thread not in unresolved_new:
                unresolved_updated_since_head.append(thread)

    # --- reviews ---------------------------------------------------------
    latest_by_reviewer = derive_latest_by_reviewer(review_data)
    effective_decision = derive_effective_decision(latest_by_reviewer)
    verify_openspec = derive_verify_openspec(review_data, event_data, head_sha)

    # --- merge state -----------------------------------------------------
    merge_state = pr.get("mergeStateStatus")
    mergeable = pr.get("mergeable")
    has_merge_conflicts = bool(merge_conflict_data.get("hasConflicts"))
    merge_blocked = has_merge_conflicts or merge_state in {
        "BEHIND",
        "BLOCKED",
        "UNKNOWN",
        "UNSTABLE",
    }

    # --- actionable list (rebased on new* / effectiveDecision) -----------
    actionable: list[str] = []
    if failed_checks:
        actionable.append("failed_checks")
    if new_issue_comments:
        actionable.append("issue_comments")
    if new_review_comments:
        actionable.append("review_comments")
    if unresolved_new or unresolved_updated_since_head:
        actionable.append("unresolved_review_threads")
    if effective_decision == "CHANGES_REQUESTED":
        actionable.append("changes_requested")
    if has_merge_conflicts:
        actionable.append("merge_conflicts")
    elif merge_blocked:
        actionable.append("merge_or_branch_state")

    summary = {
        "actionable": actionable,
        "hasActionable": bool(actionable),
        "pr": {
            "number": pr.get("number"),
            "url": pr.get("url"),
            "state": pr.get("state"),
            "isDraft": pr.get("isDraft"),
            "headRefName": pr.get("headRefName"),
            "headRefOid": pr.get("headRefOid"),
            "baseRefName": pr.get("baseRefName"),
            "baseRefOid": pr.get("baseRefOid"),
            "reviewDecision": pr.get("reviewDecision"),
            "mergeable": mergeable,
            "mergeStateStatus": merge_state,
            "labels": [
                label.get("name")
                for label in pr.get("labels", [])
                if isinstance(label, dict)
            ],
        },
        "checks": {
            "source": "commit-pinned" if pinned_combined else "pr-checks",
            "headSha": head_sha,
            "total": len(canonical_checks),
            "failed": len(failed_checks),
            "pending": len(pending_checks),
            "passed": len(passed_checks),
            "failedNames": [c.get("name") for c in failed_checks],
            "pendingNames": [c.get("name") for c in pending_checks],
        },
        "comments": {
            "issueComments": len(issue_comment_data),
            "reviewComments": len(review_comment_data),
            "newIssueComments": len(new_issue_comments),
            "newIssueCommentIds": [c.get("id") for c in new_issue_comments],
            "newReviewComments": len(new_review_comments),
            "newReviewCommentIds": [c.get("id") for c in new_review_comments],
        },
        "threads": {
            "unresolved": len(unresolved_threads),
            "unresolvedNew": len(unresolved_new),
            "unresolvedUpdatedSinceHead": len(unresolved_updated_since_head),
            "unresolvedThreadIds": [t.get("id") for t in unresolved_threads],
            "unresolvedNewThreadIds": [t.get("id") for t in unresolved_new],
        },
        "reviews": {
            "total": len(review_data),
            "newReviewIds": [r.get("id") for r in new_reviews],
            "latestByReviewer": latest_by_reviewer,
            "effectiveDecision": effective_decision,
            "verifyOpenspec": verify_openspec,
        },
        "merge": {
            "blocked": merge_blocked,
            "hasConflicts": has_merge_conflicts,
            "conflictFiles": merge_conflict_data.get("files", []),
            "conflictAnalysisAvailable": merge_conflict_data.get(
                "analysisAvailable", False
            ),
            "mergeable": mergeable,
            "mergeStateStatus": merge_state,
        },
        "totals": {
            "issueComments": len(issue_comment_data),
            "reviewComments": len(review_comment_data),
            "reviews": len(review_data),
            "unresolvedThreads": len(unresolved_threads),
            "changesRequestedReviews": sum(
                1
                for r in review_data
                if (r.get("state") or "").upper() == "CHANGES_REQUESTED"
            ),
            "approvedReviews": sum(
                1
                for r in review_data
                if (r.get("state") or "").upper() == "APPROVED"
            ),
        },
        "headPushedRecently": head_pushed_recently,
    }

    payload = {
        "repository": repo,
        "pr": pr,
        "checks": {
            "prChecks": raw_pr_checks,
            "prChecksPinned": [
                c for c in raw_pr_checks
                # gh pr checks doesn't expose per-row commit; we emit the same
                # list and rely on commit-pinned arrays below for canonical data
            ],
            "commitCheckRuns": pinned_check_runs,
            "commitStatuses": pinned_statuses,
            "statusCheckRollup": pr.get("statusCheckRollup", []),
            "headSha": head_sha,
        },
        "reviews": review_data,
        "issue_comments": issue_comment_data,
        "review_comments": review_comment_data,
        "review_threads": thread_data,
        "issue_events": event_data,
        "merge_conflicts": merge_conflict_data,
        "summary": summary,
    }

    # --- next state ------------------------------------------------------
    new_state = {
        "pr": pr.get("number"),
        "lastPolledAt": now_iso(),
        "lastHeadSha": head_sha,
        "seenIssueCommentIds": cap_seen_ids(
            list(seen_issue_ids) + [c.get("id") for c in issue_comment_data if c.get("id") is not None]
        ),
        "seenReviewCommentIds": cap_seen_ids(
            list(seen_review_comment_ids) + [c.get("id") for c in review_comment_data if c.get("id") is not None]
        ),
        "seenReviewIds": cap_seen_ids(
            list(seen_review_ids) + [r.get("id") for r in review_data if r.get("id") is not None]
        ),
        "seenReviewThreadIds": cap_seen_ids(
            list(seen_thread_ids) + [t.get("id") for t in thread_data if t.get("id") is not None]
        ),
        "lastVerifyOpenspecLabelAppliedAt": verify_openspec.get("lastLabelAppliedAt"),
        "lastVerifyOpenspecLabelRemovedAt": verify_openspec.get("lastLabelRemovedAt"),
    }

    return payload, new_state


# ---------------------------------------------------------------------------
# Top-level fetch + run
# ---------------------------------------------------------------------------


def fetch_all(pr_arg: str) -> dict[str, Any]:
    """Fetch every piece of remote state once. Raises TransientGhError."""

    repo = repo_info()
    pr = pr_view(pr_arg)
    if not pr or not pr.get("number"):
        raise TransientGhError(
            json.dumps(
                {
                    "error": "pr view returned no data",
                    "pr": pr_arg,
                    "transient": True,
                }
            )
        )

    number = int(pr["number"])
    head_sha = pr.get("headRefOid") or ""

    return {
        "repo": repo,
        "pr": pr,
        "pr_check_data": pr_checks(pr_arg),
        "commit_check_run_data": commit_check_runs(repo["owner"], repo["name"], head_sha),
        "commit_status_data": commit_combined_status(repo["owner"], repo["name"], head_sha),
        "issue_comment_data": issue_comments(repo["owner"], repo["name"], number),
        "review_comment_data": review_comments(repo["owner"], repo["name"], number),
        "review_data": reviews(repo["owner"], repo["name"], number),
        "thread_data": review_threads(repo["owner"], repo["name"], number),
        "event_data": issue_events(repo["owner"], repo["name"], number),
        "merge_conflict_data": merge_conflicts(pr),
    }


def run_once(args: argparse.Namespace, state_path: Optional[str]) -> tuple[dict[str, Any], dict[str, Any]]:
    raw = fetch_all(args.pr)
    state = load_state(state_path)
    payload, new_state = compute_payload(
        **raw,
        state=state,
        head_sha_override=args.head_sha,
        since_override=args.since,
    )
    save_state(state_path, new_state)
    return payload, new_state


# ---------------------------------------------------------------------------
# CLI
# ---------------------------------------------------------------------------


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Return one JSON payload describing actionable PR state."
    )
    parser.add_argument("pr", help="Pull request number, URL, or branch accepted by gh")
    parser.add_argument(
        "--state-file",
        default=None,
        help="Path to the persistent state file. Defaults to <git-dir>/.pr-monitor-<pr>.json.",
    )
    parser.add_argument(
        "--no-state",
        action="store_true",
        help="Disable the state file entirely (every comment counts as new).",
    )
    parser.add_argument(
        "--since",
        default=None,
        help="ISO8601 timestamp; overrides the state file's lastPolledAt for new-since detection.",
    )
    parser.add_argument(
        "--head-sha",
        default=None,
        help="Override the SHA used to pin checks (defaults to pr.headRefOid).",
    )
    parser.add_argument(
        "--watch",
        action="store_true",
        help="Poll until actionable state appears or --max-duration elapses.",
    )
    parser.add_argument(
        "--interval",
        type=int,
        default=DEFAULT_INTERVAL_SECONDS,
        help=f"Seconds between polls in --watch mode (default {DEFAULT_INTERVAL_SECONDS}).",
    )
    parser.add_argument(
        "--max-duration",
        type=int,
        default=DEFAULT_MAX_DURATION_SECONDS,
        help=f"Maximum total seconds in --watch mode (default {DEFAULT_MAX_DURATION_SECONDS}).",
    )
    return parser.parse_args(argv)


def resolve_state_path(args: argparse.Namespace) -> Optional[str]:
    if args.no_state:
        return None
    if args.state_file:
        return args.state_file
    # We need the PR number to construct the default. Defer until we have it.
    return None  # filled in after first fetch by main()


def emit_transient(exc: TransientGhError) -> None:
    try:
        body = json.loads(str(exc))
    except json.JSONDecodeError:
        body = {"error": str(exc), "transient": True}
    body.setdefault("transient", True)
    print(json.dumps(body, indent=2, sort_keys=True))


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv)

    state_path = resolve_state_path(args)

    if args.watch:
        return _run_watch(args, state_path)
    return _run_single(args, state_path)


def _ensure_state_path(args: argparse.Namespace, state_path: Optional[str], pr_number: int) -> Optional[str]:
    if args.no_state:
        return None
    if state_path:
        return state_path
    return default_state_file_path(pr_number)


def _run_single(args: argparse.Namespace, state_path: Optional[str]) -> int:
    try:
        raw = fetch_all(args.pr)
    except TransientGhError as exc:
        emit_transient(exc)
        return EXIT_TRANSIENT

    state_path = _ensure_state_path(args, state_path, int(raw["pr"]["number"]))
    state = load_state(state_path)
    payload, new_state = compute_payload(
        **raw,
        state=state,
        head_sha_override=args.head_sha,
        since_override=args.since,
    )
    save_state(state_path, new_state)
    print(json.dumps(payload, indent=2, sort_keys=True))
    return EXIT_OK


def _run_watch(args: argparse.Namespace, state_path: Optional[str]) -> int:
    started = time.monotonic()
    tick = 0
    last_payload: Optional[dict[str, Any]] = None
    while True:
        tick += 1
        try:
            raw = fetch_all(args.pr)
            state_path_resolved = _ensure_state_path(args, state_path, int(raw["pr"]["number"]))
            state = load_state(state_path_resolved)
            payload, new_state = compute_payload(
                **raw,
                state=state,
                head_sha_override=args.head_sha,
                since_override=args.since,
            )
            save_state(state_path_resolved, new_state)
            last_payload = payload
            tick_line = json.dumps(
                {
                    "tick": tick,
                    "ts": now_iso(),
                    "summary": payload["summary"],
                },
                sort_keys=True,
            )
            print(tick_line, flush=True)
            if payload["summary"]["hasActionable"]:
                print(
                    json.dumps(
                        {"final": True, "outcome": "actionable", "payload": payload},
                        sort_keys=True,
                    ),
                    flush=True,
                )
                return EXIT_OK
        except TransientGhError as exc:
            try:
                err = json.loads(str(exc))
            except json.JSONDecodeError:
                err = {"error": str(exc), "transient": True}
            print(
                json.dumps(
                    {"tick": tick, "ts": now_iso(), "transient": True, "error": err},
                    sort_keys=True,
                ),
                flush=True,
            )
            # On transient errors, sleep and try again until max-duration.
        elapsed = time.monotonic() - started
        if elapsed >= args.max_duration:
            print(
                json.dumps(
                    {
                        "final": True,
                        "outcome": "timeout",
                        "elapsedSeconds": int(elapsed),
                        "lastPayload": last_payload,
                    },
                    sort_keys=True,
                ),
                flush=True,
            )
            return EXIT_TIMEOUT
        time.sleep(max(1, args.interval))


if __name__ == "__main__":
    sys.exit(main())
