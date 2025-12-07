Helper data source which can be used to create the configuration for a set security user processor. This processor sets user-related details (such as `username`, `roles`, `email`, `full_name`, `metadata`, `api_key`, `realm` and `authentication_type`) from the current authenticated user to the current document by pre-processing the ingest. See the [set security user processor documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/ingest-node-set-security-user-processor.html) for more details.

The `api_key` property exists only if the user authenticates with an API key. It is an object containing the `id`, `name` and `metadata` (if it exists and is non-empty) fields of the API key. 

The `realm` property is also an object with two fields, name and type. When using API key authentication, the realm property refers to the realm from which the API key is created. 

The `authentication_type` property is a string that can take value from `REALM`, `API_KEY`, `TOKEN` and `ANONYMOUS`.

