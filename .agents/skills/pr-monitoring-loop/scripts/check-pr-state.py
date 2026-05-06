#!/usr/bin/env python3
"""Collect deterministic GitHub PR state for watcher subagents."""

from __future__ import annotations

import argparse
import json
import subprocess
import sys
from typing import Any


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


def run_gh(args: list[str], *, allow_failure: bool = False) -> str:
    result = subprocess.run(
        ["gh", *args],
        check=False,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
    )

    if result.returncode != 0 and not allow_failure:
        raise SystemExit(
            json.dumps(
                {
                    "error": "gh command failed",
                    "command": ["gh", *args],
                    "stderr": result.stderr.strip(),
                    "returncode": result.returncode,
                },
                indent=2,
            )
        )

    if result.returncode != 0:
        return ""

    return result.stdout


def run_git(args: list[str], *, allow_failure: bool = False) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        ["git", *args],
        check=not allow_failure,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
    )


def gh_json(args: list[str], *, default: Any, allow_failure: bool = False) -> Any:
    output = run_gh(args, allow_failure=allow_failure)
    if not output.strip():
        return default

    try:
        return json.loads(output)
    except json.JSONDecodeError as exc:
        if allow_failure:
            return default
        raise SystemExit(
            json.dumps(
                {
                    "error": "gh command returned invalid JSON",
                    "command": ["gh", *args],
                    "message": str(exc),
                    "stdout": output,
                },
                indent=2,
            )
        )


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


def issue_comments(owner: str, repo: str, number: int) -> list[dict[str, Any]]:
    return gh_json(
        [
            "api",
            f"repos/{owner}/{repo}/issues/{number}/comments",
            "--paginate",
        ],
        default=[],
        allow_failure=True,
    )


def review_comments(owner: str, repo: str, number: int) -> list[dict[str, Any]]:
    return gh_json(
        [
            "api",
            f"repos/{owner}/{repo}/pulls/{number}/comments",
            "--paginate",
        ],
        default=[],
        allow_failure=True,
    )


def reviews(owner: str, repo: str, number: int) -> list[dict[str, Any]]:
    return gh_json(
        [
            "api",
            f"repos/{owner}/{repo}/pulls/{number}/reviews",
            "--paginate",
        ],
        default=[],
        allow_failure=True,
    )


def review_threads(owner: str, repo: str, number: int) -> list[dict[str, Any]]:
    query = """
query($owner: String!, $repo: String!, $number: Int!, $cursor: String) {
  repository(owner: $owner, name: $repo) {
    pullRequest(number: $number) {
      reviewThreads(first: 100, after: $cursor) {
        pageInfo {
          hasNextPage
          endCursor
        }
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
              author {
                login
              }
              body
              createdAt
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
        args = [
            "api",
            "graphql",
            "-f",
            f"owner={owner}",
            "-f",
            f"repo={repo}",
            "-F",
            f"number={number}",
            "-f",
            f"query={query}",
        ]
        if cursor:
            args.extend(["-f", f"cursor={cursor}"])

        data = gh_json(args, default={}, allow_failure=True)
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
            "fetch",
            "--quiet",
            "origin",
            f"+refs/heads/{base_ref}:{local_base_ref}",
            f"+refs/pull/{number}/head:{local_pr_ref}",
        ]
    )

    base_sha = run_git(["rev-parse", local_base_ref]).stdout.strip()
    head_sha = run_git(["rev-parse", local_pr_ref]).stdout.strip()

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
        result = run_git(
            ["merge-tree", "--write-tree", refs["base"], refs["head"]],
            allow_failure=True,
        )
    except (KeyError, subprocess.CalledProcessError, FileNotFoundError) as exc:
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


def summarize(
    pr: dict[str, Any],
    checks: list[dict[str, Any]],
    review_data: list[dict[str, Any]],
    issue_comment_data: list[dict[str, Any]],
    review_comment_data: list[dict[str, Any]],
    thread_data: list[dict[str, Any]],
    merge_conflict_data: dict[str, Any],
) -> dict[str, Any]:
    normalized_checks = [normalize_check(check) for check in checks]
    failed_checks = [check for check in normalized_checks if check["derived"]["failed"]]
    pending_checks = [check for check in normalized_checks if check["derived"]["pending"]]
    unresolved_threads = [
        thread
        for thread in thread_data
        if not thread.get("isResolved") and not thread.get("isOutdated")
    ]

    changes_requested = [
        review
        for review in review_data
        if (review.get("state") or "").upper() == "CHANGES_REQUESTED"
    ]
    approvals = [
        review
        for review in review_data
        if (review.get("state") or "").upper() == "APPROVED"
    ]

    merge_state = pr.get("mergeStateStatus")
    mergeable = pr.get("mergeable")
    has_merge_conflicts = bool(merge_conflict_data.get("hasConflicts"))
    merge_blocked = has_merge_conflicts or merge_state in {
        "BEHIND",
        "BLOCKED",
        "UNKNOWN",
        "UNSTABLE",
    }

    actionable: list[str] = []
    if failed_checks:
        actionable.append("failed_checks")
    if issue_comment_data:
        actionable.append("issue_comments")
    if review_comment_data:
        actionable.append("review_comments")
    if unresolved_threads:
        actionable.append("unresolved_review_threads")
    if changes_requested:
        actionable.append("changes_requested")
    if has_merge_conflicts:
        actionable.append("merge_conflicts")
    elif merge_blocked:
        actionable.append("merge_or_branch_state")

    return {
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
            "total": len(normalized_checks),
            "failed": len(failed_checks),
            "pending": len(pending_checks),
            "passed": len(
                [check for check in normalized_checks if check["derived"]["passed"]]
            ),
            "failedNames": [check.get("name") for check in failed_checks],
            "pendingNames": [check.get("name") for check in pending_checks],
        },
        "reviews": {
            "total": len(review_data),
            "changesRequested": len(changes_requested),
            "approved": len(approvals),
            "approvedByLogins": sorted(
                {
                    (review.get("user") or {}).get("login")
                    for review in approvals
                    if (review.get("user") or {}).get("login")
                }
            ),
        },
        "comments": {
            "issueComments": len(issue_comment_data),
            "reviewComments": len(review_comment_data),
            "unresolvedReviewThreads": len(unresolved_threads),
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
    }


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Return one JSON payload describing actionable PR state."
    )
    parser.add_argument("pr", help="Pull request number, URL, or branch accepted by gh")
    args = parser.parse_args()

    repo = repo_info()
    pr = pr_view(args.pr)
    number = int(pr["number"])

    checks = pr_checks(args.pr)
    issue_comment_data = issue_comments(repo["owner"], repo["name"], number)
    review_comment_data = review_comments(repo["owner"], repo["name"], number)
    review_data = reviews(repo["owner"], repo["name"], number)
    thread_data = review_threads(repo["owner"], repo["name"], number)
    merge_conflict_data = merge_conflicts(pr)

    payload = {
        "repository": repo,
        "pr": pr,
        "checks": {
            "prChecks": [normalize_check(check) for check in checks],
            "statusCheckRollup": pr.get("statusCheckRollup", []),
        },
        "reviews": review_data,
        "issue_comments": issue_comment_data,
        "review_comments": review_comment_data,
        "review_threads": thread_data,
        "merge_conflicts": merge_conflict_data,
        "summary": summarize(
            pr,
            checks,
            review_data,
            issue_comment_data,
            review_comment_data,
            thread_data,
            merge_conflict_data,
        ),
    }

    print(json.dumps(payload, indent=2, sort_keys=True))
    return 0


if __name__ == "__main__":
    sys.exit(main())
