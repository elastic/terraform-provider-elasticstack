## Context

The acceptance test workflow runs a matrix of Elastic Stack versions, each in an isolated GitHub Actions job. Versions 7.17.x, 8.0.x, and 8.1.x use the Docker Hub image `elastic/elastic-agent` because Elastic does not publish agent images to `docker.elastic.co` for those releases.

Recent runs show that the `fleet` container (from Docker Hub) can stall during `docker compose up`, causing the entire job to hang until the 35-minute timeout triggers cancellation. The failed job consumed the full runner slot without producing useful diagnostics.

## Goals / Non-Goals

**Goals:**
- Detect a hung Docker image pull within a few minutes and abort the attempt
- Automatically retry the pull up to 3 times before failing the job
- Keep the compose-start step bounded so it does not consume the full job timeout

**Non-Goals:**
- Caching Docker layers or images across workflow runs
- Replacing Docker Hub with an internal registry
- Changing the fallback image logic (which versions use Docker Hub vs `docker.elastic.co`)

## Decisions

**Decision 1: Add an explicit pre-pull step with `timeout` and retry loop**
- *Rationale:* `docker compose up` does not expose a pull-timeout option. By pulling the image explicitly first, we get control over timeout and retry. When the pre-pull succeeds, `docker compose up --quiet-pull` skips the pull entirely.
- *Alternative considered:* Use `docker compose pull` with `--parallel` — rejected because it still lacks timeout/retry control.

**Decision 2: Use `timeout 90` per pull attempt**
- *Rationale:* Successful pulls of the agent image finish in ~18 seconds. 90 seconds gives ~5× headroom for slow network conditions while still detecting a true stall quickly.

**Decision 3: Add `timeout-minutes: 10` to the compose step**
- *Rationale:* A normal stack start (all pulls + container starts) takes ~76 seconds. 10 minutes gives ample margin for slow starts while preventing the job from hanging for 30+ minutes.

**Decision 4: Only run the pre-pull step when `matrix.fleetImage` is set**
- *Rationale:* Docker Hub images (the fallback cases) are the ones experiencing stalls. Images from `docker.elastic.co` have been reliable, so we avoid extra complexity for those matrix entries.

## Risks / Trade-offs

- **Risk:** The 90-second timeout could abort a valid pull on an extremely slow connection.
  - *Mitigation:* 3 retries with 30-second backoff give the pull multiple chances; 90s is 5× observed normal duration.
- **Risk:** A Docker Hub rate limit would cause all 3 retries to fail quickly.
  - *Mitigation:* This is a pre-existing risk; the retry loop does not make it worse and the faster failure gives quicker feedback.
