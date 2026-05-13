Image source for the panel. Set exactly one nested branch: `file` (uploaded Kibana file id) or `url` (external image URL). Matches the Kibana Dashboard API `kbn-dashboard-panel-type-image` `config.image_config.src` union.

For `file.file_id`, Terraform accepts only the string id returned by Kibana after an upload; creating or deleting the uploaded file is **not** handled by this resource today (use the UI, Saved Objects, or HTTP APIs). A dedicated `elasticstack_kibana_file` resource may be introduced later if practitioners need uploads as code.
