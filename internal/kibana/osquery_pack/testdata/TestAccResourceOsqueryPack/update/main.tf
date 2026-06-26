variable "suffix" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_osquery_pack" "test" {
  name        = "tf-acc-osquery-pack-updated-${var.suffix}"
  description = "Updated Terraform acceptance test pack"
  enabled     = false

  queries = {
    find_procs = {
      query    = "SELECT pid, name, path FROM processes LIMIT 10;"
      platform = ["linux"]
      version  = "1.1.0"
      snapshot = true
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
    list_users = {
      query    = "SELECT username FROM users LIMIT 5;"
      platform = ["linux", "windows"]
      version  = "2.0.0"
      snapshot = false
      removed  = false
    }
  }
}
