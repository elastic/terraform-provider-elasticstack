---
subcategory: "Security"
page_title: "Security roles guide"
description: |-
  Learn when to use Elasticsearch security roles, Kibana security roles, and API key role descriptors, with scenario-based examples and a Kibana feature privilege reference.
---

# Security roles guide

Use this guide when you need to model least-privilege access across Elasticsearch and Kibana. The examples on this page focus on common operational scenarios instead of placeholder values so that you can adapt them to real deployments.

## When to use each resource

Use [`elasticstack_elasticsearch_security_role`](../resources/elasticsearch_security_role) when you need to manage Elasticsearch-native privileges such as cluster privileges, index privileges, field-level security, document-level security, or `run_as`. This is the right resource when the access decision is entirely about Elasticsearch APIs and data.

Use [`elasticstack_kibana_security_role`](../resources/kibana_security_role) when you need to manage Kibana application access such as spaces, base privileges, and feature privileges. A Kibana role also carries an `elasticsearch {}` block, so it is the right choice when the same role should describe both the Elasticsearch privileges and the Kibana feature access a user gets in Kibana.

Use `role_descriptors` on [`elasticstack_elasticsearch_security_api_key`](../resources/elasticsearch_security_api_key) when you need to issue an API key that is narrower than the privileges of the user or role creating it. API key role descriptors use the Elasticsearch role structure and can only further restrict the owner's privileges; they never expand access beyond what the owner already has.

In practice, a common pattern is to define a human-facing Kibana role for interactive access, then mint API keys with narrower `role_descriptors` for automation that should only read from a subset of the same data.

## Scenario examples

### Data analyst

This example grants read-only access to Discover and Dashboards in an `analytics` space, plus read access to log and metric indices.

```terraform
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_security_role" "data_analyst" {
  name = "data_analyst"

  elasticsearch {
    cluster = ["monitor"]

    indices {
      names      = ["logs-*", "metrics-*"]
      privileges = ["read", "view_index_metadata"]
    }
  }

  kibana {
    feature {
      name       = "dashboard"
      privileges = ["read"]
    }

    feature {
      name       = "discover"
      privileges = ["minimal_read", "url_create", "store_search_session"]
    }

    spaces = ["analytics"]
  }
}
```

### Data ingest

This example shows a service-oriented role that can write to a data stream and manage the supporting ingest pipeline and index-template setup. It does not include a Kibana block because the workload does not need Kibana UI access.

```terraform
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_security_role" "data_ingest" {
  name = "data_ingest"

  elasticsearch {
    cluster = ["manage_ingest_pipelines", "manage_index_templates", "auto_configure"]

    indices {
      names      = ["logs-myapp-*"]
      privileges = ["write", "create_index", "auto_configure"]
    }
  }
}
```

### Security analyst

This example grants access to the Kibana security features that an analyst typically needs in a dedicated `security` space.

```terraform
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_security_role" "security_analyst" {
  name = "security_analyst"

  elasticsearch {
    cluster = ["monitor"]

    indices {
      names      = ["logs-*", "metrics-*", "traces-*", ".alerts-security.*"]
      privileges = ["read", "view_index_metadata"]
    }
  }

  kibana {
    feature {
      name       = "actions"
      privileges = ["all"]
    }

    feature {
      name       = "alerting"
      privileges = ["all"]
    }

    feature {
      name       = "osquery"
      privileges = ["all"]
    }

    feature {
      name       = "rulesSettings"
      privileges = ["all"]
    }

    feature {
      name       = "securitySolutionCases"
      privileges = ["all"]
    }

    feature {
      name       = "siem"
      privileges = ["all"]
    }

    spaces = ["security"]
  }
}
```

### DevOps read-only

This example grants read access to operational observability features in Kibana and read-only access to the underlying telemetry indices.

```terraform
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_security_role" "devops_readonly" {
  name = "devops_readonly"

  elasticsearch {
    cluster = ["monitor"]

    indices {
      names      = ["logs-*", "metrics-*", "traces-*"]
      privileges = ["read", "view_index_metadata"]
    }
  }

  kibana {
    feature {
      name       = "apm"
      privileges = ["read"]
    }

    feature {
      name       = "fleet"
      privileges = ["read"]
    }

    feature {
      name       = "infrastructure"
      privileges = ["read"]
    }

    feature {
      name       = "logs"
      privileges = ["read"]
    }

    spaces = ["operations"]
  }
}
```

### Multi-space access

This example uses two `kibana {}` blocks so the same role has broad access in non-production spaces and narrower feature-level access in production.

```terraform
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_security_role" "multi_space" {
  name = "multi_space"

  elasticsearch {
    cluster = ["monitor"]

    indices {
      names      = ["logs-*", "metrics-*", "traces-*"]
      privileges = ["read", "view_index_metadata"]
    }
  }

  kibana {
    base   = ["all"]
    spaces = ["dev", "staging"]
  }

  kibana {
    feature {
      name       = "dashboard"
      privileges = ["read"]
    }

    feature {
      name       = "discover"
      privileges = ["minimal_read", "url_create", "store_search_session"]
    }

    spaces = ["prod"]
  }
}
```

