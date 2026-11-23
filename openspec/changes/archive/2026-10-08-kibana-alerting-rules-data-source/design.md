## Context

See proposal.md for motivation. Current state that shapes the approach:

- `elasticstack_kibana_alerting_rule` lives in `internal/kibana/alertingrule` and maps `last_execution_status` / `last_execution_date` from the get-rule response. `kibanaoapi.GetAlertingRule` already fetches one rule and converts it through `ConvertResponseToModel`, which already parses the execution timestamp.
- Kibana data sources in this repo are built on `entitycore.NewKibanaDataSource`. The envelope rejects an empty resource ID before calling the read callback. List-style data sources (`spaces`, `tag`, `security_entity_store/entities`) satisfy this by returning a constant sentinel from `GetResourceID()` and doing the real work in the read callback. A `found=false` result from the callback becomes the envelope's standard not-found error.
- The generated Kibana client already exposes the rules find endpoint (`GetAlertingRulesFindWithResponse`) with `filter`, `sort_field`, `sort_order`, `page` and `per_page`. No find wrapper exists in `kibanaoapi` yet.
- Behaviour verified against a local stack: the find endpoint's default page size is 10, so unpaged reads silently truncate; `filter` uses the `alert.attributes.*` prefix; `sort_field=name` works in both directions; execution fields (`execution_status`, `scheduled_task_id`) are present on find results; malformed KQL produces an HTTP 500 rather than a 4xx. Kibana imposes no fixed `per_page` cap; the effective limit is the Elasticsearch result window (`page x per_page` must not exceed 10,000), above which Kibana returns an error. An unknown space returns HTTP 403. These were observed on Kibana 9.5.5 only.
- The Plugin Framework passes same-version state through with undefined attributes ignored, but the JSON returned by a custom state upgrader is decoded strictly against the current schema. The existing v0 (SDK-format) upgrader returns raw JSON, so it fails on keys the schema no longer defines unless it removes them.
- The `CHANGELOG.md` Unreleased section is generated from merged PRs and is not edited by hand.

## Goals / Non-Goals

**Goals:**
- Follow the repo's Kibana data source convention so connection, client and version plumbing are inherited rather than reimplemented.
- Reuse the existing single-rule fetch and response conversion so execution fields keep their existing parsing.
- Keep the resource package independent of the new data source package.

**Non-Goals:**
- No new abstraction over the envelope and no change to `entitycore`.
- No retry logic beyond what the provider's HTTP client already does.
- No workaround for rules changing between page requests within one read.
- No typed filter inputs or consumer filter (see proposal Non-Goals).

## Decisions

### Architecture

```
 data "elasticstack_kibana_alerting_rules"
        |
        v
 entitycore.NewKibanaDataSource[model]      (envelope: config -> client -> read -> state)
        |   resource ID = rule_id, or a fixed sentinel when searching
        v
 read callback (new package under internal/kibana)
        |
        +-- rule_id set ---> kibanaoapi.GetAlertingRule         (existing, single GET)
        |                          |  not found -> found=false -> envelope's not-found error
        |
        +-- else -----------> kibanaoapi.FindAlertingRules      (NEW, paged find)
                                   |  filter passed through unmodified
                                   |  sort_field=name, sort_order=asc
                                   |  loops pages until collected >= total
                                   v
                 both paths -> models.AlertingRule -> rules[] elements
                              (via the existing ConvertResponseToModel)
```

- A new package alongside `internal/kibana/alertingrule`, registered in `provider/plugin_framework.go`.
- The data source is built on the Kibana envelope. The model's `GetResourceID()` returns the configured `rule_id` when set and a constant sentinel otherwise, so the envelope's non-empty-identity check passes in search mode. The read callback chooses lookup or search by whether `rule_id` is configured on the model, never by the resource ID the envelope resolved; the single-rule GET is therefore never called with the sentinel.
- `space_id` is resolved in the read callback with the shared effective-space helper (empty becomes `default`) and the effective value is written back to `space_id` in state, as the `tag` data source does, so state, `id` and the queried space agree when `space_id` is omitted.
- The envelope turns `found=false` into a not-found error. Lookup returns `found=false` only for an unknown rule. A successful search, including one that matches no rules, returns `found=true` with an empty `rules` list.
- A lookup by `rule_id` uses the single-rule GET, with the ID in the URL path. The ID is never embedded in a query expression, so there is no query-escaping surface and no dependence on Kibana's internal saved-object ID format.
- Input validation is done by schema validators at validate time, before any API call. It rejects only `rule_id` and `filter` both set, and an empty configured `rule_id` or `filter`. Setting neither is valid (it means all rules), so an exactly-one-of validator must not be used.

