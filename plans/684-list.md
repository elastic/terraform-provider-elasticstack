# Plan for Fleet outputs data source design update

## Objective
Update the `elasticstack_fleet_outputs` data source implementation to match the revised design in dev-docs/terraform/data-sources/fleet/outputs.md, including schema, filters, mapping, and tests.

## Implementation steps
1. Inspect current data source implementation.
   - Locate the data source code under internal/fleet.
   - Identify current schema, models, read logic, and output mapping.
2. Align schema inputs with the design.
   - Add or update `output_id`, `type`, `default_integrations`, `default_monitoring`, and `space_id` inputs.
   - Ensure `space_id` remains null when unknown at plan time.
3. Align schema output structure.
   - Define `items` as a computed list of output objects.
   - Implement output attributes: `id`, `name`, `type`, `hosts`, `ca_sha256`, `ca_trusted_fingerprint`, `default_integrations`, `default_monitoring`, `config_yaml` (sensitive), `ssl`, `kafka`.
   - Set `ssl` to null when all fields are empty; normalize empty `ca_trusted_fingerprint` to null.
4. Update model translation.
   - Map `kbapi.OutputUnion` into a shared output model using the discriminator.
   - Ensure unsupported output types return a clear error.
   - Map Kafka numeric fields from optional pointers to Terraform numbers, using null when absent.
   - Map nested Kafka objects (`headers`, `hash`, `random`, `round_robin`, `sasl`) to null when missing in API.
5. Apply filtering logic.
   - Implement client-side filtering for `output_id`, `type`, `default_integrations`, and `default_monitoring`.
   - Ensure filter combinations are supported and predictable.
6. Update documentation if required by code changes.
   - Regenerate docs if schema or descriptions change.
7. Remove the old fleet_output data source.

## Testing plan
1. Unit tests for translation and filtering.
   - Table-driven tests covering each supported type (`elasticsearch`, `logstash`, `kafka`).
   - Validate empty `ca_trusted_fingerprint` normalization to null.
   - Validate `ssl` object is null when all fields are empty.
   - Validate Kafka pointer-to-number mapping (present vs absent).
   - Validate nested Kafka sub-structures map to null when absent.
   - Validate unsupported output type errors.
2. Acceptance tests (TF_ACC=1) in a real Kibana environment.
   - Space with no outputs: success and empty `items`.
   - Single output, no filters: exactly one returned.
   - Multiple outputs of a single type, no filters: all returned.
   - Multiple outputs of different supported types, no filters: all returned.
   - Multiple outputs of different types with each filter and combinations.
3. Run required tooling.
   - `make test` for unit tests.
   - `make testacc` with required env vars and any newly added cases.
   - `make lint` to ensure linting passes.
   - `make docs-generate` if schema docs change.

## Final review steps
1. Critical code review.
   - Validate schema, filtering, and mapping logic against the design.
   - Check for error handling, null handling, and unsupported type behavior.
2. Test review for coverage gaps.
   - Identify any critical code paths not exercised by unit or acceptance tests.
   - Add tests or scenarios for missed paths before merging.
3. UX review for schema and usage.
   - Evaluate attribute names, defaults, and descriptions for clarity.
   - Confirm filters and computed outputs are intuitive and consistent with other data sources.
