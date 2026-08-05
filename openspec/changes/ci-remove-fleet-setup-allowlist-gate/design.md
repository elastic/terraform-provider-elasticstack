## Context

`.github/workflows/provider.yml` (`test` job, "Matrix Acceptance Test") runs, per matrix entry:

1. Start the stack via `make docker-fleet` (starts the `fleet_settings` service, which creates a
   default Fleet Server host, a `fleet-server` agent policy, and a `fleet_server` `1.5.0` package
   policy, then registers an HTTP Fleet host before the agent starts).
2. Wait for Elasticsearch/Kibana readiness, get an ES API key.
3. `setup-fleet` step (`id: setup-fleet`): run `make setup-kibana-fleet` — but only `if:` the matrix
   version matches one of an explicit list of GA pins (`8.10.4, 8.11.4, 8.12.2, 8.13.4, 8.14.3, 8.15.5,
   8.16.6, 8.17.10, 9.0.8, 9.1.10, 9.2.8, 9.3.6`) or `endsWith(matrix.version, '-SNAPSHOT')`.
4. `make testacc`, which calls `acctest.PreCheck` — this separately ensures a default agent download
   source exists via `ensureFleetDefaultAgentDownloadSource`, independent of the `setup-fleet` step.

`make setup-kibana-fleet` (Makefile) POSTs: a default Fleet Server host, a `fleet-server` agent policy,
a `fleet_server` `1.5.0` package policy, and a default agent download source (added in #2576). Every
one of these except the download source is already created by step 1 (`fleet_settings`, added when CI
switched to `docker-compose` in #1840, Mar 2026). CI logs on versions where `setup-fleet` currently runs
(`8.10.4`, `8.17.10`, `9.5.0-SNAPSHOT`) show: host create succeeds (adds/overrides an HTTPS default
host), agent-policy create returns `409 already exists`, package-policy create returns
`409 already exists`. The download source is redundantly created twice — once implicitly not at all by
`fleet_settings` (it doesn't create one) and once by `setup-fleet`, but `acctest.PreCheck` already
creates it independently of `setup-fleet` via `ensureFleetDefaultAgentDownloadSource`, so removing
`setup-fleet` does not remove download-source coverage.

**Why the allowlist looks the way it does:** the step was introduced in #797 (Oct 2024) to test
secrets against a real Fleet Server, gated from `8.10` (`minVersionIntegrationPolicy`) onward, when CI
used GHA service containers with no `fleet_settings` equivalent. After CI moved to
`docker-compose`/`docker-fleet` in #1840 (Mar 2026), the step became redundant, but the allowlist kept
being maintained by habit — bump a patch pin, or add a promoted `-SNAPSHOT` — rather than by capability
need. `8.18` and `8.19` were promoted to GA and never swept into the list; `9.4` was tracked as
`9.4.0-SNAPSHOT` (matching the `endsWith(..., '-SNAPSHOT')` fallback) and then dropped from Fleet
coverage the moment it was promoted to the GA string `9.4.0`/`9.4.2`, because the promotion didn't add
it to the explicit list. Acceptance tests pass green in all of these gaps, which is exactly what you'd
expect if the step were fully redundant rather than gating something CI actually needs.

## Goals / Non-Goals

**Goals:**
- Remove the `setup-fleet` step from the Matrix Acceptance Test job in `provider.yml`, eliminating the
  allowlist-maintenance burden entirely (no more per-release "did we remember to add the new pin?").
- Preserve existing Fleet-dependent acceptance test coverage — nothing that currently passes should
  start failing, since `docker-fleet` + `acctest.PreCheck` already provide everything the step provided
  beyond duplicate/409 calls.
- Update the `ci-build-lint-test` capability spec so it accurately describes the acceptance test job's
  Fleet bootstrap (via `docker-fleet` + `PreCheck`) instead of a per-version-gated step that no longer
  exists.

**Non-Goals:**
- Changing `make setup-kibana-fleet` itself, or removing it from its other current callers
  (`copilot-setup-steps.yml`, `shared/elastic-stack.md`). Those workflows were not investigated as part
  of this issue and may still find the target useful for their own contexts; touching them is out of
  scope here.
- Deleting the `setup-kibana-fleet` Makefile target. It remains available for local use and for the
  other two workflows above.
- Any change to `force-install-synthetics` gating, the pre-pull fallback fleet image step, or the
  snapshot-failure PR-warning step — none of those are the allowlist this issue is about.
- The pending 9.5.0-SNAPSHOT → 9.5.0 GA promotion (tracked separately in #4403/#4404). This change
  removes the step those changes would otherwise need to keep updating; it does not itself perform that
  promotion.
- Any Terraform provider code, generated client, or documentation change.

## Decisions

### 1. Delete the step rather than drop just the allowlist condition

Remove the `setup-fleet` step block entirely (id, name, `run`, and `env`) rather than making the `if:`
unconditional (i.e., running `make setup-kibana-fleet` for every matrix entry) or narrowing it to a
smaller band.

Why:
- The step is redundant for every matrix entry today (confirmed by the `409 already exists` responses
  on agent/package policy creation), not just the ones currently missing from the list. Making it
  unconditional would keep running fully duplicate work for every version instead of removing it.
- A narrower band (e.g., the historical `8.14`–`8.15` window called out in #2576 for download-source
  flakiness) is already superseded by `acctest.PreCheck`'s `ensureFleetDefaultAgentDownloadSource`,
  which runs unconditionally for every matrix entry regardless of this step.

Alternatives considered:
- Keep the step but drop the allowlist so it always runs: rejected — this still leaves a
  fully-redundant step running on every CI matrix job (slower CI, no coverage benefit), and doesn't
  address the root cause (the step duplicates `docker-fleet`).
- Narrow the allowlist to the historically-flaky `8.14`–`8.15` band instead of removing the step:
  rejected — that flakiness was about the download source specifically, which `acctest.PreCheck`
  already guarantees independent of this step; keeping any band would still require future maintainers
  to remember to update it.

### 2. Leave other `make setup-kibana-fleet` callers untouched

Do not remove or modify the `setup-fleet`-equivalent steps in `copilot-setup-steps.yml` or
`shared/elastic-stack.md`.

Why:
- Those workflows were not part of the investigation in issue #4415 (which is scoped to the Matrix
  Acceptance Test job's allowlist specifically), and this change should not speculatively extend scope
  to workflows whose requirements haven't been examined.
- They run `make setup-kibana-fleet` unconditionally (no per-version gate), so they have no equivalent
  "silently loses coverage for one version" bug — the bug this issue is about is specific to the
  allowlist pattern in `provider.yml`.

Alternatives considered:
- Remove the equivalent step everywhere `make docker-fleet` also runs first, for consistency: rejected
  as out of scope — a reviewer or follow-up issue can evaluate those workflows independently once this
  narrower fix lands.

### 3. Update the `ci-build-lint-test` capability spec

The existing spec (`### Requirement: Acceptance test job structure`) states: "Fleet setup and forced
synthetics installation SHALL run only for configured version subsets." This sentence conflates two
independent per-version-gated steps. Split it so the Fleet-setup half is removed/updated to reflect
that Fleet bootstrap for the acceptance test job is now provided by `make docker-fleet` and
`acctest.PreCheck` (unconditionally, for every matrix entry) rather than a separate gated step, while
the synthetics-install half (unrelated to this change) is preserved unchanged.

Why:
- Leaving the old sentence in place after removing the step would make the spec describe behavior that
  no longer exists, which is exactly the kind of drift OpenSpec is meant to prevent.

Alternatives considered:
- Leave the spec sentence as-is because it's "close enough": rejected — the sentence would be actively
  wrong about Fleet setup once the step is deleted.

## Risks / Trade-offs

- [Removing the step could regress some as-yet-unobserved edge case where `docker-fleet` +
  `acctest.PreCheck` are not sufficient for a specific stack version] -> Mitigation: the step is already
  skipped today for `8.18`, `8.19`, and `9.4.2` with acceptance tests passing, and every version where it
  *does* still run shows the agent/package-policy calls returning `409 already exists` — i.e., no
  observed version depends on this step for anything `docker-fleet` doesn't already provide.
- [A future stack version could reintroduce a real need for this bootstrap step] -> Mitigation: if that
  happens, it will surface as a Fleet-dependent acceptance test failure with a clear signal (unlike the
  current silent-skip failure mode), and the step can be reinstated or replaced with the correct minimal
  fix at that time.

## Migration Plan

1. In `.github/workflows/provider.yml`, delete the `setup-fleet` step (the `id: setup-fleet` block,
   including its `if:` allowlist condition, `run: make setup-kibana-fleet`, and `env:`) from the Matrix
   Acceptance Test job.
2. Confirm no later step in the same job (`force-install-synthetics`, `tf-acceptance`, teardown/log
   steps) references the `setup-fleet` step's `id` or outputs.
3. Update `openspec/specs/ci-build-lint-test/spec.md`'s "Acceptance test job structure" requirement to
   remove the Fleet-setup-specific clause from the "SHALL run only for configured version subsets"
   sentence, keeping the synthetics-install clause intact.
4. Land the change; the next `Provider CI` run exercises every matrix entry without the `setup-fleet`
   step, relying on `docker-fleet` + `acctest.PreCheck` for Fleet bootstrap.

## Open questions

- Should `copilot-setup-steps.yml` and `shared/elastic-stack.md` eventually drop their own
  `make setup-kibana-fleet` invocations too, given the same redundancy argument applies there (both
  already run `make docker-fleet` first)? Left out of scope for this change per Non-Goals; worth a
  follow-up issue if a maintainer wants to pursue it.
- Should the pending 9.5.0-SNAPSHOT → 9.5.0 GA promotion (#4403/#4404) be sequenced before or after this
  change lands? Either order works functionally (this change removes the step those changes would
  otherwise need to keep in sync), but landing this change first means that promotion no longer needs
  to touch the Fleet-setup allowlist at all.
