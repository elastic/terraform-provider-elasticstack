provider "elasticstack" {
  elasticsearch {}
}

data "elasticstack_elasticsearch_ingest_processor_user_agent" "agent" {
  field = "agent"
}

resource "elasticstack_elasticsearch_ingest_pipeline" "my_ingest_pipeline" {
  name = "agent-ingest"

  processors = [
    data.elasticstack_elasticsearch_ingest_processor_user_agent.agent.json
  ]
}
