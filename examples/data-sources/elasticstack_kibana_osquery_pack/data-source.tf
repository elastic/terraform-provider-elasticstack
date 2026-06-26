provider "elasticstack" {
  kibana {}
}

# Read a user-managed pack created in the same root module (plan-only friendly).
resource "elasticstack_kibana_osquery_pack" "example" {
  name    = "example-osquery-pack"
  enabled = true

  queries = {
    find_procs = {
      query    = "SELECT pid, name FROM processes LIMIT 5;"
      platform = ["linux", "darwin"]
    }
  }
}

data "elasticstack_kibana_osquery_pack" "managed" {
  pack_id = elasticstack_kibana_osquery_pack.example.pack_id
}

output "managed_pack_name" {
  value = data.elasticstack_kibana_osquery_pack.managed.name
}

# Direct lookup by pack_id also works for prebuilt (read-only) packs shipped with the
# osquery_manager integration. Prebuilt pack IDs are installation-specific; find them in
# the Osquery UI or via GET /api/osquery/packs. Use read_only to confirm a pack is prebuilt.
#
# data "elasticstack_kibana_osquery_pack" "prebuilt" {
#   pack_id = "<saved_object_id>"
# }
#
# output "prebuilt_pack_read_only" {
#   value = data.elasticstack_kibana_osquery_pack.prebuilt.read_only
# }
