## Context

- **Generated client:** `generated/kbapi` is produced by oapi-codegen from `oas-filtered.yaml`, not from the upstream `oas.yaml`. `transform_schema.go` applies about 40 transforms first: removing the kbn-xsrf and API-version params, unwrapping anyOf/allOf, collapsing enum anyOf, removing broken discriminators, renaming, merging, and pruning component schemas, and resource-specific fixes. Both files contain all 532 paths, but `oas-filtered.yaml` has 1,627 component schemas against 1,821 upstream. It keeps `deprecated` and carries `x-state` on 289 nodes.
  - Every API field is a Go struct field with a json tag.
  - Every operation has generated `New<Op>Request*` constructors in `kibana.gen.go` (1,102 of them). These contain a literal `http.Method*` and an `operationPath := fmt.Sprintf("/api/...")`.
- **How resources reach the API:**
  - Kibana resources reach kbapi directly or through `internal/clients/kibanaoapi`. 39 of them are built with `entitycore.NewKibanaResource`, which receives their create, read, update, and delete logic as **function values** and calls them through struct fields.
  - Fleet resources go through `internal/clients/fleet`.
  - `elasticstack_apm_agent_configuration` and `elasticstack_apm_source_map` also call kbapi.
- **What `kibana-spec-impact` already provides:**
  - provider-registry entity discovery (`inventory.go`);
  - the deterministic pre-activation report -> agent -> Go `memory-record-from-report` pattern, with a `memory/` repo-memory branch (`create-orphan: true`).
- **Existing scanner plumbing to reuse:**
  - `issue-slots/compute.js` (open-issue cap);
  - `shared/dispatch-code-factory.md` plus `lib/producer-dispatch.js`, whose CLI already accepts the target workflow file as its third argument.
- **Dependencies:** `golang.org/x/tools` and `gopkg.in/yaml.v3` are already direct dependencies. No new library is needed.
- **Build wiring:** `make workflow-test` currently runs `go test ./scripts/kibana-spec-impact/...`.
- **gh-aw job gating** (from the compiled lock files): `safe_outputs` runs only after a successful threat-detection job, and in practice it does not process outputs when the agent job fails. `push_repo_memory` runs after `safe_outputs` and only when the agent succeeded. A run is therefore all-or-nothing: either issues are attempted and memory is pushed, or neither happens.

## Goals / Non-Goals

**Goals:**
- Deterministically discover unused Kibana API request fields per resource, with absence proven by type analysis and every fact anchored to an OAS node.
- Confine the LLM to judgement: one verdict and reason per pre-rendered gap group.
- Persist verdicts in a reviewable memory file that expires by structural fingerprint.
- Reuse the existing scanner, memory, slot, and dispatch patterns.
- Replace `kibana-spec-impact` without any period in which Kibana client changes go unwatched.

**Non-Goals:**
- Data sources, response-only fields, and missing endpoints or resources.
- Fields the provider uses but only sets to fixed values ("referenced but constant"). The spike counts them to inform a possible v2.
- An exclusions file or any maintainer-curated suppression list.
- GitHub issue existence or state lookups.
- Elasticsearch APIs.

## Decisions

### D1. OAS-first discovery with a lockstep walk

Discovery starts from the Kibana API specification and uses Go only to answer "is this field used by this resource?":

```
 1. go/packages: find kbapi client methods the resource calls -> operations
 2. from oas-filtered.yaml: those operations' request bodies + query params = API FIELDS
    (deprecated, x-state, structure read here)
 3. walk OAS schema and Go types in lockstep, matching properties by json tag
    (oneOf -> union variant types)
 4. per OAS field: is its Go field in the resource's USED set (go/types)?
 5. gap = OAS fields whose Go field is unused
```

Everything the spec keys on lives on the OAS node: the stable key (OAS location), the structural fingerprint, the deprecated exclusion, the introduced-in version, and query parameters.

The walk uses **`generated/kbapi/oas-filtered.yaml`**, the exact input the Go types were generated from, so the lockstep match is one-to-one. Group keys are JSON pointers into that file, and fingerprints are computed from it. Issue citations also name the operation as method plus path (identical upstream) and the field path, so they stay meaningful to readers of the upstream spec. A change to `transform_schema.go` that restructures a field changes its fingerprint and re-triages that group. That is intended, because the Go types change too.

