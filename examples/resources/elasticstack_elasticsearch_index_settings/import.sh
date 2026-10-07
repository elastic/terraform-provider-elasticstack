# Importing an index settings resource only hydrates the typed dynamic settings
# from the index; settings declared via `settings_json` are not imported and
# will show as additions in the next `terraform plan`.
terraform import elasticstack_elasticsearch_index_settings.my_index_settings 'cluster_uuid/concrete_index_name'