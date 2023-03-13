provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_watcher_watch" "example" {
  watch_id = "test_watch"
  active   = true

  body = jsonencode({
    "trigger" = {
      "schedule" = {
        "cron" = "0 0/1 * * * ?"
      }
    },
    "input" = {
      "none" = {}
    },
    "condition" = {
      "always" = {}
    },
    "actions" = {},
    "metadata" = {
      "example_key" = "example_value"
    },
    "throttle_period_in_millis" = 10000
  })
}

output "watch" {
  value = elasticstack_elasticsearch_watcher_watch.example.watch_id
}
