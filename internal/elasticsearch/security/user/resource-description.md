---
subcategory: "Elasticsearch"
layout: ""
page_title: "Elasticstack: elasticstack_elasticsearch_security_user Resource"
description: |-
  Adds and updates users in the native realm. These users are commonly referred to as native users.
---

# Resource: elasticstack_elasticsearch_security_user

Adds and updates users in the native realm. These users are commonly referred to as native users. See the [Elasticsearch security user API documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-put-user.html) for more details.

## Example Usage

```terraform
resource "elasticstack_elasticsearch_security_user" "user" {
  username  = "my_user"
  password  = "changeme"
  roles     = ["superuser"]
  full_name = "John Doe"
  email     = "john@example.com"
}
```
