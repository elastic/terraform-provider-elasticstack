Helper data source which can be used to create the configuration for a community ID processor. This processor computes the Community ID for network flow data as defined in the Community ID Specification. See: https://www.elastic.co/guide/en/elasticsearch/reference/current/community-id-processor.html
You can use a community ID to correlate network events related to a single flow.

The community ID processor reads network flow data from related [Elastic Common Schema (ECS)](https://www.elastic.co/guide/en/ecs/1.12) fields by default. If you use the ECS, no configuration is required.
