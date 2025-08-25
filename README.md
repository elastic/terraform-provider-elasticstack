# Terraform Provider Elastic Stack

[![Acceptance Status](https://github.com/elastic/terraform-provider-elasticstack/actions/workflows/test.yml/badge.svg)](https://github.com/elastic/terraform-provider-elasticstack/actions/workflows/test.yml)

## Use of the provider
The Elastic Stack provider allows you to manage and configure the Elastic stack (Elasticsearch, Kibana, etc) as code using `terraform`.


## Getting started

__The provider supports Elastic Stack versions 7.x+__

It is recommended to setup at least minimum security, https://www.elastic.co/guide/en/elasticsearch/reference/current/security-minimal-setup.html
in order to interact with the Elasticsearch and be able to use the provider's full capabilities.


Configuring [required providers](https://www.terraform.io/docs/language/providers/requirements.html#requiring-providers):

```terraform
terraform {
  required_version = ">= 1.0.0"
  required_providers {
    elasticstack = {
      source  = "elastic/elasticstack"
      version = "~>0.9"
    }
  }
}
```


### Authentication

The Elasticstack provider offers few different ways of providing credentials for authentication.
The following methods are supported:

* Static credentials
* Environment variables


#### Static credentials

Default static credentials can be provided by adding the `username`, `password` and `endpoints` in `elasticsearch` block:

```terraform
provider "elasticstack" {
  elasticsearch {
    username  = "elastic"
    password  = "changeme"
    endpoints = ["http://localhost:9200"]
  }
}
```

Alternatively an `api_key` can be specified instead of `username` and `password`:

```terraform
provider "elasticstack" {
  elasticsearch {
    api_key  = "base64encodedapikeyhere=="
    endpoints = ["http://localhost:9200"]
  }
}
```

#### Environment Variables

You can provide your credentials for the default connection via the `ELASTICSEARCH_USERNAME`, `ELASTICSEARCH_PASSWORD` and comma-separated list `ELASTICSEARCH_ENDPOINTS`,
environment variables, representing your user, password and Elasticsearch API endpoints respectively.

Alternatively the `ELASTICSEARCH_API_KEY` variable can be specified instead of `ELASTICSEARCH_USERNAME` and `ELASTICSEARCH_PASSWORD`.

```terraform
provider "elasticstack" {
  elasticsearch {}
}
```

## Developing the Provider

See [CONTRIBUTING.md](CONTRIBUTING.md)

## Support

We welcome questions on how to use the Elastic providers. The providers are supported by Elastic. General questions, bugs and product issues should be raised in their corresponding repositories, either for the Elastic Stack provider, or the Elastic Cloud one. Questions can also be directed to the discuss forum. https://discuss.elastic.co/c/orchestration.

We will not, however, fix bugs upon customer demand, as we have to prioritize all pending bugs and features, as part of the product's backlog and release cycles.

### Support tickets severity

Support tickets related to the Terraform provider can be opened with Elastic, however since the provider is just a client of the underlying product API's, we will not be able to treat provider related support requests as a Severity-1 (Immedediate time frame). Urgent, production-related Terraform issues can be resolved via direct interaction with the underlying project API or UI. We will ask customers to resort to these methods to resolve downtime or urgent issues.
