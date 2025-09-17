Creates or updates a Fleet Integration Policy.

It is highly recommended that all inputs and streams are provided in the
Terraform plan, even if some are disabled. Otherwise, differences may appear
between what is in the plan versus what is returned by the Fleet API.

The [Kibana Fleet UI](https://www.elastic.co/guide/en/fleet/current/add-integration-to-policy.html)
can be used as a reference for what data needs to be provided. Instead of saving
a new integration configuration, the API request can be previewed, showing what
values need to be provided for inputs and their streams.
