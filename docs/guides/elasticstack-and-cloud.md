---
subcategory: ""
page_title: "Using the Elastic Stack provider with Elastic Cloud"
description: |-
    An example of how to spin up a deployment and configure it in a single plan.
---

# Using the Elastic Stack provider with Elastic Cloud

A common scenario for using the Elastic Stack provider, is to manage & configure Elastic Cloud deployments.
In order to do that, we'll use both the Elastic Cloud provider, as well as the Elastic Stack provider.
Start off by configuring just the Elastic Cloud provider in a `provider.tf` file for example:

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
      version = "~>0.5.0"
    }
  }
}

provider "ec" {
  # You can fill in your API key here, or use an environment variable TF_VAR_ec_apikey instead
  # For details on how to generate an API key, see: https://www.elastic.co/guide/en/cloud/current/ec-api-authentication.html.
  apikey = var.ec_apikey
}
```

Notice that the Elastic Stack  provider setup with empty `elasticsearch {}` block, since we'll be using an `elasticsearch_connection` block
for each of our resources, to point to the Elastic Cloud deployment.

Next, we'll set up an Elastic Cloud `ec_deployment` resource, which represents an Elastic Stack deployment on Elastic Cloud.
We shall configure the deployment using the credentials that it outputs once created

```terraform
# Creating deployments on Elastic Cloud GCP region with elasticsearch and kibana components. One deployment is a dedicated monitor for the other. 

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

data "ec_stack" "latest" {
  version_regex = "latest"
  region        = var.region
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
