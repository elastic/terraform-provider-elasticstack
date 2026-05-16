## Context

`mapPanelFromAPI` in `internal/kibana/dashboard/models_panels.go` switches on the API panel `type` and builds a typed Terraform `panelModel`. The `default` branch currently emits `panelModel{ Type: t }` only — `id`, `grid`, and the full raw config are discarded. The Kibana API spec already defines several panel types the resource does not type today (`discover_session`, `image`, `slo_alerts`); future spec bumps will add more. Until each such type has a typed block, the resource must still survive reading dashboards that contain them.

## Goals / Non-Goals

**Goals:**
- Refresh/import of a dashboard with an unrecognized panel type produces stable state (id and grid preserved; raw config retained verbatim) and re-applies without diff.
- Subsequent writes that include the unchanged unknown panel re-send the preserved payload.
- The mechanism is invisible to users — no new HCL attribute they could misuse.

**Non-Goals:**
- Allowing users to author unknown panels in HCL.
- Typing any specific currently-unrecognized panel (image / SLO alerts / Discover session — those are separate proposals).

## Decisions

- **Preservation key**: store the raw API payload in the existing `config_json` attribute (`Optional: true, Computed: true`). The implementation reuses `config_json` as the storage vehicle for unknown-panel payloads rather than introducing a new unexported field. This decision deliberately deviates from the earlier design preference for a private mechanism.
  - **Rationale**: reusing `config_json` avoids adding a new attribute (exposed or private), eliminates the need for a separate semantic-equality normalization for a private payload, and naturally routes the payload through the same write-path codepath used for authored `config_json` panels. Since `config_json` is already `Computed: true`, practitioners can also inspect round-tripped values via `terraform show` without any new API surface.
- **Write path**: in the panel write dispatcher, when the panel's `type` matches no typed config block but the panel has a preserved raw payload from a prior read, re-marshal that payload into the API request unchanged. If no preserved payload exists (i.e., user authored an unknown type from scratch), return an error diagnostic — the existing "unsupported panel type" message.
- **Diff stability**: the preserved payload must normalize identically to its source representation. Reuse the existing `config_json` semantic-equality normalization for the catch-all branch.
- **No version pinning**: behavior is independent of stack version; the resource simply mirrors whatever the API returned.

## Risks / Trade-offs

- [Risk] Hidden state bloat for dashboards with many unknown panels → Mitigation: payloads are bounded by panel size and stored once per panel; no compounding.
- [Risk] Users can't see what the resource is silently round-tripping → Mitigation: it's still visible in `terraform show` as the stored API JSON (via the same hidden-attribute approach used elsewhere); document in changelog.
- [Risk] When we later add typed support for a previously-unknown type, existing state needs to migrate cleanly → Mitigation: typed-block additions are additive (the new block reads the same API field); preserve-or-type detection runs per-panel based on whether the typed block is set.
