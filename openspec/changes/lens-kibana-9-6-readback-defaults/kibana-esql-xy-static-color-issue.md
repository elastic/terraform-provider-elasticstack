# Filed as [elastic/kibana#291451](https://github.com/elastic/kibana/issues/291451)

Kibana persist bug: `POST /api/dashboards` accepts an ES|QL XY Y-metric `color: {type: static}` and read-backs `{type: auto}`. The provider does not treat `{type:static}` as equivalent to `{type:auto}`. Do not skip `TestAccResourceDashboardXYChart_layers` in code.

Related Terraform provider issue: https://github.com/elastic/terraform-provider-elasticstack/issues/4902.

Evidence and request vs 201 quotes: `findings.md` § "Task 2.3 follow-up".
