## 1. Practitioner can set and read a space's allowed namespace prefixes

- [x] 1.1 Add the space settings client wrapper in `internal/clients/fleet/` (read returns nil on HTTP 404, write sends the list for a given space) and verify with a unit test using a mocked Fleet API that the request is routed to the space path and 404 yields nil
- [x] 1.2 Add the `elasticstack_fleet_space_settings` resource package under `internal/fleet/` with the schema and validation from the "Schema and validation" requirement (`space_id` replace-on-change, required set of at most 10 prefixes, empty set allowed, computed `id`/`managed_by`) and verify with unit tests that 11 entries fail validation and an empty set is accepted
- [x] 1.3 Implement create/update as one write and read-after-write state mapping per the "Fleet space settings API" and "Read and state mapping" requirements (id equals `space_id`, the model carries the 9.1.0 version requirement, missing list in the API response becomes an empty set, absent `managed_by` becomes null) and verify with unit tests for the API-to-model mapping
- [x] 1.4 Register the resource in the provider and verify `make build` succeeds and the resource appears in the provider's resource list
- [x] 1.5 Add an acceptance test covering create then update of the prefixes in a non-default space (created with `elasticstack_kibana_space`) and verify it passes against a 9.1+ stack
- [x] 1.6 Add the resource example, generated documentation and a changelog entry (documenting that exactly one resource per space should exist and that destroy lifts the restriction), and verify the docs generation check (`make docs-generate` / `make check-docs`) is clean

## 2. Practitioner can import, detect drift and destroy safely

- [x] 2.1 Support import by space ID (default space included) per the "Identity and import" requirement and verify with an acceptance test that import followed by refresh populates `space_id`, prefixes and `managed_by`
- [x] 2.2 Implement not-found handling on read and destroy per the "Not-found handling on read" and "Destroy resets prefixes" requirements (read 404 removes the resource from state, other errors surface as diagnostics, destroy writes an empty set, destroy of an already-deleted space succeeds, the space itself is never deleted) and verify with unit tests for the 404 and error paths
- [x] 2.3 Add acceptance coverage for drift after an out-of-band change and for destroy resetting the set while the space remains, and verify both pass against a 9.1+ stack

## 3. Practitioner gets clear behavior on unsupported stacks and scoped connections

- [x] 3.1 Enforce the 9.1.0 minimum version on all lifecycle operations per the "Minimum stack version" requirement and verify with a unit test on the model's version requirement plus an acceptance test skipped on older stacks
- [x] 3.2 Support the resource-level `kibana_connection` override and surface API errors (for example a missing `fleet-settings-all` privilege) as diagnostics per the "Provider-level Fleet client by default with optional scoped override", "Space is neither created nor verified" and "API errors are surfaced" requirements, and verify with unit tests that the scoped client is used and errors are not swallowed
- [x] 3.3 Run the full change validation and verify `openspec validate add-fleet-space-settings-resource --strict`, `make build` and the new package's unit tests all pass
