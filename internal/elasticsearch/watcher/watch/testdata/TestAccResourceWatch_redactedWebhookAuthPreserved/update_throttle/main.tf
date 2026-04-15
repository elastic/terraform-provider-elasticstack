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
    acc_webhook = {
      webhook = {
        scheme  = "http"
        host    = "127.0.0.1"
        port    = 9
        method  = "head"
        path    = "/"
        params  = {}
        headers = {}
        auth = {
          basic = {
            username = "acc-redacted-user"
            password = "acc-redacted-webhook-secret-9f2c"
          }
        }
      }
    }
  })

  throttle_period_in_millis = 12000
}
