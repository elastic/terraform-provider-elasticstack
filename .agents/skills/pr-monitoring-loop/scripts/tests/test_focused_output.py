"""Tests for the focused output formatter.

The focused formatter strips raw data arrays and surfaces only actionable
decision data. These tests verify the shape of the default output.
"""

from __future__ import annotations

import importlib.util
from pathlib import Path

import pytest


SCRIPT = Path(__file__).resolve().parent.parent / "check-pr-state.py"


@pytest.fixture(scope="session")
def cps():
    spec = importlib.util.spec_from_file_location("check_pr_state", SCRIPT)
    assert spec and spec.loader
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


HEAD_SHA = "headsha1234567890"


def _pr(head_sha=HEAD_SHA, **overrides):
    base = {
        "number": 42,
        "url": "https://example.test/pulls/42",
        "state": "OPEN",
        "isDraft": False,
        "headRefName": "feature/x",
        "headRefOid": head_sha,
        "baseRefName": "main",
        "baseRefOid": "basesha",
        "reviewDecision": "REVIEW_REQUIRED",
        "mergeable": "MERGEABLE",
        "mergeStateStatus": "CLEAN",
        "labels": [],
        "statusCheckRollup": [],
        "title": "test PR",
        "updatedAt": "2026-05-07T00:00:00Z",
        "author": {"login": "human-dev"},
    }
    base.update(overrides)
    return base


def _comment(cid, login, body="hi", created="2026-05-07T01:00:00Z"):
    return {
        "id": cid,
        "user": {"login": login, "type": "Bot" if login.endswith("[bot]") else "User"},
        "body": body,
        "created_at": created,
        "updated_at": created,
    }


def _review_comment(cid, login, body="line note", created="2026-05-07T01:00:00Z"):
    return {
        "id": cid,
        "user": {"login": login, "type": "Bot" if login.endswith("[bot]") else "User"},
        "body": body,
        "path": "src/foo.go",
        "created_at": created,
        "updated_at": created,
    }


def _review(rid, login, state, commit_id=HEAD_SHA, submitted="2026-05-07T01:00:00Z", body=""):
    return {
        "id": rid,
        "user": {"login": login, "type": "Bot" if login.endswith("[bot]") else "User"},
        "state": state,
        "commit_id": commit_id,
        "submitted_at": submitted,
        "body": body,
    }


def _thread(tid, comments, resolved=False, outdated=False):
    return {
        "id": tid,
        "isResolved": resolved,
        "isOutdated": outdated,
        "path": "src/foo.go",
        "line": 1,
        "originalLine": 1,
        "comments": {"nodes": comments},
    }


def _thread_comment(node_id, db_id, login, created="2026-05-07T01:00:00Z"):
    return {
        "id": node_id,
        "databaseId": db_id,
        "author": {"login": login},
        "body": "please change",
        "createdAt": created,
        "updatedAt": created,
        "diffHunk": "@@",
        "path": "src/foo.go",
        "line": 1,
        "originalLine": 1,
        "outdated": False,
        "url": "https://example.test/c",
    }


def _empty_kwargs():
    return dict(
        repo={"owner": "o", "name": "r"},
        pr=_pr(),
        pr_check_data=[],
        commit_check_run_data=[],
        commit_status_data={},
        review_data=[],
        issue_comment_data=[],
        review_comment_data=[],
        thread_data=[],
        event_data=[],
        merge_conflict_data={"hasConflicts": False, "files": [], "analysisAvailable": True},
        state={},
    )


# ---------------------------------------------------------------------------
# Focused output shape
# ---------------------------------------------------------------------------


def test_focused_output_contains_pr_metadata(cps):
    kwargs = _empty_kwargs()
    payload, _ = cps.compute_payload(**kwargs)
    focused = cps.format_focused(payload)
    assert focused["pr"]["number"] == 42
    assert focused["pr"]["url"] == "https://example.test/pulls/42"
    assert focused["pr"]["title"] == "test PR"
    assert focused["pr"]["headRefOid"] == HEAD_SHA


def test_focused_output_omits_old_data(cps):
    kwargs = _empty_kwargs()
    kwargs["issue_comment_data"] = [_comment(1, "alice")]
    kwargs["review_comment_data"] = [_review_comment(10, "bob")]
    payload, _ = cps.compute_payload(**kwargs)
    focused = cps.format_focused(payload)
    assert "repository" not in focused
    assert "issue_events" not in focused
    assert "issue_comments" not in focused
    assert "review_comments" not in focused
    assert "review_threads" not in focused
    assert "merge_conflicts" not in focused
    assert "prChecks" not in focused
    assert "commitCheckRuns" not in focused
    assert "commitStatuses" not in focused


def test_focused_output_failed_checks_structured(cps):
    kwargs = _empty_kwargs()
    kwargs["commit_check_run_data"] = [
        {"name": "lint", "status": "completed", "conclusion": "failure", "html_url": "https://runs/1"},
    ]
    payload, _ = cps.compute_payload(**kwargs)
    focused = cps.format_focused(payload)
    assert focused["checks"]["failedChecks"] == [{"name": "lint", "url": "https://runs/1"}]
    assert focused["checks"]["failedNames"] == ["lint"]