*Alternative considered:* Go-types-first reachability with a reverse mapping from Go names back to OAS nodes. It was rejected because the reverse mapping duplicates oapi-codegen naming internals and drifts silently when the generator changes. The lockstep walk instead fails loudly on a mismatch.

### D2. Architecture: deterministic tool for facts, agent for judgement, gh-aw for writes

```
 workflow: kibana-api-gap-scanner.md   (concurrency: 1 run at a time, queued)
 +-----------------------------------------------------------------------------+
 | pre-activation job (deterministic, no LLM)                                  |
 |   checkout + memory branch (memory/kibana-api-gap-scanner)                  |
 |   issue-slots/compute.js  (label kibana-api-gap, cap 3)  -> slots           |
 |   go run ./scripts/kibana-api-gaps discover                                 |
 |   go/packages + oas-filtered.yaml -> all gap groups (key, fp, fields)       |
 |        minus groups covered in memory (same fingerprint)                    |
 |        ordered (resource rotation, size, key) -> triage queue (<= 15)       |
 |        each group: pre-rendered issue title/body (incl. hidden key marker)  |
 |   upload report artifact; run_agent = slots > 0 && queue non-empty          |
 +-----------------------------------------------------------------------------+
            | (skipped when run_agent false)
 +-----------------------------------------------------------------------------+
 | agent job (LLM triage only)                                                 |
 |   for each queued group in order: read OAS node, resource code, docs MCP    |
 |     verdict file|suppress + short reason                                    |
 |     file -> create-issue(pre-rendered title, body + reason)                 |
 |   stop at slots filled OR 15 triaged                                        |
 |   write verdicts.json -> go run ./scripts/kibana-api-gaps record            |
 |        validates verdicts vs report (unknown keys rejected), writes memory  |
 |   dispatch_research_factory(dispatch: true)                                 |
 +-----------------------------------------------------------------------------+
            |
 safe_outputs job: create issues (labels kibana-api-gap, triaged)
            |
 dispatch-research-factory job (shared fragment): 1 dispatch per created issue
```

The three owners:

- **`scripts/kibana-api-gaps` (Go)** owns every fact: what exists, what is absent, keys, fingerprints, ordering, rendered issue content, and memory validation and writing. It has two subcommands, `discover` and `record`, following the `kibana-spec-impact` precedent of `pre-activation` and `memory-record-from-report`.
- **The agent** owns only judgement: a verdict and a reason. It cannot add fields, because issue bodies are pre-rendered and the agent appends only the reason. `record` rejects verdicts for keys that are not in the report.
- **gh-aw** owns writes: the `create-issue` safe output, the repo-memory push, and the shared dispatch fragment.

Code reuse: `inventory.go` moves from `kibana-spec-impact`, widened to resources that call the Kibana client and narrowed to resources only. `index.go` is not carried over, because the reference graph (D3) replaces its file-level call-site scan. The memory-branch checkout and slot computation follow the schema-coverage and kibana-spec-impact patterns, and `producer-dispatch.js` is reused through its workflow-file parameter.

### D3. Component structure

**`scripts/kibana-api-gaps` (Go; replaces `scripts/kibana-spec-impact`):**

