## Why

Kibana, Fleet, and APM resources are written against the Kibana API as it existed at the time. Kibana keeps shipping new request fields, and nothing tells the provider team, so the gap between what the API can configure and what the provider exposes grows silently.

The existing `kibana-spec-impact` workflow only notices *changes* to the generated Kibana client between commits, so it cannot see the backlog of fields that were never adopted. It also relies purely on deterministic matching with no judgement about whether a field is worth exposing, which pushes every relevance decision onto maintainers.

## What Changes

- A new scheduled scanner (also run when the generated Kibana client changes) finds Kibana API request fields that a Terraform resource built on the Kibana API (Kibana, Fleet, or APM) can reach but never uses.
- **Candidate discovery is deterministic.** It proves that a field exists in the API and that the entity does not use it, and gives each candidate a stable identity based on the entity and the API spec location. Nothing is filed that has not been proven absent.
- **Relevance is decided by an LLM triage step**, which can only *remove* candidates, never add them. It suppresses a candidate only when highly confident that it should not be exposed (for example server-managed values, or fields already covered by an opaque JSON attribute). Anything uncertain is filed.
- Candidates are grouped into one issue per (entity, top-level request field).
- Every triage verdict is persisted in repository memory. Each suppressed field is recorded with a short reason, which is the sole audit trail for suppressions. A verdict (filed or suppressed) covers the group until the field's structural API definition changes; then it expires and the group is re-triaged.
- Filed issues are capped per run using the same issue-slot backpressure as the other scheduled scanners, pre-labelled `triaged`, carry a topic label, and are automatically dispatched to `research-factory`.
- **BREAKING (CI only):** `kibana-spec-impact` is removed: its workflow, helper, spec, and memory branch. The new scanner covers newly added fields, and the Go compiler covers removed or retyped fields that the provider references (regenerating the client breaks the build).
- The deterministic candidate discovery must show acceptable signal-to-noise on a small, representative set of entities before the scanner is put on a schedule. If it does not, the change is revisited instead of shipping a noisy scanner.
- **Out of scope** (explicit non-goals):
  - Elasticsearch APIs. Elasticsearch resources use a different client and spec source; they are a possible follow-up.
  - Whole Kibana API endpoints that no entity uses (that is, missing resources). That is new-entity work with its own workflow.
  - Response-only fields, which users cannot configure.
  - Terraform data sources (their value lies mainly in response fields).
  - Fields the provider references but only ever sets to fixed values (possible v2).
  - Gaps in enum values for fields the provider already exposes (possible v2).
  - A human review queue or digest of suppressed fields. Repository memory is the record.
  - Promoting issues past research automatically. `change-factory` stays a human decision.
  - Tracking semantic changes (for example enum values removed) to fields the provider already covers.

## Capabilities

### New Capabilities

- `ci-kibana-api-gap-scanner`: scheduled and client-change-triggered scanning of resources built on the Kibana API (Kibana, Fleet, APM) for unadopted Kibana API request fields; deterministic candidate discovery; confidence-gated LLM triage; per-(entity, top-level field) issue creation with slot caps; repository memory of verdicts with expiry.
- `ci-research-factory-dispatch-fanout`: producer workflows dispatch `research-factory` once per created issue, mirroring the existing `code-factory` fan-out contract.
- `gh-aw-shared-dispatch-research-factory`: the shared workflow import that provides research-factory dispatch to producer workflows, mirroring `gh-aw-shared-dispatch-code-factory`.

### Modified Capabilities

- `ci-kibana-spec-impact-issues`: **REMOVED** in full (every requirement removed), replaced by `ci-kibana-api-gap-scanner`.
- `ci-automated-issue-triaged-label`: the set of automated producer workflows that apply `triaged` at issue creation grows from four to five, adding the Kibana API gap scanner.


## Impact

- New deterministic Go helper under `scripts/`, absorbing the reusable entity inventory and Kibana client call-site discovery from `scripts/kibana-spec-impact`.
- New gh-aw workflow under `.github/workflows/` plus its compiled lock file, and a new `memory/` repo-memory branch.
- Shared workflow import for dispatching `research-factory`, alongside the existing `code-factory` dispatcher.
- Removed: `.github/workflows/kibana-spec-impact.md` and its lock file, `scripts/kibana-spec-impact`, the `ci-kibana-spec-impact-issues` spec, and the `memory/kibana-spec-impact` branch. Existing open `[kibana-spec-impact]` issues are left as they are.
- `research-factory` load increases. The `research-factory-critic-gate` change is recommended to land first, so automated research is quality-gated, but it is not a hard dependency.
- Repository labels: a new scanner topic label must be provisioned.
- LLM gateway usage: triage cost scales with the number of candidates, not with the size of the API spec.
- Docs: `dev-docs/high-level/continuous-quality-workflows.md` and `dev-docs/high-level/factory-workflows.md`.
