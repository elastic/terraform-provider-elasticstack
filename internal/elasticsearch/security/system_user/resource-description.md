Updates system user's password and enablement. See the [built-in users documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/built-in-users.html) for more details.

Since this resource is to manage built-in users, destroy will not delete the underlying Elasticsearch and will only remove it from Terraform state.
