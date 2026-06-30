provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_osquery_pack" "example" {
  name        = "example-osquery-pack"
  description = "Example Osquery pack managed by Terraform"
  enabled     = true

  queries = {
    find_procs = {
      query    = "SELECT pid, name FROM processes LIMIT 5;"
      platform = ["linux", "darwin"]
      version  = "1.0.0"

      ecs_mapping = {
        "process.name" = {
          field = "name"
        }
        "process.pid" = {
          value = "0"
        }
        "host.name" = {
          values = ["host-a", "host-b"]
        }
      }
    }
  }
}
