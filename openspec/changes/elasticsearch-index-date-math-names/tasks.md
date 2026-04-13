## 1. Schema and state model

- [ ] 1.1 Update `internal/elasticsearch/index/index/schema.go` so `name` validation uses `stringvalidator.Any(...)` with the static regex `^[a-z0-9!$%&'()+.;=@[\]^{}~_-]+$` and the date-math regex `^<[^<>]*\{[^<>]+\}[^<>]*>$`, and add the computed `concrete_name` attribute.
- [ ] 1.2 Update the index resource state/model code so `name` remains the configured value while `concrete_name` stores the managed concrete index name.
- [ ] 1.3 Update legacy/import read paths to backfill `concrete_name` from `id` and only backfill `name` when it is absent from state.

## 2. Concrete-name-aware CRUD behavior

- [ ] 2.1 Update create flow and Elasticsearch client helpers to URI-encode validated plain date math names for the Create Index API path, capture the concrete index name from the response, and compute `id` from that value.
- [ ] 2.2 Update read, update, and delete flows so all post-create API calls target the concrete managed index from state / `id`, not the configured `name`.
- [ ] 2.3 Update any Get Index helper logic that currently assumes the requested key and returned response key are identical for date math creates.

## 3. Regression coverage

- [ ] 3.1 Add focused validation tests for static names, valid plain date math names, invalid date math inputs, and provider-side URI encoding for create requests.
- [ ] 3.2 Add resource tests covering create/read stability for date math names, including preservation of `name` and persistence of `concrete_name`.
- [ ] 3.3 Add update-path regression coverage proving alias/settings/mappings updates still target the managed concrete index after creation from a date math expression.
