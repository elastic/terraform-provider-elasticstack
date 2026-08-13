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

  actions = jsonencode({
    log = {
      transform = {
        search = {
          request = {
            indices = [".monitoring-es-*"]
            body    = { query = { match_all = {} } }
          }
        }
      }
      logging = {
        level = "info"
        text  = "watch fired"
      }
    }
  })
}
