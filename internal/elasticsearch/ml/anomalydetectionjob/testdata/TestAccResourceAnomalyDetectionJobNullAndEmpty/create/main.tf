variable "job_id" {
  description = "The job ID for the anomaly detection job"
  type        = string
}

provider "elasticstack" {
  elasticsearch {}
}

resource "elasticstack_elasticsearch_ml_anomaly_detection_job" "test" {
  job_id      = var.job_id
  description = null
  groups      = []

  analysis_config = {
    detectors = [
      {
        function             = "sum"
        field_name           = "bytes"
        detector_description = "Sum of bytes"
        by_field_name        = null
        over_field_name      = null
        partition_field_name = null
        use_null             = false
        exclude_frequent     = null
        custom_rules         = []
      }
    ]

    bucket_span = "15m"
    latency     = null
    period      = null
    influencers = []

    categorization_field_name = null
    categorization_filters    = []

    per_partition_categorization = {
      enabled      = false
      stop_on_warn = false
    }

    summary_count_field_name = null
    multivariate_by_fields   = false
    model_prune_window       = null
  }

  data_description = {
    time_field  = "timestamp"
    time_format = "epoch_ms"
  }

  analysis_limits = {
    model_memory_limit            = "11MB"
    categorization_examples_limit = null
  }

  model_plot_config = {
    enabled             = true
    annotations_enabled = true
    terms               = null
  }
  allow_lazy_open                           = false
  background_persist_interval               = null
  custom_settings                           = null
  daily_model_snapshot_retention_after_days = null
  model_snapshot_retention_days             = null
  renormalization_window_days               = null
  results_index_name                        = "test-job1"
  results_retention_days                    = null
}