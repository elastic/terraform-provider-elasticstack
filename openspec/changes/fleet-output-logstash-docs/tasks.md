## 1. Example File

- [x] 1.1 Create `examples/resources/elasticstack_fleet_output/logstash.tf` with a representative logstash output configuration (hosts, ssl block with certificate_authorities/certificate/key, default_integrations, default_monitoring)

## 2. Template Update

- [x] 2.1 Add a `### Logstash output` section to `templates/resources/fleet_output.md.tmpl` that references the new example via `{{ tffile "examples/resources/elasticstack_fleet_output/logstash.tf" }}`

## 3. Verification

- [ ] 3.1 Run `make docs` (or equivalent `terraform-plugin-docs` generation) and confirm that `docs/resources/fleet_output.md` now contains the logstash example section
