variable "suffix" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_osquery_pack" "test" {
  name        = "tf-acc-osquery-pack-ds-${var.suffix}"
  description = "Terraform data source acceptance test pack"
  enabled     = true

  queries = {
    find_procs = {
      query    = "SELECT pid, name FROM processes LIMIT 5;"
      platform = ["linux", "darwin"]
      version  = "1.0.0"
      snapshot = false
      removed  = false
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

data "elasticstack_kibana_osquery_pack" "test" {
  pack_id = elasticstack_kibana_osquery_pack.test.pack_id
}
