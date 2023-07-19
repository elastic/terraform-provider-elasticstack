---
subcategory: ""
page_title: "Managing Kibana rule and connector resources"
description: |-
    An example of how to define an index connector and an index threshold rule.
---
# Managing Kibana rule and connector resources

## Prerequisites

This example assumes you have already set up your provider to access an Elastic Stack cluster or a Elastic Cloud deployment.

To use the Kibana alerting features, you must have the appropriate feature privileges.
For example, to create Stack rules such as the index threshold rule, you must have `all` privileges for the **Management > Stack Rules** feature in Kibana.
To add rule actions and test connectors, you must also have `read` privileges for the **Actions and Connectors** feature in Kibana.
For more information, refer to [Kibana alerting documentation](https://www.elastic.co/guide/en/kibana/current/alerting-setup.html#alerting-prerequisites).

If the Elasticsearch security features are enabled, to set up a data stream, index lifecycle policy, component template, and index template, you must have the appropriate cluster and index privileges.
For example, you must have the `manage_ilm` and `manage_index_templates` or `manage` cluster privileges and `manage` index privilege on the indices.
For more information about the steps required to set up a data stream, refer to [Elasticsearch data stream documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/set-up-a-data-stream.html).

## Create a data stream
You can use rules to detect complex conditions and generate alerts and actions when those conditions are met.

For example, let's take a simple data stream that contains some logs or metrics:

```terraform
// Create an ILM policy for our data stream
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

// Create a component template for mappings
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

// Create a component template for index settings
resource "elasticstack_elasticsearch_component_template" "my_settings" {
  name = "my_settings"
  template {
    settings = jsonencode({
      "lifecycle.name" = elasticstack_elasticsearch_index_lifecycle.my_lifecycle_policy.name
    })
  }
}

// Create an index template that uses the component templates
resource "elasticstack_elasticsearch_index_template" "my_index_template" {
  name           = "my_index_template"
  priority       = 500
  index_patterns = ["my-data-stream*"]
  composed_of    = [elasticstack_elasticsearch_component_template.my_mappings.name, elasticstack_elasticsearch_component_template.my_settings.name]
  data_stream {}
}

// Create a data stream based on the index template
resource "elasticstack_elasticsearch_data_stream" "my_data_stream" {
  name = "my-data-stream"

  // Make sure that template is created before the data stream
  depends_on = [
    elasticstack_elasticsearch_index_template.my_index_template
  ]
}
```

## Create an index connector

There are many different methods that you can use to be notified when the conditions of your rule are met.
In this example, we will use an index connector to write a document in an Elasticsearch index:

```terraform
resource "elasticstack_elasticsearch_index" "my_index" {
  name = "my-index"
  mappings = jsonencode({
    properties = {
      alert_date = { type = "date", format = "date_optional_time||epoch_millis" }
      rule_id    = { type = "text" }
      rule_name  = { type = "text" }
      message    = { type = "text" }
    }
  })
}
```

When you define the connector, you can optionally specify an `executionTimeField`:

```terraform
resource "elasticstack_kibana_action_connector" "index_example" {
  name              = "my_index_connector"
  connector_type_id = ".index"
  config = jsonencode({
    index              = elasticstack_elasticsearch_index.my_index.name
    executionTimeField = "alert_date"
  })
}
```

Each document that the connector creates will contain a timestamp in this `executionTimeField`, which indicates when the alert occurred.

## Create an index threshold rule

You can now create an index threshold rule that detects when your data stream exceeds a threshold and sends notifications by using the index connector.
In this example, the rule checks whether the count of all documents in the data stream exceeds 10 over a period of 1 day:

```terraform
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
    index               = elasticstack_elasticsearch_data_stream.my_data_stream.name
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
```

In this example, the `notify_when` property is set to `onActiveAlert`, which means actions run when the alert becomes active and at each check interval (every 1 minute in this case) while the rule conditions are met.
If you want fewer documents to be created, use `onActionGroupChange` or `onThrottleInterval` instead.
For more information about action frequency, refer to [Actions](https://www.elastic.co/guide/en/kibana/current/create-and-manage-rules.html#defining-rules-actions-details).

## Test your rule and connector in Kibana

After you apply these resources, you can play with them in Kibana.

For example, test the connector in the **Stack Management** app to verify that it creates documents in your index.

Then add some documents to the data stream. For example, in the **Dev Console**:

````
PUT my-data-stream/_bulk
{ "create":{ } }
{ "@timestamp": "2023-07-04T16:21:15.000Z", "field1": "host1", "field2": "test message" }
{ "create":{ } }
{ "@timestamp": "2023-07-04T16:25:42.000Z", "field1": "host2", "field2": "test message" }

````

For more details, refer to [Add documents to a data stream](https://www.elastic.co/guide/en/elasticsearch/reference/current/use-a-data-stream.html#add-documents-to-a-data-stream).

When the rule conditions are met, you'll start to see documents added to your index.

Try out different [rule action variables](https://www.elastic.co/guide/en/kibana/current/rule-action-variables.html) to customize the notification message.

To learn about more types of rules and connectors, check out [Rule types](https://www.elastic.co/guide/en/kibana/current/rule-types.html) and [Connectors](https://www.elastic.co/guide/en/kibana/current/action-types.html).

## Clean up your resources

When you've finished playing with this example, remember to destroy the rule resource in particular, since it has a low threshold and generates a lot of documents for testing purposes.

For example, comment out or delete your resource files or run the `terraform destroy` command with the appropriate target resources.