| Unit | Responsibility | Notes |
|---|---|---|
| `inventory` | Lists provider-registered **resources** and their packages | Moved from spec-impact `inventory.go`. The prefix filter is replaced by "calls the Kibana client", as decided by `reach`. |
| `load` | One `go/packages` load of the module with types and syntax | Loaded once per run and reused by every unit below |
| `reach` | Per resource: a **reference graph** rooted at **every function and function literal declared in the resource's package**. It follows every reference to another module function, whether a call or a use as a function value. Interface method calls are resolved to **all** module implementations of that method. Its outputs are the reachable function set and the kbapi client methods referenced within it. | Rooting at the package's own declarations means callbacks handed to `entitycore` as function values are always included, regardless of how `entitycore` invokes them. Traverses `kibanaoapi`, `clients/fleet`, and other module helpers. Every approximation is an **over**-approximation, so usage is over-counted: a field is counted as used whenever any reachable code touches it. The tool may miss gaps, but it never reports a false absence. |
| `opindex` | Maps each kbapi client method to an OAS operation `(method, path template)` by reading the generated `New<Op>Request*` constructors | Pure AST; no reimplementation of oapi-codegen naming |
| `oas` | Loads `generated/kbapi/oas-filtered.yaml` as a `yaml.v3` node tree; resolves `$ref` and `allOf`; exposes request-body schema and query params per operation, plus `deprecated` and `x-state` | Uses the raw node tree, not an OpenAPI loader, so fingerprinting is exact |
| `walk` | Lockstep walk of OAS schema and Go type: properties to struct fields by json tag; arrays, maps, and pointers to their element; `oneOf`/`anyOf` to union variant types through `As*`/`From*` | An unmatched json tag is a **tool error**, never silently skipped |
| `usage` | Per resource: the set of kbapi struct fields used (`types.Info` selections and composite-literal keys) within that resource's `reach` set; plus **opaque-covered types**: targets of whole-value `json.Unmarshal`/`Marshal`, and used `json.RawMessage` or map-typed fields | Computed per resource, not module-wide, so another resource's usage cannot hide a gap |
| `group` | Builds gap groups per (resource, top-level field name), keyed `<resource>#<OAS JSON pointer>`, whose members are the unused fields in the subtree; drops deprecated and response-only fields | If the top-level field itself is unused, its whole subtree forms one group. The same field name in several operation bodies or union variants (create and update, or every variant of a union body) is merged into one group keyed by the smallest pointer (so a new smaller sibling re-keys and re-triages the group once, accepted), kept apart for body properties versus query parameters, with a fingerprint over all constituent pointers and fingerprints; this was added after the spike, where per-variant groups made one resource report 506 groups |
| `fingerprint` | SHA-256 of a canonical JSON projection of the resolved subtree over `type`, `format`, `properties`, `required`, `enum`, `items`, combinators, `additionalProperties`, and `deprecated`; excludes `description`, `title`, `example(s)`, and `x-*` | Implements the "structure only" fingerprint |
| `memory` | Reads and writes `memory/kibana-api-gap-scanner/state.json`: `groups{key: {fingerprint, verdict, reason, date}}` and `resources{name: last_triaged}` | The issue reference in a `file` record is the group key itself |
| `render` | Issue title `<resource>: <top-level field>`; body listing fields, OAS pointers, the `x-state` version when present, and a hidden marker `<!-- kibana-api-gap key=… fp=… -->` | The agent supplies only the reason paragraph |
| cmd `discover` | Takes slots; outputs `report.json` with the ordered triage queue (at most 15), rendered issues, and `run_agent` | Runs in pre-activation |
| cmd `record` | Takes `verdicts.json` and the report; writes a validated memory update. Rejects unknown keys, verdicts out of queue order, verdicts after the slots are filled with `file` verdicts, verdict lists that stop early (neither filling the slots with `file` verdicts nor covering the whole queue), and empty reasons; updates `last_triaged` (an RFC 3339 timestamp per resource, so same-day runs rotate) | Run by the agent, like `memory-record-from-report` |

**Workflow and CI components:**

| Component | Change |
|---|---|
| `.github/workflows/kibana-api-gap-scanner.md` (+ lock) | New. Imports `shared/setup-dev.md` and `shared/dispatch-research-factory.md`. Triggers: daily, push, and dispatch, with a non-cancelling `concurrency` group. GitHub keeps at most one pending run per group, so a burst of triggers collapses to one queued run. That is enough, because each run rescans everything. Safe outputs: `create-issue` (prefix `[kibana-api-gap] `, labels `kibana-api-gap` and `triaged`, max 3) and `noop`. |
| `.github/workflows/shared/dispatch-research-factory.md` | New. Mirrors the code-factory fragment and passes `research-factory-issue.lock.yml` as `producer-dispatch.js`'s existing third argument; `producer-dispatch.js` itself is unchanged. |
| `Makefile` `workflow-test` | Swaps `./scripts/kibana-spec-impact/...` for `./scripts/kibana-api-gaps/...` |
| Removed | `kibana-spec-impact.md` + lock, `scripts/kibana-spec-impact/`, the `memory/kibana-spec-impact` branch |

### D4. Data flow and state ownership

```
 durable inputs                     pre-activation                 agent job
 --------------                     --------------                 ---------
 repo @ SHA (Go, oas-filtered) --> discover --> report.json --artifact--> read queue
 memory branch state.json ------->     |          (queue <= 15,             |
 open kibana-api-gap issues -> slots ->|           rendered issues,         triage loop
                                       |           fingerprints)            |
                                                                     verdicts.json
                                                                            |
                                              record (validates vs report) -+
                                                    |                       |
                                         /tmp/gh-aw/repo-memory/.../state.json
                                                    |                create-issue x N
                                    gh-aw repo-memory push            (safe output)
                                                    v                       |
                                         memory branch updated      issues created
                                                                            |
                                                         dispatch-research-factory
```

