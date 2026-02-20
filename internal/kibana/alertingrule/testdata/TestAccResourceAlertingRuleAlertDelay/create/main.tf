variable "name" {
  type = string
}

resource "elasticstack_kibana_alerting_rule" "autoops_service_crashloopbackoff" {
  name         = var.name
  rule_type_id = ".es-query"
  consumer     = "alerts"
  enabled      = true
  interval     = "1m"
  tags         = ["autoops"]
  alert_delay  = 1

  params = jsonencode({
    aggType                    = "count",
    esQuery                    = "{\n    \"query\": {\n        \"bool\": {\n            \"filter\": [\n                {\n                    \"term\": {\n                        \"data_stream.dataset\": \"kubernetes.state_container\"\n                    }\n                },\n                {\n                    \"term\": {\n                        \"kubernetes.container.status.reason\": \"CrashLoopBackOff\"\n                    }\n                },\n                {\n                    \"term\": {\n                        \"kubernetes.namespace\": \"autoops\"\n                   }\n                }\n            ]\n        }\n    }\n}",
    excludeHitsFromPreviousRun = true,
    groupBy                    = "top",
    index                      = ["metrics-*,metrics-*:metrics-*"],
    searchType                 = "esQuery",
    size                       = 1,
    termField                  = ["kubernetes.pod.name", "orchestrator.cluster.name"],
    termSize                   = 10,
    threshold                  = [10],
    thresholdComparator        = ">",
    timeField                  = "@timestamp",
    timeWindowSize             = 10,
    timeWindowUnit             = "m",
    sourceFields = [
      { label = "container.id", searchPath = "container.id" },
      { label = "host.hostname", searchPath = "host.hostname" },
      { label = "host.id", searchPath = "host.id" },
      { label = "host.name", searchPath = "host.name" },
      { label = "kubernetes.pod.uid", searchPath = "kubernetes.pod.uid" }
    ]
  })
}