provider "elasticstack" {
  elasticsearch {}
}

# Self-managed PostgreSQL connector
resource "elasticstack_elasticsearch_connector" "postgres" {
  connector_id = "music-catalog"
  service_type = "postgresql"
  name         = "music catalog"
  description  = "Indexes the music catalog database."
  index_name   = "search-music"
  language     = "english"

  pipeline = {
    name                   = "search-default-ingestion"
    extract_binary_content = false
    reduce_whitespace      = true
    run_ml_inference       = false
  }

  scheduling = {
    full = {
      enabled  = true
      interval = "0 0 * * * ?"
    }
    incremental = {
      enabled  = false
      interval = "0 0 0 * * ?"
    }
    access_control = {
      enabled  = false
      interval = "0 0 0 * * ?"
    }
  }

  features = {
    sync_rules = {
      basic = {
        enabled = true
      }
      advanced = {
        enabled = false
      }
    }
    document_level_security = {
      enabled = false
    }
    incremental_sync = {
      enabled = false
    }
    native_connector_api_keys = {
      enabled = false
    }
  }

  # configuration_values requires the connector service to have registered a
  # configuration schema for `service_type = "postgresql"`. Uncomment after
  # the connector service has booted.
  #
  # configuration_values = {
  #   host     = { string = "db.internal" }
  #   port     = { number = 5432 }
  #   username = { string = "indexer" }
  #   password = { secret_value = var.postgres_password }
  #   database = { string = "music" }
  # }
}
