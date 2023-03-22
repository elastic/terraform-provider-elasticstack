provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_watch" "example" {
  watch_id = "test_watch"
  active   = true

  trigger = jsonencode({
    "schedule" = {
      "cron" = "0 0/1 * * * ?"
    }
  })
  input = jsonencode({
    "none" = {}
  })
  condition = jsonencode({
    "always" = {}
  })
  actions = jsonencode({})
  metadata = jsonencode({
    "example_key" = "example_value"
  })
  throttle_period_in_millis = 10000
}

output "watch" {
  value = elasticstack_elasticsearch_watch.example.watch_id
}
