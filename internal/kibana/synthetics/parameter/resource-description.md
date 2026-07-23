Creates or updates a Kibana synthetics parameter.

See [Working with secrets and sensitive values](https://www.elastic.co/docs/solutions/observability/synthetics/work-with-params-secrets)
and [API docs](https://www.elastic.co/docs/api/doc/kibana/group/endpoint-synthetics)

Parameters are scoped to a Kibana space. Set `space_id` to the target space identifier; when omitted, the resource uses the `default` space (`space_id` is computed as `"default"`). Changing `space_id` forces replacement of the parameter.

The computed `id` is a composite identifier: `<space_id>/<parameter_uuid>`, where the UUID is assigned by Kibana.

Import accepts a bare parameter UUID (treated as the `default` space, with `id` set to `default/<uuid>`) or the composite form `<space_id>/<parameter_uuid>`.

**Example** (parameter in a named space):

```terraform
provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_synthetics_parameter" "example" {
  space_id    = "my-space"
  key         = "example_key"
  value       = "example_value"
  description = "Example description in a named space"
  tags        = ["tag-a", "tag-b"]
}
```
