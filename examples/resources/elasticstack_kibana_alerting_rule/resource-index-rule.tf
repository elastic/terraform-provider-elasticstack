provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

// Prerequisites: data stream and index connector referenced by this rule (self-contained snippet).

resource "elasticstack_elasticsearch_index_lifecycle" "my_lifecycle_policy" {
  name = "my_lifecycle_policy"

  hot {
    min_age = "1h"
    set_priority {
      priority = 10
    }
    rollover {
      max_age = "1d"
    }
    readonly {}
  }

  delete {
    min_age = "2d"
    delete {}
  }
}

resource "elasticstack_elasticsearch_component_template" "my_mappings" {
  name = "my_mappings"
  template {
    mappings = jsonencode({
      properties = {
        field1       = { type = "keyword" }
        field2       = { type = "text" }
        "@timestamp" = { type = "date" }
      }
    })
  }
}

resource "elasticstack_elasticsearch_component_template" "my_settings" {
  name = "my_settings"
  template {
    settings = jsonencode({
      "lifecycle.name" = elasticstack_elasticsearch_index_lifecycle.my_lifecycle_policy.name
    })
  }
}

resource "elasticstack_elasticsearch_index_template" "my_index_template" {
  name           = "my_index_template"
  priority       = 500
  index_patterns = ["my-data-stream*"]
  composed_of = [
    elasticstack_elasticsearch_component_template.my_mappings.name,
    elasticstack_elasticsearch_component_template.my_settings.name
  ]
  data_stream {}
}

resource "elasticstack_elasticsearch_data_stream" "my_data_stream" {
  name = "my-data-stream"

  depends_on = [
    elasticstack_elasticsearch_index_template.my_index_template
  ]
}

resource "elasticstack_kibana_action_connector" "index_example" {
  name              = "my_index_connector"
  connector_type_id = ".index"
  config = jsonencode({
    index              = elasticstack_elasticsearch_data_stream.my_data_stream.name
    executionTimeField = "alert_date"
  })
}

resource "elasticstack_kibana_alerting_rule" "DailyDocumentCountThresholdExceeded" {
  name         = "DailyDocumentCountThresholdExceeded"
  consumer     = "alerts"
  rule_type_id = ".index-threshold"
  interval     = "1m"
  enabled      = true
  notify_when  = "onActiveAlert"

  params = jsonencode({
    aggType             = "count"
    thresholdComparator = ">"
    timeWindowSize      = 1
    timeWindowUnit      = "d"
    groupBy             = "all"
    threshold           = [10]
    index               = [elasticstack_elasticsearch_data_stream.my_data_stream.name]
    timeField           = "@timestamp"
  })

  actions {
    id    = elasticstack_kibana_action_connector.index_example.connector_id
    group = "threshold met"
    params = jsonencode({
      "documents" : [{
        "rule_id" : "{{rule.id}}",
        "rule_name" : "{{rule.name}}",
        "message" : "{{context.message}}"
      }]
    })
  }

  actions {
    id    = elasticstack_kibana_action_connector.index_example.connector_id
    group = "recovered"
    params = jsonencode({
      "documents" : [{
        "rule_id" : "{{rule.id}}",
        "rule_name" : "{{rule.name}}",
        "message" : "Recovered"
      }]
    })
  }
}
