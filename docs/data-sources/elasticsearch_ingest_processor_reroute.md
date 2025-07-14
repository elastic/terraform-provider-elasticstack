---
subcategory: "Ingest"
layout: ""
page_title: "Elasticstack: elasticstack_elasticsearch_ingest_processor_reroute Data Source"
description: |-
  Helper data source to create a processor which reroutes a document to a different data stream, index, or index alias.
---

# Data Source: elasticstack_elasticsearch_ingest_processor_reroute

Reroutes a document to a different data stream, index, or index alias. This processor is useful for routing documents based on data stream routing rules.

See: https://www.elastic.co/guide/en/elasticsearch/reference/current/reroute-processor.html

## Example Usage

```terraform
provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_reroute" "reroute" {
  destination = "logs-generic-default"
  dataset     = "generic"
  namespace   = "default"
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "reroute-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_reroute.reroute.json
  ]
}
```

## Schema

### Optional

- `dataset` (String) The destination dataset to route the document to.
- `description` (String) Description of the processor.
- `destination` (String) The destination data stream, index, or index alias to route the document to.
- `if` (String) Conditionally execute the processor
- `ignore_failure` (Boolean) Ignore failures for the processor.
- `namespace` (String) The destination namespace to route the document to.
- `on_failure` (List of String) Handle failures for the processor.
- `tag` (String) Identifier for the processor.

### Read-Only

- `id` (String) Internal identifier of the resource.
- `json` (String) JSON representation of this data source.