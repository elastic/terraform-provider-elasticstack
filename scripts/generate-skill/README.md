# generate-skill

Generates an agent Skill directory (`dist/skill/elasticstack-terraform/`) that teaches coding agents how to write Terraform configuration against the `elastic/elasticstack` provider.

## Why

Manually maintaining a skill for 100+ resources and data sources is not tractable and would drift instantly. This generator uses `docs/` (tfplugindocs output) as the sole source of truth and layers in hand-seeded static content from `assets/`.

## Design principles

- **Progressive disclosure.** The top-level `SKILL.md` is a small router. The per-entity detail, the entity index, the provider block reference, and the gotchas live in separate files that the agent loads only when a task calls for them.
- **Docs as source of truth.** Per-entity reference files are the docs pages copied verbatim — no re-rendering, no drift.
- **Deterministic.** No timestamps or network calls in content. Rerunning the generator on the same inputs produces byte-identical output.
- **No third-party dependencies.** Standard library only.

## Output layout

```
dist/skill/elasticstack-terraform/
├── SKILL.md                         # Copied verbatim from assets/SKILL.md
├── GENERATED.md                     # Provenance (generated)
└── references/
    ├── index.md                     # assets/references/index.md with {{ENTITIES}} substituted
    ├── context-checklist.md         # Copied verbatim from assets/
    ├── provider.md                  # Copied verbatim from assets/
    ├── gotchas.md                   # Copied verbatim from assets/
    ├── elastic-docs.md              # Copied verbatim from assets/
    ├── resources/<short_name>.md    # Copied verbatim from docs/resources/
    └── data-sources/<short_name>.md # Copied verbatim from docs/data-sources/
```

## What you can hand-edit

Everything under `assets/` is hand-editable Markdown. The generator copies it into the output tree unchanged, except for `references/index.md` where `{{ENTITIES}}` is replaced with the generated entity list. Edit `assets/SKILL.md`, `assets/references/provider.md`, etc. freely — no Go changes required.

The per-entity files under `references/resources/` and `references/data-sources/` are copied verbatim from `docs/`. To change them, update the docs source and rerun the generator.

## Running

```sh
make skill-generate      # writes to dist/skill/elasticstack-terraform/
make skill-test          # runs unit tests
```

Custom invocation:

```sh
go run ./scripts/generate-skill \
  -docs  docs \
  -assets scripts/generate-skill/assets \
  -out   dist/skill/elasticstack-terraform \
  -provider-version 0.14.4 \
  -v
```

## Inputs

- `docs/resources/*.md` and `docs/data-sources/*.md` — tfplugindocs output, copied verbatim into the skill. Frontmatter `description` is also used to build the entity index.
- `scripts/generate-skill/assets/` — hand-edited Markdown, copied verbatim. `assets/references/index.md` must contain the `{{ENTITIES}}` placeholder.

## Editing the skill

- To change wording or add guidance to non-per-entity pages: edit the corresponding file under `assets/` and rerun `make skill-generate`.
- To add a new hand-authored page: drop a file into `assets/<rel_path>` and link to it from `assets/SKILL.md`.
- To change how the entity index list is rendered: edit `emit.go` (`renderEntityList`).

## Tests

`entities_test.go` covers the docs frontmatter parser and entity discovery logic.
