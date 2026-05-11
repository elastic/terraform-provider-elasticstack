## Context

The markdown panel API has two storage modes — inline ("by value") and linked library item ("by reference") — modeled as an `anyOf` in `kbn-dashboard-panel-type-markdown.config`. The TF resource currently only handles the by-value subset and is missing two by-value fields (`settings.open_links_in_new_tab`, `hide_border`). Mirroring the `lens_dashboard_app_config` pattern (which already exposes `by_value` and `by_reference` sub-blocks) keeps the resource internally consistent.

## Goals / Non-Goals

**Goals:**
- Surface the full markdown API surface through typed attributes.
- Mirror the `by_value` / `by_reference` shape used by `lens_dashboard_app_config` so the union pattern is recognizable.
- Keep `config_json` available for markdown as REQ-010 already specifies.

**Non-Goals:**
- Support a "library item" resource for markdown content. Practitioners would still create the library item out-of-band (Kibana UI or future resource).
- Validate markdown content syntax. Kibana renders whatever it receives.

## Decisions

- **Schema shape**: `markdown_config = object({ by_value = object({ content, settings, description, hide_title, title, hide_border }), by_reference = object({ ref_id, description, hide_title, title, hide_border }) })`. Mirrors `lens_dashboard_app_config`.
  - *Rejected*: keep the flat current shape and inline a `mode = "value"|"reference"` discriminator. Less consistent with the existing lens-dashboard-app pattern; harder to add per-mode validation.
- **`settings` block**: required when `by_value` is set, with `open_links_in_new_tab` (bool, no default — let Kibana apply its default of `true` if unset). Marking the inner block required mirrors the API but allows the bool inside to remain optional.
- **`ref_id` validation**: required string when `by_reference` is set; no library-item existence check (analogous to other reference shapes).
- **Pre-release breakage**: per project guidance, no migration; existing `markdown_config = { content = ..., title = ... }` configurations will need to be rewritten to `markdown_config = { by_value = { content = ..., title = ..., settings = { open_links_in_new_tab = true } } }`.

## Risks / Trade-offs

- [Risk] Breaking change for any pre-release adopters of `markdown_config` → Accepted per project guidance; flagged in the proposal.
- [Risk] Acceptance test for `by_reference` requires a markdown library item; library-item creation is not yet a TF resource → Mitigation: create via the Kibana saved-objects API in test setup, mirroring how lens-by-reference acceptance tests already work.
