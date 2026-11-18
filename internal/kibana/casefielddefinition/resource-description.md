Manages a single field in the Kibana Cases field library. Requires Elastic Stack 9.6.0 or later.

**Technical Preview:** the underlying Kibana API (`/api/cases/field_definitions`) is marked Technical Preview in the Kibana OpenAPI specification, even though the Field Library feature is generally available in the Kibana UI as of 9.5. The public API is only available from Elastic Stack 9.6.0 (Stack 9.5 serves only internal routes). The API may change in a future release.

The field's shape (type, label, control, metadata, validation) is declared as YAML in `definition`. The field `name` and the YAML `type` are immutable after creation:

- Changing the top-level `name` attribute (together with the matching `name` key in `definition`) replaces the resource.
- Editing only the `name` key inside `definition` so that it no longer matches `name` is rejected by the API with a 400 error.
- Editing the YAML `type` is rejected by the API with a 409 `field_identity_immutable` error.
- Deleting a field that is still referenced by a case template or custom field is rejected with a 409 error.

If `name` is omitted from configuration, it is taken from the YAML on create and then kept in state; editing only the `name` key inside `definition` afterwards is rejected by the API (400) rather than replacing the resource. Set `name` explicitly, or change both together, to rename (replace) a field.
