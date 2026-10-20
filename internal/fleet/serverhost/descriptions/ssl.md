TLS configuration Fleet Server presents to Elastic Agents. Requires Elastic Stack 9.1.0 or later.

~> **Warning:** Changing SSL settings on a Fleet Server host that is already enrolled can disconnect connected Elastic Agents and may require them to be re-enrolled.

~> **Note:** Kibana replaces the whole `ssl` object on update and this resource reads any `ssl` already stored on a host. When upgrading the provider, add the existing SSL settings of a host that already has `ssl` configured to your configuration; otherwise Terraform plans to remove them. The `agent_*` and `es_*` SSL fields are not managed by this resource; updating `ssl` on a host that has any of them stored is rejected so they are not erased.