## Field security and document-level security

Use `field_security` when users should see only selected fields from matching documents. This is useful for redacting PII such as email addresses, phone numbers, or social security numbers while still allowing analysts to query the rest of the record.

Use `query` for document-level security when users should only see a subset of documents that match a tenant, business unit, or similar boundary. The example below combines both patterns on the same indices block so a role can be restricted to one tenant and still hide sensitive fields within those documents.

```terraform
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_role" "field_and_doc_security" {
  name = "field_and_doc_security"

  indices {
    names      = ["customer-orders-*"]
    privileges = ["read", "view_index_metadata"]

    field_security {
      grant  = ["customer_id", "tenant_id", "order_id", "order_total", "order_status", "@timestamp"]
      except = ["customer_email", "customer_phone", "customer_ssn"]
    }

    query = jsonencode({
      term = {
        tenant_id = "tenant-a"
      }
    })
  }
}
```

## Composing with API keys

The following example reuses the same access pattern as the data analyst role, then creates an API key whose `role_descriptors` narrow that access to only `logs-myapp-*`. This demonstrates the important rule for API keys: role descriptors can reduce privileges relative to the owning user or role, but they cannot grant additional access.

```terraform
provider "elasticstack" {
  elasticsearch {}
  kibana {}
}

resource "elasticstack_kibana_security_role" "data_analyst" {
  name = "data_analyst"

  elasticsearch {
    cluster = ["monitor"]

    indices {
      names      = ["logs-*", "metrics-*"]
      privileges = ["read", "view_index_metadata"]
    }
  }

  kibana {
    feature {
      name       = "dashboard"
      privileges = ["read"]
    }

    feature {
      name       = "discover"
      privileges = ["minimal_read", "url_create", "store_search_session"]
    }

    spaces = ["analytics"]
  }
}
```

```terraform
provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_security_api_key" "data_analyst_logs_only" {
  name = "data-analyst-logs-only"

  role_descriptors = jsonencode({
    logs_only = {
      cluster = ["monitor"]
      indices = [
        {
          names      = ["logs-myapp-*"]
          privileges = ["read", "view_index_metadata"]
        }
      ]
    }
  })
}
```

## Kibana feature privilege reference

The following table covers commonly used Kibana features and example privilege strings that the Kibana features API exposes. Some deployments and Kibana versions expose versioned or alias IDs for these features; use the feature IDs shown in your deployment's `GET /api/features` response when configuring a role. For the complete list for your environment, query [`GET /api/features`](https://www.elastic.co/guide/en/kibana/current/features-api-get.html).

| Feature name | Available privileges |
| --- | --- |
| `discover` | `all`, `read`, `minimal_all`, `minimal_read`, `url_create`, `store_search_session` |
| `dashboard` | `all`, `read`, `minimal_all`, `minimal_read`, `url_create`, `store_search_session`, `generate_report`, `download_csv_report` |
| `visualize` | `all`, `read`, `minimal_all`, `minimal_read`, `url_create`, `generate_report` |
| `ml` | `all`, `read` |
| `apm` | `all`, `read`, `settings_save` |
| `fleet` | `all`, `read`, `agents_all`, `agents_read`, `agent_policies_all`, `agent_policies_read`, `settings_all`, `settings_read`, `epm_all`, `epm_read`, `integrations_all`, `integrations_read`, `agent_policies_read_integrations`, `fleet_proxies_read`, `fleet_proxies_all`, `remote_synced_integrations_read`, `remote_synced_integrations_all`, `remote_elasticsearch_output_read`, `remote_elasticsearch_output_all`, `secret_fleet_read`, `secret_fleet_all`, `package_settings_read`, `package_settings_all` |
| `siem` | `all`, `read` |
| `securitySolutionCases` | `all`, `read`, `cases_delete`, `cases_settings`, `create_comment`, `case_reopen`, `cases_assign`, `cases_manage_templates` |
| `observabilityCases` | `all`, `read`, `cases_delete`, `cases_settings`, `create_comment`, `case_reopen`, `cases_assign`, `cases_manage_templates` |
| `osquery` | `all`, `read`, `live_queries_read`, `live_queries_all`, `saved_queries_read`, `saved_queries_all`, `packs_read`, `packs_all`, `runs_saved_queries` |
| `rulesSettings` | `all`, `read`, `security_solution_exceptions_all`, `security_solution_investigation_guide_edit`, `security_solution_custom_highlighted_fields_edit`, `security_solution_enable_disable_rules`, `security_solution_manual_run_rules`, `security_solution_rules_management_settings` |
| `actions` | `all`, `read`, `endpoint_security_execute` |
| `alerting` | `all`, `read` |
| `canvas` | `all`, `read`, `generate_report` |
| `maps` | `all`, `read`, `url_create`, `store_in_session`, `create_alert`, `save_query`, `generate_report` |
| `infrastructure` | `all`, `read` |
| `logs` | `all`, `read` |
