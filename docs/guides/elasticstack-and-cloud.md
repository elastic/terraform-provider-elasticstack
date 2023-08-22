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
      version = "~>0.7"
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
