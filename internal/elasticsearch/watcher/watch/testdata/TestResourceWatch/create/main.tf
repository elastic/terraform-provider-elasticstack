variable "watch_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_watch" "test" {
  watch_id = var.watch_id
  active   = false

  trigger = <<EOF
  {
    "schedule" : { "cron" : "0 0/1 * * * ?" }
  }
EOF
}
