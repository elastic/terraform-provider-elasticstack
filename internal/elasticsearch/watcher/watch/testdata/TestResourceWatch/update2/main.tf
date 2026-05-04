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
  input     = jsonencode({ simple = { count = 2, environment = "staging" } })
  condition = jsonencode({ script = { source = "return true", lang = "painless" } })
  actions   = jsonencode({ log = { logging = { level = "info", text = "example logging text" } } })
  metadata  = jsonencode({ env = "staging", priority = 2 })

  transform = jsonencode({ script = { source = "return ctx.payload", lang = "painless" } })

  throttle_period_in_millis = 15000
}