| State | Lives in | Writer | Lifetime |
|---|---|---|---|
| Gap groups, fingerprints, queue, rendered issues | `report.json` (artifact) | `discover` only | One run |
| Verdicts and reasons | `verdicts.json` (agent temp) | Agent | One run; folded into memory |
| Group verdict records, `last_triaged` per resource | `memory/kibana-api-gap-scanner` branch, `state.json` | `record` only (invoked by the agent, pushed by gh-aw) | Durable; a record expires on fingerprint mismatch, evaluated by `discover` |
| Filed issues | GitHub | `create-issue` safe output only | Durable; never read back by the scanner |
| Slot count | Computed fresh each run from open `kibana-api-gap` issues | `issue-slots/compute.js` | One run |

- **Coverage is decided only in `discover`.** A group is skipped when `state.json` has a record for its key with an equal fingerprint.
- **Stale records are kept.** Records whose group no longer appears in discovery (the gap was fixed) are ignored; there is no garbage collection in v1.
- **The agent never decides facts.** Group membership, fields, and issue text come from the report; the agent adds only a verdict and reason, and `record` refuses keys outside the queue.
- **All-or-nothing runs** (gh-aw job gating). If the agent fails, no issues are created and no memory is pushed, so the next run re-triages cleanly with no duplicates. If the agent succeeds, memory is pushed after issue creation has been attempted. A blocked or failed `create-issue` therefore still leaves a `file` record, which is the accepted limitation in the spec.
- **`discover` is deterministic:** the same SHA and memory produce a byte-identical report, which the unit tests rely on.

### D5. Failure handling: fail loudly on unknown shapes, never guess absence

| Failure | Where | Handling | Result |
|---|---|---|---|
| `go/packages` load or type-check fails | `discover` | Non-zero exit; pre-activation fails and `run_agent` is never set | No triage, no issues, memory untouched |
| The lockstep walk finds a json tag with no OAS match, or the reverse | `walk` | **Tool error** naming the type, field, and operation; the run fails rather than guessing | As above; signals that kbapi and `oas-filtered.yaml` are out of sync or that `walk` has a gap |
| A kbapi method has no matching `New<Op>Request` constructor | `opindex` | Tool error | As above |
| Unsupported schema shape (e.g. `not`, exotic `allOf`) | `walk` | That **subtree** is treated as covered and listed in the report's `skipped_shapes` diagnostics; the run continues | No false gap is possible; visible in the artifact |
| No slots, or an empty queue | `discover` | `run_agent=false` with a gate reason, as in kibana-spec-impact | Agent skipped |
| A verdict for an unknown key, out of queue order, after the slots are filled, a list that stops early, or with an empty reason | `record` | The whole `verdicts.json` is rejected with a precise error; the agent fixes it and re-runs `record` | Memory only ever takes valid verdicts |
| Agent dies or times out mid-triage | Agent job | gh-aw processes no safe outputs and pushes no memory for a failed agent | No issues, memory unchanged; the next run re-triages those groups |
| Memory push fails after successful triage | gh-aw repo-memory | The job fails visibly | Red run. Issues from this run exist without records, so the next run may re-file them. Rare, because the push is a plain git push to a dedicated branch; the key marker makes any duplicate obvious. |
| `create-issue` fails or is blocked by threat detection | `safe_outputs` | None | Accepted limitation: the group stays covered until its fingerprint changes |
| Two triggers at the same time | Workflow | Queueing `concurrency` group (no cancel-in-progress) | Runs are serialised |

**Ordering inside the agent:** for each group, triage, then append the verdict, then call `create-issue` if the verdict is `file`. `record` runs once at the end, after all verdicts, and must succeed before the agent finishes. Running `record` incrementally would add nothing, because memory is only pushed when the whole agent job succeeds.

**Accepted ordering asymmetry:** `create-issue` is a gh-aw safe output that runs after the agent finishes, while `record` validates the verdicts inside the agent. A `record` failure therefore stops the agent, and gh-aw then processes no safe outputs, so no issue is created from an invalid run. The residual asymmetry is the reverse: issues for `file` verdicts are created before memory is pushed, so if the memory push fails the issues exist without records. This is accepted: the exposure is bounded by the slot cap (at most 3 issues per run), the key marker makes duplicates obvious, and the behaviour is verified live in task 4.3.

