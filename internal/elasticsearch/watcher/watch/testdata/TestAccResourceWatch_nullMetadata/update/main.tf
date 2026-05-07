variable "watch_id" {
  type = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_watch" "test" {
  watch_id = var.watch_id
  active   = false

  trigger   = jsonencode({ schedule = { cron = "0 0/1 * * * ?" } })
  input     = jsonencode({ none = {} })
  condition = jsonencode({ always = {} })
  actions   = jsonencode({})
  metadata  = jsonencode({ example_key = "example_value" })
}
