# Documentation

Provider docs are generated; keep them in sync with code changes.

## Where docs come from

- Templates: `templates/`
- Examples: `examples/`
- Generated output: `docs/`

Canonical contributor guidance: “Updating Documentation” in [`contributing.md`](./contributing.md#updating-documentation).

## Terraform version used for docs generation

`make docs-generate` reads the Terraform CLI version pinned in the repository root `.terraform-version` file and passes that version to `tfplugindocs`.

This keeps docs generation deterministic across contributor machines and CI. Docs generation does not rely on whichever Terraform version happens to be installed locally.

The pinned version is maintained in `.terraform-version`. Renovate is expected to keep that file up to date over time.

## Regenerating docs

- Generate docs: `make docs-generate`
- Note: `make lint` also runs `docs-generate`.
- If you change provider schemas, docs templates, or examples, regenerate docs and commit the resulting `docs/` updates.

