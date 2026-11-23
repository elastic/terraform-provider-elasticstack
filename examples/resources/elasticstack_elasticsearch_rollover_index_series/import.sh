# Import identifier: <cluster_uuid>/<rollover_alias>/<prefix>
# State id is <cluster_uuid>/<rollover_alias>. The three-part import string is not stored as id.
# prefix is taken from this identifier and is not inferred from the current writer.
# deletion_protection is initialized to true. A later apply must set it to false before destroy.
# Healthy create and import read an ordinary alias and do not require an extra flag.
terraform import elasticstack_elasticsearch_rollover_index_series.orders '<cluster_uuid>/orders-write/orders'