**Observability:**
- the report artifact: all groups, the queue, `skipped_shapes`, the gate reason, and resources with zero mapped operations;
- step summaries from `discover` and `record`;
- the memory branch as the audit trail.

### D6. Testing approach

| Layer | What | Where / how | Covers |
|---|---|---|---|
| **Unit, synthetic fixtures (main layer)** | `oas`, `walk`, `fingerprint`, `opindex`, `reach`, `usage`, `group`, `render`, `memory` | `go test ./scripts/kibana-api-gaps/...` against a synthetic fixture module that mimics the generated client, its OAS input, and representative resource shapes (including callback-style resources like `entitycore`) | Unused field becomes a candidate and a used one does not; nested grouping; deprecated exclusion; opaque passthrough; query params; union branches; unmatched-tag error; unsupported shape -> covered subtree plus diagnostic; fingerprint stable under description edits and changed by enum, property, or required changes; byte-identical reruns |
| **Unit, commands** | `discover` and `record` end to end on fixtures | Same package, with temp memory files | Zero slots -> no agent; ordering (never-triaged first, `last_triaged`, size, key tie-breaks); the 15 cap; skipping groups covered by memory; expiry on fingerprint change; `record` rejecting unknown keys, too many `file` verdicts, and empty reasons; idempotent re-record |
| **Integration, real repo** | `discover` over the actual provider and `oas-filtered.yaml` | A Go test **skipped under `-short`**, run by `make workflow-test` in CI | The walk handles every real shape with no unmatched tags; every `entitycore`-based resource maps at least one operation; the report is deterministic; a smoke check that known-covered fields of a well-covered resource produce no candidates |
| **Static / compile** | New workflow, shared fragment, removed workflow | The existing `gh aw compile` pairing check | Triggers, concurrency, labels including `triaged`, and the import compiling in |
| **Spike / signal-to-noise gate (manual)** | `discover` on three resources: SLO, synthetics monitor, Fleet agent policy | Run locally and reviewed by hand | Candidate counts, the share of union or passthrough noise, and a count of constant-only references for the v2 decision. **Go/no-go before the schedule is enabled.** |
| **Live, manual** | One `workflow_dispatch` run, plus one deliberately failed agent run | Real runs | Triage reasons, the rendered issue, the research-factory dispatch, memory branch contents; confirms that a failed agent run creates no issues and pushes no memory |

- **Triage quality** cannot be unit-tested. The design reduces it to one verdict and reason over pre-rendered facts, and records every suppression with its reason for review.
- **Real-repo variety** is covered by the integration test, which fails loudly on unknown shapes instead of silently mis-reporting.
- No live Elastic Stack is needed.

### D7. Rollout in a single PR, verified live after merge

```
 step 0  spike: discover (resource-filtered) on 3 resources -> signal-to-noise review
         (conditional go, user-approved; see "Spike outcome" below)
           |
 step 1  ONE PR: scripts/kibana-api-gaps + tests, shared dispatch-research-factory
         fragment, scanner workflow with schedule + push + dispatch,
         REMOVE kibana-spec-impact (workflow, lock, script, Makefile line),
         docs: continuous-quality-workflows.md + factory-workflows.md
           |
 step 2  after merge: live dispatch runs (task 4.3): verify triage, issues,
         dispatch, memory branch, failed-agent all-or-nothing; tune the prompt if needed
           |
 step 3  after merge, manual: delete the memory/kibana-spec-impact branch;
         archive this change so the ci-kibana-spec-impact-issues removal reaches the main specs
```

- **What shipped and the trade-off:**
  - The original plan used two PRs: dispatch-only first, then schedule plus the `kibana-spec-impact` removal. Review and the spike's reflection and merge fixes made the scanner deterministic and unit-tested, so it shipped as one PR instead.
  - Cost: the live behaviour (triage in order, issue creation, dispatch, memory, failed-agent all-or-nothing) is first exercised after merge, with the schedule already enabled. Exposure is bounded: at most 3 open `kibana-api-gap` issues (slots), at most 15 groups triaged per run, and at most 3 research-factory dispatches per day.
  - Benefit: there is never a coverage gap and never two scanners filing at once, and no intermediate state to maintain.
