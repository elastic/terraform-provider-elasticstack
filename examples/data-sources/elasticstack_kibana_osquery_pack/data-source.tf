provider "elasticstack" {
  kibana {}
}

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

data "elasticstack_kibana_osquery_pack" "example" {
  pack_id = elasticstack_kibana_osquery_pack.example.pack_id
}

output "pack_name" {
  value = data.elasticstack_kibana_osquery_pack.example.name
}
