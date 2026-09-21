# Provider configuration

## required_providers

Always emit this, pinned:

```terraform
terraform {
  required_version = ">= 1.0.0"
  required_providers {
    elasticstack = {
      source  = "elastic/elasticstack"
      version = "~> 0.14"
    }
  }
}
```

## Minimal provider block

The provider supports four independent subsystems. Configure only what you use.

```terraform
provider "elasticstack" {
  elasticsearch {}
  kibana {}
  fleet {}
}
```

Empty blocks pull credentials from environment variables. This is the preferred path — it keeps secrets out of HCL and state.

## Environment variables

| Subsystem | Endpoint env | Auth envs |
|---|---|---|
| Elasticsearch | `ELASTICSEARCH_ENDPOINTS` | `ELASTICSEARCH_USERNAME` + `ELASTICSEARCH_PASSWORD`, or `ELASTICSEARCH_API_KEY`, or `ELASTICSEARCH_BEARER_TOKEN` |
| Kibana | `KIBANA_ENDPOINT` | `KIBANA_USERNAME` + `KIBANA_PASSWORD`, or `KIBANA_API_KEY` |
| Fleet | `FLEET_ENDPOINT` | `FLEET_USERNAME` + `FLEET_PASSWORD`, `FLEET_API_KEY`, or a service token |

## Inline credentials (only when env vars are impractical)

```terraform
provider "elasticstack" {
  elasticsearch {
    endpoints = ["https://es.example:9243"]
    api_key   = var.es_api_key
  }
  kibana {
    endpoints = ["https://kibana.example:9243"]
    api_key   = var.kibana_api_key
  }
}
```

Always source secrets from variables, never inline literals.

## Multiple stacks

Use provider aliases when managing more than one deployment:

```terraform
provider "elasticstack" {
  alias = "prod"
  elasticsearch { endpoints = [var.prod_es] }
}

provider "elasticstack" {
  alias = "staging"
  elasticsearch { endpoints = [var.staging_es] }
}

resource "elasticstack_elasticsearch_index" "prod_idx" {
  provider = elasticstack.prod
  name     = "events"
}
```

## Elastic Cloud integration

When also using `elastic/ec` to create deployments, feed outputs into `elasticstack`:

```terraform
provider "elasticstack" {
  elasticsearch {
    username  = ec_deployment.this.elasticsearch_username
    password  = ec_deployment.this.elasticsearch_password
    endpoints = [ec_deployment.this.elasticsearch.https_endpoint]
  }
  kibana {
    username  = ec_deployment.this.elasticsearch_username
    password  = ec_deployment.this.elasticsearch_password
    endpoints = [ec_deployment.this.kibana.https_endpoint]
  }
}
```
