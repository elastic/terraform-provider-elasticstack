variable "watch_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_watch" "test" {
  watch_id = var.watch_id
  active   = false

  trigger = jsonencode({ schedule = { cron = "0 0/1 * * * ?" } })

  input = jsonencode({
    search = {
      request = {
        indices                = [".monitoring-es-*", ".monitoring-logs-*"]
        rest_total_hits_as_int = false
        search_type            = "query_then_fetch"
        body                   = { query = { match_all = {} } }
      }
    }
  })
}
