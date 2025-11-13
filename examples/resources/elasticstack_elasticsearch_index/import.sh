# Importing a index resource is minimal and may result in seemingly unnecessary plan changes. 
# Index settings are *not* included in the import, and so any settings defined in the elasticstack_elasticsearch_index
# resource definition will show up as an addition in the next `terraform plan` operation. 
# Applying these settings 'changes' should be safe, resulting in no actual change to the backing index. 
terraform import elasticstack_elasticsearch_index.my_index <cluster_uuid>/<index_name>

