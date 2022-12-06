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
      version = "~> 0.5.0"
    }
  }
}
```


### Authentication

The Elasticstack provider offers few different ways of providing credentials for authentication.
The following methods are supported:

* Static credentials
* Environment variables
* Each `elasticsearch` resource supports `elasticsearch_connection` block - allowing using the same provider to configure many different clusters at the same time


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


#### Per resource credentials

See docs related to the specific resources.


## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements)).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To install the provider locally into the `~/.terraform.d/plugins/...` directory one can use `make install` command. This will allow to refer this provider dirrecty in the Terraform configuration without needing to download it from the registry.

To generate or update documentation, run `make gen`. All the generated docs will have to be committed to the repository as well.

In order to run the full suite of Acceptance tests, run `make testacc`.

If you have [Docker](https://docs.docker.com/get-docker/) installed, you can use following command to start the Elasticsearch container and run Acceptance tests against it:

```sh
$ make docker-testacc
```

To clean up the used containers and to free up the assigned container names, run `make docker-clean`.


### Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0.0
- [Go](https://golang.org/doc/install) >= 1.19


### Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the `make install` command:
```sh
$ make install
```


### Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.