- **Compatibility:**
  - **First-run backlog:** the first scheduled run sees the whole historical backlog; the slot cap (at most 3 open) and the triage budget (15 per run) throttle it with no seeding.
  - **Memory bootstrap:** the branch is created as an orphan on the first `record` (`create-orphan: true`).
  - **Existing `[kibana-spec-impact]` issues** stay open. They lack the `kibana-api-gap` label, so they do not consume slots. Overlap with new gap issues is possible for a few resources; it is visible and harmless.
  - **research-factory load:** at most 3 dispatches per day at steady state. The `research-factory-critic-gate` change is recommended to land first.
- **Rollback:**
  - If live verification (task 4.3) shows a problem: disable the schedule and push triggers (leaving dispatch), or disable the scanner; fix forward.
  - To restore the old scanner: revert the PR. The `kibana-spec-impact` helper comes back from git and its baseline is re-seeded from current `main`. Its old dedupe fingerprints (and the deleted memory branch, if already removed) are lost, so it may re-file a few entities once.

**Spike outcome (task 2.3):** conditional go, user-approved. The whole-repository group count fell from 1,632 to 1,051 after treating types handed to `reflect` as opaque-covered, and to 244 after merging same-field groups across create, update, and union bodies (keeping query-parameter and body groups apart); the three spike resources (SLO, synthetics monitor, Fleet agent policy) produced 21 groups, matching the 21 logical fields counted by hand, with about 57% plausible real gaps and the rest in the triage suppress categories.

## Risks / Trade-offs

| Concern | Risk | Mitigation |
|---|---|---|
| **Performance of `discover`** | A full-module load with types, plus a 12 MB generated `kibana.gen.go` and a large `oas-filtered.yaml`, could take minutes and several GB of RAM | One load per run, reused throughout; standard `ubuntu-latest` runners (16 GB) suffice; the integration test is skipped under `-short`. If load time exceeds about 5 minutes, load only the resource packages and their dependencies. |
| **Reference graph over-approximates** | Resolving interface calls to all implementations, and following every function reference, pulls in more code than a resource really executes. Usage is over-counted, so some real gaps are missed. Operations can also be over-counted, so a resource may be credited with an operation a shared helper references for another resource. | Missing gaps is the safe direction (no false absence). Operation over-counting yields candidates that are true statements about reachable code, and triage can suppress them as irrelevant. The report lists resources with zero mapped operations and per-resource operation counts, so outliers show up during the spike. |
| **Triage model and prompt** | Weak suppression criteria make the filter too lenient or too aggressive | The same OpenRouter Claude setup as schema-coverage (`anthropic/claude-sonnet-5`, `--effort high`). The prompt lists concrete suppress categories (server-managed or computed state, internal or telemetry fields, runtime actions that make no sense as IaC, fields covered by an opaque attribute that `usage` missed) and says "if in doubt, file". Every suppression and its reason is in memory for review. |
| **Prompt injection via the spec** | OAS descriptions come from upstream Kibana and are shown to the triage agent | Low risk, because upstream is Elastic's own spec. The agent is read-only (`contents: read`, `issues: read`); writes are limited to `create-issue` (max 3) and `record`, whose input is validated. An injected "file everything" is still capped by slots and keys. |
| **Generator coupling** | An oapi-codegen upgrade changes how unions, maps, or constructors are emitted | `walk` and `opindex` fail loudly on unknown shapes or missing constructors, and the real-repo integration test catches this in the PR that regenerates kbapi |
| **Cost** | LLM triage on every run | At most 15 short triage judgements per run, and none when no slots are available. Once the backlog is gated by open issues, steady-state cost is near zero. The daily credits guard applies. |
| **Report size** | Hundreds of groups in the artifact | Only the queue (at most 15) and its rendered issues go to the agent; the full group list stays in the artifact for diagnostics |
| **Silent loss when issue creation fails** | A `file` record stands but no issue exists, so the gap stays hidden until its fingerprint changes | Accepted limitation (spec). Threat-detection blocks and safe-output failures are visible in the run. |
| **Duplicate issue after a failed memory push** | A run whose triage succeeded but whose memory push failed leaves issues without records, so the next run may re-file them | Rare (plain git push to a dedicated branch); the run fails visibly; the key marker in every issue makes duplicates easy to spot |
| **Throughput is bounded by open issues** | At most 3 open `kibana-api-gap` issues at any time, so the backlog drains only as issues close | Intended backpressure, matching the other scanners; the cap is a single constant |
