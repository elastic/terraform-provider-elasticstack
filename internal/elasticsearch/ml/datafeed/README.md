# ML Datafeed Resource

This resource creates and manages Machine Learning datafeeds in Elasticsearch. Datafeeds retrieve data from Elasticsearch for analysis by an anomaly detection job. Each anomaly detection job can have only one associated datafeed.

## Key Features

- **Complete API Coverage**: Supports all ML datafeed API options including:
  - Index patterns and queries for data retrieval
  - Aggregations for time-based data summaries
  - Chunking configuration for long time periods
  - Delayed data check configuration
  - Runtime mappings and script fields
  - Frequency and query delay settings
  - Scroll size optimization
  - Custom headers support

- **Orchestration**: Manages datafeed lifecycle including:
  - Automatic stopping before updates (as required by Elasticsearch)
  - Restarting after successful updates
  - Proper cleanup on deletion

- **Validation**: Comprehensive field validation including:
  - Datafeed ID format validation (lowercase alphanumeric, hyphens, underscores)
  - Duration format validation for frequency and query_delay
  - JSON validation for complex objects

## Implementation Details

This resource follows the Plugin Framework architecture and includes:
- Complete schema definition with all API fields
- Proper model conversion between API and Terraform types
- Comprehensive CRUD operations
- Extensive acceptance tests
- Import functionality for existing datafeeds

## API References

- [Create Datafeed API](https://www.elastic.co/guide/en/elasticsearch/reference/current/ml-put-datafeed.html)
- [Update Datafeed API](https://www.elastic.co/guide/en/elasticsearch/reference/current/ml-update-datafeed.html)
- [Delete Datafeed API](https://www.elastic.co/guide/en/elasticsearch/reference/current/ml-delete-datafeed.html)
- [Get Datafeed API](https://www.elastic.co/guide/en/elasticsearch/reference/current/ml-get-datafeed.html)

## Special Considerations

1. **Update Orchestration**: Datafeeds must be stopped before updating and can be restarted afterward
2. **Job Association**: Each datafeed is associated with exactly one anomaly detection job
3. **Index Permissions**: Requires read permissions on the source indices
4. **ML Privileges**: Requires `manage_ml` cluster privilege