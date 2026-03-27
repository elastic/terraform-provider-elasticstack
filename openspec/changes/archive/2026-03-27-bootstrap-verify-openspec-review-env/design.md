## Context

The `ci-aw-openspec-verification` workflow currently relies on markdown instructions to prepare the review environment, and those instructions only cover `npm ci` plus `npx openspec`. That was sufficient while verification stayed narrowly focused on OpenSpec CLI usage, but it breaks down once agent-driven review work or repo scripts need Go tooling that matches `go.mod` or Terraform tooling that matches normal CI behavior. The repository's `lint` job already defines the expected bootstrap shape: checkout, Node from `package.json` engines, Go from `go.mod`, Terraform CLI, and then repo-native setup commands.

This change should make the review environment runner-independent without turning the verification workflow into a second full lint job. The agent should start from a workspace that already has the required toolchains and repository dependencies available.

## Goals / Non-Goals

**Goals:**
- Make the `verify-openspec` review environment able to run repo-standard OpenSpec, Go, and Terraform commands without depending on whatever toolchain the base runner happens to provide.
- Align the review workflow's bootstrap layers with the `lint` job closely enough that future Go version bumps are picked up automatically from `go.mod`.
- Keep the agent prompt focused on verification behavior instead of ad hoc environment recovery.
- Preserve the existing review, archive, and cleanup behavior outside the bootstrap path.

**Non-Goals:**
- Running the full `lint` job or `make check-lint` as part of review verification.
- Changing review scoring, approval rules, archive behavior, or label cleanup semantics.
- Reworking unrelated deterministic gating logic already covered by other `ci-aw-openspec-verification` changes.

## Decisions

Bootstrap the core toolchains in the workflow before agent reasoning.
The workflow should provision Node using `package.json` engines, Go using `go.mod`, and Terraform CLI in the review job before the agent starts. This mirrors the `lint` job's initial setup while letting the Node version follow the repository's declared engine range, and it removes dependence on runner-default Go versions.

Alternative considered: keep toolchain setup in the prompt.
Rejected because prompt-only setup is easy to drift from CI, harder to audit, and still leaves the run vulnerable to outdated default toolchains before the agent corrects them.

Run `make setup` in the agent workspace after toolchain setup.
The workflow should run `make setup` in the same workspace the agent will use, so `npx openspec` resolves locally and Go commands see downloaded module dependencies through the repository's standard bootstrap path. This keeps the review workflow aligned with existing repo setup conventions without escalating all the way to `make check-lint`.

Alternative considered: run only `npm ci` or a narrower custom subset of setup commands.
Rejected because that fixes OpenSpec CLI availability but does not clearly commit the workflow to the repository's existing bootstrap contract, and it leaves more room for drift from the `make setup` path maintainers already use.

Alternative considered: run `make check-lint`.
Rejected because it couples review bootstrap to full lint execution, adds significant runtime, and changes the purpose of the verification workflow.

Use repository-standard bootstrap commands instead of bespoke installation logic where practical.
The workflow should prefer repo-native setup commands or their direct equivalents from `Makefile` so review setup follows the same maintenance path as CI expectations.

Alternative considered: hard-code independent dependency installation commands in the workflow.
Rejected because it would duplicate bootstrap logic and make future dependency changes easier to miss.

## Risks / Trade-offs

- Extra bootstrap work on each qualifying review run -> Limit setup to the toolchain and dependency steps needed for verification, not the full lint target.
- Review setup may still drift if `lint` changes substantially -> Anchor the workflow changes to the same Node, Go, and Terraform setup pattern used in `.github/workflows/test.yml`.
- `make setup` may install slightly more than verification strictly needs -> Accept modest extra setup cost in exchange for lower drift and fewer ad hoc environment failures.
- Toolchain setup must happen in the agent workspace, not a separate pre-activation job -> Keep dependency installation in the review job so the agent sees the prepared workspace directly.

## Migration Plan

- Update the `ci-aw-openspec-verification` delta spec to require repo-standard runtime provisioning plus `make setup` before agent verification.
- Update `.github/workflows/openspec-verify-label.md` to add the `lint`-like toolchain setup layers and revise repository setup instructions accordingly.
- Recompile `.github/workflows/openspec-verify-label.lock.yml` with `gh aw compile`.
- Validate that the generated workflow still preserves the existing review and cleanup flow while exposing the new bootstrap behavior.

## Open Questions

- None.
