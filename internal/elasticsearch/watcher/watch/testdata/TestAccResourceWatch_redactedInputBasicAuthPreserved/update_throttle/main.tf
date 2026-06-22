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

  input = jsonencode({
    http = {
      request = {
        scheme  = "http"
        method  = "get"
        host    = "127.0.0.1"
        port    = 9
        path    = "/"
        params  = {}
        headers = {}
        auth = {
          basic = {
            username = "acc-redacted-input-user"
            password = "acc-redacted-input-secret-4d7e"
          }
        }
      }
    }
  })

  throttle_period_in_millis = 12000
}
