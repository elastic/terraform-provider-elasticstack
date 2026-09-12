## Context

`.github/workflows/provider.yml`'s `test` job hardcodes `strategy.matrix.version` (21 GA entries plus a trailing `-SNAPSHOT`) and an `include:` block that layers `runner: ubuntu-22.04` and a Docker Hub `fleetImage` fallback onto five specific `8.0.x`–`8.4.x` entries. A separate `if:` condition force-installs synthetics `1.2.2` for four exact patch strings (`8.14.3`, `8.15.5`, `8.16.6`, `8.17.10`). All of this is edited by hand.

There is no single public "all versions" API (see the issue's source-of-truth table). The two sources that are complete, public, and stable are:

- `elastic/elasticsearch` git tags matching `vX.Y.Z` — gives every published GA patch, so "latest patch per minor" is a pure function of the tag list.
- `https://snapshots.elastic.co/latest/master.json` — gives today's unreleased-next-minor SNAPSHOT label.

A workflow that fetched these live at Provider CI start would make matrix membership non-reproducible for a given commit SHA. The existing `changelog-generation.yml` workflow already solves an analogous problem (a generated, pinned artifact — `CHANGELOG.md`'s `## [Unreleased]` section — proposed by a scheduled PR, re-triggered via `GH_AW_CI_TRIGGER_TOKEN`, no-op when unchanged). This change reuses that shape for the version matrix instead of inventing a new one.

The other repository-specific wrinkle: today's per-version `include:`/`if:` overrides are keyed to *exact patch strings*. The moment automation bumps `8.14.3` to a later `8.14.x` patch, the exact-string `force-install-synthetics` condition silently stops firing for that entry — a correctness regression introduced by the very automation this change adds. The overrides must become minor-range rules.

## Goals / Non-Goals

**Goals:**
- Desired stack-version list is a pure, deterministic function of `elastic/elasticsearch` GA tags (8.x/9.x, latest patch per minor) and the current SNAPSHOT label.
- The list is pinned in git so a given commit SHA tests a fixed, reproducible set of versions.
- A scheduled workflow (+ `workflow_dispatch`) proposes updates as a single standing PR, re-using the branch/no-op/re-trigger shape already established by `changelog-generation.yml`.
- Per-version environment rules in `provider.yml` (fleet image, runner, forced synthetics install) survive an automated patch bump without being re-edited, by matching on minor-version ranges instead of exact patch strings.
- The version-bump PR runs full Provider CI against the *proposed* list and is auto-approved (not auto-merged) once Provider Gate is green, via a new `scripts/auto-approve` category.
- `acceptance-test-isolation` REQ-ACC-002 stops pinning a literal patch string that automation will immediately invalidate.

**Non-Goals:**
- Auto-merging the version-bump PR. Merge stays with whatever already exists (humans / branch protection rulesets).
- Computing the matrix at Provider CI run time from a live fetch (non-reproducible for a given SHA).
- Tracking every patch of a minor, or 7.x/10.x lines.
- Treating `artifacts-api.elastic.co`, GCS `releases/current/*`, Elastic Cloud `stack/versions`, or Docker registry tag listing as the version catalog (per the issue's research, each is incomplete, frozen, auth-walled, or a runtime constraint rather than a catalog).
- Chasing SNAPSHOT build IDs — only the SNAPSHOT *label* (e.g. `9.6.0-SNAPSHOT`) is pinned; CI already pulls the newest image for a given label.
- Changing acceptance-test sharding, Fleet bootstrap, or the existing snapshot-failure PR comment beyond what the range-rule rewrite requires.

## Decisions

### Decision 1: Pin a flat JSON array of version strings, not a full matrix

`.github/versions/acceptance-test-matrix.json` is a JSON array of version strings, sorted ascending by release line, e.g.:

```json
[
  "8.0.1", "8.1.3", "8.2.3", "8.3.3", "8.4.3", "8.5.3", "...",
  "8.19.21", "9.0.8", "9.1.10", "9.2.8", "9.3.8", "9.4.6", "9.5.3",
  "9.6.0-SNAPSHOT"
]
```

**Why:** The issue is explicit that pinning must not YAML-surgery `strategy.matrix.version`/`include`, and that environment rules (runner, fleet image, synthetics) must become version-*range* rules rather than exact-patch pins. Keeping the pinned artifact to bare version strings — with no `runner`/`fleetImage` fields — forces those environment rules to live as range expressions in the workflow itself (Decision 3), which is what makes them survive a patch bump. A richer pinned structure (e.g. baking `runner`/`fleetImage` into the JSON) was considered and rejected: it would just move the "exact version → override" staleness problem from hand-edited YAML into generator code, instead of eliminating it.

### Decision 2: `scripts/version-matrix` Go engine, invoked the same way as `scripts/changelog`

A new Go module `scripts/version-matrix/` owns:
- `compute()`: list `elastic/elasticsearch` tags via the GitHub API (using the workflow token for rate limit), filter to `^v(8|9)\.\d+\.\d+$`, group by minor, keep the max patch per minor, and fetch `https://snapshots.elastic.co/latest/master.json` for the SNAPSHOT label.
- the compose-stack image-existence probe (Decision 4).
- `diff()`/`write()`: compare the computed list against the pinned artifact and rewrite it when different.

The workflow invokes it as `go run ./scripts/version-matrix <subcommand>` after `actions/checkout` + `actions/setup-go`, mirroring the `scripts/changelog` contract (`ci-changelog-generation` capability) instead of an `actions/github-script` module. Unlike the changelog engine, no `.github/scripts/workflows/lib/*.js` predecessor exists here, so there is no migration/parity concern — this is a fresh Go tool from the start.

**Why:** Consistency with the repository's established pattern of "shared logic as a Go module under `scripts/`, invoked via `go run` from a checkout+setup-go step" (`scripts/changelog`, `scripts/auto-approve`), and it keeps GA-tag pagination, semver comparison, and HTTP calls unit-testable outside of GitHub Actions.

### Decision 3: Per-version environment rules match numeric major.minor, not string prefixes

The `test` job's `runs-on`, fleet-image selection, and forced-synthetics condition match **integer** major and minor components of `matrix.version`, not `startsWith(matrix.version, '8.1.')` (or any other dotted-prefix check). `'8.10.4'` starts with `'8.1.'`, so prefix matching would send 8.10–8.19 to `ubuntu-22.04` and the Docker Hub Agent image.

Range bounds stay as data next to the workflow (or in the `load-matrix` helper), not in the pinned JSON:

- Runner: major `8` and minor `0`–`4` → `ubuntu-22.04`, else `ubuntu-latest`
- Fleet image: major `8` and minor `0`–`1` → Docker Hub `elastic/elastic-agent`, else `docker.elastic.co/elastic-agent/elastic-agent`
- Forced synthetics: major `8` and minor `14`–`17`

GitHub Actions expressions cannot parse semver. The preceding `load-matrix` job (or an equivalent helper it invokes) SHALL parse each pinned version and attach the derived flags (`runner`, `fleetImage` / pre-pull, `forceSynthetics`) so the `test` job reads those fields instead of re-deriving them with `startsWith`. The pinned artifact remains a flat version-string array; the flags are runtime derivatives of the range rules, not committed overrides.

`shard: [0, 1]` stays a static cross-product axis.

**Why:** Numeric major.minor is the only range check that survives both patch bumps and two-digit minors. Evaluating the rules in `load-matrix` keeps them out of the pin (Decision 1) without using unsafe GHA string prefixes. A promoted `X.Y.0-SNAPSHOT` → `X.Y.0` entry keeps matching any rule for that minor automatically.

### Decision 4: Compose-stack image probe falls back, not fails, including on SNAPSHOT promotion

Before including a newly computed **GA** version in the desired list, the generator checks that every image `docker-compose.yml` will pull for that version resolves:

- `docker.elastic.co/elasticsearch/elasticsearch:<version>`
- `docker.elastic.co/kibana/kibana:<version>`
- the Fleet/Agent image for that version: Docker Hub `elastic/elastic-agent:<version>` when major `8` minor `0`–`1`, otherwise `docker.elastic.co/elastic-agent/elastic-agent:<version>`

If any of those manifests does not resolve, the run does not fail. Fallback depends on what was previously pinned for that minor:

- **Patch bump** (pin already has a GA `X.Y.*`): keep the previously pinned GA for `X.Y`.
- **SNAPSHOT promotion** (pin has `X.Y.*-SNAPSHOT`, computed list wants GA `X.Y.*`): keep `X.Y.*-SNAPSHOT` as that minor's entry and **do not** append master's newer SNAPSHOT label on this run. The list still has exactly one SNAPSHOT (the line that is not pullable as GA yet). Next successful probe promotes `X.Y` and then adds the new master SNAPSHOT.
- **Brand-new minor** with no previous pin: omit that minor for this run.

Other minors' successful bumps still land in the same computed list.

**Why:** Compose starts Elasticsearch, Kibana, and Agent; probing only Elasticsearch would propose versions that still fail `docker compose up`. Promotion has no previous *GA* pin to fall back to — keeping the SNAPSHOT label preserves coverage and the "exactly one SNAPSHOT" rule instead of dropping the line or emitting two SNAPSHOT entries.

### Decision 5: Standing PR reuses the `generated-changelog` branch/no-op/re-trigger shape

Branch `acceptance-test-version-matrix` (stable, force-pushed on each change, mirroring `generated-changelog`), commits authored by `github-actions[bot]`, empty re-trigger commit pushed with `GH_AW_CI_TRIGGER_TOKEN` (because `GITHUB_TOKEN`-authored pushes do not run Provider CI, so Gate never goes green and auto-approve never fires), PR body stating the all-or-nothing "make main match the computed list" intent, `no-changelog` label (it is not user-facing), lookup-existing-PR-by-branch idempotency, and no-op (no commit, no branch push, existing PR left untouched) when the computed list matches the pinned artifact.

**Why:** This is a proven, already-reviewed pattern in this repository (`ci-changelog-generation`); reusing it minimizes new surface area and keeps the two scheduled-PR workflows operationally consistent for maintainers.

### Decision 6: New `version-matrix` auto-approve category, gated identically to `generated-changelog`

Selector: same-repository PR whose head branch is exactly `acceptance-test-version-matrix`. Gates: every commit authored by `github-actions[bot]`; every changed file path exactly `.github/versions/acceptance-test-matrix.json`. No diff-threshold gate (a full matrix rewrite touching one JSON file is not a Copilot-style unbounded diff). Approval requires global gates (including a green Provider Gate) to pass, same as every other category.

**Why:** Directly mirrors the existing `generated-changelog` category (`ci-pr-auto-approve`), which is the closest existing precedent for "bot-authored, single-file, branch-name-identified" auto-approval, and satisfies REQ-012 (category extensibility) without touching existing categories.

## Risks / Trade-offs

- **GitHub API rate limits on tag listing.** `elastic/elasticsearch` has thousands of tags; the workflow authenticates with a token to get the higher rate limit, and the generator only needs to run once daily.
- **`snapshots.elastic.co` availability.** If the SNAPSHOT endpoint is briefly unavailable, the generator should fail the run (not silently drop the SNAPSHOT entry from the pinned list), leaving the existing pinned artifact untouched — same "no-op on uncertainty" posture as Decision 4's image-probe fallback.
- **Full-matrix PRs are expensive.** A version-bump PR runs the entire Provider CI matrix (now provider-impacting via change-classification) even when only one minor's patch changed. This is intentional per the issue ("all-or-nothing: a red new GA holds the whole delta") and matches the cost the current hand-edited process already pays.
- **Range bounds are still a manual allowlist.** The numeric minor ranges (`8.0`–`8.4` runner, `8.0`–`8.1` Fleet image, `8.14`–`8.17` synthetics) stay hand-maintained because they encode closed historical lines. They must not be implemented as `startsWith('8.1.')`-style prefixes. They only need a human edit when a *new* minor needs a new range rule.

## Open questions

- None raised by the issue or comment history at authoring time; the issue's own "Solution" section is prescriptive enough to serve as the spine of this design.
