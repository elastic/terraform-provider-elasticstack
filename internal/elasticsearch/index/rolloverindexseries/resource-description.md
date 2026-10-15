Manages one alias-owned Elasticsearch index series for ILM rollover.

When `rollover_alias` is missing, Terraform creates `<prefix>-000001` and attaches that alias as its only write alias. The create request does not set index settings or mappings. Put those on an index template that matches the series, including `index.lifecycle.name` and `index.lifecycle.rollover_alias`. The template must not also define `rollover_alias` in its aliases, because Elasticsearch then reports a duplicate alias during rollover.

When the alias already exists, Terraform adopts it without an extra flag and without changing indices or aliases, but only when it is an ordinary index alias with one effective writer whose name ends in `-<digits>`. Adoption includes every index currently attached to the alias, including indices whose names do not start with `prefix`. Adopting the alias acquires deletion rights over every one of those indices once `deletion_protection` is disabled. `deletion_protection` defaults to `true`, and destroy is rejected while it remains enabled.

Keep `rollover_alias` for this series alone. Do not also manage that alias with `elasticstack_elasticsearch_alias`, and do not manage its member indices with `elasticstack_elasticsearch_index`. Those resources would compete to own the same alias and indices.
