---
subcategory: ""
page_title: "Using the Elastic Stack provider with Elastic Cloud"
description: |-
    An example of how to spin up a deployment and configure it in a single plan.
---

# Using the Elastic Stack provider with Elastic Cloud



## Creating deployments

A common scenario for using the Elastic Stack provider, is to manage & configure Elastic Cloud deployments.
In order to do that, we'll use both the Elastic Cloud provider, as well as the Elastic Stack provider.

Start off by configuring just the Elastic Cloud provider in a `provider.tf` file for example:

```terraform
terraform {
  required_version = ">= 1.0.0"

  required_providers {
    ec = {
      source = "elastic/ec"
    }
    elasticstack = {
      source  = "elastic/elasticstack"
      version = "~>0.6"
    }
  }
}

provider "ec" {
  # You can fill in your API key here, or use an environment variable TF_VAR_ec_apikey instead
  # For details on how to generate an API key, see: https://www.elastic.co/guide/en/cloud/current/ec-api-authentication.html.
  apikey = var.ec_apikey
}
```

Note that the provider needs to be configured with an API key which has to be created in the [Elastic Cloud console](https://www.elastic.co/guide/en/cloud/current/ec-api-authentication.html). With the provider configured, we can now use it to create some deployments. Here, we're creating two deployments -- one for your data, and the other is setup as a separate monitor. 

```terraform
# Creating deployments on Elastic Cloud GCP region with elasticsearch and kibana components. One deployment is a dedicated monitor for the other. 

data "ec_stack" "latest" {
  version_regex = "latest"
  region        = var.region
}

resource "ec_deployment" "monitoring" {
  region                 = var.region
  name                   = "my-monitoring-deployment"
  version                = data.ec_stack.latest.version
  deployment_template_id = var.deployment_template_id

  elasticsearch {}
  kibana {}
}

resource "ec_deployment" "cluster" {
  region                 = var.region
  name                   = "my-deployment"
  version                = data.ec_stack.latest.version
  deployment_template_id = var.deployment_template_id

  observability {
    deployment_id = ec_deployment.monitoring.id
    ref_id        = ec_deployment.monitoring.elasticsearch[0].ref_id
  }

  elasticsearch {}
  kibana {}
}

provider "elasticstack" {
  # Use our Elastic Cloud deployment outputs for connection details.
  # This also allows the provider to create the proper relationships between the two resources.
  elasticsearch {
    endpoints = ["${ec_deployment.cluster.elasticsearch[0].https_endpoint}"]
    username  = ec_deployment.cluster.elasticsearch_username
    password  = ec_deployment.cluster.elasticsearch_password
  }
}

provider "elasticstack" {
  # Use our Elastic Cloud deployment outputs for connection details.
  # This also allows the provider to create the proper relationships between the two resources.
  elasticsearch {
    endpoints = ["${ec_deployment.monitoring.elasticsearch[0].https_endpoint}"]
    username  = ec_deployment.monitoring.elasticsearch_username
    password  = ec_deployment.monitoring.elasticsearch_password
  }
  alias = "monitoring"
}
```

Notice that the Elastic Stack  provider setup with empty `elasticsearch {}` block, since we'll be using an `elasticsearch_connection` block
for each of our resources, to point to the Elastic Cloud deployment.



## Managing stack resources and configuration

Now we can add resources to these deployments as follows:

```terraform
# Defining a user for ingesting
resource "elasticstack_elasticsearch_security_user" "user" {
  username = "ingest_user"

  # Password is cleartext here for comfort, but there's also a hashed password option
  password = "mysecretpassword"
  roles    = ["editor"]

  # Set the custom metadata for this user
  metadata = jsonencode({
    "env"    = "testing"
    "open"   = false
    "number" = 49
  })
}

# Configuring my cluster with an index template as well.
resource "elasticstack_elasticsearch_index_template" "my_template" {
  name = "my_ingest_1"

  priority       = 42
  index_patterns = ["server-logs*"]

  template {
    alias {
      name = "my_template_test"
    }

    settings = jsonencode({
      number_of_shards = "3"
    })

    mappings = jsonencode({
      properties : {
        "@timestamp" : { "type" : "date" },
        "username" : { "type" : "keyword" }
      }
    })
  }
}

# Defining a user for viewing monitoring
resource "elasticstack_elasticsearch_security_user" "monitoring_user" {
  # Explicitly select the monitoring provider here
  provider = elasticstack.monitoring

  username = "monitoring_viewer"

  # Password is cleartext here for comfort, but there's also a hashed password option
  password = "mysecretpassword"
  roles    = ["reader"]

  # Set the custom metadata for this user
  metadata = jsonencode({
    "env"    = "testing"
    "open"   = false
    "number" = 49
  })
}
```

Note that resources can be targed to certain deployments using the `provider` attribute. 

## Example: Creating an index threshold rule

You can use rules to detect complex conditions and generate alerts and actions when those conditions are met.

For example, let's take a simple data stream that contains some logs or metrics:

```terraform
provider "elasticstack" {
  elasticsearch {}
}

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

You can define an index threshold rule to detect when your data stream exceeds a threshold.
In this simple example, the rule checks whether the count of all documents in the data stream exceeds 10 over a period of 1 day:

```terraform
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_kibana_alerting_rule" "my_rule" {
  name         = "my_rule"
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
}
```

NOTE: When you've finished playing with this example, remember to destroy the rule resource, since it has a low threshold and generates a lot of documents for testing purposes.

There are many different methods that you can use to be notified when the threshold is met and when it recovers.
In this example, the rule uses an index connector to write a document in an Elasticsearch index when the rule conditions are met:

```terraform
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_kibana_action_connector" "index_example" {
  name              = "my_index_connector"
  connector_type_id = ".index"
  config = jsonencode({
    index              = elasticstack_elasticsearch_index.my_index.name
    executionTimeField = "alert_date"
  })
}
```

The `executionTimeField` is optional; in this case we set it so that each document will contain a timestamp that indicates when the alert occurred.
For example, the index connector can generate documents in a simple index like this:

```terraform
provider "elasticstack" {
  elasticsearch {}
}

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

After you apply these resources, add some documents to the data stream:
https://www.elastic.co/guide/en/elasticsearch/reference/current/use-a-data-stream.html#add-documents-to-a-data-stream

For example:

````
PUT my-data-stream/_bulk
{ "create":{ } }
{ "@timestamp": "2023-07-04T16:21:15.000Z", "field1": "host1", "field2": "test message" }
{ "create":{ } }
{ "@timestamp": "2023-07-04T16:25:42.000Z", "field1": "host2", "field2": "test message" }

````

When the rule conditions are met, you'll start to see documents added to `my-index`.
Try out different rule action variables to customize the notification message:
https://www.elastic.co/guide/en/kibana/current/rule-action-variables.html

To learn about more types of rules and connectors, check out https://www.elastic.co/guide/en/kibana/current/rule-types.html and https://www.elastic.co/guide/en/kibana/current/action-types.html.


