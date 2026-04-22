## Context

The repo's docs target is currently:

```make
 docs-generate:
 	go tool github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name terraform-provider-elasticstack
```

`tfplugindocs` builds the provider and extracts schema using Terraform CLI (`terraform init` and `terraform providers schema -json`). Upstream supports two relevant control points:

- `--tf-version <x.y.z>`: download/use that exact Terraform version
- `--providers-schema <path>`: skip Terraform invocation entirely and consume a precomputed schema file

For this change we intentionally keep the existing `tfplugindocs`-managed flow and only take control of version selection.

## Goals / Non-Goals

**Goals:**
- Make docs generation deterministic across developer machines.
- Keep the implementation small and close to current tooling.
- Make the Terraform version choice explicit in repository-owned configuration.
- Keep CI and local docs generation aligned on the same Terraform version.
- Use a repository convention that Renovate can manage without custom regex-manager logic.

**Non-Goals:**
- Rebuilding docs generation around precomputed schema JSON files.
- Pinning every Terraform usage in the repo in this same change (for example `terraform fmt` can be addressed separately if desired).
- Changing the generated docs structure, templates, or examples.

## Decisions

### Use `.terraform-version` as the single source of truth

A root `.terraform-version` file SHALL define the Terraform CLI version used by docs generation and relevant CI validation jobs. This makes the Terraform version a repository-level toolchain decision rather than a docs-target-local detail, and it aligns with Renovate's built-in support for `.terraform-version` updates.

This is preferred over a Makefile-local variable because the version becomes visible as a top-level toolchain file, can be reused by CI and local tooling, and does not require custom Renovate regex-manager configuration.

### Pass `--tf-version` to `tfplugindocs` using `.terraform-version`

`docs-generate` SHALL read `.terraform-version` and pass the resulting pinned version via `tfplugindocs generate --tf-version <version>`. This uses upstream-supported behavior while keeping the tool invocation deterministic even on machines whose local Terraform installation does not respect `.terraform-version` automatically.

This is preferred over `--providers-schema` because the latter would require the repo to reimplement provider build + plugin layout + schema extraction orchestration that `tfplugindocs` already handles correctly.

### Align CI with `.terraform-version`

The lint/docs validation path in CI SHALL install the same Terraform version declared in `.terraform-version`. This prevents situations where contributors regenerate docs with one CLI version while CI validates with another.

The most likely implementation is for the workflow to read `.terraform-version` and pass that value into `hashicorp/setup-terraform`, but the exact wiring can follow the current workflow source conventions.

### Document the policy where contributors look for docs guidance

Contributor-facing documentation under `dev-docs/high-level/documentation.md` SHALL explain that docs generation uses the Terraform version pinned in `.terraform-version` via `tfplugindocs`, rather than whichever Terraform binary happens to be installed locally. The docs SHOULD also note that Renovate maintains `.terraform-version` over time.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| `.terraform-version` broadens the scope from docs-only to a repo-level Terraform convention | Accept this intentionally because the shared file simplifies CI/local alignment and future tooling integration |
| CI and local config diverge if one side is updated without the other | Update both in the same change and capture the requirement in OpenSpec specs/docs |
| Contributors assume their local Terraform binary or version manager alone controls docs generation | Document that `docs-generate` reads `.terraform-version` and passes it explicitly to `tfplugindocs` |
| Renovate does not pick up `.terraform-version` as expected in this repo's inherited config | Verify Renovate behavior/config as part of implementation and adjust only if the built-in manager is being suppressed |

## Migration Plan

1. Add `.terraform-version` with the current latest stable Terraform release.
2. Wire `docs-generate` to read `.terraform-version` and pass it to `--tf-version`.
3. Update CI Terraform setup for lint/docs validation to read the same version.
4. Update contributor docs, Renovate expectations, and OpenSpec requirements.
5. Verify docs generation and docs freshness checks still behave as expected.

## Open Questions

1. **Initial pinned version** — use the current latest stable Terraform release in `.terraform-version` (currently `1.14.9`) unless verification during implementation shows a repo-specific incompatibility.
