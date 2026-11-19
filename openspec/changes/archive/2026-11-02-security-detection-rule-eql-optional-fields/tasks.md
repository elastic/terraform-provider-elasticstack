## 1. EQL rules manage event_category_override and timestamp_field

- [x] 1.1 `elasticstack_kibana_security_detection_rule` accepts optional `event_category_override` and `timestamp_field` for EQL rules (REQ-037): set values reach Kibana on create and update, null values are omitted from the request, removing either attribute clears it in Kibana with an in-place update, and non-EQL rule types never send them; `tiebreaker_field` keeps the same request behavior
- [x] 1.2 Read and import populate `event_category_override` and `timestamp_field` from Kibana (REQ-037): absent fields map to null, and a value set outside Terraform lands in state so the next plan shows it as drift; `tiebreaker_field` follows the same read mapping
- [x] 1.3 The generated resource documentation describes both attributes, including Kibana's defaults (`event.category`, `@timestamp`) and the difference between `timestamp_field` and `timestamp_override`
