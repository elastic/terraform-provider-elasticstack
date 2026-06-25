provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_osquery_saved_query" "list_processes" {
  saved_query_id = "list_processes"
  query          = "SELECT pid, name, cmdline FROM processes LIMIT 100;"
  description    = "List running processes with command lines"
  platform       = ["linux", "darwin"]
  interval       = 3600
  version        = "1.0.0"

  ecs_mapping = {
    "process.name" = {
      field = "name"
    }
    "process.command_line" = {
      field = "cmdline"
    }
    "event.category" = {
      value = "process"
    }
    "event.type" = {
      values = ["start", "end"]
    }
  }
}
