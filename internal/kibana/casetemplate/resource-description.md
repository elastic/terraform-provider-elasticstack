Manages a single Kibana Cases template. Requires Elastic Stack 9.5.2 or later.

**Technical Preview:** the underlying Kibana API (`/api/cases/templates`) is marked Technical Preview in the Kibana OpenAPI specification and may change in a future release.

The templates routes are also gated by the Kibana setting `xpack.cases.templates.enabled`. It defaults to `true`, but when it is set to `false` the routes are removed and Kibana answers with a 404, which is indistinguishable from a server that is too old at the HTTP level (only the response body differs). If operations fail with a 404 against a supported stack version, check this setting.

Case defaults (`name`, `severity`, `category`, `tags`, `assignees`, `connector`, `settings`) and the template's custom fields (inline definitions or `$ref` entries into the field library, see `elasticstack_kibana_case_field_definition`) are declared as YAML in `definition_yaml`. The `connector` and `settings` keys (including `settings.syncAlerts`) belong inside the YAML; they are not separate resource attributes.

Every update creates a new template version in Kibana, so edits to `definition_yaml` that only change whitespace, comments, or mapping-key order are suppressed at plan time and do not trigger an update. Unlike the field definition resource, `description` and `tags` are cleared when they are removed from configuration. Destroying the resource soft-deletes all versions of the template; cases already created from it keep their pinned template reference.
