---
subcategory: "Security"
layout: ""
page_title: "Elasticstack: elasticstack_elasticsearch_security_role_mapping Resource"
description: |-
  Manage role mappings. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-put-role-mapping.html
---

# Resource: elasticstack_elasticsearch_security_role_mapping

Manage role mappings. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-put-role-mapping.html

## Example Usage

```terraform
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role_mapping" "example" {
  name    = "test_role_mapping"
  enabled = true
  roles = [
    "admin"
  ]
  rules = jsonencode({
    any = [
      { field = { username = "esadmin" } },
      { field = { groups = "cn=admins,dc=example,dc=com" } },
    ]
  })
}

output "role" {
  value = elasticstack_elasticsearch_security_role_mapping.example.name
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The distinct name that identifies the role mapping, used solely as an identifier.
- `rules` (String) The rules that determine which users should be matched by the mapping. A rule is a logical condition that is expressed by using a JSON DSL.

### Optional

- `elasticsearch_connection` (Block List, Max: 1, Deprecated) Elasticsearch connection configuration block. This property will be removed in a future provider version. Configure the Elasticsearch connection via the provider configuration instead. (see [below for nested schema](#nestedblock--elasticsearch_connection))
- `enabled` (Boolean) Mappings that have `enabled` set to `false` are ignored when role mapping is performed.
- `metadata` (String) Additional metadata that helps define which roles are assigned to each user. Keys beginning with `_` are reserved for system usage.
- `role_templates` (String) A list of mustache templates that will be evaluated to determine the roles names that should granted to the users that match the role mapping rules.
- `roles` (Set of String) A list of role names that are granted to the users that match the role mapping rules.

### Read-Only

- `id` (String) Internal identifier of the resource

<a id="nestedblock--elasticsearch_connection"></a>
### Nested Schema for `elasticsearch_connection`

Optional:

- `api_key` (String, Sensitive) API Key to use for authentication to Elasticsearch
- `bearer_token` (String, Sensitive) Bearer Token to use for authentication to Elasticsearch
- `ca_data` (String) PEM-encoded custom Certificate Authority certificate
- `ca_file` (String) Path to a custom Certificate Authority certificate
- `cert_data` (String) PEM encoded certificate for client auth
- `cert_file` (String) Path to a file containing the PEM encoded certificate for client auth
- `endpoints` (List of String, Sensitive) A list of endpoints where the terraform provider will point to, this must include the http(s) schema and port number.
- `es_client_authentication` (String, Sensitive) ES Client Authentication field to be used with the JWT token
- `headers` (Map of String, Sensitive) A list of headers to be sent with each request to Elasticsearch.
- `insecure` (Boolean) Disable TLS certificate validation
- `key_data` (String, Sensitive) PEM encoded private key for client auth
- `key_file` (String) Path to a file containing the PEM encoded private key for client auth
- `password` (String, Sensitive) Password to use for API authentication to Elasticsearch.
- `username` (String) Username to use for API authentication to Elasticsearch.

## Import

Import is supported using the following syntax:

```shell
terraform import elasticstack_elasticsearch_security_role_mapping.my_role_mapping <cluster_uuid>/<role mapping name>
```
