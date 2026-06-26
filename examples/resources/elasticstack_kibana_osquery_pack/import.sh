# Import user-managed Osquery packs only. Prebuilt/read-only packs cannot be imported
# into the resource; read them with the elasticstack_kibana_osquery_pack data source.
# pack_id is the Kibana saved_object_id (UUID) for the pack.
terraform import elasticstack_kibana_osquery_pack.example <space_id>/<pack_id>
