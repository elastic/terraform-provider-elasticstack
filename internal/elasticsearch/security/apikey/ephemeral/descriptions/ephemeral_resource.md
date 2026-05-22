Creates an Elasticsearch API key during each Terraform plan and apply without persisting credentials to state.

See the [security API create API key documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-create-api-key.html) and [create cross-cluster API key documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/security-api-create-cross-cluster-api-key.html) for more details.

Use the managed [`elasticstack_elasticsearch_security_api_key`](/docs/resources/elasticsearch_security_api_key) resource when credentials should remain in Terraform state.
