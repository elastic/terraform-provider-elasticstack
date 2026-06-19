## Context

The shared `ml.IDValidator()` helper at `internal/elasticsearch/ml/idvalidator.go:44` enforces two
constraints on all ML resource ID attributes:

1. `stringvalidator.LengthBetween(1, 64)` — length 1 to 64 characters.
2. `stringvalidator.RegexMatches(pathIDRegexp, ...)` — lowercase alphanumeric, hyphens,
   underscores, dots; must start and end with alphanumeric.

The Elasticsearch API documentation for `PUT /_ml/datafeeds/{datafeed_id}` and
`PUT /_ml/filters/{filter_id}` does not specify a maximum length — only the character-class
restriction. The 64-character limit is a conservative provider-side addition with no documented
API basis for datafeed IDs or filter IDs.

The human direction for this proposal explicitly specifies **Approach B** — introduce a new,
narrowly-scoped `IDValidatorWithoutLength()` validator for datafeed IDs and filter IDs, leaving
the shared `IDValidator()` unchanged for other ML resources (`job_id`, `calendar_id`, etc.).

## Goals / Non-Goals

**Goals:**
- Allow users to manage existing Elasticsearch ML datafeeds and filters whose IDs exceed 64
  characters through Terraform.
- Preserve the character-class validation (regex) for `datafeed_id` and `filter_id`.
- Keep `IDValidator()` unchanged so other ML resources are not affected.

**Non-Goals:**
- Modify `IDValidator()` or any other ML resource validator.
- Remove any character-class validation.
- Address the 64-character limit for `job_id`, `calendar_id`, `filter_id` in calendar resources,
  or other ML identifiers — those remain separate questions.
- Backport to older provider versions.

## Decisions

- **New validator name**: `IDValidatorWithoutLength()` — descriptive, parallel to `IDValidator()`,
  signals the intentional omission of the upper-bound length constraint. It still enforces
  `stringvalidator.LengthAtLeast(1)` (non-empty) and the character-class regex.
- **Scope**: `datafeed_id` in `ml_datafeed` and `ml_datafeed_state`, and `filter_id` in
  `ml_filter`. These are the three attributes where Elasticsearch has confirmed no length
  restriction applies.
- **IDValidator() stays unchanged**: preserving correctness for resources where the 64-char limit
  may still be valid (no confirmed API evidence it is wrong for `job_id`, `calendar_id`, etc.).
- **Spec updates**: Delta specs for `elasticsearch-ml-datafeed`,
  `elasticsearch-ml-datafeed-state`, and `elasticsearch-ml-filter` remove the 64-character
  upper-bound from their ID attribute descriptions and validation requirements.

## Open questions

- Does Elasticsearch enforce any upper-bound length restriction on anomaly detection `job_id`,
  `calendar_id`, or `filter_id` in calendar resources that differs from datafeeds and filters?
  Checking the ES source (e.g. `MlStrings` / `MlField.java` in the ES monorepo) would confirm
  whether the 64-char limit was ever intentional for any ML entity type. This answer would
  determine whether a future change should apply a similar fix to `IDValidator()` itself.
- Are there acceptance tests that assert a 64-char `datafeed_id` or `filter_id` is the upper
  limit? If so, they must be updated during implementation.

## Risks / Trade-offs

- [Risk] If ES does enforce 64-char limits for `filter_id` in some undocumented path, relaxing
  the validator would allow users to declare IDs that ES then rejects at the API level.
  Mitigation: the research found no evidence of such a limit; the API error would surface clearly
  on apply rather than silently misconfiguring the resource.
- [Risk] The `datafeed_state` resource fix is coupled with `datafeed` — they share `datafeed_id`
  semantics. Both must be updated in the same change to avoid a split where users can create but
  not manage state for long-named datafeeds.
