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

  # Omits indices, rest_total_hits_as_int, and search_type.
  input = jsonencode({
    search = {
      request = {
        body = { query = { match_all = {} } }
      }
    }
  })
}
