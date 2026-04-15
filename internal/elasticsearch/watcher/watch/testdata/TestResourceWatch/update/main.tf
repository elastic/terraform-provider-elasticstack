variable "watch_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_watch" "test" {
  watch_id = var.watch_id
  active   = true

  trigger   = jsonencode({ schedule = { cron = "0 0/2 * * * ?" } })
  input     = jsonencode({ simple = { name = "example" } })
  condition = jsonencode({ never = {} })
  actions   = jsonencode({ log = { logging = { level = "info", text = "example logging text" } } })
  metadata  = jsonencode({ example_key = "example_value" })

  transform = jsonencode({
    search = {
      request = {
        body                   = { query = { match_all = {} } }
        indices                = []
        rest_total_hits_as_int = true
        search_type            = "query_then_fetch"
      }
    }
  })

  throttle_period_in_millis = 10000
}
