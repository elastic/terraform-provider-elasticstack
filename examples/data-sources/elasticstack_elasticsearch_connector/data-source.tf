provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_connector" "postgres" {
  connector_id = "music-catalog"
  service_type = "postgresql"
  name         = "music catalog"
  description  = "Indexes the music catalog database."
  index_name   = "search-music"
  language     = "english"

  pipeline {
    name                   = "search-default-ingestion"
    extract_binary_content = false
    reduce_whitespace      = true
    run_ml_inference       = false
  }

  scheduling {
    full {
      enabled  = true
      interval = "0 0 * * * ?"
    }
    incremental {
      enabled  = false
      interval = "0 0 0 * * ?"
    }
    access_control {
      enabled  = false
      interval = "0 0 0 * * ?"
    }
  }

  features {
    sync_rules {
      basic {
        enabled = true
      }
      advanced {
        enabled = false
      }
    }
    document_level_security {
      enabled = false
    }
    incremental_sync {
      enabled = false
    }
    native_connector_api_keys {
      enabled = false
    }
  }
}

data "elasticstack_elasticsearch_connector" "lookup" {
  connector_id = elasticstack_elasticsearch_connector.postgres.connector_id
}

output "connector_status" {
  value = data.elasticstack_elasticsearch_connector.lookup.status
}

output "connector_configuration" {
  value = data.elasticstack_elasticsearch_connector.lookup.configuration
}

output "connector_last_synced" {
  value = data.elasticstack_elasticsearch_connector.lookup.last_synced
}
