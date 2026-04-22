#!/usr/bin/env python3

import argparse
import json
import subprocess
import sys
from typing import Any


REVIEW_THREADS_QUERY = """
query($owner: String!, $name: String!, $number: Int!, $after: String) {
  repository(owner: $owner, name: $name) {
    pullRequest(number: $number) {
      reviewThreads(first: 100, after: $after) {
        nodes {
          id
          isResolved
          isOutdated
          comments(first: 100) {
            nodes {
              author {
                login
              }
              body
              createdAt
              path
              outdated
              url
            }
          }
        }
        pageInfo {
          endCursor
          hasNextPage
        }
      }
    }
  }
}
""".strip()


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description=(
            "Collect the current PR check state plus review activity in one "
            "deterministic JSON payload."
        )
    )
    parser.add_argument("pr_ref", help="PR number, URL, or branch name")
    parser.add_argument("--repo", help="GitHub repo in owner/name format")
    return parser.parse_args()


def run_gh(args: list[str]) -> Any:
    result = subprocess.run(
        ["gh", *args],
        capture_output=True,
        text=True,
        check=False,
    )
    if result.returncode != 0:
        if result.stdout:
            sys.stderr.write(result.stdout)
        if result.stderr:
            sys.stderr.write(result.stderr)
        raise SystemExit(result.returncode)
    return json.loads(result.stdout)


def sort_items(items: list[dict[str, Any]], *keys: str) -> list[dict[str, Any]]:
    return sorted(
        items,
        key=lambda item: tuple(str(item.get(key) or "") for key in keys),
    )


def fetch_review_threads(repo_args: list[str], owner: str, repo_name: str, pr_number: int) -> list[dict[str, Any]]:
    threads: list[dict[str, Any]] = []
    after: str | None = None

    while True:
        gh_args = [
            "api",
            *repo_args,
            "graphql",
            "-F",
            f"owner={owner}",
            "-F",
            f"name={repo_name}",
            "-F",
            f"number={pr_number}",
            "-f",
            f"query={REVIEW_THREADS_QUERY}",
        ]
        if after:
            gh_args.extend(["-F", f"after={after}"])

        payload = run_gh(gh_args)
        review_threads = (
            payload.get("data", {})
            .get("repository", {})
            .get("pullRequest", {})
            .get("reviewThreads", {})
        )
        threads.extend(review_threads.get("nodes", []))

        page_info = review_threads.get("pageInfo", {})
        if not page_info.get("hasNextPage"):
            break
        after = page_info.get("endCursor")
        if not after:
            break

    return threads


def main() -> int:
    args = parse_args()
    repo_args = ["--repo", args.repo] if args.repo else []

    pr = run_gh(
        [
            "pr",
            "view",
            args.pr_ref,
            *repo_args,
            "--json",
            "number,url,headRefOid,reviewDecision,isDraft,state,title",
        ]
    )
    repo = run_gh(["repo", "view", *repo_args, "--json", "owner,name"])

    owner = repo["owner"]["login"]
    repo_name = repo["name"]
    pr_number = pr["number"]

    checks = run_gh(
        [
            "pr",
            "checks",
            args.pr_ref,
            *repo_args,
            "--json",
            "bucket,completedAt,description,event,link,name,startedAt,state,workflow",
        ]
    )
    issue_comments = run_gh(
        [
            "api",
            *repo_args,
            f"repos/{owner}/{repo_name}/issues/{pr_number}/comments?per_page=100",
        ]
    )
    reviews = run_gh(
        [
            "api",
            *repo_args,
            f"repos/{owner}/{repo_name}/pulls/{pr_number}/reviews?per_page=100",
        ]
    )
    review_comments = run_gh(
        [
            "api",
            *repo_args,
            f"repos/{owner}/{repo_name}/pulls/{pr_number}/comments?per_page=100",
        ]
    )
    thread_nodes = fetch_review_threads(repo_args, owner, repo_name, pr_number)

    normalized_threads = []
    for thread in thread_nodes:
        comments = thread.get("comments", {}).get("nodes", [])
        normalized_threads.append(
            {
                "id": thread.get("id"),
                "isResolved": bool(thread.get("isResolved")),
                "isOutdated": bool(thread.get("isOutdated")),
                "comments": sort_items(comments, "createdAt", "url"),
            }
        )

    checks = sort_items(checks, "workflow", "name", "startedAt", "link")
    issue_comments = sort_items(issue_comments, "created_at", "html_url")
    reviews = sort_items(reviews, "submitted_at", "html_url")
    review_comments = sort_items(review_comments, "created_at", "html_url")

    checks_by_bucket: dict[str, int] = {}
    for check in checks:
        bucket = str(check.get("bucket") or "unknown")
        checks_by_bucket[bucket] = checks_by_bucket.get(bucket, 0) + 1

    reviews_by_state: dict[str, int] = {}
    for review in reviews:
        state = str(review.get("state") or "UNKNOWN").upper()
        reviews_by_state[state] = reviews_by_state.get(state, 0) + 1

    unresolved_threads = [
        thread for thread in normalized_threads if not thread["isResolved"]
    ]
    unresolved_review_comments = sum(
        len(thread["comments"]) for thread in unresolved_threads
    )

    payload = {
        "summary": {
            "pr": {
                "number": pr["number"],
                "url": pr["url"],
                "title": pr.get("title"),
                "state": pr.get("state"),
                "isDraft": pr.get("isDraft"),
                "headRefOid": pr.get("headRefOid"),
                "reviewDecision": pr.get("reviewDecision"),
            },
            "checks": {
                "total": len(checks),
                "byBucket": checks_by_bucket,
            },
            "reviews": {
                "total": len(reviews),
                "byState": reviews_by_state,
            },
            "issueComments": {
                "total": len(issue_comments),
            },
            "reviewComments": {
                "total": len(review_comments),
                "unresolvedThreads": len(unresolved_threads),
                "unresolvedThreadComments": unresolved_review_comments,
            },
        },
        "checks": checks,
        "issueComments": issue_comments,
        "reviews": reviews,
        "reviewComments": review_comments,
        "reviewThreads": normalized_threads,
    }

    json.dump(payload, sys.stdout, indent=2, sort_keys=True)
    sys.stdout.write("\n")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
