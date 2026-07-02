## MODIFIED Requirements

### Requirement: Stream vars semantic equality strips server-managed `data_stream.*` keys (REQ-DATASTREAM-VARS)

Fleet 9.5 injects server-managed `data_stream.type` and `data_stream.dataset` keys into the compiled stream `vars` of input-type packages. These keys are synthesised by Fleet's enrichment pipeline and are not user-configurable. The provider SHALL normalise stream `vars` by stripping all server-managed keys before performing semantic equality comparisons so that their presence in the API response does not trigger a `"Provider produced inconsistent result after apply"` error.

The server-managed keys that SHALL be stripped are:

- `data_stream.type`
- `data_stream.dataset`

Stripping SHALL occur on both sides of the comparison (plan-side and API-side stream `vars`) immediately before the `StringSemanticEquals` call in the `compareStreams` function. The stripped values are used only for comparison; they SHALL NOT be removed from the persisted state or from the value sent to the API. If the input JSON is null or unknown, the stripping helper SHALL return the input unchanged. Diag errors from the stripping helper SHALL abort the comparison and propagate to the caller.

The stripping is applied at stream level only. Input-level `vars` and the top-level `vars_json` attribute are unaffected.

#### Scenario: Stream vars with injected `data_stream.type` are semantically equal to plan-side vars without it

- GIVEN a stream whose plan-side `vars` JSON is `{"project_id":"my-project","subscription_name":"my-sub","tags":["forwarded"],"topic":"my-topic"}`
- AND the API-returned stream `vars` is `{"data_stream.dataset":"gcp_pubsub.generic","data_stream.type":"logs","project_id":"my-project","subscription_name":"my-sub","tags":["forwarded"],"topic":"my-topic"}`
- WHEN `compareStreams` evaluates semantic equality
- THEN the comparison SHALL return `true` (semantically equal)
- AND Terraform SHALL NOT produce a `"Provider produced inconsistent result after apply"` error

#### Scenario: Stream vars without server-managed keys are compared normally

- GIVEN a stream whose plan-side `vars` JSON is `{"threshold":42}` and the API-side `vars` is `{"threshold":99}`
- WHEN `compareStreams` evaluates semantic equality
- THEN the comparison SHALL return `false` (the user-defined key differs)

#### Scenario: Strip helper is a no-op on null/unknown vars

- GIVEN a stream `vars` value that is null or unknown
- WHEN the server-managed key stripping helper is called
- THEN the input SHALL be returned unchanged with no diagnostics

### Requirement: `defaults` attribute null ⇄ populated-object transition is semantically equal (REQ-DEFAULTS-COMPUTED)

Fleet 9.5 now populates a `defaults` block in the package policy GET response for input-type packages. Prior to 9.5, this block was not returned. The `defaults` attribute is purely computed from package information — the user never configures it directly. When the planned value of `defaults` is `null` (because the attribute was absent from the plan) and the API returns a populated `defaults` object, the provider SHALL treat this transition as semantically equal and SHALL NOT produce a `"Provider produced inconsistent result after apply"` error.

The semantic equality rule is applied in `InputValue.ObjectSemanticEquals`: if either side's `defaults` is null or unknown, the `defaults` component SHALL be skipped in the equality check (treated as equal). When both sides have a fully known `defaults`, the equality check SHALL also treat them as semantically equal (since `defaults` is purely server-managed and any value returned by the API is a valid resolved state for a null plan). This ensures robustness against future Fleet enrichment changes that alter the `defaults` content across applies.

No schema version bump is required — `defaults` is already `Computed: true` in the schema. No state upgrader is required. The fix is entirely within the semantic-equality comparison layer.

#### Scenario: `defaults` goes from null in plan to populated object after apply

- GIVEN a plan where `inputs["gcp-gcp-pubsub"].defaults` is `null`
- AND the Fleet API returns a populated `defaults` object after apply: `{"streams":{"gcp_pubsub.gcp":{"enabled":true,"vars":{"data_stream.dataset":"gcp_pubsub.generic","tags":["forwarded"]}}},"vars":null}`
- WHEN `InputValue.ObjectSemanticEquals` evaluates the input
- THEN the comparison SHALL return `true` (semantically equal)
- AND Terraform SHALL NOT produce a `"Provider produced inconsistent result after apply"` error on the `defaults` attribute

#### Scenario: Fully-known `defaults` on both sides does not block equality

- GIVEN a plan where `inputs["foo"].defaults` is a known populated object
- AND the API returns a `defaults` object with different content
- WHEN `InputValue.ObjectSemanticEquals` evaluates
- THEN the comparison SHALL treat the `defaults` component as semantically equal (since `defaults` is server-managed) and SHALL continue to compare `vars` and `streams`

#### Scenario: 9.4 behavior unchanged (defaults remains null)

- GIVEN a Fleet 9.4 API response where `defaults` is not returned
- WHEN the resource is read and `defaults` remains null in state
- THEN no plan diff SHALL appear on subsequent plans
- AND the stripping and defaults-equality fixes SHALL be no-ops (the keys are absent; null ⇄ null is already equal)

### Requirement: Acceptance test coverage for Fleet 9.5 enrichment (REQ-DATASTREAM-ACC)

The acceptance test suite SHALL include (or continue to pass, if already present) test cases that exercise both failure modes fixed by this change:

1. `TestAccResourceIntegrationPolicyGCPPubSub` — verifies that an input-type package policy can be applied and subsequently planned with no diff on a Fleet 9.5 stack, where stream `vars` contain injected `data_stream.*` keys and `defaults` is populated in the API response.
2. `TestAccResourceIntegrationPolicySecrets` (both subtests) — verifies that secret-valued input vars and multi-valued secrets on an input-type package policy apply cleanly on Fleet 9.5 without post-apply inconsistency.

Both tests SHALL be skipped when the target Elastic Stack version is strictly below `9.5.0`.

#### Scenario: GCP PubSub policy apply and re-plan on 9.5 produce no diff

- GIVEN `TestAccResourceIntegrationPolicyGCPPubSub` running against a 9.5.0+ stack
- WHEN the Terraform apply and a subsequent plan run
- THEN no `"Provider produced inconsistent result after apply"` error occurs
- AND the subsequent plan SHALL show no changes