def test_focused_output_includes_new_issue_comments_bodies(cps):
    kwargs = _empty_kwargs()
    kwargs["issue_comment_data"] = [
        _comment(1, "alice", body="first"),
        _comment(2, "bob", body="second"),
    ]
    payload, _ = cps.compute_payload(**kwargs)
    focused = cps.format_focused(payload)
    assert len(focused["comments"]["newIssueComments"]) == 2
    assert focused["comments"]["newIssueComments"][0]["author"] == "alice"
    assert focused["comments"]["newIssueComments"][0]["body"] == "first"
    assert focused["comments"]["newIssueComments"][1]["id"] == 2


def test_focused_output_includes_thread_details(cps):
    thread = _thread(
        "thread1",
        [_thread_comment("c1node", 555, "human-dev")],
    )
    kwargs = _empty_kwargs()
    kwargs["thread_data"] = [thread]
    payload, _ = cps.compute_payload(**kwargs)
    focused = cps.format_focused(payload)
    detail = focused["threadDetails"]["thread1"]
    assert detail["path"] == "src/foo.go"
    assert detail["line"] == 1
    assert detail["resolved"] is False
    assert detail["outdated"] is False
    assert len(detail["comments"]) == 1
    assert detail["comments"][0]["author"] == "human-dev"
    assert detail["comments"][0]["body"] == "please change"
    assert detail["comments"][0]["databaseId"] == 555


def test_focused_output_does_not_duplicate_thread_comments_in_new_review_comments(cps):
    thread = _thread(
        "thread1",
        [_thread_comment("c1node", 555, "human-dev")],
    )
    kwargs = _empty_kwargs()
    kwargs["thread_data"] = [thread]
    kwargs["review_comment_data"] = [_review_comment(555, "human-dev", body="please change")]
    payload, _ = cps.compute_payload(**kwargs)
    focused = cps.format_focused(payload)
    # The comment is in threadDetails, so it should NOT appear standalone
    assert focused["comments"]["newReviewComments"] == []


def test_focused_output_requires_openspec_true_when_ready(cps):
    kwargs = _empty_kwargs()
    payload, _ = cps.compute_payload(**kwargs)
    focused = cps.format_focused(payload)
    assert focused["verifyOpenspec"]["runState"] == "none"
    assert focused["verifyOpenspec"]["requiresOpenspecVerification"] is True


def test_focused_output_requires_openspec_false_when_checks_pending(cps):
    kwargs = _empty_kwargs()
    kwargs["commit_check_run_data"] = [
        {"name": "build", "status": "in_progress", "conclusion": None},
    ]
    payload, _ = cps.compute_payload(**kwargs)
    focused = cps.format_focused(payload)
    assert focused["verifyOpenspec"]["requiresOpenspecVerification"] is False


def test_focused_output_requires_openspec_false_when_changes_requested(cps):
    kwargs = _empty_kwargs()
    kwargs["review_data"] = [_review(1, "alice", "CHANGES_REQUESTED")]
    payload, _ = cps.compute_payload(**kwargs)
    focused = cps.format_focused(payload)
    assert focused["verifyOpenspec"]["requiresOpenspecVerification"] is False


def test_focused_output_requires_openspec_false_when_approved(cps):
    verify_report_body = (
        "## Verification Report: `some-change`\n\n"
        "Generated by OpenSpec verify (label) for issue #42"
    )
    kwargs = _empty_kwargs()
    kwargs["review_data"] = [
        _review(7, "github-actions[bot]", "APPROVED", body=verify_report_body),
    ]
    payload, _ = cps.compute_payload(**kwargs)
    focused = cps.format_focused(payload)
    assert focused["verifyOpenspec"]["runState"] == "approved"
    assert focused["verifyOpenspec"]["requiresOpenspecVerification"] is False


def test_focused_output_new_reviews_include_body_for_changes_requested(cps):
    kwargs = _empty_kwargs()
    kwargs["review_data"] = [
        _review(1, "alice", "CHANGES_REQUESTED", body="fix the thing"),
        _review(2, "bob", "APPROVED"),
    ]
    payload, _ = cps.compute_payload(**kwargs)
    focused = cps.format_focused(payload)
    new_reviews = focused["reviews"]["newReviews"]
    assert len(new_reviews) == 2
    alice = [r for r in new_reviews if r["author"] == "alice"][0]
    bob = [r for r in new_reviews if r["author"] == "bob"][0]
    assert alice["body"] == "fix the thing"
    assert "body" not in bob


def test_focused_output_total_issue_comments_naming(cps):
    kwargs = _empty_kwargs()
    kwargs["issue_comment_data"] = [_comment(1, "alice")]
    payload, _ = cps.compute_payload(**kwargs)
    focused = cps.format_focused(payload)
    assert focused["comments"]["totalIssueComments"] == 1
    assert focused["comments"]["totalReviewComments"] == 0


def test_full_payload_flag_restores_all_data(cps, monkeypatch, capsys):
    def fake_fetch_all(pr_arg):
        return {
            "repo": {"owner": "o", "name": "r"},
            "pr": _pr(),
            "pr_check_data": [],
            "commit_check_run_data": [],
            "commit_status_data": {},
            "issue_comment_data": [],
            "review_comment_data": [],
            "review_data": [],
            "thread_data": [],
            "event_data": [],
            "merge_conflict_data": {"hasConflicts": False, "files": [], "analysisAvailable": True},
        }

    monkeypatch.setattr(cps, "fetch_all", fake_fetch_all)
    rc = cps.main(["42", "--full-payload"])
    assert rc == cps.EXIT_OK
    out = capsys.readouterr().out
    import json
    data = json.loads(out)
    assert "summary" in data
    assert "repository" in data
    assert "issue_comments" in data
