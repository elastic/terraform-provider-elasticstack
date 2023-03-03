terraform {
  required_providers {
    elasticstack = {
      source  = "elastic/elasticstack"
      version = "0.5.0"
    }
  }
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_watcher_watch" "example" {
  watch_id = "test"
  active   = true

  body = jsonencode({
    "trigger" = {
      "schedule" = {
        "daily" = {
          "at" = "noon"
        }
      }
    },
    "input": {
      "none": {}
    },
    "condition": {
      "always": {}
    },
    "actions": {}
  })
}

# output "watch" {
#   value = elasticstack_elasticsearch_watcher_watch.example.watch_id
# }
