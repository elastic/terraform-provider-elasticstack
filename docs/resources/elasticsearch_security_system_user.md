---
subcategory: "Security"
layout: ""
page_title: "Elasticstack: elasticstack_elasticsearch_security_system_user Resource"
description: |-
  Updates system user's password and enablement.
---

# Resource: elasticstack_elasticsearch_security_system_user

Updates system user's password and enablement. See, https://www.elastic.co/guide/en/elasticsearch/reference/current/built-in-users.html
Since this resource is to manage built-in users, destroy will not delete the underlying Elasticsearch and will only remove it from Terraform state.

## Example Usage

```terraform
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_system_user" "kibana_system" {
  username = "kibana_system"

  // For details on how to generate the hashed password see https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-put-user.html#security-api-put-user-request-body
  password_hash = "$2a$10$rMZe6TdsUwBX/TA8vRDz0OLwKAZeCzXM4jT3tfCjpSTB8HoFuq8xO"

  elasticsearch_connection {
    endpoints = ["http://localhost:9200"]
    username  = "elastic"
    password  = "changeme"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `username` (String) An identifier for the system user (see https://www.elastic.co/guide/en/elasticsearch/reference/current/built-in-users.html).

### Optional

- `elasticsearch_connection` (Block List, Deprecated) Elasticsearch connection configuration block. (see [below for nested schema](#nestedblock--elasticsearch_connection))
- `enabled` (Boolean) Specifies whether the user is enabled. The default value is true.
- `password` (String, Sensitive) The user's password. Passwords must be at least 6 characters long.
- `password_hash` (String, Sensitive) A hash of the user's password. This must be produced using the same hashing algorithm as has been configured for password storage (see https://www.elastic.co/guide/en/elasticsearch/reference/current/security-settings.html#hashing-settings).

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
