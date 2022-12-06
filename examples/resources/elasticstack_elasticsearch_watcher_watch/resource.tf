provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_watcher_watch" "example" {
  watch_id = "test_watch"
  active   = true

  body = jsonencode({
    "trigger" = {
      "schedule" = {
        "daily" = {
          "at" = "noon"
        }
      }
    },
  })
}

output "watch" {
  value = elasticstack_elasticsearch_watcher_watch.example.watch_id
}
