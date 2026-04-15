variable "watch_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_watch" "test" {
  watch_id = var.watch_id
  active   = true

  trigger = <<EOF
  {
    "schedule" : { "cron" : "0 0/2 * * * ?" }
  }
EOF

  input = <<EOF
  {
    "simple" : {
      "name" : "example"
    }
  }
EOF

  condition = <<EOF
  {
    "never" : {}
  }
EOF

  actions = <<EOF
  {
    "log" : {
      "logging" : {
        "level" : "info",
        "text" : "example logging text"
      }
    }
  }
EOF

  metadata = <<EOF
  {
    "example_key" : "example_value"
  }
EOF

  transform = <<EOF
  {
    "search" : {
      "request" : {
        "body" : {
          "query" : {
            "match_all" : {}
          }
        },
        "indices": [],
        "rest_total_hits_as_int" : true,
        "search_type": "query_then_fetch"
      }
    }
  }
EOF

  throttle_period_in_millis = 10000
}
