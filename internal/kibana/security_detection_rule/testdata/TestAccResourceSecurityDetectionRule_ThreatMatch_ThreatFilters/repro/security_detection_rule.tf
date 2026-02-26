variable "name" {
  type = string
}

provider "elasticstack" {
  kibana {}
}

resource "elasticstack_kibana_security_detection_rule" "test" {
  name        = var.name
  description = "Repro for threat_filters drift (issue #1751)"
  type        = "threat_match"

  index    = ["auditbeat-*", "endgame-*", "filebeat-*", "logs-*", "packetbeat-*", "winlogbeat-*"]
  query    = "url.full:*"
  language = "kuery"

  threat_index          = ["filebeat-*", "logs-ti_*"]
  threat_query          = "@timestamp >= \"now-30d/d\" and event.module:(threatintel or ti_*) and threat.indicator.url.full:* and not labels.is_ioc_transform_source:\"true\""
  threat_indicator_path = "threat.indicator"

  threat_mapping = [
    {
      entries = [
        {
          field = "url.full"
          type  = "mapping"
          value = "threat.indicator.url.full"
        }
      ]
    },
    {
      entries = [
        {
          field = "url.original"
          type  = "mapping"
          value = "threat.indicator.url.original"
        }
      ]
    }
  ]

  threat_filters = [
    jsonencode({
      "$state" = { store = "appState" }
      meta = {
        disabled = false
        key      = "event.category"
        negate   = false
        params   = { query = "threat" }
        type     = "phrase"
      }
      query = { match_phrase = { "event.category" = "threat" } }
    }),
    jsonencode({
      "$state" = { store = "appState" }
      meta = {
        disabled = false
        key      = "event.kind"
        negate   = false
        params   = { query = "enrichment" }
        type     = "phrase"
      }
      query = { match_phrase = { "event.kind" = "enrichment" } }
    }),
    jsonencode({
      "$state" = { store = "appState" }
      meta = {
        disabled = false
        key      = "event.type"
        negate   = false
        params   = { query = "indicator" }
        type     = "phrase"
      }
      query = { match_phrase = { "event.type" = "indicator" } }
    }),
  ]

  rule_id    = "f3e22c8b-ea47-45d1-b502-b57b6de950b3"
  severity   = "high"
  risk_score = 73
  from       = "now-65m"
  interval   = "1h"
}

