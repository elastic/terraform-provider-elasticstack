## Why

Kibana Cases "templates" let teams pre-fill case defaults (severity, category, tags, assignees, connector, settings) and custom-field values so analysts create consistent cases without re-entering the same data every time. Stack 9.5 introduced a standalone templates API, `POST /api/cases/templates` and `GET`/`PUT`/`DELETE /api/cases/templates/{template_id}`, distinct from the legacy `templates` array embedded in `/api/cases/configure`. Today the provider has no way to manage these templates, so teams that declare their case configuration as code either fall back to manual UI changes or `terraform_data`/`local-exec`, neither of which gives state management or drift detection.

This change is split out from [#5191](https://github.com/elastic/terraform-provider-elasticstack/issues/5191) and is a companion to the field library resource added in [#5236](https://github.com/elastic/terraform-provider-elasticstack/issues/5236) (`elasticstack_kibana_case_field_definition`), which a template's YAML `fields` array may reference by `$ref`.

Source: [elastic/terraform-provider-elasticstack#5193](https://github.com/elastic/terraform-provider-elasticstack/issues/5193). An implementation-research comment on the issue evaluated two approaches and recommended the opaque-YAML approach below; a maintainer follow-up comment then resolved the research comment's open questions against a live 9.5.5 stack. See design.md for the full evaluation and the specific points where the follow-up comment's decisions refine the research comment's recommendation.

## What Changes

- Add a new Terraform resource `elasticstack_kibana_case_template` that manages a single case template through `POST`/`GET`/`PUT`/`DELETE /api/cases/templates{,/{template_id}}`.
- The template's case defaults and custom-field declarations (including `$ref` entries into the field library) are declared as a YAML string in `definition_yaml`; `owner` scopes the template to a solution (Stack Management/Observability/Security) and forces replacement; `name`, `description`, `tags`, and `is_enabled` are additional top-level, request-level attributes (not part of the YAML).
- `definition_yaml` uses the repo's `customtypes.NormalizedYamlType` for plan-time YAML syntax validation and semantic-equality reconciliation; the resource adds explicit plan-time diff suppression on top of that type so whitespace-, comment-, and key-order-only edits produce an empty plan and skip the `PUT` call entirely, since every `PUT` (including a no-op resubmit of unchanged YAML) creates a new template version server-side.
- Exposes computed `template_id` and `template_version` only; `latest_version` is not exposed because the resource always manages the latest version and historical-version management is out of scope.
- `tags` is modeled as `List[String]` (not a set): the API preserves submission order and duplicate entries.
- `description` and `tags` are plain Optional (not Computed): the API omits them from the response when not supplied on write, so the model maps an omitted config directly to a null/absent state without needing `UseStateForUnknown` reconciliation.
- Built on the repo's `entitycore.KibanaResource` envelope (matching `casefielddefinition`, `agentbuilderworkflow`), with space-aware routing and a minimum Elastic Stack version of 9.5.2 (the backport commit that shipped kibana#280144 landed after the 9.5.0 and 9.5.1 tags).
- Add documentation, an example, and acceptance tests for the resource, with a Technical Preview note (the OpenAPI spec's own `x-state` on create/update/delete) and a note that the feature additionally requires the `xpack.cases.templates.enabled` setting (defaults to `true`).

Non-goals:

- No data source for listing/searching templates (`GET /api/cases/templates` with paging/sorting/filtering).
- No changes to the legacy `templates` array embedded in `/api/cases/configure`.
- No client-side structural validation of the YAML `fields`/`$ref`/`display`/`validation` keys; that stays server-side, consistent with the field-definition resource.
- No exposure of historical template versions (`GET ?version=N`) as a Terraform-managed concept.

## Capabilities

### New Capabilities
- `kibana-case-template`: Schema, identity/import, CRUD, YAML diff suppression, version gating, and error-mapping requirements for the `elasticstack_kibana_case_template` resource.

### Modified Capabilities
<!-- None. This is a net-new, additive resource with no interaction with existing specs. -->

## Impact

- New Kibana resource package under `internal/kibana/`; the required API operations and types already exist in the generated Kibana client (`generated/kbapi`), so no client regeneration is expected.
- New thin `internal/clients/kibanaoapi` wrapper for the template CRUD calls.
- User-facing documentation and an example.
- Acceptance tests requiring Elastic Stack 9.5.2+ with `xpack.cases.templates.enabled` (default `true`); tests that reference a field-library field via `$ref` additionally require `elasticstack_kibana_case_field_definition`.
