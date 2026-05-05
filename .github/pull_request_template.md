## Changelog
Provide a changelog entry in the exact format below so it passes the PR changelog check.

Customer impact: <none, fix, enhancement, breaking>
Summary: <single line summary>

Expected format:
- `Customer impact:` must be one of `none`, `fix`, `enhancement`, or `breaking`
- `Summary:` is required unless `Customer impact: none`
- `### Breaking changes` is required only when `Customer impact: breaking`
- `<!-- /breaking-changes -->` can optionally end the `### Breaking changes` block early to prevent trailing PR content from entering the changelog

Good example:
```md
## Changelog
Customer impact: enhancement
Summary: Add support for configuring Kibana alert snooze schedules.
```

Good example (breaking):
```md
## Changelog
Customer impact: breaking
Summary: Remove deprecated role mapping compatibility behavior.
### Breaking changes
Deprecated role mapping compatibility behavior has been removed.
<!-- /breaking-changes -->
```

> **Delete the examples above before submitting.** The parser matches the first `Customer impact:` and `Summary:` lines it finds.

## Detailed changes
Describe the intent of this PR and the approach taken to implement it. Include any notable design decisions, tradeoffs, limitations, or follow-up work that would help reviewers understand the change.
