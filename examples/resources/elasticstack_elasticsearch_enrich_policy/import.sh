# NOTE: while importing index resource, keep in mind, that some of the default index settings will be imported into the TF state too
# You can later adjust the index configuration to account for those imported settings
terraform import elasticstack_elasticsearch_enrich_policy.policy1 <cluster_uuid>/<policy_name>
