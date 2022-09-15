---
subcategory: ""
page_title: "Using the Elastic Stack provider with Elastic Cloud"
description: |-
    An example of how to spin up a deployment and configure it in a single plan.
---

# Using the Elastic Stack provider with Elastic Cloud

A common scenario for using the Elastic Stack provider, is to manage & configure Elastic Cloud deployments.
In order to do that, we'll use both the Elastic Cloud provider, as well as the Elastic Stack provider.
Start off by configuring the two providers in a `provider.tf` file for example:

```terraform
terraform {
  required_version = ">= 1.0.0"

  required_providers {
    ec = {
      source  = "elastic/ec"
      version = "~>0.3.0"
    }
    elasticstack = {
      source  = "elastic/elasticstack"
      version = "~>0.4.0"
    }
  }
}
provider "ec" {
  # You can fill in your API key here, or use an environment variable instead
  apikey = "<api key>"
}

provider "elasticstack" {
  # In this example, connectivity to Elasticsearch is defined per resource
  elasticsearch {}
}
```

Notice that the Elastic Stack  provider setup with empty `elasticsearch {}` block, since we'll be using an `elasticsearch_connection` block
for each of our resources, to point to the Elastic Cloud deployment.

Next, we'll set up an Elastic Cloud `ec_deployment` resource, which represents an Elastic Stack deployment on Elastic Cloud.
We shall configure the deployment using the credentials that it outputs once created

```terraform
# Creating a deployment on Elastic Cloud GCP region,
# with elasticsearch and kibana components.
resource "ec_deployment" "cluster" {
  region                 = "gcp-us-central1"
  name                   = "mydeployment"
  version                = data.ec_stack.latest.version
  deployment_template_id = "gcp-storage-optimized"

  elasticsearch {}

  kibana {}
}

data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "gcp-us-central1"
}

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

  # Use our Elastic Cloud deployemnt outputs for connection details.
  # This also allows the provider to create the proper relationships between the two resources.
  elasticsearch_connection {
    endpoints = ["${ec_deployment.cluster.elasticsearch[0].https_endpoint}"]
    username  = ec_deployment.cluster.elasticsearch_username
    password  = ec_deployment.cluster.elasticsearch_password
  }
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

  elasticsearch_connection {
    endpoints = ["${ec_deployment.cluster.elasticsearch[0].https_endpoint}"]
    username  = ec_deployment.cluster.elasticsearch_username
    password  = ec_deployment.cluster.elasticsearch_password
  }
}
```
