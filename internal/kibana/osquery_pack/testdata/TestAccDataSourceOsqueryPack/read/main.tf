variable "suffix" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
  kibana {}
  fleet {}
}

resource "elasticstack_fleet_agent_download_source" "default" {
  name      = "Osquery Pack DS Agent Download Source ${var.suffix}"
  source_id = "agent-download-source-osquery-pack-ds-${var.suffix}"
  default   = false
  host      = "https://artifacts.elastic.co/downloads/elastic-agent"
  space_ids = ["default"]
}

resource "elasticstack_fleet_agent_policy" "test_policy" {
  name               = "Osquery Pack DS Agent Policy ${var.suffix}"
  namespace          = "testacc"
  description        = "Osquery pack data source acceptance test agent policy"
  monitor_logs       = true
  monitor_metrics    = true
  skip_destroy       = false
  download_source_id = elasticstack_fleet_agent_download_source.default.source_id
}

resource "elasticstack_fleet_integration" "osquery_manager" {
  name    = "osquery_manager"
  version = "1.28.1"
  force   = true
}

resource "elasticstack_fleet_integration_policy" "osquery_manager" {
  name                = "Osquery Manager DS ${var.suffix}"
  namespace           = "testacc"
  agent_policy_id     = elasticstack_fleet_agent_policy.test_policy.policy_id
  integration_name    = elasticstack_fleet_integration.osquery_manager.name
  integration_version = elasticstack_fleet_integration.osquery_manager.version
}

resource "elasticstack_kibana_osquery_pack" "test" {
  name        = "tf-acc-osquery-pack-ds-${var.suffix}"
  description = "Terraform data source acceptance test pack"
  enabled     = true

  policy_ids = [elasticstack_fleet_agent_policy.test_policy.policy_id]
  shards = {
    (elasticstack_fleet_agent_policy.test_policy.policy_id) = 100
  }

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

  depends_on = [elasticstack_fleet_integration_policy.osquery_manager]
}

data "elasticstack_kibana_osquery_pack" "test" {
  pack_id = elasticstack_kibana_osquery_pack.test.pack_id
}
