## Context

The `schema-coverage-rotation` workflow now asks the agent to run repository-local Go commands from `scripts/schema-coverage-rotation`, but the authored workflow source under `.github/workflows-src/schema-coverage-rotation/workflow.md.tmpl` does not provision the repository toolchain before agent reasoning begins. Unlike `openspec-verify-label`, it does not install Go from `go.mod`, export Go path variables for AWF chroot mode, install Node from `package.json`, run `make setup`, or declare the Go and Node ecosystems in `network.allowed`.

That omission makes the workflow depend on runner-default toolchains and package-network policy. The current failure mode is that the default Go toolchain is too old for the repository's Go commands, but the same gap also leaves Node-backed repository bootstrap undefined for future prompt steps.

## Goals / Non-Goals

**Goals:**
- Provision the schema-coverage rotation workflow with the repository's Go and Node toolchains before agent reasoning begins.
- Export `GOROOT`, `GOPATH`, and `GOMODCACHE` so AWF chroot-mode commands can reuse the configured Go installation and module cache.
- Run `make setup` deterministically so repository-local CLI and dependency setup is complete before the agent executes `go run ./scripts/schema-coverage-rotation ...`.
- Allow the `defaults`, `node`, and `go` ecosystems in the workflow network policy so repository bootstrap and agent-invoked commands can access expected package-manager hosts.
- Keep the bootstrap contract aligned with the existing `openspec-verify-label` workflow where the same repository toolchain pattern is already established.

**Non-Goals:**
- Changing schema-coverage issue-slot gating, memory selection, or issue-creation behavior.
- Moving schema-coverage analysis into deterministic pre-activation steps.
- Adding Terraform CLI setup unless a later implementation detail proves it is actually required for this workflow.
- Changing the repository's declared Go or Node versions outside the workflow setup behavior.

## Decisions

### Mirror the repository bootstrap pattern already used by `openspec-verify-label`

The schema-coverage rotation workflow should add deterministic setup steps before agent reasoning in the same core order used by `openspec-verify-label`: install Go from `go.mod`, export `GOROOT`, `GOPATH`, and `GOMODCACHE`, install Node from `package.json`, then run `make setup`.

This addresses the immediate Go-version failure while also ensuring the workflow uses the repository's declared toolchain and standard bootstrap path rather than whatever the base runner happens to provide.

Alternative considered: add only `actions/setup-go`.
Rejected because the workflow also depends on repository bootstrap via `make setup`, which in turn relies on the Node/OpenSpec toolchain and should be prepared deterministically rather than left implicit.

Alternative considered: tell the agent to install toolchains in prompt instructions.
Rejected because toolchain preparation is deterministic environment setup, not reasoning work, and should complete before the expensive agent path starts.

### Export Go environment variables for AWF chroot mode

The workflow should export `GOROOT`, `GOPATH`, and `GOMODCACHE` to `GITHUB_ENV` immediately after `actions/setup-go`. This matches the existing AWF pattern in `openspec-verify-label` and makes the prepared Go workspace visible to agent-executed commands that run in chroot mode.

Alternative considered: rely on Go defaults after setup-go.
Rejected because the workflow already has a proven repository pattern for AWF Go-path export, and the schema-coverage workflow should use the same handoff contract.

### Declare ecosystem-based network access for bootstrap and agent commands

The workflow should explicitly set `network.allowed` to include `defaults`, `node`, and `go`. This keeps the policy at the repository's package-ecosystem level instead of depending on ad hoc domain allowlists, and it covers both `make setup` and any agent-invoked Node or Go commands that still need package-manager access.

Alternative considered: allow only `go`.
Rejected because the requested repository bootstrap includes `make setup`, which depends on the Node ecosystem as well.

## Risks / Trade-offs

- Broader network policy than today -> Keep the allowlist limited to `defaults`, `node`, and `go` rather than introducing arbitrary domains.
- Slightly longer workflow startup time -> Accept the deterministic setup cost to avoid agent failures and repeated retries caused by missing toolchains.
- Tight alignment with `openspec-verify-label` could drift if that workflow evolves later -> Keep the schema-coverage requirements focused on the observable bootstrap contract instead of a brittle textual copy.

## Migration Plan

1. Update `.github/workflows-src/schema-coverage-rotation/workflow.md.tmpl` to add deterministic Go setup, Go path export, Node setup, `make setup`, and `network.allowed: [defaults, node, go]`.
2. Recompile `.github/workflows/schema-coverage-rotation.md` and `.github/workflows/schema-coverage-rotation.lock.yml` from the authored workflow source.
3. Run the relevant OpenSpec and workflow validation checks to confirm the generated workflow and requirements stay in sync.
