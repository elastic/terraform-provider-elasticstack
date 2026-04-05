## Context

The `openspec-verify-label` workflow already prepares the repository with `actions/setup-go` and `make setup`, but the AWF agent runs inside a sandboxed environment with its own firewall policy and chroot-oriented environment handoff. Recent runs show that agent-executed `go test` commands still attempt live requests to `proxy.golang.org`, which means the current setup is not sufficient to reuse the Go dependencies downloaded during bootstrap.

This change touches both the workflow frontmatter and the review-environment contract in the `ci-aw-openspec-verification` capability. A short design is useful because the failure can be addressed in two complementary ways: by allowing the Go ecosystem through the AWF firewall and by exporting the Go workspace/cache variables that make the runner-prepared module cache visible to chroot-mode commands.

## Goals / Non-Goals

**Goals:**
- Allow AWF agent-executed Go commands in the review workflow to access the Go ecosystem when module resolution must reach the network.
- Export the Go environment variables needed for AWF chroot mode to reuse the prepared Go workspace and module cache.
- Keep the workflow aligned with the repository's existing bootstrap path of `actions/setup-go` followed by `make setup`.
- Update the OpenSpec requirement text so the expected AWF behavior is explicit and testable.

**Non-Goals:**
- Converting the repository to checked-in Go vendoring with `go mod vendor`.
- Redesigning unrelated OpenSpec verification, review submission, archive, or label-cleanup behavior.
- Changing the repository's declared Go version or general dependency-management strategy outside this workflow.

## Decisions

### Allow the Go ecosystem in the AWF firewall

The workflow should add `go` to `network.allowed` rather than relying on hand-maintained domain entries. AWF documents ecosystem-level allowlists for package managers, and the observed failures are specifically blocked requests to `proxy.golang.org`. Using the `go` ecosystem keeps the workflow configuration concise and should cover the Go module hosts the toolchain expects.

Alternative considered: allow only `proxy.golang.org` or a small set of Go domains. Rejected because Go module resolution may involve more than one host, and ecosystem identifiers are the AWF-maintained abstraction for this exact problem.

### Export Go cache and workspace paths alongside GOROOT

The workflow should continue exporting `GOROOT`, but also export `GOPATH` and `GOMODCACHE` after `actions/setup-go` runs. `GOROOT` only exposes the toolchain location; it does not tell chroot-mode commands where the prepared Go workspace and downloaded module cache live. Exporting all three variables gives the agent the best chance to reuse the dependencies prepared during `make setup`.

Alternative considered: export only `GOROOT` and rely on Go defaults for the rest. Rejected because that is the current behavior and it demonstrably still leads to live module fetch attempts from inside AWF.

### Treat network access and cache reuse as complementary safeguards

This change should not assume that cache reuse alone will always prevent outbound module requests. The workflow should preserve both protections: reuse prepared dependencies when possible, and allow Go ecosystem egress when the agent still needs to resolve modules or checksums. This keeps the review job resilient across different agent commands and future workflow changes.

Alternative considered: rely exclusively on firewall access once `go` is allowed. Rejected because it gives up on the already-prepared dependency cache and makes the review workflow more dependent on network availability than necessary.

## Risks / Trade-offs

- Allowing `go` in AWF expands the review job's outbound network surface -> Mitigation: use the AWF ecosystem identifier instead of broad custom domains, and keep the rest of the network policy unchanged.
- Exporting `GOPATH` and `GOMODCACHE` may still be insufficient if a future AWF runtime changes mount behavior -> Mitigation: keep the network allowlist in place as a fallback and validate the workflow on a real run.
- The spec now encodes a more detailed AWF contract -> Mitigation: keep the requirement text tied to observable workflow behavior (`network.allowed`, `GITHUB_ENV`, `make setup`) rather than undocumented internals.
