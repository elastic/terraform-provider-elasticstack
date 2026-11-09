Creates or updates a Fleet Integration Policy.

The [Kibana Fleet UI](https://www.elastic.co/guide/en/fleet/current/add-integration-to-policy.html)
can be used as a reference for what data needs to be provided. Instead of saving
a new integration configuration, the API request can be previewed, showing what
values need to be provided for inputs and their streams.

Some Kibana features store data in an integration policy's input- or
stream-level `config` (for example the API keys Kibana's APM plugin creates for
APM Server to read source maps and agent configuration, or the packs assigned
by Osquery). This data is not managed by this resource and is preserved when
the resource updates the policy. If another client modifies the policy while
the resource is updating it, the resource reports an error instead of restoring
that data from a copy that may be out of date.
