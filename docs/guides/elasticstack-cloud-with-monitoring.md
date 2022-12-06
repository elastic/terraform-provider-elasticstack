---
subcategory: ""
page_title: "Using the Elastic Stack provider with multiple Elastic Cloud deployments"
description: |-
    An example of how to spin up multiple Elastic Cloud deployments and configure them using multiple Elastic Stack provider instances.
---

# Using the Elastic Stack provider with multiple Elastic Cloud deployments

Using aliased Elastic Stack providers allows managing multiple Elastic Cloud deployments (or self hosted Elasticsearch clusters).
In this example, we use both the Elastic Cloud provider, as well as the Elastic Stack provider.
We start off by configuring just the Elastic Cloud provider in a `provider.tf` file, for example:

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
```

Next, we'll set up two Elastic Cloud `ec_deployment` resources, which represent Elastic Stack deployments on Elastic Cloud.
The `monitoring` deployment is configured as a dedicated monitoring deployment, with the `cluster` deployment configured to ship
monitoring data to the `monitoring` deployment.

We also configure two instances of the Elastic Stack provider, including an alias for the instance connected to the `monitoring` deployment.

Finally, we configure the Elastic Stack resources. When provisioning monitoring resources, we include an `provider = elasticstack.monitoring`
attribute to target the intended deployment. Aliased providers can be configured on a per-resource or per-module basis.
For more information consult the [documentation](https://developer.hashicorp.com/terraform/language/providers/configuration#alias-multiple-provider-configurations)

```terraform
# Creating a deployment on Elastic Cloud GCP region,
# with elasticsearch and kibana components.

resource "ec_deployment" "monitoring" {
  region                 = "gcp-us-central1"
  name                   = "my-monitoring-deployment"
  version                = data.ec_stack.latest.version
  deployment_template_id = "gcp-storage-optimized"

  elasticsearch {}
  kibana {}
}

resource "ec_deployment" "cluster" {
  region                 = "gcp-us-central1"
  name                   = "mydeployment"
  version                = data.ec_stack.latest.version
  deployment_template_id = "gcp-storage-optimized"

  observability {
    deployment_id = ec_deployment.monitoring.id
    ref_id        = ec_deployment.monitoring.elasticsearch[0].ref_id
  }

  elasticsearch {}

  kibana {}
}

data "ec_stack" "latest" {
  version_regex = "latest"
  region        = "gcp-us-central1"
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