### Component structure

| Unit | Responsibility | Interface |
|---|---|---|
| New data source package (`internal/kibana/alertingrules`) | Data source constructor and registration | `NewDataSource()` calls the envelope with the schema and read callback |
| Data source model | Holds `space_id`, `rule_id`, `filter`, `id`, `rules` and the connection list; supplies the envelope identity accessors | `GetResourceID` returns `rule_id`, or the sentinel when unset |
| Data source schema | Declares inputs and the Computed `rules` nested attribute; validators enforce mutual exclusion and non-empty strings | Plain framework schema; the envelope adds `kibana_connection` |
| Read callback | Chooses lookup or search from the model, calls the client, maps results to `rules[]`, sets `id` | Returns `(model, found, diags)` per the envelope contract |
| Rule element mapper | Converts a `models.AlertingRule` into a `rules` element; formats `last_execution_date` as RFC3339; maps an empty tag list to an empty set | Pure function |
| `kibanaoapi.FindAlertingRules` (new) | Pages through the find endpoint with sort and filter, converts each element, returns all of them | `(ctx, client, spaceID, filter) -> ([]models.AlertingRule, diags)` |
| `kibanaoapi.GetAlertingRule` (existing) | Single-rule fetch | Unchanged; 404 maps to not found |

- The resource package loses only the two schema attributes, the two model fields and the code that sets them. It does not import the new package.
- `ConvertResponseToModel` is unchanged; it still parses the execution fields because the data source consumes them.
- The data source gets a docs template, an example and a generated docs page like other data sources.

### Data flow

```
 Terraform config (space_id, rule_id | filter)
        |
        v
 [validate]  rule_id xor filter, non-empty        (no API call on failure)
        |
        v
 [envelope]  resolve client from kibana_connection; resolve space (default "default")
        |
        v
 [read]  rule_id? --yes--> GET rule in space ----> 404 => found=false => not-found error
        |                                    \--> 1 rule
        \-no--> find(space, filter, sort=name asc, page=N)
                   loop: append page; stop when collected >= total or a page is empty
        |
        v
 [map]   each rule -> {id, name, rule_type_id, consumer, enabled, tags,
                       scheduled_task_id, last_execution_status,
                       last_execution_date (RFC3339)}
        |
        v
 [state] rules = mapped list; id = "<space>/<rule_id>" or "<space>"
```

- The data source holds no state between reads. Each refresh recomputes `rules` and `id` from Kibana, so execution fields are current at read time. The data source only issues GET requests.
- The client sets `per_page` explicitly to a provider constant of 100 and pages until the collected count reaches `total` or a page comes back empty; the empty-page guard prevents a loop if rules are deleted mid-read. Because Kibana cannot page past 10,000 results, a first page reporting a `total` above 10,000 fails the read with a diagnostic asking the practitioner to narrow `filter`, instead of returning a truncated list or hitting Kibana's window error part-way.
- Results keep Kibana's server-side ordering by `name`; rules with equal names have no guaranteed relative order.
- `last_execution_date` reuses the existing timestamp parse and is formatted in UTC with a fixed three-digit fractional second (`2006-01-02T15:04:05.000Z`), which reproduces Kibana's own value byte for byte so state never drifts between equivalent representations. The resource's old layout (`2006-01-02 15:04:05.999 -0700 MST`) is not reused. A missing or unparseable time maps to null.

### Failure handling

| Failure | Where it surfaces | Behaviour |
|---|---|---|
| Both `rule_id` and `filter`, or an empty string for either | Schema validators | Error at validate time; no API call |
| Unknown `rule_id` (404) | Client maps to not found | Envelope not-found error naming the data source, ID and space |
| Unknown or inaccessible space | API 403 (observed) or 404 during read | Error diagnostic; never treated as an empty search result |
| Authorization failure (401 or 403) | Client | Error diagnostic with status and body; no partial results |
| Filter rejected by Kibana | Client | Error diagnostic including the response body and naming `filter` as the likely cause, because malformed KQL is reported as an HTTP 500 |
| Failure on page N of a multi-page search | Client | The whole read fails; earlier pages are discarded |
| Rule that never executed | Mapper | Null `last_execution_status` and `last_execution_date`, no error |
| Unparseable timestamp | Mapper | `last_execution_date` null, no error |
| Empty search result | Read | `found=true` with an empty `rules` list, no error |
| Reported total above 10,000 | Client, after the first page | Error diagnostic asking to narrow `filter`; no truncated list |
| Search endpoint unavailable on an older Kibana | Client | The API error is surfaced as an error diagnostic; no version gate is applied |

