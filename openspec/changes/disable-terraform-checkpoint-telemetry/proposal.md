## Why

The `code-factory` and `reproducer-factory` agentic workflows run Terraform acceptance tests
(`TF_ACC=1 go test ...`) inside the GH-AW sandbox. The real `terraform` binary embedded by
`terraform-plugin-testing` calls HashiCorp's checkpoint service
(`checkpointapi.hashicorp.com`) on startup for an upgrade/security-bulletin check. That domain is
not in either workflow's AWF network allowlist, so the sandbox's egress firewall drops the call —
reported against [PR #4470](https://github.com/elastic/terraform-provider-elasticstack/pull/4470)
as a blocked-domain failure that aborted a targeted acceptance-test run before it could complete.

The same report also flagged `host.docker.internal` as blocked. That access path is handled by the
AWF sandbox's host-access mechanism (`sandbox.agent.legacy-security: enable` plus
`--allow-host-service-ports`/`--enable-host-access`/`--allow-host-ports`), which commit `2162bb4c`
("Fixup test stack", 2026-07-31) already added to both `code-factory-issue` and
`reproducer-factory-issue`, predating the PR #4470 run that reported the failure. `host.docker.internal`
is also already present in `GH_AW_ALLOWED_DOMAINS` via gh-aw's `defaults` network preset. This
proposal treats that half of the report as already resolved on `main` and focuses on the
still-reproducible `checkpointapi.hashicorp.com` gap.

## What Changes

- Disable the Terraform CLI's checkpoint telemetry call (`CHECKPOINT_DISABLE=1`) for every
  acceptance-test invocation the `code-factory` and `reproducer-factory` agents run inside the
  GH-AW sandbox, removing the blocked outbound call outright instead of adding a narrow domain
  allowlist entry for a telemetry endpoint.
- Update the documented `go test` acceptance-test invocation in both workflows' agent prompts to
  show `CHECKPOINT_DISABLE=1` alongside the existing `ELASTICSEARCH_ENDPOINTS`/`KIBANA_ENDPOINT`
  connection parameters.
- Set `CHECKPOINT_DISABLE` at the job level (workflow-level `env:`) for both workflows so it applies
  regardless of the exact `go test` invocation the agent constructs, not only when the agent copies
  the documented example verbatim.

## Capabilities

### New Capabilities
<!-- None. -->

### Modified Capabilities
- `ci-code-factory-issue-intake`: acceptance-test environment for the implementation agent now
  disables Terraform checkpoint telemetry so the sandboxed `go test` run cannot be blocked by the
  egress firewall on `checkpointapi.hashicorp.com`.
- `ci-reproducer-factory-issue-intake`: acceptance-test environment for the reproduction agent
  gets the same `CHECKPOINT_DISABLE=1` treatment for parity with `code-factory`.

## Impact

- `.github/workflows/code-factory-issue.md` and its compiled `.lock.yml`: add `CHECKPOINT_DISABLE`
  to the workflow-level `env:` block and to the documented `go test` example in the agent prompt.
- `.github/workflows/reproducer-factory-issue.md` and its compiled `.lock.yml`: same treatment,
  reusing the workflow's existing top-level `env:` block.
- No changes to `network.allowed` on either workflow, and no changes to Terraform provider source,
  provider tests, generated clients, or `shared/elastic-stack.md`'s proxy/service setup.
