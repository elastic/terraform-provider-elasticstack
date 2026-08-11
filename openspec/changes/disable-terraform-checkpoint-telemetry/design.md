## Context

The `code-factory` and `reproducer-factory` GH-AW workflows import `shared/elastic-stack.md` to
provision a live Elastic Stack and run `TF_ACC=1 go test ...` against it inside the agentic
sandbox. The sandbox enforces an egress allowlist (`network.allowed` compiled into
`GH_AW_ALLOWED_DOMAINS`/`awf-config.json`). The `terraform-plugin-testing` module used by
acceptance tests embeds a real `terraform` binary, and that binary calls HashiCorp's checkpoint
service on startup unless `CHECKPOINT_DISABLE` is set. `checkpointapi.hashicorp.com` is not in
either workflow's allowlist, so the sandbox firewall drops the call and the reported run
(quoted from PR #4470) failed before the acceptance test could complete.

Investigation of `main` (`.github/workflows/shared/elastic-stack.md`, `code-factory-issue.md`,
`reproducer-factory-issue.md`, and their compiled `.lock.yml`s) confirms the `host.docker.internal`
part of the report is separate and already addressed: commit `2162bb4c` ("Fixup test stack",
2026-07-31) added `sandbox.agent.legacy-security: enable` plus `--allow-host-service-ports`,
`--enable-host-access`, and `--allow-host-ports 80,443,8080` to both workflows, and
`host.docker.internal` ships in `GH_AW_ALLOWED_DOMAINS` via gh-aw's `defaults` network preset
regardless of this repo's config. That commit predates the PR #4470 run that reported the failure.

## Goals / Non-Goals

**Goals:**
- Eliminate the blocked outbound call to `checkpointapi.hashicorp.com` from sandboxed acceptance
  test runs in `code-factory` and `reproducer-factory`.
- Apply the fix at the source (disable the telemetry call) rather than growing the egress
  allowlist for a purely informational HashiCorp check.
- Keep the fix effective regardless of the exact `go test` command the agent constructs, not only
  when it copies the documented example verbatim.

**Non-Goals:**
- Re-litigating the `host.docker.internal` access path — treated as already resolved by commit
  `2162bb4c`; see Open Questions for the one item still worth confirming.
- Changing `shared/elastic-stack.md`'s proxy/service topology or `docker-compose.yml`'s bind-address
  configuration.
- Changing the non-sandboxed `provider.yml` CI acceptance-test job or any gh-aw upstream behavior
  (squid firewall implementation, `defaults`/`terraform` network presets, `awf` host-access flags).

## Decisions

### Disable checkpoint telemetry at the source (`CHECKPOINT_DISABLE=1`) rather than allowlisting the domain

Set `CHECKPOINT_DISABLE=1` as a job-level environment variable on both `code-factory-issue.md` and
`reproducer-factory-issue.md`, and reflect it in each workflow's documented `go test` example so a
human reading the prompt sees the same behavior the job environment already guarantees.

Why:
- It removes the outbound call entirely instead of routing around the firewall — no new domain to
  keep matching squid's SNI/Host-based ACL as HashiCorp rotates IPs behind that hostname, and no
  telemetry egress from CI at all.
- It also avoids the same checkpoint call for developers running acceptance tests locally **if they opt in** by setting `CHECKPOINT_DISABLE=1`, without requiring a docker-compose change in this proposal's scope.
- A job-level `env:` var applies no matter what `go test -run ...` invocation the agent types,
  whereas fixing only the documented example snippet would leave the gap open if the agent
  constructs its own command.

This was evaluated against allowlisting `checkpointapi.hashicorp.com` in `network.allowed`
(considered and rejected in the issue's research comment): that approach is a one-line change to a
shared component, but it grows the egress allowlist for a call that provides no CI value and adds
one more hostname squid has to keep resolving correctly. Disabling the call at the source avoids
both costs. Both approaches were already evaluated in the linked research; this proposal does not
re-open that exploration.

### Set the env var per-workflow rather than in `shared/elastic-stack.md`

`shared/elastic-stack.md`'s only `env:` block is scoped to its own "Setup Elastic Stack" step
(`make docker-fleet`), which never invokes `terraform` — the checkpoint call happens later, when
the agent runs `go test` under its own shell. Setting `CHECKPOINT_DISABLE` there would not reach
the agent's later invocation. Both consuming workflows already have (or, for `code-factory-issue.md`,
can add) a workflow-level `env:` block that applies to the whole job, so that is the correct place
for a variable the agent's shell needs to inherit.

## Risks / Trade-offs

- [Job-level `env:` may not propagate into the agent's bash-tool shell the way it does for regular
  `run:` steps, depending on how gh-aw wires the agent step's environment] -> Mitigation: the
  implementer should verify with a real `code-factory`/`reproducer-factory` run (or a compiled
  `.lock.yml` inspection) that `CHECKPOINT_DISABLE` is visible to the agent's shell, and fall back to
  embedding it directly in the documented `go test` example if the job-level var does not propagate.
- [Some other, not-yet-identified HashiCorp endpoint could still be hit by a future Terraform CLI
  feature] -> Mitigation: none in this proposal's scope; tracked as an open question below.

## Open questions

- Was the PR #4470 sandbox run that produced the quoted firewall message actually on a checkout
  that predates commit `2162bb4c` (2026-07-31)? If so, the `host.docker.internal` portion of this
  issue is already closed by that commit and only the checkpoint-domain gap remains.
- Should `CHECKPOINT_DISABLE` also be set for the non-sandboxed `provider.yml` CI acceptance-test
  job (outside the GH-AW workflows), for consistency and to trim unnecessary egress there too, or
  is that out of scope for this issue?
- Are there other HashiCorp/Terraform CLI telemetry endpoints (e.g. Terraform Cloud/registry
  version checks) that could hit the same block under `TF_ACC_TERRAFORM_VERSION` overrides used for
  provider-defined-action tests, and would `CHECKPOINT_DISABLE` cover those too?
