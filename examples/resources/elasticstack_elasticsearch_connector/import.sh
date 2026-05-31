# Import using the bare connector ID
terraform import elasticstack_elasticsearch_connector.postgres music-catalog

# Or using the composite ID
terraform import elasticstack_elasticsearch_connector.postgres <cluster_uuid>/music-catalog
