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

Next, we'll set up two Elastic Cloud `ec_deployment` resources, which represent Elastic Stack deployments on Elastic Cloud.
The `monitoring` deployment is configured as a dedicated monitoring deployment, with the `cluster` deployment configured to ship
monitoring data to the `monitoring` deployment.

We also configure two instances of the Elastic Stack provider, including an alias for the instance connected to the `monitoring` deployment.

Finally, we configure the Elastic Stack resources. When provisioning monitoring resources, we include an `provider = elasticstack.monitoring`
attribute to target the intended deployment. Aliased providers can be configured on a per-resource or per-module basis.
For more information consult the [documentation](https://developer.hashicorp.com/terraform/language/providers/configuration#alias-multiple-provider-configurations)

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
