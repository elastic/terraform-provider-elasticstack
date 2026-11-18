## ADDED Requirements

### Requirement: Daily AI-credits guardrail override (REQ-018)

The authored workflow frontmatter for `.github/workflows/openspec-verify-label.md` SHALL set `max-daily-ai-credits: -1`, disabling gh-aw's per-workflow rolling 24-hour AI-credits guardrail (default `5000` AI Credits/day) for this workflow. This override SHALL be documented inline via a frontmatter comment explaining why the daily cap is disabled, co-located with the existing per-run `max-ai-credits: -1` override and its rationale.

#### Scenario: Accumulated daily credit usage does not block activation

- **GIVEN** this workflow's aggregated AI Credits usage across completed runs in the trailing 24-hour window would otherwise exceed the default `5000` AI Credits guardrail threshold
- **WHEN** a `verify-openspec` labeled pull request event triggers this workflow
- **THEN** the deterministic pre-activation gate SHALL NOT skip the agent job on account of the daily AI-credits guardrail

#### Scenario: Lock file stays paired with the override

- **GIVEN** the `max-daily-ai-credits: -1` frontmatter override is added to the workflow source
- **WHEN** the change is merged
- **THEN** the regenerated `.lock.yml` produced by `gh aw compile` SHALL reflect the disabled daily guardrail, consistent with REQ-001
