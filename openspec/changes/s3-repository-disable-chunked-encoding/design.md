## Context

`internal/elasticsearch/snapshot/repository/schema.go` defines the `s3` repository type block
(`s3Block()`, ~line 266). It already exposes several settings that Elasticsearch's S3 repository
type documents as **client** settings rather than native repository settings — `endpoint` and
`path_style_access` — via the deprecated-but-functional "overriding client settings in repository
settings" mechanism (repository `settings` values are merged into the named S3 client, repository
value winning). `disable_chunked_encoding` and `always_sign_requests` are two more S3 client
settings of the same shape:

- `s3.client.CLIENT_NAME.disable_chunked_encoding` (Elastic docs: "Planned for deprecation").
  Boolean, defaults to `false`. When `true`, disables chunked encoding for S3 uploads — required
  by some S3-compatible stores that reject the `x-amz-content-sha256` chunked-encoding header
  Elasticsearch's SDK v2-based client now sends, per #4875.
- `s3.client.CLIENT_NAME.always_sign_requests` (Elastic docs: "Generally available since 9.5").
  Boolean, defaults to `false`. Forces the additional payload-signature mechanism even over HTTPS;
  addresses a related plain-HTTP-integrity incompatibility with some S3-compatible stores, without
  relying on a setting Elastic plans to deprecate.

Confirmed via `elastic-docs` (`s3-repository-settings` page, fetched during this proposal): both
settings are documented only under "Client settings," not "Repository settings," and the docs page
does not state whether `GET _snapshot/{repo}` echoes client-setting overrides back in its response.
This is the same gap already called out in the codebase for `endpoint`/`path_style_access`
(`settingsToS3` in `read.go`, ~line 262): "Whether Elasticsearch returns them via the raw settings
overlay is version-dependent and difficult to determine empirically."

## Goals / Non-Goals

**Goals:**
- Let practitioners set `disable_chunked_encoding` and `always_sign_requests` on the `s3` block so
  the #4875 workaround (previously only possible via direct API calls) is available through
  Terraform.
- Follow the exact attribute pattern already used for `server_side_encryption`/`path_style_access`
  (optional+computed bool, default `false`, unconditional write) so the diff is mechanical and
  low-risk.
- Prevent `terraform plan` drift after apply if the API does not echo these settings back on GET,
  using the same prior-state fallback already implemented for `endpoint`/`path_style_access`.

**Non-Goals:**
- Do not add `unsafely_incompatible_with_s3_conditional_writes` (a different, conditional-writes
  incompatibility class; a possible separate follow-up per issue research).
- Do not build a generic S3 client-setting passthrough map (an alternative approach considered and
  rejected in issue research: no precedent in this resource, loses type validation, harder drift
  detection for arbitrary keys).
- Do not change the default value of either attribute (both stay `false`, matching Elasticsearch)
  or auto-enable either for all S3-compatible setups.
- Do not add an Elasticsearch minimum-version schema-level guard unless implementation-time
  verification shows one is strictly required (see Open questions).

## Decisions

- **Mechanism**: add both as `schema.BoolAttribute{Optional: true, Computed: true, Default:
  booldefault.StaticBool(false)}` on `s3Block()`, matching `server_side_encryption` and
  `path_style_access` exactly.
- **Scope**: include both `disable_chunked_encoding` and `always_sign_requests` in this same
  change. Per maintainer direction on the issue, `always_sign_requests` is explicitly in scope
  alongside `disable_chunked_encoding` — it addresses the same S3-compatible-storage incompatibility
  class without depending on a setting Elastic plans to deprecate. This does not change the spine
  of the fix, only its attribute count.
- **Write path**: set both unconditionally in `s3ToSettings` (`write.go`), like
  `server_side_encryption`/`path_style_access` — not via `setIfNotEmpty`, since these are booleans
  with a resolved schema default, never null/unknown at write time.
- **Read path**: extend `settingsToS3` (`read.go`) with the same prior-state fallback already
  applied to `path_style_access`: capture a `*Fallback` value from prior state before mapping,
  and use `boolSetting(s, settingX, xFallback)` so an omitted GET key inherits state instead of
  reverting to `false`. This is a defensive default matching the existing REQ-017 precedent for
  `endpoint`/`path_style_access`, not an empirically confirmed requirement — see Open questions.
- **Data source**: mirror both attributes as computed-only on the data source `s3` block, the same
  way `path_style_access`/`server_side_encryption` are exposed there today.
- **Spec shape**: add REQ-019 (schema + write-path mapping, covering both attributes) and REQ-020
  (read-side drift prevention, covering both attributes), following the REQ-016/REQ-017 pattern
  already in `openspec/specs/elasticsearch-snapshot-repository/spec.md` for `endpoint`/
  `path_style_access`. Update the Schema HCL sketch for the `s3` block.

## Risks / Trade-offs

- [Risk] `disable_chunked_encoding` is documented "Planned for deprecation" by Elastic. →
  Mitigation: it is the only currently-shipped fix for the #4875 error class; `always_sign_requests`
  (not deprecated) is added alongside it for the related class, and removal/migration is Elastic's
  timeline to set, not this provider's to pre-empt.
- [Risk] Both settings rely on the deprecated "overriding client settings in repository settings"
  mechanism, not a native repository setting. → Mitigation: this provider already depends on the
  same mechanism for `endpoint`/`path_style_access`; no new mechanism is introduced.
- [Risk] If GET does not echo these settings back, the prior-state fallback masks any future
  out-of-band change to the underlying client setting (e.g. edited directly via `elasticsearch.yml`
  or another tool) — Terraform would not detect that drift. → Mitigation: this is the same accepted
  trade-off already made for `endpoint`/`path_style_access`; consistent behavior across all four
  client-setting-backed attributes is preferable to inconsistent drift handling within one block.

## Open questions

- Does `GET _snapshot/{repo}` echo back `disable_chunked_encoding` and `always_sign_requests` once
  set, or do they need the same prior-state fallback as `endpoint`/`path_style_access`? — Not
  resolved by documentation: the `elastic-docs` S3 repository settings page lists both only under
  "Client settings" and does not state GET-response echo behavior, mirroring the existing
  uncertainty already documented for `endpoint`/`path_style_access` in `settingsToS3`. **Default
  decision for this proposal**: apply the fallback defensively (see Decisions) since that is the
  established, lower-risk precedent in this file. Implementation SHOULD attempt empirical
  verification against a real cluster if feasible and simplify to a direct `boolSetting(s, key,
  false)` read (no fallback) if GET is confirmed to echo the value. This remains **blocking** for a
  fully drift-free implementation per issue research, but is not blocking for authoring this
  proposal.
- Is a minimum Elasticsearch version guard needed, or is this setting supported across all versions
  this provider targets? — `disable_chunked_encoding` is a long-standing S3 client setting with no
  documented minimum version. `always_sign_requests` reached GA in Elasticsearch 9.5; the docs page
  does not state whether it existed (e.g. as a technical preview) in earlier supported versions.
  Per maintainer direction, this remains an open question for discovery during implementation; if
  unresolved, treat `always_sign_requests` as valid schema on all supported versions (Elasticsearch
  ignores unknown settings keys) rather than adding a `RequiresReplace`/version-gated validator
  speculatively.
