# The cluster settings resource is a singleton identified by <cluster_uuid>/cluster-settings.
# Find <cluster_uuid> with the elasticstack_elasticsearch_info data source or the Elasticsearch GET / API.
# After import, only the id is stored in state. Declare the persistent and/or transient setting
# blocks you want to manage before running terraform plan or terraform apply.
terraform import elasticstack_elasticsearch_cluster_settings.my_cluster_settings <cluster_uuid>/cluster-settings
