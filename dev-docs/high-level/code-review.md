# Code review

This page is for **maintainers** who use automation alongside normal pull request review.

## What the `verify-openspec` workflow does

When you add the pull request label **`verify-openspec`**, a bot can:

- Check that the PR lines up with **one** in-progress OpenSpec change (the work tracked under `openspec/changes/…`).
- Post a **pull request review** summarizing what it found.
- If it **approves** the review, **archive** that OpenSpec change and **push** the result to the PR branch so the branch matches what you would get after a normal archive—without you running the steps locally.

Nothing runs until someone applies that label. The workflow is only meant for PRs where you explicitly want this verify-and-maybe-archive pass.

## Where the details live

- **Behavior and requirements** (what must be true, when the bot skips work, what “approve” implies):  
  [`openspec/changes/aw-openspec-verification/specs/ci-aw-openspec-verification/spec.md`](../../openspec/changes/aw-openspec-verification/specs/ci-aw-openspec-verification/spec.md)  
  (When that change is archived, the same capability will live under [`openspec/specs/`](../../openspec/specs/) per project process.)

- **How it is implemented** (instructions for the automation, triggers, compilation):  
  - Source: [`.github/workflows/openspec-verify-label.md`](../../.github/workflows/openspec-verify-label.md)  
  - Generated Actions workflow (do not edit by hand): [`.github/workflows/openspec-verify-label.lock.yml`](../../.github/workflows/openspec-verify-label.lock.yml)

Repository admins may need to adjust **GitHub Actions** settings so workflows can open or update pull requests and push to PR branches; exact needs depend on your org. If pushes from the bot do not trigger your usual CI, see [Triggering CI](https://github.github.io/gh-aw/reference/triggering-ci/) in the GitHub Agentic Workflows docs.

For how OpenSpec changes and specs fit into everyday contribution work, see [`openspec-requirements.md`](./openspec-requirements.md).
