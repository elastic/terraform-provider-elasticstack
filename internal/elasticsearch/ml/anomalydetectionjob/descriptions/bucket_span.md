The size of the interval that the analysis is aggregated into, typically between 15m and 1h.

If the anomaly detector is expecting to see data at near real-time frequency, then the `bucket_span` should be set to a value around 10 times the time between ingested documents. For example, if data comes every second, `bucket_span` should be 10s; if data comes every 5 minutes, `bucket_span` should be 50m.

For sparse or batch data, use larger `bucket_span` values.
