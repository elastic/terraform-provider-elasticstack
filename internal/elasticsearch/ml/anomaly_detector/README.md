# ML Anomaly Detection Job Resource

This resource creates and manages Machine Learning anomaly detection jobs in Elasticsearch. Anomaly detection identifies unusual patterns in data based on historical data patterns.

## Key Features

- **Complete API Coverage**: Supports all ML anomaly detection job API options including:
  - Analysis configuration with detectors, influencers, and bucket span
  - Data description for time-based data
  - Analysis limits for memory management
  - Model plot configuration for detailed views
  - Datafeed configuration for data ingestion
  - Custom settings and retention policies

- **Job Lifecycle Management**: 
  - Create new anomaly detection jobs
  - Update job configurations (limited fields)
  - Delete jobs (automatically closes jobs before deletion)
  - Import existing jobs

- **Framework Migration**: Built using Terraform Plugin Framework for better performance and type safety

## Supported Operations

### Create Job
- PUT `/_ml/anomaly_detectors/{job_id}`
- Supports all job configuration options
- Includes optional datafeed configuration

### Read Job
- GET `/_ml/anomaly_detectors/{job_id}`
- Retrieves current job configuration and status

### Update Job
- POST `/_ml/anomaly_detectors/{job_id}/_update`
- Updates modifiable job properties:
  - description
  - groups
  - model_plot_config
  - analysis_limits.model_memory_limit
  - renormalization_window_days
  - results_retention_days
  - custom_settings
  - background_persist_interval

### Delete Job
- POST `/_ml/anomaly_detectors/{job_id}/_close` (if needed)
- DELETE `/_ml/anomaly_detectors/{job_id}`

## Configuration Examples

### Basic Count Detector
```hcl
resource "elasticstack_elasticsearch_ml_anomaly_detector" "basic" {
  job_id      = "basic-count-job"
  description = "Basic count anomaly detection"

  analysis_config = {
    bucket_span = "15m"
    detectors = [
      {
        function = "count"
        detector_description = "Count anomalies"
      }
    ]
  }

  data_description = {
    time_field  = "@timestamp"
    time_format = "epoch_ms"
  }

  analysis_limits = {
    model_memory_limit = "10mb"
  }
}
```

### Advanced Multi-Detector Job
```hcl
resource "elasticstack_elasticsearch_ml_anomaly_detector" "advanced" {
  job_id      = "advanced-web-analytics"
  description = "Advanced web analytics anomaly detection"
  groups      = ["web", "analytics"]

  analysis_config = {
    bucket_span = "15m"
    detectors = [
      {
        function = "count"
        by_field_name = "client_ip"
        detector_description = "High request count per IP"
      },
      {
        function = "mean"
        field_name = "response_time"
        over_field_name = "url.path"
        detector_description = "Response time anomalies by URL"
      },
      {
        function = "distinct_count"
        field_name = "user_id"
        detector_description = "Unique user count anomalies"
      }
    ]
    influencers = ["client_ip", "url.path", "status_code"]
  }

  data_description = {
    time_field  = "@timestamp"
    time_format = "epoch_ms"
  }

  analysis_limits = {
    model_memory_limit = "100mb"
    categorization_examples_limit = 10
  }

  model_plot_config = {
    enabled = true
    annotations_enabled = true
  }

  datafeed_config = {
    datafeed_id = "datafeed-advanced-web-analytics"
    indices = ["web-logs-*"]
    query = jsonencode({
      bool = {
        filter = [
          {
            range = {
              "@timestamp" = {
                gte = "now-7d"
              }
            }
          }
        ]
      }
    })
    frequency = "30s"
    query_delay = "60s"
    scroll_size = 1000
  }

  model_snapshot_retention_days = 30
  results_retention_days = 90
  daily_model_snapshot_retention_after_days = 7
}
```

### Categorization Job
```hcl
resource "elasticstack_elasticsearch_ml_anomaly_detector" "categorization" {
  job_id      = "log-categorization"
  description = "Log message categorization job"

  analysis_config = {
    bucket_span = "1h"
    categorization_field_name = "message"
    categorization_filters = [
      "\\b\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\b",  # IP addresses
      "\\b[A-Fa-f0-9]{8,}\\b"                             # Hex values
    ]
    detectors = [
      {
        function = "count"
        by_field_name = "mlcategory"
        detector_description = "Log category count anomalies"
      }
    ]
    per_partition_categorization = {
      enabled = true
      stop_on_warn = true
    }
  }

  data_description = {
    time_field  = "@timestamp"
    time_format = "epoch_ms"
  }

  analysis_limits = {
    model_memory_limit = "200mb"
    categorization_examples_limit = 20
  }
}
```

## Field Validation

The resource includes comprehensive validation:

- **job_id**: Must contain only lowercase alphanumeric characters, hyphens, and underscores
- **bucket_span**: Must be a valid time interval (e.g., "15m", "1h")
- **detector.function**: Must be one of the supported ML functions
- **memory_limit**: Must be a valid memory size format
- **time_format**: Supports epoch, epoch_ms, or custom patterns

## Import Support

Existing ML anomaly detection jobs can be imported:

```bash
terraform import elasticstack_elasticsearch_ml_anomaly_detector.example existing-job-id
```

## Error Handling

The resource handles various error scenarios:

- **Job not found**: Gracefully removes resource from state
- **Insufficient ML capacity**: Provides clear error messages
- **Configuration conflicts**: Validates detector configurations
- **Memory limits**: Warns about memory usage patterns

## Best Practices

1. **Memory Sizing**: Start with conservative memory limits and increase as needed
2. **Bucket Span**: Choose appropriate bucket spans based on data frequency
3. **Detectors**: Use specific field names for better anomaly detection
4. **Influencers**: Include relevant fields that might influence anomalies
5. **Datafeeds**: Use appropriate query delays for real-time data

## Limitations

- Some job properties cannot be updated after creation (analysis_config structure)
- Jobs must be closed before deletion (handled automatically)
- Datafeed creation is included but separate datafeed management is recommended for complex scenarios