HTTP errors use the repo's existing diagnostic helpers so messages match other Kibana calls. No retries are added: a failed data source read surfaces immediately and the next plan retries.

### Testing approach

| Layer | Covers | Where |
|---|---|---|
| Unit: element mapper | RFC3339 formatting, null handling for a never-run rule, empty tags to an empty set | New package |
| Unit: schema and validators | `rule_id` / `filter` mutual exclusion, empty-string rejection, and that setting neither is accepted | New package |
| Unit: read callback routing | Lookup is chosen by configured `rule_id`, not by the resolved resource ID; the search path never calls the single-rule GET with the sentinel; a successful empty search returns `found=true`; the effective `default` space is written to state | New package |
| Unit: `FindAlertingRules` | Multi-page paging with an explicit `per_page`, empty-page stop, the over-10,000 total error, failure on page N discarding results, error mapping, using an `httptest` server in place of Kibana | `internal/clients/kibanaoapi`, alongside existing alerting rule client tests |
| Contract: envelope | The `entitycore_contract_test.go` pattern used by sibling packages, confirming the model satisfies the identity interface | New package |
| Acceptance: data source | Lookup by ID, search by filter, empty result, unknown ID error, ordering by name | New package `acc_test.go` with `testdata` configs |
| Unit: state upgrade | The existing v0 (SDK-format) upgrade succeeds on state that contains the two removed attributes and drops them | `internal/kibana/alertingrule/state_upgrade_test.go` |
| Acceptance: resource | Drop assertions on the removed attributes; add a check that a follow-up plan is empty after the rule executes; add a step that reads current-version prior-shape state (with the two attributes) with the new schema | `internal/kibana/alertingrule/acc_test.go` |

Paging beyond one page is impractical in an acceptance test, so it is covered by the unit test with a stub server. Acceptance assertions on `last_execution_*` must tolerate null because a new rule may not have executed yet.

### Migration and rollout

- The removal is breaking and ships in a release that permits breaking changes, with the data source in the same PR so the replacement always exists where the removal does.
- The generated changelog reads merged PRs and is not edited by hand, so the PR description must name both removed attributes, mark the change breaking and show a data source example (a requirement in the resource delta spec).
- Schema `Version` stays at 1 and no new upgrader is added. State already at version 1 is passed through by the framework with the removed attributes ignored (verified in the framework source and re-checked by the prior-shape acceptance step). State at version 0 goes through the existing SDK-format upgrader, whose output is decoded strictly, so that upgrader is changed to delete the two keys before returning. This is a committed decision, not a fallback.
- Resource docs drop the two attributes and note the data source as the replacement; the new data source gets a doc page, example and template.
- Rollback is reverting the PR; the attributes are recomputed from Kibana on the next read, so no configuration or data is lost.
- The canonical spec is not edited while the change is active. After the delta specs are synced to `openspec/specs/` (sync or archive), the two attributes are removed by hand from the Schema listing of `openspec/specs/kibana-alerting-rule/spec.md`, because deltas cannot edit that section, and the resulting canonical spec is validated.

### Security and performance

- `filter` is sent to the practitioner's own Kibana using the provider's credentials, so it cannot read anything the provider could not already read.
- The data source omits `params`, `actions` and artifacts, which can hold sensitive values, so exposure is limited to names, tags, identifiers and run status.
- A search with no `filter` in a large space reads every page on each refresh. The docs recommend setting `filter` for large spaces.

## Risks / Trade-offs

- [Breaking removal of two attributes] → Called out in the PR description and migration notes; the data source ships in the same PR.
- [Offset paging while rules change] → A rule may be skipped or repeated within one read; accepted as inherent to offset paging.
- [Raw KQL `filter`] → Users can write invalid or version-sensitive queries; Kibana's error body is surfaced, and typed filters can be added later without breaking configs.
- [Sentinel resource ID for search mode] → A small workaround for the envelope, consistent with the other list-style Kibana data sources.
- [Sorting by `name`] → Rules with equal names have no guaranteed relative order.
- [State compatibility] → Version-1 state is tolerated by the framework; the version-0 upgrader is changed to drop the removed keys; both are covered by tests.
- [Result window limit] → Searches matching more than 10,000 rules fail with a diagnostic asking to narrow `filter`; they cannot be listed in full.
- [No minimum Kibana version enforced] → Behaviour was observed on Kibana 9.5.5 only; the find endpoint belongs to the same alerting API family as the resource, and on a stack without it the API error is surfaced.
- [Large unfiltered searches] → Documented guidance to use `filter`; no design change.
