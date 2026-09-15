## Context

`elasticstack_kibana_spaces` (`internal/kibana/spaces`) is a read-only data source that lists every Kibana space via the Get All Spaces API and maps each entry into the `spaces` computed list attribute (`openspec/specs/kibana-spaces/spec.md`, REQ-004–REQ-005). Neither the Kibana API contract nor the provider's mapping code makes any promise about list order — the spec never mentions ordering, and the mapping loop in the data source simply forwards whatever order the API returns.

`internal/kibana/spaces/data_source_test.go` was nonetheless written assuming the built-in `default` space is always at index 0:

- `TestAccSpacesDataSource` checks `spaces.0.id == "default"` plus several other `spaces.0.*` fields for the default space's full attribute set.
- `TestAccSpacesDataSource_multipleSpaces` checks `spaces.0.id == "default"` (with a code comment asserting "Default space is always the first element."), while its custom space assertions already use the order-independent `testCheckSpaceAttrByID(spaceID, attr, value)` helper.
- `TestAccSpacesDataSource_noDescription`, `TestAccSpacesDataSource_withImageURL`, and `TestAccSpacesDataSource_withKibanaConnection` each check `spaces.0.id == "default"` once, alongside id-based checks for their own custom space where applicable.

On Kibana `9.6.0-SNAPSHOT`, the API (or the provider's flatten) started returning `acc_test_space` before `default`, breaking `TestAccSpacesDataSource_withImageURL` in CI ([run 34672008631](https://github.com/elastic/terraform-provider-elasticstack/actions/runs/34672008631)). The other four tests share the identical assumption and are equally exposed; they simply have not been observed failing yet, and 9.6 snapshot CI jobs run `continue-on-error` so the failure did not block the run.

`testCheckSpaceAttrByID` already exists in the same file and is order-independent: it scans `spaces.#` and matches by `spaces.<i>.id`, so it is the natural fix for every remaining `spaces.0.*` assertion against `default`.

## Goals / Non-Goals

**Goals:**
- Make every acceptance test assertion in `internal/kibana/spaces/data_source_test.go` about the `default` space (or any other named space) order-independent, using `testCheckSpaceAttrByID` or an equivalent id-based lookup.
- Document in `openspec/specs/kibana-spaces/spec.md` that the `spaces` list has no ordering contract, so future tests are written correctly from the start and this class of regression is caught in spec review.
- Fix all five affected tests in one pass, not just the one that has already failed in CI, since they share the exact same bug.

**Non-Goals:**
- Changing the data source's read/mapping behavior, schema, or API usage. The `elasticstack_kibana_spaces` data source itself is not buggy; only the tests are.
- Introducing a stable sort order into the data source's read path. The issue explicitly frames this as optional ("optionally sort the data-source list if consumers also assume a stable order"); no evidence in the issue indicates practitioners depend on order, and adding one would be an unrequested behavior change to production code that this change does not propose.
- Auditing acceptance tests outside `internal/kibana/spaces/data_source_test.go` for the same pattern. Other data sources are out of scope for this change.

## Decisions

- **Reuse `testCheckSpaceAttrByID` rather than adding a new helper.** It already exists in the same file, is already exercised by three of the five affected tests for the custom space, and does exactly what is needed for the `default` space too: scan `spaces.#` and match by `id`. Introducing a second, parallel helper would duplicate logic for no benefit.
- **Fix all five affected tests, not only `TestAccSpacesDataSource_withImageURL`.** All five share byte-for-byte the same `spaces.0.id == "default"` assumption; fixing only the one that failed in CI would leave the same latent bug in the other four, ready to resurface on the next Kibana version that changes ordering.
- **Do not add ordering to the data source.** The provider has never documented or guaranteed spaces order, and no reported practitioner impact depends on it. Scope stays limited to correcting the test assumption, consistent with the issue calling this "a test assumption, not a product bug."
- **Document the no-ordering contract in the spec, not just fix the tests silently.** Recording "the `spaces` list has no ordering guarantee; tests MUST assert by id" in `openspec/specs/kibana-spaces/spec.md` gives future contributors (and reviewers) a normative reason to reject a new `spaces.0.*`-style assertion, rather than relying on tribal knowledge of this one incident.

## Risks / Trade-offs

- [Risk] If a future Kibana version changes ordering again in a way not covered by `testCheckSpaceAttrByID`'s `spaces.#` scan (for example, pagination or a changed list-key convention), the fix could still be incomplete. Mitigation: `testCheckSpaceAttrByID` already handles an arbitrary count and arbitrary position, so this is a low-probability residual risk, not a known gap.
- [Risk] Editing `TestAccSpacesDataSource`'s multiple chained checks for the default space (color, disabled_features, etc.) to go through per-attribute id lookups is slightly more verbose than the single fixed-index block it replaces. Mitigation: this is a one-time verbosity cost in exchange for correctness; `testCheckSpaceAttrByID` already establishes the pattern elsewhere in the same file.
