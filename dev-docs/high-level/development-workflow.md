# Development workflow

For full contributor guidance (setup, PR expectations), see [`contributing.md`](./contributing.md).

## Typical change loop

1. Prepare an OpenSpec proposal before writing provider code.
   Start with `openspec-explore` when the problem or scope is still fuzzy and you want to investigate the codebase, compare approaches, or clarify requirements without implementing yet.
   Example: "Use `openspec-explore` to think through how dashboard schema alignment should work before we formalize the change."
   Use `openspec-propose` when you are ready to generate an implementation-ready change in one pass. It creates the change plus the key artifacts such as `proposal.md`, `design.md`, and `tasks.md`.
   Example: "Use `openspec-propose` to create `dashboard-api-schema-alignment` with proposal, design, and tasks."
   Use `openspec-new-change` when you want to scaffold the change first and then create artifacts incrementally.
   Example: "Use `openspec-new-change` to start `dashboard-api-schema-alignment` and show me the first artifact template."
   The goal of this step is an approved change under `openspec/changes/<change-name>/` that is ready to review.

2. Open a proposal PR.
   Send the OpenSpec artifacts for review before implementation. In most cases this PR should contain only the proposal artifacts under `openspec/changes/<change-name>/`.
   This is the point to resolve scope, requirements, and design questions before code lands.

3. Implement the approved proposal.
   Use `openspec-apply-change` to read the change context, work through the task list, make the code changes, and update task checkboxes as work completes.
   Example: "Use `openspec-apply-change` for `dashboard-api-schema-alignment` and implement the remaining tasks."
   Use `openspec-continue-change` if the change is not fully apply-ready yet, or if review/implementation feedback means you need to create the next artifact before continuing.
   Example: "Use `openspec-continue-change` for `dashboard-api-schema-alignment` and create the next required artifact."
   Use `openspec-implementation-loop` when you want a more automated end-to-end loop around a single approved change, including implementation, local review, push, and optional PR handling.
   Example: "Use `openspec-implementation-loop` for `dashboard-api-schema-alignment` in PR mode."
   During implementation, add or update acceptance tests for new behavior and bug fixes. For bugs, verify the new test fails first so it reproduces the original issue.
   Make small, reviewable changes. Keep generated artifacts up to date (docs and generated clients when applicable). Run the narrowest tests that prove correctness, then broaden as appropriate.
   The System User resource (see `internal/elasticsearch/security/system_user` referenced from [`coding-standards.md`](./coding-standards.md)) is the canonical example for new resources. Follow it.

4. Verify the implementation against the spec.
   Run `openspec-verify-change` to check completeness, correctness, and coherence against the approved change artifacts.
   Example: "Use `openspec-verify-change` for `dashboard-api-schema-alignment` and report any gaps before we open the implementation PR."
   Address any verification findings before moving on.

5. Open the implementation PR.
   Once the change is implemented and verified, open a separate PR for the provider code and any related generated artifacts.
   Link back to the approved proposal change so reviewers can compare the implementation with the agreed requirements.

## Common make targets

The canonical list is the root `Makefile`, but the usual ones are:

- `make lint`
- `make test`
- `make testacc` (requires Docker and `TF_ACC=1`)
- `make docs-generate`
