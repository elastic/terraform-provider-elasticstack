## Context

The repository already has `code-factory` issue intake automation that turns trusted labeled issues into implementation pull requests. It also uses OpenSpec as the normal entry point for feature and workflow changes, but there is no equivalent automation that turns a labeled issue into a reviewable OpenSpec change proposal.

`change-factory` should stay close to the proven `code-factory` intake shape while changing the agent contract from implementation to proposal authoring. The triggering issue remains the source of truth, but the output is an OpenSpec change under `openspec/changes/<id>/` and a linked pull request, not provider code.

## Goals / Non-Goals

**Goals:**

- Create a label-gated issue workflow for trusted maintainers to request OpenSpec change proposals.
- Reuse deterministic gating patterns for trigger qualification, actor trust, and duplicate linked pull request suppression.
- Keep the agent focused on OpenSpec artifact creation: proposal, design, tasks, and delta specs.
- Bootstrap only the tooling required for OpenSpec authoring and workflow validation.
- Produce exactly one linked pull request per triggering issue.

**Non-Goals:**

- Implement the requested provider/workflow behavior inside the same run.
- Provision the Elastic Stack or run acceptance tests.
- Add a GitHub-comment or GitHub-Discussion exploration loop in the first version.
- Replace existing manual OpenSpec proposal authoring workflows.

## Decisions

### Use a dedicated `change-factory` issue label and branch namespace

The workflow should trigger from `issues.opened` and `issues.labeled` events using a dedicated `change-factory` label. It should create proposal branches under `change-factory/issue-<issue-number>`.

Alternatives considered:

- Reuse `code-factory`: rejected because implementation and proposal-generation contracts have different toolchains, outputs, and guardrails.
- Use a sub-label such as `code-factory:proposal`: rejected because the behavior is distinct enough to deserve separate duplicate detection, labeling, and safe-output policy.

### Keep deterministic pre-activation before agent reasoning

The workflow should adapt the existing deterministic gates from `code-factory`: event qualification, actor trust, duplicate linked pull request detection, and a consolidated gate reason. This keeps permission and idempotency decisions outside the model prompt.

Alternatives considered:

- Let the agent inspect existing pull requests and decide whether to continue: rejected because duplicate suppression should be stable across reruns and independent of model behavior.

### Generate a pull request containing only OpenSpec change artifacts

The agent should derive an OpenSpec change id from the issue title/body, create or update `openspec/changes/<id>/`, validate the artifacts, and open a single pull request labeled `change-factory`. The prompt should explicitly prohibit implementation changes unless needed for repository-authored workflow metadata in a future extension.

Alternatives considered:

- Create a comment with proposed requirements instead of a PR: rejected for v1 because PR review provides normal diff review, CI, and `verify-openspec` follow-up.
- Create canonical specs directly under `openspec/specs/`: rejected because new work should flow through active changes.

### Bootstrap OpenSpec tooling without Elastic Stack setup

The workflow should configure Node from `package.json` and install repository npm dependencies so the OpenSpec CLI is available. It should not start the Elastic Stack, create ES API keys, set up Fleet, or run acceptance tests. Go setup should be included only if the chosen validation/generation path requires repository Go tooling, such as workflow-source compilation checks.

Alternatives considered:

- Run the full `code-factory` environment: rejected because it is slow and gives the agent unnecessary access to runtime services.
- Avoid all repository setup and rely on global tooling: rejected because OpenSpec is pinned in `package.json`.

### Defer exploration loops

If issue context is too ambiguous to produce a coherent proposal, the v1 workflow should require exactly one `add-comment` on the triggering issue listing missing facts, then `noop` with a brief completion note — not `noop` alone. Interactive exploration through further issue comments or GitHub Discussions should be a later, separately specified workflow/state.

Alternatives considered:

- Drive `/openspec-explore` through issue comments immediately: rejected for v1 because it requires state tracking, bot-loop prevention, comment trust policy, and clear rules for which comments are authoritative.

## Risks / Trade-offs

- Ambiguous issues may produce weak proposals → Prefer one clarifying `add-comment` plus `noop` over speculative artifacts or `noop`-only completion when core scope is unclear.
- Proposal quality depends on issue context → The prompt should require preserving assumptions and open questions in the generated artifacts.
- Duplicate detection can miss manually created PRs without canonical metadata → Require deterministic branch, label, and explicit issue linkage in the PR body.
- Omitting Go setup may break workflow-generation validation if the implementation uses `make check-workflows` → Choose validation commands during implementation based on the final workflow source path, and include Go only when needed.
