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
    acc_script_webhook = {
      webhook = {
        scheme = "http"
        host   = "127.0.0.1"
        port   = 9
        method = "head"
        path   = "/"
        params = {}
        headers = {
          "Content-Type" = "application/json"
          "Authorization" = {
            source = "return 'Bearer acc-script-header-3a91'"
            lang   = "painless"
          }
        }
      }
    }
  })

  throttle_period_in_millis = 12000
}
